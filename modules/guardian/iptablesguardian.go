package guardian

import (
	"errors"
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/modules/metrics"
	"justdevelop.it/goaway/obj"
	"justdevelop.it/goaway/repos/verdict"
	"justdevelop.it/goaway/service"
	"justdevelop.it/goaway/utils"
	"log"
	"net"
	"strings"
	"time"
)

var goawayIpBanChain = "goawayIpban"
var goawayGlobalChain = "goawayGlobal"
var goawayLogChain = "goawayLog"
var filterTable = "filter"

type IPTablesGuardian struct {
	cfg                *ini.File
	logger             *log.Logger
	repo               verdict.Repository
	metricsEnabled     bool
	metricsLogFile     string
	banEnabled         bool
	globalRulesEnabled bool
	metricsChan        chan []obj.Metric
	run                bool
}

func New(cfg *ini.File, sm service.ServiceManager) *IPTablesGuardian {
	logFile := cfg.Section("guardian").Key("logFile").String()
	globalRules := cfg.Section("guardian").Key("enableGlobalRules").MustBool(false)
	banEnabled := cfg.Section("guardian").Key("banEnabled").MustBool(true)

	fmt.Printf("%v\n", globalRules)
	time.Sleep(time.Second * 2)

	logger := utils.NewLogger(logFile)
	repo := verdict.NewRepo(cfg, sm, logger)

	g := &IPTablesGuardian{
		cfg:                cfg,
		logger:             logger,
		repo:               repo,
		metricsEnabled:     cfg.Section("metrics").Key("enabled").MustBool(false),
		metricsLogFile:     cfg.Section("metrics").Key("logFile").String(),
		globalRulesEnabled: globalRules,
		banEnabled:         banEnabled,
		run:                true,
	}
	return g
}

func (g *IPTablesGuardian) Run() {
	g.logger.Println("Guardian Started...")
	if !g.banEnabled {
		g.logger.Println("Banning disabled")
	}

	pollInterval, err := time.ParseDuration(g.cfg.Section("guardian").Key("pollInterval").MustString("1s"))
	tickChan := time.NewTicker(pollInterval).C
	utils.CheckAndPanic(err)
	for g.run {
		for range tickChan {
			rules, err := g.repo.FetchRules()

			if err != nil {
				utils.LogErrorMessage(err.Error())
				log.Fatal(err)
			}
			if len(rules) > 0 {
				if err := g.ApplyRules(rules); err != nil {
					fmt.Println(err)
					utils.LogErrorMessage(err.Error())
				}
			}
		}
	}
}

func (g *IPTablesGuardian) Setup() {
	err := g.setupipTables()
	utils.CheckAndLogError(err)
}

func (g *IPTablesGuardian) SetMetricsChan(c chan []obj.Metric) {
	g.metricsChan = c
}

func (g *IPTablesGuardian) Quit() {
	g.run = false
}

/** we have 2 iptable chain goawayIpban and goawayGlobal **/
func (g *IPTablesGuardian) setupipTables() error {
	ipTable, err := iptables.New()
	if err != nil {
		return err
	}

	//create if doesn't exist
	g.logger.Println("Clearing or creating " + goawayIpBanChain)
	err = ipTable.ClearChain(filterTable, goawayIpBanChain)
	if err != nil {
		return err
	}
	g.logger.Println("Clearing or creating " + goawayGlobalChain)
	err = ipTable.ClearChain(filterTable, goawayGlobalChain)
	if err != nil {
		return err
	}

	g.logger.Println("Set up chain links and global rules " + goawayGlobalChain)
	//append chain to input and forward
	existInput, err := ipTable.Exists(filterTable, "INPUT", "-j", goawayIpBanChain)
	if !existInput {
		ipTable.Append(filterTable, "INPUT", "-j", goawayIpBanChain)
	}
	existForward, err := ipTable.Exists(filterTable, "FORWARD", "-j", goawayIpBanChain)
	if !existForward {
		ipTable.Append(filterTable, "FORWARD", "-j", goawayIpBanChain)
	}

	existInputG, err := ipTable.Exists(filterTable, "INPUT", "-j", goawayGlobalChain)
	if !existInputG {
		ipTable.Append(filterTable, "INPUT", "-j", goawayGlobalChain)
	}

	/***** GLOBAL goaway chain RULES*****/
	/***
	https://javapipe.com/iptables-ddos-protection
	http://www.mad-hacking.net/documentation/linux/security/iptables/rate-limiting.xml
	*/
	rules := []string{}
	if g.globalRulesEnabled {
		//todo might be worth logging dropped ips
		rules = []string{
			/*NULL Packets */
			//"-p tcp -m tcp --tcp-flags ALL NONE -j DROP",
			//"-p tcp -m tcp --tcp-flags ALL ALL -j DROP",
			/*XMAS Packets are malformed packets of data and as a rule of thumb you should block these*/
			//"-p tcp -m tcp --tcp-flags ALL ALL -j DROP",
			//drop new packet not syn for synflood
			// can be bit dangerous: require NEW packets to have SYN. This will break if a TCP connection is continued after the respective state in iptables already timed out.
			//"-p tcp --dport 80 ! --syn -m conntrack --ctstate NEW -j DROP",
			// Limit concurrent connections per source IP
			// check with sudo netstat -plan|grep :80|awk {'print $5'}|cut -d: -f 1|sort|uniq -c|sort -nk 1
			"-p tcp --dport 80 -m connlimit --connlimit-above 80 -j DROP",
			//Limit NEW TCP connections per second per source IP ###
			//conntrack module info cat /proc/sys/net/ipv4/netfilter/ip_conntrack_count cat /proc/sys/net/ipv4/netfilter/ip_conntrack_max
			//hashlimit info /proc/net/ipt_hashlimit
			"-p tcp --dport 80 -m conntrack --ctstate NEW -m hashlimit --hashlimit-name HTTP --hashlimit 100/s --hashlimit-burst 200 --hashlimit-mode srcip --hashlimit-htable-expire 10000 -j ACCEPT",
			"-p tcp --dport 80 -m conntrack --ctstate NEW -j DROP",
			//Limit RST packets
			"-p tcp --dport 80 -m tcp --tcp-flags RST RST -m limit --limit 4/s --limit-burst 2 -j ACCEPT",
			"-p tcp --dport 80 -m tcp --tcp-flags RST RST -j DROP",
		}
	}

	for _, rule := range rules {
		args := strings.Split(rule, " ")
		if len(args) > 0 {
			err = ipTable.Append(filterTable, goawayGlobalChain, args...)
			if err != nil {
				return err
			}
		} else {
			return errors.New("firewall rule not parsed correctly " + fmt.Sprint(rule))
		}
	}

	return nil
}

// Proto returns the protocol used by this IPTables.
func (g *IPTablesGuardian) ApplyRules(ipRules []verdict.IpSentence) error {
	for _, rule := range ipRules {
		ipTable, err := iptables.New()

		trial := net.ParseIP(rule.Ip)
		if trial.To4() == nil {
			fmt.Printf("%v is not an IPv4 address skipping \n", rule.Ip)
			return errors.New("Not a valid Ip")
		}
		if err != nil {
			return err
		}
		exist, err := ipTable.Exists(filterTable, goawayIpBanChain, "-s", rule.Ip, "-j", "DROP")

		if !rule.IsExpired() && rule.RuleActive {
			if err != nil {
				return err
			}
			if !exist && g.banEnabled {
				fmt.Printf("Append DROP rule %s", rule.Ip)
				err := ipTable.Append(filterTable, goawayIpBanChain, "-s", rule.Ip, "-j", "DROP")
				if err != nil {
					return err
				}
				//metric logic
				if g.metricsEnabled {
					tags := map[string]string{
						"action": "ban",
						"ip":     rule.Ip,
					}
					metrics.AddTagsVal("guardian", tags, 1)
				}
			}
		} else {
			if exist {
				fmt.Printf("Removing DROP rule %s", rule.Ip)
				err := ipTable.Delete(filterTable, goawayIpBanChain, "-s", rule.Ip, "-j", "DROP")

				if err != nil {
					return err
				}

				if g.metricsEnabled {
					tags := map[string]string{
						"action": "unban",
						"ip":     rule.Ip,
					}
					metrics.AddTagsVal("guardian", tags, 1)
				}
			}
		}
	}
	return nil
}

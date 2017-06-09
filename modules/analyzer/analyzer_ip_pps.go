package analyzer

import (
	"fmt"
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/modules/metrics"
	"justdevelop.it/goaway/obj"
	"justdevelop.it/goaway/repos/evidence"
	"justdevelop.it/goaway/repos/verdict"
	"justdevelop.it/goaway/service"
	"justdevelop.it/goaway/utils"
	"log"
	"strconv"
	"strings"
	"time"
)

type AnalyzePps struct {
	//implements...
	Analyzer
	cfg     *ini.File
	banFreq int
	//fields
	evidenceCollection map[string]*obj.Evidence
	banTtl             time.Duration
	evidenceRepo       *evidence.RepositoryRedis
	logger             *log.Logger
	metricsChan        chan []obj.Metric
	quitChan           chan bool
	run                bool
	verdictRepo        verdict.Repository
	whitelistIps       []string
}

func Pps(cfg *ini.File, sm service.ServiceManager) *AnalyzePps {
	logFile := cfg.Section("analyzer").Key("logFile").String()
	logger := utils.NewLogger(logFile)

	banFreq := cfg.Section("analyzer_ippps").Key("ipPpsthreshold").MustInt()
	ttl, err := time.ParseDuration(cfg.Section("analyzer").Key("banTtl").MustString("1800s"))
	if err != nil {
		ttl, _ = time.ParseDuration("1800s")
	}
	repo := evidence.NewRepositoryRedis(cfg, sm, logger)
	verdictRepo := verdict.NewRepo(cfg, sm, logger)

	a := &AnalyzePps{
		cfg:                cfg,
		banFreq:            banFreq,
		evidenceCollection: make(map[string]*obj.Evidence),
		banTtl:             ttl,
		evidenceRepo:       repo,
		logger:             logger,
		verdictRepo:        verdictRepo,
		whitelistIps:       cfg.Section("whitelist_ips").KeyStrings(),
		run:                true,
	}
	return a
}

func (a *AnalyzePps) Run() {
	fmt.Println("Staring Analyzer (PPS)")
	for a.run {
		verdicts := a.analyze()
		if len(verdicts) > 0 {

			tags := map[string]string{
				"key": a.GetName(),
			}
			metrics.AddTagsVal("analysis-result", tags, len(verdicts))
			go func() {
				defer utils.HandleErrors()
				a.PublishVerdict(verdicts)
			}()
		}
		time.Sleep(time.Millisecond * 200)
	}
}

const IpPpsName = "ip_pps"

func (a *AnalyzePps) PublishVerdict(result []*verdict.IpSentence) {

	//remove whitelist
	for i, v := range result {
		if utils.InStringSlice(a.whitelistIps, v.Ip){
			//https://github.com/golang/go/wiki/SliceTricks
			result = append(result[:i], result[i+1:]...)
		}
	}
	err := a.verdictRepo.Publish(result)
	utils.CheckAndLogError(err)
}

func (a *AnalyzePps) Setup() {

}

func (a *AnalyzePps) SetMetricsChan(c chan []obj.Metric) {
	a.metricsChan = c
}

func (a *AnalyzePps) Quit() {
	a.run = false
}

func (a *AnalyzePps) GetName() string {
	return IpPpsName
}

// Here we get any packet with scores above the threshold and return them to the caller for banning.
// Any packets with a lower threshold and discarded after 5 seconds by the janitor
func (a *AnalyzePps) analyze() []*verdict.IpSentence {
	ipPps, err := a.evidenceRepo.GetIpPpsAboveScore(strconv.Itoa(a.banFreq))
	sentences := []*verdict.IpSentence{}

	if err != nil {
		utils.LogErrorMessage(err.Error())
	} else {
		for _, z := range ipPps {
			ipTime := strings.SplitN(z.Member.(string), ":", 2)
			metricsTags := map[string]string{}
			metricsTags["ip"] = ipTime[1]
			timeInt, err := strconv.ParseInt(ipTime[0], 10, 64)

			msg := fmt.Sprintf("Analyzed Pps: Ip %s score %f %s", ipTime[1], z.Score, time.Unix(timeInt, 0).String())
			a.logger.Println(msg)

			sentences = append(
				sentences,
				&verdict.IpSentence{
					Ip: ipTime[1],
					Ttl: a.banTtl,
					BannedBy: a.GetName(),
					DateTime: time.Now().Format(utils.MYSQLDATETIME),
					Reason: a.GetName() + ":" + strconv.FormatFloat(z.Score, 'f', 0, 64),
				},
			)

			if err == nil {
				metrics.SumInSecond(
					"analysis-ip-pps-score",
					ipTime[1],
					time.Unix(timeInt, 0),
					map[string]string{"ip": ipTime[1]},
					z.Score)
			}
			err = a.evidenceRepo.Remove(z)
			utils.CheckAndLogError(err)

		}
	}

	return sentences
}

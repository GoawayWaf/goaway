package recorder

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"justdevelop.it/goaway/modules/metrics"
	"justdevelop.it/goaway/obj"
	"justdevelop.it/goaway/repos/evidence"
	"justdevelop.it/goaway/repos/whitelist"
	"justdevelop.it/goaway/service"
	"justdevelop.it/goaway/utils"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"sync"
)

//todo interfaces
type PacketRecorder struct {
	dataStore    *evidence.RepositoryRedis
	cfg          *ini.File
	logger       *log.Logger
	device       string
	snapshot_len int32
	promiscuous  bool
	timeout      time.Duration
	handle       *pcap.Handle
	// BPF filter when capturing packets
	filter string

	rsaKeyPaths    []string
	packetsLogRepo *evidence.RepositoryPacketsMysql
	whitelistIps   []string
	whitelistNets   []string
	packetSource   *gopacket.PacketSource
	run            bool
	metricsChan    chan []obj.Metric
	forwardedForIpRegex   string
}

var flushChannel chan map[string]packetCount
var flushTicker *time.Ticker

func NewPacketRecorder(cfg *ini.File, sm service.ServiceManager) *PacketRecorder {

	logFile := cfg.Section("recorder").Key("logFile").String()
	logger := utils.NewLogger(logFile)

	var rsaKeyPaths []string
	arr := cfg.Section("analyzer_rsa_keys").Keys()
	for _, val := range arr {
		rsaKeyPaths = append(rsaKeyPaths, cfg.Section("analyzer_rsa_keys").Key(val.String()).MustString(""))
	}
	iface := cfg.Section("recorder").Key("interface").MustString("any")

	packetsMysqlRepo := evidence.NewRepo(cfg, sm, logger)
	whitelistRepo := whitelist.NewRepo(cfg, sm, logger)

	whitelistIps, err := whitelistRepo.GetAllOfType("ip")
	utils.CheckAndLogError(err)
	whitelistNetworks, err := whitelistRepo.GetAllOfType("net")
	utils.CheckAndLogError(err)

	return &PacketRecorder{
		dataStore:      evidence.NewRepositoryRedis(cfg, sm, logger),
		cfg:            cfg,
		logger:         logger,
		device:         iface,
		snapshot_len:   1024,
		promiscuous:    false,
		timeout:        pcap.BlockForever, //200 * time.Millisecond, //timeout for pcap library check docs
		filter:         "ip",
		rsaKeyPaths:    rsaKeyPaths,
		packetsLogRepo: packetsMysqlRepo,
		whitelistIps:   whitelistIps,
		whitelistNets:  whitelistNetworks,
		run:            true,
		forwardedForIpRegex:   cfg.Section("recorder").Key("forwardedForIpRegex").String(),
	}
}

func (r *PacketRecorder) Setup() {
	fmt.Println("Setup")

	d, _ := time.ParseDuration("2s")
	flushTicker = time.NewTicker(d)

	flushChannel = make(chan map[string]packetCount)

	excludeIps := []string{"127.0.0.1"}
	excludeNetworks := []string{"169.254.0.0/16"}

	ipExcludeFilter := getExcludeIpBpfFilter(excludeIps)
	netExcludeFilter := getExcludeNetBpfFilter(excludeNetworks)
	//http://www.netresec.com/?page=Blog&month=2015-11&post=BPF-is-your-Friend
	//todo configurable ports
	inboundPacketsForWebserverPorts := "inbound and tcp and (dst port 80 or dst port 443 or dst port 8080)"
	//we ignore syn ack and empty packets, the attacks we are interested to analyze are http resources load
	onlyHttpIpv4PacketsWithData := "(((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0)"
	filter := fmt.Sprintf(
		"%s%s%s and %s",
		inboundPacketsForWebserverPorts,
		ipExcludeFilter,
		netExcludeFilter,
		onlyHttpIpv4PacketsWithData,
	)

	r.filter = filter
	r.logger.Println(filter)
}
func (r *PacketRecorder) Run() {
	fmt.Println("Run Recorder")
	// Open device
	var err error
	r.handle, err = pcap.OpenLive(r.device, r.snapshot_len, r.promiscuous, r.timeout)
	utils.CheckAndPanic(err)

	err = r.handle.SetBPFFilter(r.filter)
	utils.CheckAndPanic(err)

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(r.handle, r.handle.LinkType())

	//flush full packet data to redis
	go func(rec *PacketRecorder) {
		defer utils.HandleErrors()
		rec.ListenFlushPacketCountChan()
	}(r)

	//flush full packet data to mysql
	go func(rec *PacketRecorder) {
		defer utils.HandleErrors()
		for r.run {
			select {
			case <-flushTicker.C:
				r.packetsLogRepo.FlushData()
			}
		}
	}(r)

	for r.run {
		select {
		case packet := <-packetSource.Packets():
			r.record(packet)
		default:
			fmt.Println("No packets to record")
			time.Sleep(time.Second * 5)
		}
	}

	close(flushChannel)
	flushTicker.Stop()
	r.handle = nil
}

func (r *PacketRecorder) Quit() {
	r.run = false
}

func (r *PacketRecorder) SetMetricsChan(c chan []obj.Metric) {
	r.metricsChan = c
}

func Debug(filter string) {
	intf, err := pcap.FindAllDevs()
	utils.CheckAndPanic(err)
	for _, v := range intf {
		fmt.Println(v)
	}

	// Open device
	handle, err := pcap.OpenLive("any", 1024, false, pcap.BlockForever)
	utils.CheckAndPanic(err)
	fmt.Println("Executing filter: " + filter)
	time.Sleep(500 * time.Millisecond)
	err = handle.SetBPFFilter(filter)
	utils.CheckAndPanic(err)
	defer handle.Close()

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		_, oki := ipLayer.(*layers.IPv4)

		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		_, okt := tcpLayer.(*layers.TCP)
		if oki && okt {
			fmt.Println(packet)
		} else {

			fmt.Println("Packet layers not found")
		}
	}
}

/*
  we get the packet and put in a buffer keyed per timestamp which we flush every time we have a new timestamp
  like that we have in redis every time the full score per second and we can remove the entries beow the score treshold with 1 command
  and timestamp is absolute so in sync with the analyzer

   option 2 we increse timeout in pcap so it process packets every second

*/
var lastTimestamp int64
var packetsCountBuffer = map[string]packetCount{}

func (r *PacketRecorder) record(packet gopacket.Packet) {

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	ip, oki := ipLayer.(*layers.IPv4)

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	tcpx, okt := tcpLayer.(*layers.TCP)
	if oki && okt {
		packetTime := time.Now()
		currentIntTimeStamp := packetTime.Unix()

		logMsg := fmt.Sprintf("From port %s:%d to %s:%d %d %s\n", ip.SrcIP, tcpx.SrcPort, ip.DstIP, tcpx.DstPort, currentIntTimeStamp, fmt.Sprint(packetTime))
		fmt.Println(logMsg)

		/* Packet sum on each second (timestamp) for quick detection */

		/* if timestamp changed we flush the packet buffer */
		if currentIntTimeStamp > lastTimestamp {
			flushChannel <- packetsCountBuffer
			packetsCountBuffer = map[string]packetCount{}
		}
		lastTimestamp = currentIntTimeStamp

		/* Full Packet data into db */
		packetInfo := newPacketData(
			string(tcpx.Payload),
			ip.SrcIP.String(),
			len(packet.Data()),
			packetTime,
			r.forwardedForIpRegex,
		)

		/** we log it but don't count it for analysis*/
		isWhiteliStedIp := utils.InStringSlice(r.whitelistIps, packetInfo.RealIp)
		/** we log it but don't count it for analysis*/
		isWhiteliStedNet := isInWhitelistedNetwork(packetInfo.RealIp, r.whitelistNets)
		/** todo multipart forms split the request in lot of packets which have no method so we only count the first **/
		/** we should handle this case with another threshold */
		/** altough this skips ssl encrypted records as well so decryption needs to work */
		if packetInfo.RequestMethod != "" && !isWhiteliStedIp && !isWhiteliStedNet {
			fmt.Println("Counting", packetInfo.RealIp)
			incrementPacketBufferCount(packetInfo.RealIp, currentIntTimeStamp)
		} else {
			fmt.Println("Skipping", packetInfo.RealIp)
		}


		fields := make(map[string]interface{})
		fields["value"] = 1.0
		fields["ip"] = packetInfo.RealIp
		metrics.AddMeasurement("recorder-packet-score", map[string]string{}, fields, packetTime)
		r.packetsLogRepo.LazyRecord(packetInfo)
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("Error decoding some part of the packet:", err)
	}
}

type packetCount struct {
	Ip        string
	Timestamp int64
	Count     float64
}

func (p *packetCount) getKey() string {
	key := strconv.FormatInt(p.Timestamp, 10) + ":" + p.Ip
	return key
}

var packetCountMux sync.Mutex
func incrementPacketBufferCount(ip string, ts int64) {

	packetCountMux.Lock()
	p := packetCount{
		Ip:        ip,
		Timestamp: ts,
	}
	if pcount, exists := packetsCountBuffer[p.getKey()]; exists {
		pcount.Count += 1.0
		packetsCountBuffer[p.getKey()] = pcount
	} else {
		p.Count = 1.0
		packetsCountBuffer[p.getKey()] = p
	}
	packetCountMux.Unlock()
}

func (r *PacketRecorder) ListenFlushPacketCountChan() {

	for countBuf := range flushChannel {
		for i, v := range countBuf {
			r.dataStore.Record(v.Count, i)
		}
	}
}

func newPacketData(
	payload string,
	packetIp string,
	size int,
	t time.Time,
	forwardedForIpRegex string,
) *obj.PacketData {
	hostReg, _ := regexp.Compile(`Host: (.*)\n`)
	methodUriReg, _ := regexp.Compile(`(GET|POST) (.*) HTTP`)
	uaReg, _ := regexp.Compile(`User-Agent: (.*)\n`)
	realIpReg, err := regexp.Compile(forwardedForIpRegex)
	utils.CheckAndPanic(err)

	// Using FindStringSubmatch you are able to access the
	// individual capturing groups
	methodUri := methodUriReg.FindStringSubmatch(payload)
	host := hostReg.FindStringSubmatch(payload)
	ua := uaReg.FindStringSubmatch(payload)
	forwardedFor := realIpReg.FindStringSubmatch(payload)

	info := obj.PacketData{}
	if len(host) > 0 {
		info.Host = strings.TrimSpace(host[1])
	}
	if len(methodUri) > 1 {
		info.RequestMethod = strings.TrimSpace(methodUri[1])
		info.RequestUri = strings.TrimSpace(methodUri[2])
	}
	if len(ua) > 0 {
		info.UserAgent = strings.TrimSpace(ua[1])
	}
	info.RealIp = packetIp
	if len(forwardedFor) > 0 {
  		info.RealIp = forwardedFor[1]
	}
	info.ReceivedIp = packetIp
	info.DateTime = t
	info.PacketSize = size
	info.Payload = payload
	return &info
}
func isInWhitelistedNetwork(ip string, nets []string) bool{
	for _, netAddr := range nets {
		var err error
		isWhiteliStedNet, err := utils.CIDRMatch(ip, netAddr)
		utils.CheckAndLogError(err)
		if isWhiteliStedNet {
			return true
		}
	}
	return false
}
func getLocalNetworksBpfFilter() string {
	addresses := utils.GetIp4InterfacesNetworkAdress()
	result := ""
	for _, v := range addresses {
		result += fmt.Sprintf(" and not src net %s", v)
	}

	return result
}
func getExcludeIpBpfFilter(addresses []string) string {
	result := ""
	for _, v := range addresses {
		result += fmt.Sprintf(" and not host %s", v)
	}
	return result
}

func getExcludeNetBpfFilter(addresses []string) string {
	result := ""
	for _, v := range addresses {
		result += fmt.Sprintf(" and not net %s", v)
	}
	return result
}

func DecryptSsl(ciphertext string) string {
	privateKeyFile, err := ioutil.ReadFile("../conf/rsa")
	block, _ := pem.Decode(privateKeyFile)
	if block == nil {
		utils.LogErrorMessage("bad key data: not PEM-encoded")
		return ""
	}

	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		utils.LogErrorMessage(fmt.Sprintf("unknown key type %q, want %q", got, want))
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	utils.LogErrorMessage(err.Error())
	hash := sha256.New()
	label := []byte("")
	textBytes, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, []byte(ciphertext), label)

	utils.CheckAndLogError(err)
	_ = bytes.Index(textBytes, []byte{0})
	//text := string(textBytes[:n])
	return string(textBytes)
}

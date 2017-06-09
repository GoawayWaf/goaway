package utils

import (
"github.com/google/gopacket"
"github.com/google/gopacket/layers"
"github.com/google/gopacket/pcap"
"log"
"net"
"time"
	"fmt"
	"golang.org/x/net/ipv4"
)

var (
	device       string = "eth0"
	snapshot_len int32  = 1024
	promiscuous  bool   = false
	err          error
	timeout      time.Duration = 30 * time.Second
	handle       *pcap.Handle
	buffer       gopacket.SerializeBuffer
)

func generateIp() net.IP {
	ip := make(net.IP, 4)
	ip[0] = 66
	ip[1] = 66
	ip[2] = 0
	for i := 3; i < 4; i++ {
		ip[i] = byte(r.Intn(254))
	}
	return ip
}
var packetCount = 0
func sendpacket() {
	// Open device
	handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
	if err != nil {log.Fatal(err) }
	defer handle.Close()

	// Send raw bytes over wire
	payloadString := `GET /features HTTP/1.1
	Host: www.ehost.stage.wzdev.co
	User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36
	Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8`

	rawBytes := []byte(payloadString)

	payload := gopacket.Payload(rawBytes)

	// This time lets fill out some information
	ipLayer := &layers.IPv4{
		SrcIP: net.IP{127, 0, 0, 1},
		DstIP: net.IP{127, 0, 0, 1},
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}
	ethernetLayer := &layers.Ethernet{
		SrcMAC: net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA},
		DstMAC: net.HardwareAddr{0xBD, 0xBD, 0xBD, 0xBD, 0xBD, 0xBD},
	}
	tcpLayer := &layers.TCP{
		SrcPort: layers.TCPPort(4321),
		DstPort: layers.TCPPort(80),
		Window:  1505,
		Urgent:  0,
		Seq:     11050,
		Ack:     0,
		ACK:     false,
		SYN:     false,
		FIN:     false,
		RST:     false,
		URG:     false,
		ECE:     false,
		CWR:     false,
		NS:      false,
		PSH:     false,
	}

	options:= gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	tcpLayer.SetNetworkLayerForChecksum(ipLayer)
	ipHeaderBuf := gopacket.NewSerializeBuffer()
	err := ipLayer.SerializeTo(ipHeaderBuf, options)
	if err != nil {
		panic(err)
	}

	// And create the packet with the layers
	buffer = gopacket.NewSerializeBuffer()
	err = gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		ipLayer,
		tcpLayer,
		payload,
	)
	if err != nil {
		panic(err)
	}

	//send packet try1
	var packetConn net.PacketConn
	var rawConn *ipv4.RawConn
	packetConn, err = net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		panic(err)
	}
	rawConn, err = ipv4.NewRawConn(packetConn)
	if err != nil {
		panic(err)
	}
	outgoingPacket := buffer.Bytes()

	ipHeader, err := ipv4.ParseHeader(ipHeaderBuf.Bytes())
	if err != nil {
		panic(err)
	}

	err = rawConn.WriteTo(ipHeader, outgoingPacket, nil)
	log.Printf("packet of length %d sent!\n", 1)

	// Send our packet try2

	err = handle.WritePacketData(outgoingPacket)
	if err != nil {
		log.Fatal(err)
	} else {
		packetCount+=1
	}
	fmt.Println(packetCount)
	fmt.Println(string(outgoingPacket))
}

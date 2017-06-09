package utils

import (
	"net"
	"strings"
	"fmt"
	"encoding/binary"
)

func GetIp4InterfacesNetworkAdress() []string{
	addresses,_ := net.InterfaceAddrs()
	networkaddresses := []string{}
	for _,val := range addresses {
		ipnet := strings.Split(val.String(), "/")
		addr := net.ParseIP(ipnet[0])
		ip := addr.To4()
		if ip != nil && ip.String() != "127.0.0.1" && len(ipnet) > 1 {
			network := ip.Mask(ip.DefaultMask())
			networkaddresses = append(
				networkaddresses,
				fmt.Sprintf("%s/%s", network.String(), ipnet[1]),
			)
		}
	}
	return networkaddresses
}

func CIDRMatch(ip string, netAddr string) (bool, error){

	_, cidrnet, err := net.ParseCIDR(netAddr)
	if err != nil {
		return false, err
	}
	myaddr := net.ParseIP(ip)
	if cidrnet.Contains(myaddr) {
		return true, nil
	}
	return false, nil
}

func Ip2int(ip net.IP) uint32 {
	if ip != nil {
		if len(ip) == 16 {
			return binary.BigEndian.Uint32(ip[12:16])
		}
		return binary.BigEndian.Uint32(ip)
	}
	return 0
}

func Int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

package obj

import "time"

type PacketData struct {
	Id            int
	ReceivedIp    string
	RealIp        string
	DateTime      time.Time
	Host          string
	UserAgent     string
	RequestUri    string
	RequestMethod string
	PacketSize    int
	CreatedAt     string
	Payload       string
}

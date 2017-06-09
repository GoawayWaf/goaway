package obj

type LogLine struct {
	Id             int
	Ip             string
	DateTime       string
	Host           string
	Protocol       string
	UserAgent      string
	RequestUri     string
	Referrer       string
	Status         int
	BodyPacketSize int
	PacketSize     int
	CpuUsage       float64
	MemUsage       float64
	CreatedAt      string
}

package utils

import (
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/obj"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Generate(cfg *ini.File) {

	logDir := cfg.Section("recorder").Key("logDir").String()
	fileName := filepath.Join(logDir, "generated.access.json.log")
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to open file", fileName)
	} else {

		logLine := &obj.LogLine{
			Host:           "static.websitebuilder.qa.wzdev.co",
			Protocol:       "Http",
			UserAgent:      "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
			RequestUri:     "/logo.png",
			Referrer:       "",
			BodyPacketSize: 100,
			PacketSize:     100,
			CpuUsage:       0.00,
			MemUsage:       0.00,
		}

		defer f.Close()
		l, _ := GetLastLine(fileName)
		f.Seek(l, os.SEEK_CUR)

		var i int = 0
		randIpCache := []string{}
		var maxIpCount = 1
		for true {

			//set random data for a more realistic dataset
			logLine.Id = i
			var ip string
			if maxIpCount > 0 && len(randIpCache) <= maxIpCount {
				ip = randIp()
				randIpCache = append(randIpCache, ip)
			} else {
				ip = randIpCache[r.Intn(len(randIpCache)-1)]
			}
			logLine.Ip = ip
			logLine.DateTime = time.Now().Format(LOGDATEFORMAT)

			b, _ := json.Marshal(logLine)
			content := string(b) + "\n"
			f.WriteString(content)
			i++
			time.Sleep(time.Millisecond * 5)
			fmt.Print(fmt.Sprintf("Generated %s with %d rows\r", fileName, i))
		}
	}
}

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func randIp() string {
	ip := make([]string, 4)
	for i := 0; i < 4; i++ {
		ip[i] = strconv.Itoa(r.Intn(256))
	}
	return strings.Join(ip, ".")
}

package recorder

import (
	"testing"
	"fmt"
	"time"
)

func TestA(t *testing.T){
	plain := DecryptSsl("")
	fmt.Println(plain)
}
func TestForwardedIp(t *testing.T){

	payload := `GET /_sitebuilder/img/logos/holding-logo.png HTTP/1.1
User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.111 Safari/537.36
Referer: http://dritech-cleaning.com/
Accept: */*
Connection: Keep-Alive
Accept-Encoding: gzip
Accept-Language: pt-BR,en,*
Host: static.sitebuilder.com
X-Forwarded-For: 106.93.12.12

`
	data := newPacketData(string(payload), "130.211.7.234", 100, time.Now(), "X-Forwarded-For: ([0-9.]+),*")
	fmt.Println(data.RealIp)
}

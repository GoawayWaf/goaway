package verdict

import (
	"justdevelop.it/goaway/utils"
	"log"
	"time"
)

type IpSentence struct {
	Ip       string `json:"ip"`
	IpInt    uint32 `json:"ipint"`
	Ttl      time.Duration `json:"ttl"`
	DateTime string `json:"datetime"`
	BannedBy string `json:"bannedby"`
	RuleActive bool `json:"ruleactive"`
	Reason   string `json:"reason"`
}

func (r *IpSentence) IsExpired() bool {
	inserted, err := time.Parse(utils.MYSQLDATETIME, r.DateTime)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now()
	expires := inserted.Add(r.Ttl)
	if now.After(expires) {
		return true
	}
	return false
}

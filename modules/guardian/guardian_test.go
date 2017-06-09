package guardian

import (
	"justdevelop.it/goaway/repos/verdict"
	"testing"
	"time"
)

var enf IPTablesGuardian

func init() {

	enf = IPTablesGuardian{}
}

func TestA(t *testing.T) {

	ruleOne := verdict.IpSentence{
		Ip:       "81.2.3.5",
		Ttl:      300,
		BannedBy: "test",
		DateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	rules := []verdict.IpSentence{ruleOne}
	if err := enf.ApplyRules(rules); err != nil {
		t.Fatal(err)
	}
}
func TestBan(t *testing.T) {

	ruleOne := verdict.IpSentence{
		Ip:       "81.2.3.5",
		Ttl:      0,
		BannedBy: "test",
		DateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	rules := []verdict.IpSentence{ruleOne}
	if err := enf.ApplyRules(rules); err != nil {
		t.Fatal(err)
	}
}

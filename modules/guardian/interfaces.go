package guardian

import "justdevelop.it/goaway/repos/verdict"

type Guardian interface {
	ApplyRules(ipRules []verdict.IpSentence) error
}

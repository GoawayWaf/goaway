package analyzer

import (
	"justdevelop.it/goaway/modules"
)

type Analyzer interface {
	modules.LongRunner
	GetName() string
}

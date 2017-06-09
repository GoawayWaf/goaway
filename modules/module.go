package modules

import "justdevelop.it/goaway/obj"

type Moduler interface {
	Run()
	Setup()
	SetMetricsChan(c chan []obj.Metric)
}

type LongRunner interface {
	Moduler
	Quit()
}

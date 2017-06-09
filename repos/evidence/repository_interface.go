package evidence

import (
	"justdevelop.it/goaway/obj"
	"time"
)

type Repository interface {
	Record(batch []string)
	Get(t time.Time) (result *obj.Evidence, err error)
	GetBetween(s time.Time, e time.Time) (evidence *obj.Evidence, err error)
	GetAverageRps(since time.Time) float64
	AnyPotentialAttackedHosts(since time.Time, avgRps float64) bool
	Migrate() error
}

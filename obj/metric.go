package obj

import "time"

type InfluxDbMetric struct {
	Measurement string
	Time        time.Time
	Tags        map[string]string
	Fields      map[string]interface{}
}

func (m *InfluxDbMetric) GetTime() time.Time{
	return m.Time
}

type Metric interface {
	GetTime() time.Time
}

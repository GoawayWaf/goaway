package metrics

// todo refactor to use channels so we have a routine runnign and we send events to it to log

import (
	"time"
	"sync"
)

var keyValMux sync.Mutex

type keyValMetric struct {
	Group  string
	Time   time.Time
	Tags   map[string]string
	Fields map[string]interface{}
}

var keyValMetricsBuffer []*keyValMetric

func resetKeyValMetrics() {
	keyValMetricsBuffer = make([]*keyValMetric, 0)
}

func initKeyValMetrics() {
	keyValMux.Lock()
	if len(keyValMetricsBuffer) < 1 {
		keyValMetricsBuffer = make([]*keyValMetric, 0)
	}
	keyValMux.Unlock()
}

func AddKeyVal(group string, key string, val interface{}) {
	tags := make(map[string]string)
	tags["key"] = key
	fields := make(map[string]interface{})
	fields["value"] = val
	AddMeasurement(group, tags, fields, time.Now())
}
func AddTagsVal(group string, tags map[string]string, val interface{}) {
	fields := make(map[string]interface{})
	fields["value"] = val
	AddMeasurement(group, tags, fields, time.Now())
}
func AddMeasurement(group string, tags map[string]string, val map[string]interface{}, t time.Time) {
	initKeyValMetrics()
	keyValMux.Lock()
	g := keyValMetric{Group: group, Tags: tags, Fields: val, Time: t}
	keyValMetricsBuffer = append(keyValMetricsBuffer, &g)
	keyValMux.Unlock()
}

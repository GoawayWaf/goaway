package metrics

// todo refactor to use channels so we have a routine runnign and we send events to it to log

import (
	"time"
	"strconv"
	"sync"
)

func resetSumMetrics() {
	sumMetricsBuffer = make(map[string]*keyValMetric)
}

var sumMetricMux sync.Mutex

var sumMetricsBuffer map[string]*keyValMetric

func initSumMetrics() {
	sumMetricMux.Lock()
	if len(sumMetricsBuffer) < 1 {
		sumMetricsBuffer = map[string]*keyValMetric{}
	}
	sumMetricMux.Unlock()
}
/**
writes a sum at a second instead of a single value at nanosecond
so the flush to the database is much lighter
 */
func SumInSecond(group string, key string, secondsTime time.Time, tags map[string]string, val interface{}) {
	initSumMetrics()
	sumMetricMux.Lock()
	fields := make(map[string]interface{})
	fields["value"] = val
	bufferKey := group + ":" + key + ":" + strconv.FormatInt(secondsTime.Unix(), 10)
	if metric, ok := sumMetricsBuffer[bufferKey]; ok {
		addVal, isfloat := metric.Fields["value"].(float64)
		currentVal, isfloatCur := val.(float64)
		if isfloat && isfloatCur{
			currentVal+= addVal
		} else { //we assume is int
			val = float64(val.(int))
			currentVal+= addVal
		}
		fields["value"] = currentVal
	}
	g:= keyValMetric{Group: group, Tags: tags, Fields: fields, Time: secondsTime}
	sumMetricsBuffer[bufferKey] = &g
	sumMetricMux.Unlock()
}


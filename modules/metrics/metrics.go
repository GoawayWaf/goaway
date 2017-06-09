package metrics

import (
	"github.com/influxdata/influxdb/client/v2"
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/utils"
	"os"
	"time"
	"fmt"
	"justdevelop.it/goaway/service"
)

func Init(cfg *ini.File, inf *service.InfluxDb) {
	if cfg.Section("metrics").Key("driver").String() == "influxdb" {
		interval := cfg.Section("metrics_driver_influxdb").Key("writeInterval").String()

		duration, _ := time.ParseDuration(interval)
		go writeToDbInterval(inf, duration)

	}
}

func MetricsBlackholeTicker() {
	tickChan := time.NewTicker(time.Second).C
	for {
		select {
		case <-tickChan:
			keyValMux.Lock()
			resetKeyValMetrics()
			keyValMux.Unlock()
			sumMetricMux.Lock()
			resetSumMetrics()
			sumMetricMux.Unlock()
		}
	}
}
func writeToDbInterval(db *service.InfluxDb, duration time.Duration) {

	tickChan := time.NewTicker(duration).C
	for {
		select {
		case <-tickChan:
			FlushMetricsToDb(db)
		}
	}
}

func FlushMetricsToDb(db *service.InfluxDb) {

	var bp client.BatchPoints
	var err error

	keyValMux.Lock()
	keyValData := keyValMetricsBuffer
	resetKeyValMetrics()
	keyValMux.Unlock()

	sumMetricMux.Lock()
	sumMetricsData := sumMetricsBuffer
	resetSumMetrics()
	sumMetricMux.Unlock()

	bp, err = db.NewBatchPoints()
	if len(keyValData) > 0 {
		for _, val := range keyValData {

			tags := val.Tags
			tags["hostname"], _ = os.Hostname()

			influxtable := val.Group
			db.AddPointToBatch(bp, influxtable, tags, val.Fields, val.Time)
			utils.CheckAndLogError(err)
		}
	}
	if len(sumMetricsData) > 0 {
		for _, val := range sumMetricsData {

			tags := val.Tags
			tags["hostname"], _ = os.Hostname()

			influxtable := val.Group
			db.AddPointToBatch(bp, influxtable, tags, val.Fields, val.Time)
			utils.CheckAndLogError(err)
		}
	}
	err = db.WriteBatchPoints(bp)
	if err != nil {
		utils.LogErrorMessage(err.Error() + " " + fmt.Sprint(bp.Points()))
	}
}

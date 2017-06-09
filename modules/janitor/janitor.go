package janitor

import (
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/obj"
	"justdevelop.it/goaway/repos/evidence"
	"justdevelop.it/goaway/service"
	"justdevelop.it/goaway/utils"
	"log"
	"strconv"
	"time"
)

type Janitor struct {
	cfg          *ini.File
	logger       *log.Logger
	run          bool
	metricsChan  chan []obj.Metric
	evidenceRepo *evidence.RepositoryPacketsMysql
	redisRepo    *evidence.RepositoryRedis
}

func New(cfg *ini.File, sm service.ServiceManager) *Janitor {
	logger := utils.NewLogger(cfg.Section("janitor").Key("logFile").String())
	return &Janitor{
		cfg:          cfg,
		run:          true,
		logger:       logger,
		evidenceRepo: evidence.NewRepo(cfg, sm, logger),
		redisRepo:    evidence.NewRepositoryRedis(cfg, sm, logger),
	}
}

func (j *Janitor) Run() {
	j.logger.Println("Starting Janitor")


	redisInterval := time.Second * 5
	ticker2 := time.NewTicker(redisInterval)
	j.logger.Println("Clear Redis Every ", redisInterval)

	evidenceDuration, err := time.ParseDuration(j.cfg.Section("janitor").Key("evidenceDuration").MustString("1800s"))
	utils.CheckAndLogError(err)

	evidenceShardDuration, err := time.ParseDuration(j.cfg.Section("janitor").Key("evidenceShardDuration").MustString("30m"))
	utils.CheckAndLogError(err)

	archiveDuration, err := time.ParseDuration(j.cfg.Section("janitor").Key("evidenceArchiveDuration").MustString("0s"))
	utils.CheckAndLogError(err)

	archiveShardDuration, err := time.ParseDuration(
		j.cfg.Section("janitor").Key("evidenceArchiveShardDuration").MustString("1h"),
	)
	utils.CheckAndLogError(err)

	ticker1 := time.NewTicker(evidenceShardDuration)
	j.logger.Println("Evidence Db Max Duration", evidenceDuration)
	j.logger.Println("Evidence Clear Db Every ", evidenceShardDuration)

	ticker3 := time.NewTicker(archiveShardDuration)
	j.logger.Println("Archive Max Duration: ", archiveDuration)
	j.logger.Println("Archive Clear Every: ", archiveShardDuration)

	for j.run {
		select {
		case <-ticker1.C:
			var err error
			if archiveDuration > time.Second {
				j.logger.Println("Clearing Db evidence and archive")
				err = j.evidenceRepo.TrimAndArchive(time.Now().Add(-evidenceDuration))

			} else {
				j.logger.Println("Clearing Db evidence")
				err = j.evidenceRepo.Trim(time.Now().Add(-evidenceDuration))
			}
			utils.CheckAndLogError(err)
		case <-ticker3.C:
			if archiveDuration > time.Second {
				err = j.evidenceRepo.TrimArchive(time.Now().Add(-archiveDuration))
				j.logger.Println("Clearing evidence Archive Db")
			}

		case <-ticker2.C:
			j.logger.Println("Clearing Redis evidence")
			//remove pps records below current timestamp so we don't fill up redis with
			//irrelevant data
			err := j.redisRepo.RemoveBelowLex(strconv.FormatInt(time.Now().Unix(), 10))
			utils.CheckAndLogError(err)
		}
	}

	ticker1.Stop()
	ticker2.Stop()
}

func (j *Janitor) Setup() {

}

func (j *Janitor) Quit() {
	j.run = false
}

func (j *Janitor) SetMetricsChan(c chan []obj.Metric) {
	j.metricsChan = c
}

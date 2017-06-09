package evidence

import (
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/utils"
	"log"
	"justdevelop.it/goaway/service"
)

func NewRepo(cfg *ini.File, sm service.ServiceManager, logger *log.Logger) *RepositoryPacketsMysql {
	driver := cfg.Section("evidence").Key("driver").String()
	dbPacket := &RepositoryPacketsMysql{}
	if driver == "mysql" {

			dbPacket = &RepositoryPacketsMysql{
				DbConn: sm.GetMysqlDb(),
				Table:  "packets_log",
				Logger: logger,
			}

	} else {
		utils.LogErrorMessage("No support for " + driver + " driver")
		panic("No support for " + driver + " driver")
	}
	return dbPacket
}

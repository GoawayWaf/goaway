package verdict

import (
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/utils"
	"log"
	"justdevelop.it/goaway/service"
)

var db Repository

func NewRepo(cfg *ini.File, sm service.ServiceManager, logger *log.Logger) Repository {
	driver := cfg.Section("verdict").Key("driver").String()

	if db != nil {
		return db
	}
	if driver == "mysql" {

		defaultSentencActive := cfg.Section("verdict").Key("ban_rule_default_active").MustBool(true)

		db = &RepositoryMysql{
			DbConn: sm.GetMysqlDb(),
			Table:  cfg.Section("verdict").Key("table").MustString("requests"),
			Logger: logger,
			defaultSentenceActive: defaultSentencActive,
		}
	} else {
		utils.LogErrorMessage("No support for " + driver + " driver")
		panic("No support for " + driver + " driver")
	}
	return db
}

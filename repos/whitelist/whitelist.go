package whitelist

import (
	"gopkg.in/ini.v1"
	"justdevelop.it/goaway/utils"
	"log"
	"justdevelop.it/goaway/service"
)

type Repository interface {
	GetAllOfType(wtype string) ([]string, error)
	Migrate() error
}

func NewRepo(cfg *ini.File, sm service.ServiceManager, logger *log.Logger) *RepositoryMysql {
	driver := cfg.Section("whitelist").Key("driver").String()
	repo := &RepositoryMysql{}
	if driver == "mysql" {

		repo = &RepositoryMysql{
			DbConn: sm.GetMysqlDb(),
			Table: cfg.Section("whitelist").Key("table").String(),
			Logger: logger,
		}

	} else {
		utils.LogErrorMessage("No support for " + driver + " driver")
		panic("No support for " + driver + " driver")
	}
	return repo
}

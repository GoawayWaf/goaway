package service

import (
	"fmt"
	"database/sql"
	"gopkg.in/ini.v1"

)
type MysqlDb struct {
	conn *sql.DB
}

func NewMysqlDb(cfg *ini.File) (*MysqlDb, error) {

	dbConn, err := getMysqlClient(cfg)
	if err != nil {
		return nil, err
	}

	return &MysqlDb{conn: dbConn}, err
}
func getMysqlClient(cfg *ini.File) (*sql.DB, error){

	return sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8",
			cfg.Section("driver_mysql").Key("user").String(),
			cfg.Section("driver_mysql").Key("password").String(),
			cfg.Section("driver_mysql").Key("host").String(),
			cfg.Section("driver_mysql").Key("port").String(),
			cfg.Section("driver_mysql").Key("database").String(),
		))
}

func (m *MysqlDb) GetConn() *sql.DB{
	return m.conn
}
/***
to reload the connection and make sure no query is executed when conn is closed we use mutex

 */
func (m *MysqlDb) Reload(cfg *ini.File) error{

	dbConn, err := getMysqlClient(cfg)
	if err != nil {
		return err
	}
	//clone the old connection
	oldConn := *m.conn
	//replace connection with new
	m.conn = dbConn
	//then close old connection
	err = oldConn.Close()
	if err != nil {
		return err
	}

	return nil
}

package service

import (
	"gopkg.in/ini.v1"
)

type ServiceManager interface {
	Reload(*ini.File) error
	GetMysqlDb() *MysqlDb
	GetRedisDb() *RedisDb
	GetInfluxDb() *InfluxDb
}

type Sm struct {
	databases *databases
}
type databases struct {
	MysqlDb *MysqlDb
	RedisDb *RedisDb
	InfluxDb *InfluxDb
}

func (sm *Sm) Reload(cfg *ini.File) error {

	var err error
	if sm.databases.MysqlDb.GetConn() != nil {
		err = sm.databases.MysqlDb.Reload(cfg)
	}
	if sm.databases.RedisDb.GetConn() != nil {
		err = sm.databases.RedisDb.Reload(cfg)
	}
	if sm.databases.InfluxDb.GetConn() != nil {
		err = sm.databases.InfluxDb.Reload(cfg)
	}
	return err
}

func (sm *Sm) GetMysqlDb() *MysqlDb{
	return sm.databases.MysqlDb
}

func (sm *Sm) GetRedisDb() *RedisDb{
	return sm.databases.RedisDb
}

func (sm *Sm) GetInfluxDb() *InfluxDb{
	return sm.databases.InfluxDb
}

func New(cfg *ini.File) (ServiceManager, error){

	databaseClients, err := newDatabases(cfg)
	if err != nil {
		return nil, err
	}
	sm := &Sm{
		databases: databaseClients,
	}
	return sm, err
}

func newDatabases(cfg *ini.File) (*databases, error) {

	mysqldb, err := NewMysqlDb(cfg)

	if err != nil {
		return nil, err
	}
	redisdb, err := NewRedisDb(cfg)
	if err != nil {
		return nil, err
	}
	influxdb, err := NewInfluxDb(cfg)
	if err != nil {
		return nil, err
	}

	db := &databases{
		MysqlDb: mysqldb,
		RedisDb: redisdb,
		InfluxDb: influxdb,
	}
	return db, err
}

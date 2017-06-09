package service

import (
	"fmt"
	"gopkg.in/redis.v5"
	"gopkg.in/ini.v1"

	"time"
)
type RedisDb struct {
	conn *redis.Client
}

func NewRedisDb(cfg *ini.File) (*RedisDb, error) {

	dbConn, err := getRedisClient (cfg)
	return &RedisDb{conn: dbConn}, err
}

func getRedisClient(cfg *ini.File)  (*redis.Client, error){
	host := cfg.Section("driver_redis").Key("host").String()
	password := cfg.Section("driver_redis").Key("password").MustString("")
	port := cfg.Section("driver_redis").Key("port").MustInt(6379)

	dbConn := redis.NewClient(&redis.Options{
		Addr:               fmt.Sprintf("%s:%d", host, port),
		Password: password, // no password set
		DB:       0,        // use default DB
		DialTimeout:        10 * time.Second,
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
		PoolSize:           10,
		PoolTimeout:        30 * time.Second,
		IdleTimeout:        500 * time.Millisecond,
		IdleCheckFrequency: 500 * time.Millisecond,
	})
	result := dbConn.Ping()

	return dbConn, result.Err()
}

func (r *RedisDb) GetConn() *redis.Client{
	return r.conn
}

func (m *RedisDb) Reload(cfg *ini.File) error{
	dbConn, err := getRedisClient(cfg)
	if err != nil {
		return err
	}
	//clone the old connection
	oldConn := *m.conn

	//assign the new one
	m.conn = dbConn

	//close the old one
	err = oldConn.Close()
	if err != nil {
		return err
	}
	return nil
}

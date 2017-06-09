package service

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"time"
	"gopkg.in/ini.v1"

	"justdevelop.it/goaway/utils"
)

type InfluxDb struct {
	conn     client.Client
	database string
	username string
}

func NewInfluxDb(cfg *ini.File) (*InfluxDb, error) {

	// Make client
	conn, err := getInfluxClient(cfg)
	adapter := &InfluxDb{
		conn: conn,
		database: cfg.Section("driver_influxdb").Key("database").String(),
	}
	return adapter, err
}

func getInfluxClient(cfg *ini.File) (client.Client, error){
	host:= 	cfg.Section("driver_influxdb").Key("host").String()
	port:= cfg.Section("driver_influxdb").Key("port").String()
	conn, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", host, port),
		Username: cfg.Section("driver_influxdb").Key("user").String(),
		Password: cfg.Section("driver_influxdb").Key("password").String(),
	})
	if err != nil {
		return nil, err
	}
	_, _, err = conn.Ping(time.Second * 5)
	if err != nil {
		utils.LogErrorMessage("Connection to InfuxDb failed")
		return nil, nil
	}
	return conn, err

}
func (inf *InfluxDb) GetConn() client.Client{
	return inf.conn
}

func (inf *InfluxDb) Reload(cfg *ini.File) error{
	dbConn, err := getInfluxClient(cfg)
	if err != nil {
		return err
	}
	//clone the old connection
	oldConn := inf.conn

	//assign the new one
	inf.conn = dbConn

	//close the old one
	err = oldConn.Close()

	return nil
}

func (inf *InfluxDb) NewBatchPoints() (client.BatchPoints, error) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  inf.database,
		Precision: "ns",
	})
	return bp, err
}
func (inf *InfluxDb) AddPointToBatch(
	bp client.BatchPoints,
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
) error {
	point, err := client.NewPoint(
		name,
		tags,
		fields,
		t,
	)
	if err != nil {
		return err
	}
	bp.AddPoint(point)

	return err
}

func (inf *InfluxDb) WriteBatchPoints(bp client.BatchPoints) error {
	return inf.conn.Write(bp)
}

func (inf *InfluxDb) Query(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: inf.database,
	}
	if response, err := inf.conn.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, err
}

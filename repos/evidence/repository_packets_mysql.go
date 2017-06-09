package evidence

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"justdevelop.it/goaway/obj"
	"justdevelop.it/goaway/utils"
	"log"
	"strconv"
	"strings"
	"time"
	"justdevelop.it/goaway/service"
	"sync"
)

type RepositoryPacketsMysql struct {
	Logger *log.Logger
	Table  string
	DbConn *service.MysqlDb
	batchBuffer []*obj.PacketData
	mu sync.Mutex
}

func (db *RepositoryPacketsMysql) mysql_escape(source string) (string, error) {
	var j int = 0
	if len(source) == 0 {
		return "", errors.New("source is null")
	}
	tempStr := source[:]
	desc := make([]byte, len(tempStr)*2)
	for i := 0; i < len(tempStr); i++ {
		flag := false
		var escape byte
		switch tempStr[i] {
		case '\r':
			flag = true
			escape = tempStr[i]
			break
		case '\032':
			flag = true
			escape = 'Z'
			break
		default:
		}
		if flag {
			desc[j] = '\\'
			desc[j+1] = escape
			j = j + 2
		} else {
			desc[j] = tempStr[i]
			j = j + 1
		}
	}
	return string(desc[0:j]), nil
}
func (db *RepositoryPacketsMysql) LazyRecord(data *obj.PacketData) {
	db.mu.Lock()
	db.batchBuffer = append(db.batchBuffer, data)
	db.mu.Unlock()
}
func  (db *RepositoryPacketsMysql) FlushData() {
	db.mu.Lock()
	data := db.batchBuffer
	db.batchBuffer = nil
	db.mu.Unlock()
	db.Record(data)
}

func (db *RepositoryPacketsMysql) Record(batch []*obj.PacketData) {
	if len(batch) > 0 {

		msg := fmt.Sprintf("Packet recorder: flushing %d rows to db", len(batch))
		db.Logger.Println(msg)
		fmt.Println(msg)

		query := "INSERT INTO " + db.Table + " (Ip, DateTime, Host, " +
			"UserAgent, RequestUri, RequestMethod," +
			"PacketSize, Payload, CreatedAt) VALUES "
		insertBatch := []string{}
		for _, packetData := range batch {

			//smallest valid IP is 0.0.0.0
			if len(packetData.ReceivedIp) >= 7 {
				x, _ := db.mysql_escape(packetData.UserAgent)

				logData := []string{
					//strconv.Quote(fmt.Sprintf("%d",l.Id)),
					strconv.Quote(packetData.RealIp),
					strconv.Quote(packetData.DateTime.Format(utils.MYSQLDATETIME)),
					strconv.Quote(packetData.Host),
					strconv.Quote(x),
					strconv.Quote(packetData.RequestUri),
					strconv.Quote(packetData.RequestMethod),
					strconv.Quote(fmt.Sprintf("%d", packetData.PacketSize)),
					strconv.Quote(packetData.Payload),
					strconv.Quote(time.Now().Format(utils.MYSQLDATETIME)),
				}

				insertBatch = append(insertBatch, "("+strings.Join(logData, ",")+")")
			}

		}
		query = query + strings.Join(insertBatch, ",")
		if query != "" {
			_, err := db.DbConn.GetConn().Exec(query)
			utils.CheckAndLogError(err)
		}
	}
}

func (db *RepositoryPacketsMysql) Get(timeSince time.Time) (packets []*obj.PacketData, err error) {

	timeAgo := timeSince.Format(utils.MYSQLDATETIME)
	query := `SELECT id, Ip, DateTime, Host, UserAgent,
	 RequestUri, RequestMethod, PacketSize
	 FROM requests WHERE DateTime >= '` + timeAgo + `'`

	rows, err := db.DbConn.GetConn().Query(query)
	if err != nil {
		utils.LogErrorMessage(err.Error())
	}
	for rows.Next() {

		l := obj.PacketData{}
		err := rows.Scan(
			&l.Id,
			&l.ReceivedIp,
			&l.DateTime,
			&l.Host,
			&l.UserAgent,
			&l.RequestUri,
			&l.RequestMethod,
			&l.PacketSize)
		if err != nil {
			utils.LogErrorMessage(err.Error())
		} else {
			packets = append(packets, &l)
		}

	}
	return packets, err
}

func (db *RepositoryPacketsMysql) Trim(before time.Time) error {
	//todo table name from config and comma separated drivers mysql,redis
	query := fmt.Sprintf("DELETE FROM packets_log WHERE CreatedAt <= '%s'", before.Format(utils.MYSQLDATETIME))
	fmt.Println(query)
	_, err := db.DbConn.GetConn().Exec(query)

	if err != nil {
		utils.LogErrorMessage(err.Error())
		fmt.Println(err)
	}
	return err
}

func (db *RepositoryPacketsMysql) TrimArchive(before time.Time) error {
	//todo table name from config and comma separated drivers mysql,redis
	query := fmt.Sprintf("DELETE FROM packets_log_archive WHERE CreatedAt <= '%s'", before.Format(utils.MYSQLDATETIME))
	fmt.Println(query)
	_, err := db.DbConn.GetConn().Exec(query)

	if err != nil {
		utils.LogErrorMessage(err.Error())
		fmt.Println(err)
	}
	return err
}

func (db *RepositoryPacketsMysql) TrimAndArchive(before time.Time) error {
	//todo table name from config and comma separated drivers mysql,redis
	//todo add archive trim x days
	date := before.Format(utils.MYSQLDATETIME)

	queryA := fmt.Sprintf("INSERT INTO packets_log_archive SELECT * FROM packets_log WHERE CreatedAt <= '%s'", date)
	fmt.Println(queryA)
	_, err := db.DbConn.GetConn().Exec(queryA)

	if err != nil {
		return err
	}
	err = db.Trim(before)
	return err
}

func (db *RepositoryPacketsMysql) Migrate() error {

	var err error
	//create table and archive table
	for _, tname := range []string{db.Table, db.Table + "_archive"} {
		queryF := `CREATE TABLE IF NOT EXISTS %s(
  id bigint(16) NOT NULL AUTO_INCREMENT,
  Ip varchar(225) NOT NULL DEFAULT '',
  DateTime varchar(225) NOT NULL DEFAULT '',
  Host varchar(225) NOT NULL DEFAULT '',
  UserAgent text NOT NULL,
  RequestUri text NOT NULL,
  RequestMethod varchar(225) NOT NULL DEFAULT '',
  PacketSize int(11) NOT NULL,
  Payload text NOT NULL,
  CreatedAt varchar(20) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY DateTimeIndex (DateTime),
  KEY IpIndex (Ip),
  KEY HostIndex (Host),
  KEY MethodIndex (RequestMethod),
  KEY UserAgentIndex (UserAgent(100))
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`
		query := fmt.Sprintf(queryF, tname)
		fmt.Println(query)
		_, err = db.DbConn.GetConn().Exec(query)

		if err != nil {
			utils.LogErrorMessage(err.Error())
			fmt.Println(err)
		}
	}


	return err
}

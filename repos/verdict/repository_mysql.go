package verdict

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"justdevelop.it/goaway/utils"
	"log"
	"time"
	"justdevelop.it/goaway/service"
	"net"
)

type RepositoryMysql struct {
	Repository
	Logger *log.Logger
	Table  string
	DbConn *service.MysqlDb
	defaultSentenceActive bool
}

func (db *RepositoryMysql) Store(sentence *IpSentence) (int64, error){

	sentence.RuleActive = db.defaultSentenceActive
	query := `
	INSERT INTO ip_rules (ip, ttl, banned_by, rule_active, reason, created_at, updated_at) VALUES(?,?,?,?,?,?,?)`

	res, err := db.DbConn.GetConn().Exec(
		query, sentence.Ip,
		sentence.Ttl.Seconds(),
		sentence.BannedBy,
		sentence.RuleActive,
		sentence.Reason,
		sentence.DateTime,
		time.Now().Format(utils.MYSQLDATETIME),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

func (db *RepositoryMysql) Publish(result []*IpSentence) error{

	fmt.Println(fmt.Sprintf("Publishing %d results to db", len(result)))


	//insert if duplicate concatenates bannedBy if same value not already in comma separeted list
	query := `
	INSERT INTO ip_rules (ip, ttl, banned_by, rule_active, reason, created_at, updated_at) VALUES(?,?,?,?,?,?,?)
	ON DUPLICATE KEY UPDATE ttl=?,
	banned_by=TRIM( BOTH ',' FROM (
	CONCAT_WS(',', banned_by,
	CASE WHEN LOCATE(?, banned_by) THEN '' END ,
	CASE WHEN NOT LOCATE(?, banned_by) THEN ? END) ) ), reason=CONCAT_WS(',',reason,?), updated_at=?;`

	for v := range result {
		//to enable / disable guardian enforcement but still see result and isnert them manually
		result[v].RuleActive = db.defaultSentenceActive

		_, err := db.DbConn.GetConn().Exec(
			query, result[v].Ip,
			result[v].Ttl.Seconds(),
			result[v].BannedBy,
			result[v].RuleActive,
			result[v].Reason,
			result[v].DateTime,
			result[v].DateTime,

			result[v].Ttl.Seconds(),
			result[v].BannedBy,
			result[v].BannedBy,
			result[v].BannedBy,
			result[v].Reason,
			time.Now().Format(utils.MYSQLDATETIME),
		)
		return err
	}
	return nil
}

func (db *RepositoryMysql) FetchRules() (result []IpSentence, err error) {

	query := "SELECT ip, ttl, banned_by, rule_active, reason, updated_at FROM ip_rules "
	rows, err := db.DbConn.GetConn().Query(query)
	if err != nil {
		fmt.Println(err)
		utils.PanicMessage(err.Error())
	}
	for rows.Next() {

		var rule = IpSentence{}
		var stringNull = sql.NullString{}
		var stringReasonNull = sql.NullString{}
		var ttlDuration = ""
		if err := rows.Scan(
			&rule.Ip,
			&ttlDuration,
			&stringNull,
			&rule.RuleActive,
			&stringReasonNull,
			&rule.DateTime,
		); err != nil {
			utils.LogErrorMessage(err.Error())
			return result, err
		}
		rule.Ttl, err = time.ParseDuration(ttlDuration + "s")
		rule.BannedBy = stringNull.String
		rule.Reason = stringReasonNull.String
		if rule.Ip != "" {
			rule.IpInt = utils.Ip2int(net.ParseIP(rule.Ip))
			result = append(result, rule)
		}
	}
	rows.Close()
	return
}

func (db *RepositoryMysql) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS ip_rules (
  id int(11) NOT NULL AUTO_INCREMENT,
  ip varchar(255) DEFAULT NULL,
  ttl int(11) DEFAULT NULL,
  banned_by varchar(255) DEFAULT '',
  rule_active boolean DEFAULT 0,
  reason varchar(255) NULL,
  created_at datetime DEFAULT NULL,
  updated_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY ip_rules_ip_IDX (ip),
  KEY ip_rules_updated_at_IDX (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`
	_, err := db.DbConn.GetConn().Exec(query)
	fmt.Println(query)

	return err
}

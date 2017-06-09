package whitelist

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"justdevelop.it/goaway/obj"
	"justdevelop.it/goaway/utils"
	"log"
	"justdevelop.it/goaway/service"
)

type RepositoryMysql struct {
	Logger      *log.Logger
	Table       string
	DbConn      *service.MysqlDb
	batchBuffer []*obj.PacketData
}

func (db *RepositoryMysql) GetAllOfType(wtype string) ([]string, error) {
	query := "SELECT address FROM " + db.Table + " WHERE type = ?"
	rows, err := db.DbConn.GetConn().Query(query, wtype)
	utils.CheckAndLogError(err)

	results := make([]string, 0)
	for rows.Next() {

		var address string
		err := rows.Scan(&address)
		if err != nil {
			utils.LogErrorMessage(err.Error())
		} else {
			results = append(results, address)
		}

	}
	return results, err
}

func (db *RepositoryMysql) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS %s (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  address varchar(18) DEFAULT NULL,
  type enum('ip','net') DEFAULT 'ip',
  date datetime DEFAULT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`
	query = fmt.Sprintf(query, db.Table)
	fmt.Println(query)
	_, err := db.DbConn.GetConn().Exec(query)

	if err != nil {
		utils.LogErrorMessage(err.Error())
		fmt.Println(err)
	}
	return err
}

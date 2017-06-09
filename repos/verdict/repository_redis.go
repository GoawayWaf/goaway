package verdict

import (
	"fmt"
	_"encoding/json"
	"justdevelop.it/goaway/service"
	_"justdevelop.it/goaway/utils"
)

type RepositoryRedis struct {
	dbConn *service.RedisDb
}
func (db *RepositoryRedis) Publish(result []*IpSentence, whitelistIps map[string]string) {
	/*for _, v := range result {
		if _, inMap := whitelistIps[result[v].Ip]; !inMap{
			js, err := json.Marshal(v)
			utils.CheckError(err)
			reply := db.dbConn.GetConn().HSet("ip_rules", v.Ip, js)
			reply.Result()
		}
	}*/
}
func (db *RepositoryRedis) FetchRules() (result []IpSentence, error error) {

	fmt.Println("Fetching db")
	reply := db.dbConn.GetConn().HGetAll("ip_rules")
	rules, err := reply.Result()
	if err != nil {
		error = err
		return
	}
	for _, data := range rules {
		fmt.Println(rules)
		fmt.Print(data)

	}

	return
}

func (db *RepositoryRedis) Migrate() error {
	return nil
}

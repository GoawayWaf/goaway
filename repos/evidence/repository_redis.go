package evidence

import (
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/ini.v1"
	"gopkg.in/redis.v5"
	"log"
	"justdevelop.it/goaway/service"
)

func NewRepositoryRedis(cfg *ini.File, sm service.ServiceManager, logger *log.Logger) *RepositoryRedis {

	repo := RepositoryRedis{
		DbConn: sm.GetRedisDb(),
		Logger: logger,
		Table: "gpr_pps",
	}

	return &repo
}

type RepositoryRedis struct {
	Logger *log.Logger
	Table  string
	DbConn *service.RedisDb
}

func (r *RepositoryRedis) Record(score float64, member string) {
	r.DbConn.GetConn().ZIncr(r.Table, redis.Z{Score: score, Member: member})
}

func (r *RepositoryRedis) GetAllIpPps() ([]redis.Z, error) {
	opt := redis.ZRangeBy{
		Min: "0",
		Max: "+inf",
	}
	ipPps, err := r.DbConn.GetConn().ZRevRangeByScoreWithScores(r.Table, opt).Result()
	if err != nil {
		return nil, err
	}
	return ipPps, nil

}
func (r *RepositoryRedis) GetIpPpsAboveScore(score string) ([]redis.Z, error) {
	opt := redis.ZRangeBy{
		Min: score,
		Max: "+inf",
	}
	ipPps, err := r.DbConn.GetConn().ZRevRangeByScoreWithScores(r.Table, opt).Result()
	if err != nil {
		return nil, err
	}
	return ipPps, nil

}

func (r *RepositoryRedis) RemoveBelowScore(maxScore string) error {
	_, err := r.DbConn.GetConn().ZRemRangeByScore(r.Table, "0", maxScore).Result()
	return err
}

func (r *RepositoryRedis) RemoveBelowLex(maxLex string) error {
	cmnd := redis.NewIntCmd(
		"zremrangebylex",
		r.Table,
		"-",
		"[" + maxLex,
	)
	r.DbConn.GetConn().Process(cmnd)
	_, err := cmnd.Result()

	return err
}

func (r *RepositoryRedis) Remove(z redis.Z) error {
	_, err := r.DbConn.GetConn().ZRem(r.Table, z.Member).Result()

	return err
}

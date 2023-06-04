package config

import (
	"github.com/go-redis/redis/v8"
	sessdb "github.com/kataras/iris/v12/sessions/sessiondb/redis"
	"time"
)

// 返回redis实例
func NewRedis() *redis.Client {
	var database *redis.Client
	//项目配置

	cmsConfig := InitConfig()
	if cmsConfig != nil {
		//iris.New().Logger().Info("  hello  ")
		rd := cmsConfig.Redis
		//iris.New().Logger().Info(rd)
		database = redis.NewClient(&redis.Options{
			Network:    rd.NetWork,
			Addr:       rd.Addr + ":" + rd.Port,
			Password:   rd.Password,
			DB:         rd.Database,
			MaxRetries: 3,
			//MinRetryBackoff:   redis.DefaultRedisTimeout,
		})
	} else {
		panic(" InitConfig  error ")
	}
	return database
}

func NewSessStorage() *sessdb.Database {
	var database *sessdb.Database
	//项目配置
	//syncOnce.Do(func() {
	cmsConfig := InitConfig()
	//dump.Printf(cmsConfig)
	if cmsConfig != nil {
		rd := cmsConfig.Redis
		database = sessdb.New(sessdb.Config{
			Network:   rd.NetWork,
			Addr:      rd.Addr + ":" + rd.Port,
			Password:  rd.Password,
			Database:  "2",
			MaxActive: 10,
			Timeout:   10 * time.Second,
			Prefix:    rd.Prefix,
		})
	} else {
		panic(" InitConfig  error ")
	}
	//})
	return database
}

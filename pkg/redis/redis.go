package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/lgbya/go-dump"
	"luckypay/pkg/config"
)

var Client *redis.Client

func Init() {
	cmsConfig := config.Instance
	if cmsConfig != nil {
		rd := cmsConfig.Redis
		dump.Printf(rd)
		Client = redis.NewClient(&redis.Options{
			Network:    rd.NetWork,
			Addr:       rd.Addr,
			Password:   rd.Password,
			DB:         rd.Db,
			MaxRetries: 3,
			//MinRetryBackoff:   redis.DefaultRedisTimeout,
		})
	} else {
		panic("InitConfig  error")
	}

}

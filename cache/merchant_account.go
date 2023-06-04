package cache

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"luckypay/repositories"
	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var MerchantAccount = newMerchantAccount()

func newMerchantAccount() *merchantAccount {
	return &merchantAccount{}
}

type merchantAccount struct {
	RedisClient *redis.Client
}

func (s *merchantAccount) RefreshCache() (merchantData []model.MerchantAccount) {
	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData = repositories.MerchantAccountRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		//fmt.Println("%+v", val)
		jsons, err := json.Marshal(val)
		if err != nil {
			logrus.Error(err)
		}
		_, err = redisServer.Set(ctx, "merchantAccount:"+val.LoginName, jsons, 0).Result()
		if err != nil {
			logrus.Error(err)
			return
		}
	}
	return
}

func (s *merchantAccount) RefreshOne(accountId int64) {
	merchantData := model.MerchantAccount{}
	err := sqls.DB().Where("accountId = ?", accountId).First(&merchantData).Error
	if err != nil {
		logrus.Error(err)
		return
	}

	jsons, err := json.Marshal(merchantData)
	if err != nil {
		logrus.Error(err)
		return
	}
	_, err = redisServer.Set(ctx, "merchantAccount:"+merchantData.LoginName, jsons, 0).Result()
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

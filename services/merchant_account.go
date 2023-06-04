package services

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"luckypay/repositories"
	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var MerchantAccountService = newMerchantAccountService()

func newMerchantAccountService() *merchantAccountService {
	return &merchantAccountService{}
}

type merchantAccountService struct {
	RedisClient *redis.Client
}

func (s *merchantAccountService) Get(id int64) *model.MerchantAccount {
	return repositories.MerchantAccountRepository.Get(sqls.DB(), id)
}

func (s *merchantAccountService) Take(where ...interface{}) *model.MerchantAccount {
	return repositories.MerchantAccountRepository.Take(sqls.DB(), where...)
}

func (s *merchantAccountService) Find(cnd *sqls.Cnd) []model.MerchantAccount {
	return repositories.MerchantAccountRepository.Find(sqls.DB(), cnd)
}

func (s *merchantAccountService) FindOne(cnd *sqls.Cnd) *model.MerchantAccount {
	return repositories.MerchantAccountRepository.FindOne(sqls.DB(), cnd)
}

func (s *merchantAccountService) RefreshCache() (merchantData []model.MerchantAccount) {
	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData = repositories.MerchantAccountRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		jsons, err := json.Marshal(val)
		if err != nil {
			logrus.Error(err)
			continue
		}
		_, err = redisServer.Set(ctx, "merchantAccount:"+val.LoginName, jsons, 0).Result()
		if err != nil {
			logrus.Error(err)
			continue
		}
	}
	return
}

func (s *merchantAccountService) RefreshOne(accountId int64) {
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

func (s *merchantAccountService) GetCacheByLoginName(loginName string) (merchant model.MerchantAccount, err error) {

	key := "merchantAccount:" + loginName
	jsons, err := redisServer.Get(ctx, key).Result()
	if err != nil {
		return merchant, err
	}

	err = json.Unmarshal([]byte(jsons), &merchant)
	if err != nil {
		//panic(err.Error())
		return merchant, err
	}
	return merchant, err
}

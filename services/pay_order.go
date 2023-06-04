package services

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"luckypay/repositories"
	"strconv"

	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var PayOrder = newPayOrder()

func newPayOrder() *payOrder {
	return &payOrder{}
}

type payOrder struct {
	RedisClient *redis.Client
}

func (s *payOrder) Get(id int64) *model.PlatformPayOrder {
	return repositories.PayRepository.Get(sqls.DB(), id)
}

func (s *payOrder) Take(where ...interface{}) *model.PlatformPayOrder {
	return repositories.PayRepository.Take(sqls.DB(), where...)
}

func (s *payOrder) Find(cnd *sqls.Cnd) []model.PlatformPayOrder {
	return repositories.PayRepository.Find(sqls.DB(), cnd)
}

func (s *payOrder) FindOne(cnd *sqls.Cnd) *model.PlatformPayOrder {
	return repositories.PayRepository.FindOne(sqls.DB(), cnd)
}

func (s *payOrder) RefreshCache() {

	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData := repositories.MerchantRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//fmt.Println("%+v", val)
		cnd = sqls.NewCnd().Where("merchantId = ?", val.MerchantID).Desc("merchantId")
		payOrders := repositories.PayRepository.Find(sqls.DB(), cnd)
		jsons, errs := json.Marshal(payOrders)
		//fmt.Println(jsons)
		if errs != nil {
			//fmt.Println()
			iris.New().Logger().Info(errs.Error())
			return
		}
		_, errs = redisServer.Set(ctx, "payOrder:n:"+val.MerchantNo, jsons, 0).Result()

		if errs != nil {
			logrus.Error(errs)
			return
		}
		_, errs = redisServer.Set(ctx, "payOrder:i:"+strconv.FormatInt(val.MerchantID, 10), jsons, 0).Result()
		if errs != nil {
			logrus.Error(errs)
			return
		}
	}
	return
}

func (s *payOrder) GetCacheByMerchanNo(merchantNo string) (merchantRates []model.PlatformPayOrder, res bool) {
	key := "payOrder:n:" + merchantNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	//parseMerchantRates := []model.MerchantRate
	err := json.Unmarshal([]byte(jsons), &merchantRates)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	res = true
	return
}

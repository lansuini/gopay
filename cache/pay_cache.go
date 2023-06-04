package cache

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"time"
)

var PayCache = newPayCache()

func newPayCache() *payCache {
	return &payCache{}
}

type payCache struct {
	RedisClient *redis.Client
}

func (s *payCache) GetCacheByMerchantOrderNo(merchantNo string, merchantOrderNo string) (payOrder model.PlatformPayOrder, res bool) {

	key := "payorder:m:" + merchantNo + ":" + merchantOrderNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	err := json.Unmarshal([]byte(jsons), &payOrder)
	if err != nil {
		iris.New().Logger().Info(err.Error())
		return
	}
	res = true
	return
}

func (s *payCache) RefreshOne(platformOrderNo string) {
	var order model.PlatformPayOrder
	err := sqls.DB().Where("platformOrderNo = ?", platformOrderNo).First(&order).Error
	if err != nil {
		logrus.Error(platformOrderNo + "刷新代付订单失败-" + err.Error())
		return
	}
	key1 := "payorder:" + platformOrderNo
	key2 := "payorder:m:" + order.MerchantNo + ":" + order.MerchantOrderNo

	jsons, err := json.Marshal(order)
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo1 : ", err.Error())
	}
	_, err = redisServer.Set(ctx, key1, jsons, 7*24*time.Hour).Result()
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo2 : ", err.Error())
	}
	_, err = redisServer.Set(ctx, key2, jsons, 7*24*time.Hour).Result()
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo3 : ", err.Error())
	}
}

func (s *payCache) SetCacheByPlatformOrderNo(platformOrderNo string, payStruct model.PlatformPayOrder) {
	key1 := "payorder:" + platformOrderNo
	key2 := "payorder:m:" + payStruct.MerchantNo + ":" + payStruct.MerchantOrderNo

	jsons, err := json.Marshal(payStruct)
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo1 : ", err.Error())
	}
	_, err = redisServer.Set(ctx, key1, jsons, 7*24*time.Hour).Result()
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo2 : ", err.Error())
	}
	_, err = redisServer.Set(ctx, key2, jsons, 7*24*time.Hour).Result()
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo3 : ", err.Error())
	}

}

func (s *payCache) GetCacheByPlatformOrderNo(platformOrderNo string) (payOrder model.PlatformPayOrder, res bool) {
	key := "payorder:" + platformOrderNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	err := json.Unmarshal([]byte(jsons), &payOrder)
	if err != nil {
		logrus.Error(platformOrderNo, "-payCache-GetCacheByPlatformOrderNo-", err.Error())
		return
	}
	res = true
	return

}

func (s *payCache) GetReqPayOrderParams(orderData model.PlatformPayOrder) (payOrderParams model.ReqPayOrder) {

	payOrderParams.MerchantNo = orderData.MerchantNo
	payOrderParams.MerchantOrderNo = orderData.MerchantOrderNo
	payOrderParams.OrderAmount = orderData.RealOrderAmount
	payOrderParams.PayType = orderData.PayType
	payOrderParams.PayModel = orderData.PayModel
	payOrderParams.CardType = orderData.CardType
	payOrderParams.BankCode = orderData.BankCode

	return
}

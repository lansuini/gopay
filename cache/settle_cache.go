package cache

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"

	//"luckypay/channels"
	"luckypay/config"
	"luckypay/utils"
	"strconv"
	"time"

	"github.com/mlogclub/simple/sqls"
	"luckypay/model"
)

var SettleCache = newSettleCache()

func newSettleCache() *settleCache {
	return &settleCache{}
}

type settleCache struct {
	RedisClient *redis.Client
}

func (s *settleCache) GetCacheByMerchantOrderNo(merchantNo string, merchantOrderNo string) (payOrder model.PlatformSettlementOrder, res bool) {

	key := "settlementorder:m:" + merchantNo + ":" + merchantOrderNo
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

func (s *settleCache) GetPlatformOrderNo(capLetter string) string {
	redisServer = config.NewRedis()
	now := time.Now()
	ymdkey := utils.GetFormatTime(now)

	fmt.Print("ymdkey:", ymdkey, "\n")
	p1 := utils.GetYMDHISTime(now)
	fmt.Print("p1:", p1, "\n")

	p2, _ := redisServer.Get(ctx, ymdkey).Result()
	if p2 == "" {
		p2 = utils.RandomString(3)
		_, err := redisServer.Set(ctx, ymdkey, p2, 48*time.Hour).Result()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	p3 := strconv.FormatInt(utils.GetTimeTick64(), 10)
	p4 := utils.RandomString(3)
	orderNo := capLetter + p1 + p2 + p3 + p4
	redisServer.SAdd(ctx, ymdkey+":settlementorder", orderNo).Result()
	redisServer.Expire(ctx, ymdkey+":settlementorder", 48*time.Hour)
	fmt.Print("orderNo:", orderNo, "\n")
	return orderNo

}

func (s *settleCache) SetCacheByPlatformOrderNo(platformOrderNo string, payStruct model.PlatformSettlementOrder) {
	key1 := "settlementorder:" + platformOrderNo
	key2 := "settlementorder:m:" + payStruct.MerchantNo + ":" + payStruct.MerchantOrderNo

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

func (s *settleCache) GetCacheByPlatformOrderNo(platformOrderNo string) (payOrder model.PlatformSettlementOrder, res bool) {
	key := "settlementorder:" + platformOrderNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	err := json.Unmarshal([]byte(jsons), &payOrder)
	if err != nil {
		logrus.Error(platformOrderNo, "-settleCache-GetCacheByPlatformOrderNo-", err.Error())
		return
	}
	res = true
	return

}

func (s *settleCache) RefreshOne(platformOrderNo string) {
	var order model.PlatformSettlementOrder
	err := sqls.DB().Where("platformOrderNo = ?", platformOrderNo).First(&order).Error
	if err != nil {
		logrus.Error(platformOrderNo + "刷新代付订单失败-" + err.Error())
		return
	}
	key1 := "settlementorder:" + platformOrderNo
	key2 := "settlementorder:m:" + order.MerchantNo + ":" + order.MerchantOrderNo

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

func (s *settleCache) DelCacheByPlatformOrderNo(platformOrderNo string, MerchantNo string, MerchantOrderNo string) {
	key1 := "settlementorder:" + platformOrderNo
	key2 := "settlementorder:m:" + MerchantNo + ":" + MerchantOrderNo
	redisServer.Del(ctx, key1)
	redisServer.Del(ctx, key2)

}

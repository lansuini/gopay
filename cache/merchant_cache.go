package cache

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"luckypay/repositories"
	"strconv"
	"time"

	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var MerchantCache = newMerchantCache()

type DayStatSearch struct {
	Counts                int64   `gorm:"column:counts;" json:"counts"`
	OrderAmounts          float64 `gorm:"column:orderAmounts;" json:"orderAmounts"`
	ServiceCharges        float64 `gorm:"column:serviceCharges;" json:"serviceCharges"`
	ChannelServiceCharges float64 `gorm:"column:channelServiceCharges;" json:"channelServiceCharges"`
	AgentFees             float64 `gorm:"column:agentFees;" json:"agentFees"`
}

func newMerchantCache() *merchantCache {
	return &merchantCache{}
}

type merchantCache struct {
	RedisClient *redis.Client
}

func (s *merchantCache) RefreshCache() (merchantData []model.Merchant) {
	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData = repositories.MerchantRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		//fmt.Println("%+v", val)
		jsons, err := json.Marshal(val)
		if err != nil {
			logrus.Error("merchant:n:"+val.MerchantNo, "merchantCache-RefreshCache-json.Marshal error: ", err)
			return
		}
		_, err = redisServer.Set(ctx, "merchant:n:"+val.MerchantNo, jsons, 0).Result()

		if err != nil {
			logrus.Error("merchant:n:"+val.MerchantNo, "merchantCache-RefreshCache-redis-Set error: ", err)
			return
		}
		_, err = redisServer.Set(ctx, "merchant:i:"+strconv.FormatInt(val.MerchantID, 10), jsons, 0).Result()
		if err != nil {
			logrus.Error("merchant:i:"+strconv.FormatInt(val.MerchantID, 10), "merchantCache-RefreshCache-redis-Set error: ", err)
			return
		}
	}
	return
}

func (s *merchantCache) RefreshOne(merchantNo string) {
	merchantData := model.Merchant{}
	err := sqls.DB().Table("merchant").Where("merchantNo = ?", merchantNo).First(&merchantData).Error
	if err != nil {
		logrus.Error(merchantNo, "merchantService RefreshOne Fail: ", err.Error())
	}

	jsons, err := json.Marshal(merchantData)
	if err != nil {
		logrus.Error(merchantNo, "merchantService RefreshOne Fail: ", err.Error())
	}
	_, err = redisServer.Set(ctx, "merchant:n:"+merchantNo, jsons, 0).Result()
	if err != nil {
		logrus.Error(merchantNo, "merchantService RefreshOne Fail: ", err.Error())
		return
	}
	_, err = redisServer.Set(ctx, "merchant:i:"+strconv.FormatInt(merchantData.MerchantID, 10), jsons, 0).Result()
	if err != nil {
		logrus.Error(merchantNo, "merchantService RefreshOne Fail: ", err.Error())
		return
	}

	return
}

func (c *merchantCache) SetCacheMerchantData() {
	merchants := []model.Merchant{}
	err := sqls.DB().Table("merchant").Find(&merchants).Error
	if err != nil {
		logrus.Error("GetCacheMerchantData error : " + err.Error())
		return
	}
	jsons, err := json.Marshal(merchants)
	if err != nil {
		logrus.Error("SetCacheMerchantData-error", err)
		return
	}

	_, err = redisServer.Set(ctx, "merchants", jsons, 0).Result()

	if err != nil {
		logrus.Error("GetCacheMerchantData error2 : " + err.Error())
		return
	}

}

func (s *merchantCache) GetCacheByMerchantNo(merchantNo string) (merchant model.Merchant, res bool) {

	key := "merchant:n:" + merchantNo
	jsons, weds := redisServer.Get(ctx, key).Result()
	if weds != nil {
		return merchant, false
	}

	err := json.Unmarshal([]byte(jsons), &merchant)
	if err != nil {
		//panic(err.Error())
		return merchant, false
	}
	//fmt.Println("%+v", merchant)
	return merchant, true
}

func (s *merchantCache) GetCacheByDaySettleAmountLimit(merchantNo string) (amount float64, err error) {
	err = nil
	key := "merchant:settle:tc:" + time.Now().Format("20060102") + ":" + merchantNo
	cmd := redisServer.Get(ctx, key)
	if err = cmd.Err(); err != nil {
		if err == redis.Nil {
			logrus.Error("GetCacheByDaySettleAmountLimit key does not exists")
			return 0, nil
		}
		return 0, err
	}
	amount, err = cmd.Float64()
	if err != nil {
		return 0, err
	}
	return amount, nil
}

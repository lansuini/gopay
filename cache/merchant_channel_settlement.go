package cache

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"strconv"
	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var MerchantChannelSettlement = newMerchantChannelSettlement()

func newMerchantChannelSettlement() *merchantChannelSettlement {
	return &merchantChannelSettlement{}
}

type merchantChannelSettlement struct {
	RedisClient *redis.Client
}

func (s *merchantChannelSettlement) RefreshCache() {

	merchantData := []model.Merchant{}
	err := sqls.DB().Table("merchant").Find(&merchantData).Error
	if err != nil {
		logrus.Error("merchantChannelSettlement RefreshCache Find Error-", err)
		return
	}

	for _, merchant := range merchantData {
		merchantChannelSettlements := []model.MerchantChannelSettlement{}
		err = sqls.DB().Where("merchantId = ?", merchant.MerchantID).Find(&merchantChannelSettlements).Error
		if err != nil {
			logrus.Error("merchantChannelSettlement RefreshCache Find Error2-", err)
			continue
		}
		jsons, errs := json.Marshal(merchantChannelSettlements)
		//fmt.Println(jsons)
		if errs != nil {
			//fmt.Println()
			iris.New().Logger().Info(errs.Error())
			return
		}
		_, errs = redisServer.Set(ctx, "merchantChannelSettle:n:"+merchant.MerchantNo, jsons, 0).Result()

		if errs != nil {
			iris.New().Logger().Info(errs.Error())
			return
		}
		_, errs = redisServer.Set(ctx, "merchantChannelSettle:i:"+strconv.FormatInt(merchant.MerchantID, 10), jsons, 0).Result()
		if errs != nil {
			iris.New().Logger().Info(errs.Error())
			return
		}
	}
	return
}

func (s *merchantChannelSettlement) RefreshOne(merchantNo string, merchantId int64) {
	merchantChannels := []model.MerchantChannelSettlement{}
	err := sqls.DB().Table("merchant_channel_settlement").Where("merchantNo = ?", merchantNo).Find(&merchantChannels).Error
	if err != nil {
		logrus.Error("merchant_channel-RefreshOne: "+merchantNo, err)
		return
	}

	jsons, errs := json.Marshal(merchantChannels)
	//fmt.Println(jsons)
	if errs != nil {
		//fmt.Println()
		logrus.Error(merchantNo, "merchant_channel-RefreshOne: json.Marshal", err)
		return
	}
	_, errs = redisServer.Set(ctx, "merchantChannelSettle:n:"+merchantNo, jsons, 0).Result()

	if errs != nil {
		iris.New().Logger().Info(errs.Error())
		return
	}
	_, errs = redisServer.Set(ctx, "merchantChannelSettle:i:"+strconv.FormatInt(merchantId, 10), jsons, 0).Result()
	if errs != nil {
		iris.New().Logger().Info(errs.Error())
		return
	}
	return
}

func (s *merchantChannelSettlement) GetCacheByMerchantNo(merchantNo string) (merchantRates []model.MerchantChannelSettlement, res bool) {
	key := "merchantChannelSettle:n:" + merchantNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	//parseMerchantRates := []model.MerchantRate
	err := json.Unmarshal([]byte(jsons), &merchantRates)
	if err != nil {
		iris.New().Logger().Info(err.Error())
		return
	}
	res = true
	return
}

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

var ChannelMerchant = newChannelMerchant()

func newChannelMerchant() *channelMerchant {
	return &channelMerchant{}
}

type channelMerchant struct {
	RedisClient *redis.Client
}

func (s *channelMerchant) RefreshCache() {
	var channelMerchants []model.ChannelMerchant
	err := sqls.DB().Table("channel_merchant").Find(&channelMerchants).Error
	if err != nil {
		logrus.Error("ChannelMerchant RefreshCache fail : " + err.Error())
	}

	for _, val := range channelMerchants {
		merchantData := model.ChannelMerchant{}
		err := sqls.DB().Table("channel_merchant").Where("channelMerchantNo = ?", val.ChannelMerchantNo).First(&merchantData).Error
		if err != nil {
			logrus.Error(val.ChannelMerchantNo, "ChannelMerchant 查询失败：", err.Error())
			return
		}
		jsons, errs := json.Marshal(merchantData)
		//fmt.Println(jsons)
		if errs != nil {
			//fmt.Println()
			iris.New().Logger().Info(errs.Error())
			continue
		}
		_, errs = redisServer.Set(ctx, "channelMerchant:n:"+val.ChannelMerchantNo, jsons, 0).Result()

		if errs != nil {
			iris.New().Logger().Info(errs.Error())
			continue
		}
		_, errs = redisServer.Set(ctx, "channelMerchant:i:"+strconv.FormatInt(val.ChannelMerchantID, 10), jsons, 0).Result()
		if errs != nil {
			iris.New().Logger().Info(errs.Error())
			continue
		}
	}
	return
}

func (s *channelMerchant) RefreshOne(merchantNo string) {
	merchantData := model.ChannelMerchant{}
	err := sqls.DB().Table("channel_merchant").Where("channelMerchantNo = ?", merchantNo).First(&merchantData).Error
	if err != nil {
		logrus.Error(merchantNo, "ChannelMerchant 查询失败：", err.Error())
		return
	}
	jsons, errs := json.Marshal(merchantData)
	//fmt.Printf(string(jsons))

	_, errs = redisServer.Set(ctx, "channelMerchant:n:"+merchantData.ChannelMerchantNo, string(jsons), 0).Result()

	if errs != nil {
		logrus.Info(errs.Error())
		return
	}
	_, errs = redisServer.Set(ctx, "channelMerchant:i:"+strconv.FormatInt(merchantData.ChannelMerchantID, 10), string(jsons), 0).Result()
	if errs != nil {
		logrus.Info(errs.Error())
		return
	}

	return
}

func (s *channelMerchant) GetCacheByChannelMerchantNo(ChannelMerchantNo string) (cacheData model.ChannelMerchant, res bool) {

	key := "channelMerchant:n:" + ChannelMerchantNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	//parseMerchantRates := []model.MerchantRate
	err := json.Unmarshal([]byte(jsons), &cacheData)
	if err != nil {
		iris.New().Logger().Info(err.Error())
		return
	}
	res = true
	return
}

func (s *channelMerchant) GetCacheByChannelMerchantId(ChannelMerchantId int) (cacheData model.ChannelMerchant, res bool) {

	ChannelMerchantIdStr := strconv.Itoa(ChannelMerchantId)
	key := "channelMerchant:n:" + ChannelMerchantIdStr
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	//parseMerchantRates := []model.MerchantRate
	err := json.Unmarshal([]byte(jsons), &cacheData)
	if err != nil {
		iris.New().Logger().Info(err.Error())
		return
	}
	res = true
	return
}

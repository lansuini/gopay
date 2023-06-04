package cache

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"luckypay/repositories"
	"strconv"
	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var ChannelMerchantRate = newChannelMerchantRate()

func newChannelMerchantRate() *channelMerchantRate {
	return &channelMerchantRate{}
}

type channelMerchantRate struct {
	RedisClient *redis.Client
}

// GetServiceChargeSettle

func (s *channelMerchantRate) RefreshCache() {
	cnd := sqls.NewCnd().Desc("channelMerchantId")
	merchantData := repositories.ChannelMerchantRepository.Find(sqls.DB(), cnd)
	//merchantData = repositories.ChannelMerchantRateRepository.Find(sqls.DB(), cnd)

	//_, weds := redisServer.Set(ctx, "test", "4535", 60*time.Second).Result()
	//if weds != nil {
	//	fmt.Println(weds)
	//	return
	//}

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		//fmt.Println("%+v", val)
		cnd = sqls.NewCnd().Where("channelMerchantId = ?", val.ChannelMerchantID).Desc("channelMerchantId")
		channelMerchantRate := repositories.ChannelMerchantRateRepository.Find(sqls.DB(), cnd)
		jsons, errs := json.Marshal(channelMerchantRate)
		//fmt.Println(jsons)
		if errs != nil {
			fmt.Println(errs.Error())
		}
		_, cerr := redisServer.Set(ctx, "channelMerchantRate:n:"+val.ChannelMerchantNo, jsons, 0).Result()

		if cerr != nil {
			fmt.Println(cerr)
			return
		}
		_, cerr = redisServer.Set(ctx, "channelMerchantRate:i:"+strconv.FormatInt(val.ChannelMerchantID, 10), jsons, 0).Result()
		if cerr != nil {
			fmt.Println(cerr)
			return
		}
	}
	return
}

func (s *channelMerchantRate) RefreshOne(merchantNo string) {
	//fmt.Println(merchantNo)
	merchantData := model.ChannelMerchant{}
	err := sqls.DB().Table("channel_merchant").Where("channelMerchantNo = ?", merchantNo).First(&merchantData).Error
	if err != nil {
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-FindOne-null")
		return
	}
	cnd := sqls.NewCnd().Where("channelMerchantNo = ?", merchantNo)
	rateData := repositories.ChannelMerchantRateRepository.Find(sqls.DB(), cnd)
	jsons, err := json.Marshal(rateData)
	if err != nil {
		fmt.Println(err.Error())
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-json.Marshal-", err.Error())
		return
	}

	_, err = redisServer.Set(ctx, "channelMerchantRate:n:"+merchantNo, jsons, 0).Result()

	if err != nil {
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-redisServer.Set-", err.Error())
		fmt.Println(err.Error())
		return
	}
	_, err = redisServer.Set(ctx, "channelMerchantRate:i:"+strconv.FormatInt(merchantData.ChannelMerchantID, 10), jsons, 0).Result()
	if err != nil {
		logrus.Error(merchantData.ChannelMerchantID, "-merchantRate-RefreshOne-redisServer.Set-", err.Error())
		fmt.Println(err.Error())
		return
	}

}

func (s *channelMerchantRate) GetCacheByMerchantNo(merchantNo string) (channelRate []model.ChannelMerchantRate, res bool) {
	key := "channelMerchantRate:n:" + merchantNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	//parseChannelMerchantRates := []model.ChannelMerchantRate
	err := json.Unmarshal([]byte(jsons), &channelRate)
	if err != nil {
		iris.New().Logger().Info(err.Error())
		return
	}
	res = true
	return
}

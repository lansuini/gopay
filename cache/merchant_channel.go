package cache

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

var MerchantChannel = newMerchantChannel()

func newMerchantChannel() *merchantChannel {
	return &merchantChannel{}
}

type merchantChannel struct {
	RedisClient *redis.Client
}

func (s *merchantChannel) RefreshCache() {

	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData := repositories.MerchantRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//fmt.Println("%+v", val)
		cnd = sqls.NewCnd().Where("merchantId = ?", val.MerchantID).Desc("merchantId")
		merchantChannel := repositories.MerchantChannelRepository.Find(sqls.DB(), cnd)
		jsons, errs := json.Marshal(merchantChannel)
		//fmt.Println(jsons)
		if errs != nil {
			//fmt.Println()
			iris.New().Logger().Info(errs.Error())
			return
		}
		_, errs = redisServer.Set(ctx, "merchantChannel:n:"+val.MerchantNo, jsons, 0).Result()

		if errs != nil {
			iris.New().Logger().Info(errs.Error())
			return
		}
		_, errs = redisServer.Set(ctx, "merchantChannel:i:"+strconv.FormatInt(val.MerchantID, 10), jsons, 0).Result()
		if errs != nil {
			iris.New().Logger().Info(errs.Error())
			return
		}
	}
	return
}

func (s *merchantChannel) RefreshOne(merchantNo string, merchantId int64) {
	merchantChannels := []model.MerchantChannel{}
	err := sqls.DB().Table("merchant_channel").Where("merchantNo = ?", merchantNo).Find(&merchantChannels).Error
	if err != nil {
		logrus.Error("merchant_channel-RefreshOne: " + merchantNo + err.Error())
		return
	}

	jsons, errs := json.Marshal(merchantChannels)
	//fmt.Println(jsons)
	if errs != nil {
		//fmt.Println()
		iris.New().Logger().Info(errs.Error())
		return
	}
	_, errs = redisServer.Set(ctx, "merchantChannel:n:"+merchantNo, jsons, 0).Result()

	if errs != nil {
		iris.New().Logger().Info(errs.Error())
		return
	}
	_, errs = redisServer.Set(ctx, "merchantChannel:i:"+strconv.FormatInt(merchantId, 10), jsons, 0).Result()
	if errs != nil {
		iris.New().Logger().Info(errs.Error())
		return
	}
	return
}

func (s *merchantChannel) GetCacheByMerchantNo(merchantNo string) (merchantRates []model.MerchantChannel, res bool) {
	key := "merchantChannel:n:" + merchantNo
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

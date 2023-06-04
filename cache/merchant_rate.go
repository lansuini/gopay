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

var MerchantRate = newMerchantRate()

func newMerchantRate() *merchantRate {
	return &merchantRate{}
}

type merchantRate struct {
	RedisClient *redis.Client
}

func (s *merchantRate) RefreshCache() {
	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData := repositories.MerchantRepository.Find(sqls.DB(), cnd)
	//merchantData = repositories.MerchantRateRepository.Find(sqls.DB(), cnd)

	//_, weds := redisServer.Set(ctx, "test", "4535", 60*time.Second).Result()
	//if weds != nil {
	//	fmt.Println(weds)
	//	return
	//}

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		//fmt.Println("%+v", val)
		cnd = sqls.NewCnd().Where("merchantId = ?", val.MerchantID).Desc("merchantId")
		merchantRate := repositories.MerchantRateRepository.Find(sqls.DB(), cnd)
		jsons, errs := json.Marshal(merchantRate)
		//fmt.Println(jsons)
		if errs != nil {
			fmt.Println(errs.Error())
		}
		_, cerr := redisServer.Set(ctx, "merchantRate:n:"+val.MerchantNo, jsons, 0).Result()

		if cerr != nil {
			fmt.Println(cerr)
			return
		}
		_, cerr = redisServer.Set(ctx, "merchantRate:i:"+strconv.FormatInt(val.MerchantID, 10), jsons, 0).Result()
		if cerr != nil {
			fmt.Println(cerr)
			return
		}
	}
	return
}

func (s *merchantRate) RefreshOne(merchantNo string) {
	//fmt.Println(merchantNo)
	cnd := sqls.NewCnd().Where("merchantNo = ?", merchantNo)
	merchantData := repositories.MerchantRepository.FindOne(sqls.DB(), cnd)
	if merchantData.MerchantID == 0 {
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-FindOne-null")
		return
	}
	cnd = sqls.NewCnd().Where("merchantNo = ?", merchantNo)
	rateData := repositories.MerchantRateRepository.Find(sqls.DB(), cnd)
	jsons, err := json.Marshal(rateData)
	if err != nil {
		fmt.Println(err.Error())
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-json.Marshal-", err.Error())
		return
	}

	_, err = redisServer.Set(ctx, "merchantRate:n:"+merchantNo, jsons, 0).Result()

	if err != nil {
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-redisServer.Set-", err.Error())
		fmt.Println(err.Error())
		return
	}
	_, err = redisServer.Set(ctx, "merchantRate:i:"+strconv.FormatInt(merchantData.MerchantID, 10), jsons, 0).Result()
	if err != nil {
		logrus.Error(merchantData.MerchantID, "-merchantRate-RefreshOne-redisServer.Set-", err.Error())
		fmt.Println(err.Error())
		return
	}

}

func (s *merchantRate) GetCacheByMerchantNo(merchantNo string) (merchantRates []model.MerchantRate, res bool) {
	key := "merchantRate:n:" + merchantNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		logrus.Info("缓存获取失败")
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

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

var MerchantAmount = newMerchantAmount()

func newMerchantAmount() *merchantAmount {
	return &merchantAmount{}
}

type merchantAmount struct {
	RedisClient *redis.Client
}

func (s *merchantAmount) RefreshCache() {
	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData := repositories.MerchantRepository.Find(sqls.DB(), cnd)
	//merchantData = repositories.MerchantAmountRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		//fmt.Println("%+v", val)
		cnd = sqls.NewCnd().Where("merchantId = ?", val.MerchantID).Desc("merchantId")
		merchantAmount := repositories.MerchantAmountRepository.FindOne(sqls.DB(), cnd)
		jsons, errs := json.Marshal(merchantAmount)
		//fmt.Println(jsons)
		if errs != nil {
			fmt.Println(errs.Error())
		}
		_, cerr := redisServer.Set(ctx, "merchantAmount:n:"+val.MerchantNo, jsons, 0).Result()

		if cerr != nil {
			fmt.Println(cerr)
			return
		}
		_, cerr = redisServer.Set(ctx, "merchantAmount:i:"+strconv.FormatInt(val.MerchantID, 10), jsons, 0).Result()
		if cerr != nil {
			fmt.Println(cerr)
			return
		}
	}
	return
}

func (s *merchantAmount) RefreshOne(merchantNo string) error {
	merchantAmount := model.MerchantAmount{}
	err := sqls.DB().Where("merchantNo = ?", merchantNo).First(&merchantAmount).Error
	if err != nil {
		logrus.Error(merchantNo + "查询merchantAmount失败：" + err.Error())
		return err
	}

	jsons, err := json.Marshal(merchantAmount)
	//fmt.Println(jsons)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, cerr := redisServer.Set(ctx, "merchantAmount:n:"+merchantNo, jsons, 0).Result()

	if cerr != nil {
		fmt.Println(cerr)
		return err
	}
	_, cerr = redisServer.Set(ctx, "merchantAmount:i:"+strconv.FormatInt(merchantAmount.MerchantID, 10), jsons, 0).Result()
	if cerr != nil {
		fmt.Println(cerr)
		return err
	}
	return nil
}

func (s *merchantAmount) GetCacheByMerchantNo(merchantNo string) (merchantAmount model.MerchantAmount, res bool) {
	key := "merchantAmount:n:" + merchantNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		logrus.Info("缓存获取失败")
		return
	}
	//parseMerchantAmounts := []model.MerchantAmount
	err := json.Unmarshal([]byte(jsons), &merchantAmount)
	if err != nil {
		iris.New().Logger().Info(err.Error())
		return
	}
	res = true
	return
}

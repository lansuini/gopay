package services

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"luckypay/repositories"
	"strconv"
	"time"

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

func (s *merchantRate) Get(id int64) *model.MerchantRate {
	return repositories.MerchantRateRepository.Get(sqls.DB(), id)
}

func (s *merchantRate) Take(where ...interface{}) *model.MerchantRate {
	return repositories.MerchantRateRepository.Take(sqls.DB(), where...)
}

func (s *merchantRate) Find(cnd *sqls.Cnd) []model.MerchantRate {
	return repositories.MerchantRateRepository.Find(sqls.DB(), cnd)
}

func (s *merchantRate) FindOne(cnd *sqls.Cnd) *model.MerchantRate {
	return repositories.MerchantRateRepository.FindOne(sqls.DB(), cnd)
}

func (s *merchantRate) GetServiceCharge(merchantRateData []model.MerchantRate, orderData model.ReqPayOrder, productType string) (serviceCharge float64, res bool) {

	res = false
	if len(merchantRateData) == 0 {
		return
	}
	orderAmount := orderData.OrderAmount

	if productType == "Settlement" {
		orderData.PayType = "D0Settlement"
	}
	serviceCharge = 0.00
	fialReason := ""
	for _, v := range merchantRateData {

		if v.MerchantNo == "" || orderData.MerchantNo == "" {
			fialReason = "empaty MerchantNo"
			fmt.Println(fialReason)
			continue
		}

		if v.MerchantNo != orderData.MerchantNo {
			fialReason = "different MerchantNo"
			fmt.Println(fialReason)
			continue
		}

		if v.Status != "Normal" {
			fialReason = "Status unNoraml"
			fmt.Println(fialReason)
			continue
		}

		if v.BeginTime != "" {
			beginTime, _ := time.Parse("2006-01-02 15:04:05", v.BeginTime)
			if beginTime.Unix() > time.Now().Unix() {
				fialReason = "BeginTime not valid"
				fmt.Println(fialReason)
				continue
			}
		}

		if v.EndTime != "" {
			endTime, _ := time.Parse("2006-01-02 15:04:05", v.EndTime)
			if endTime.Unix() < time.Now().Unix() {
				fialReason = "EndTime not valid"
				fmt.Println(fialReason)
				continue
			}

		}

		if v.ProductType != productType {
			fialReason = "ProductType not valid"
			continue
		}

		if v.PayType != orderData.PayType {
			fialReason = "PayType not valid"
			fmt.Println(fialReason)
			continue
		}

		if productType == "Pay" && v.CardType != orderData.CardType {
			fialReason = "CardType not valid"
			fmt.Println(fialReason)
			continue
		}

		if productType == "Pay" && v.BankCode != orderData.BankCode {
			fialReason = "BankCode not valid"
			fmt.Println(fialReason)
			continue
		}

		if v.Rate > 0 || v.Fixed > 0 {
			if v.RateType == "Rate" {
				serviceCharge = v.Rate * orderAmount
			} else if v.RateType == "FixedValue" {
				serviceCharge = v.Fixed
			} else if v.RateType == "Mixed" {
				serviceCharge = v.Rate*orderAmount + v.Fixed
			} else {
				continue
			}
		}

		if v.MinServiceCharge > 0 && serviceCharge < v.MinServiceCharge {
			serviceCharge = v.MinServiceCharge
		}

		if v.MaxServiceCharge > 0 && serviceCharge > v.MaxServiceCharge {
			serviceCharge = v.MaxServiceCharge
		}

		return serviceCharge, true
	}
	return
}

func (s *merchantRate) GetServiceChargeSettle(merchantRateData []model.MerchantRate, orderData model.ReqSettlement, productType string) (serviceCharge float64, res bool) {

	res = false
	if len(merchantRateData) == 0 {
		return
	}
	orderAmount := orderData.OrderAmount

	serviceCharge = 0.00
	fialReason := ""
	for _, v := range merchantRateData {

		if v.MerchantNo == "" || orderData.MerchantNo == "" {
			fialReason = "empaty MerchantNo"
			fmt.Println(fialReason)
			continue
		}

		if v.MerchantNo != orderData.MerchantNo {
			fialReason = "different MerchantNo"
			fmt.Println(fialReason)
			continue
		}

		if v.Status != "Normal" {
			fialReason = "Status unNoraml"
			fmt.Println(fialReason)
			continue
		}

		if v.BeginTime != "" {
			beginTime, _ := time.Parse("2006-01-02 15:04:05", v.BeginTime)
			if beginTime.Unix() > time.Now().Unix() {
				fialReason = "BeginTime not valid"
				fmt.Println(fialReason)
				continue
			}
		}

		if v.EndTime != "" {
			endTime, _ := time.Parse("2006-01-02 15:04:05", v.EndTime)
			if endTime.Unix() < time.Now().Unix() {
				fialReason = "EndTime not valid"
				fmt.Println(fialReason)
				continue
			}

		}

		if v.ProductType != productType {
			fialReason = "ProductType not valid"
			continue
		}

		if v.PayType != "D0Settlement" {
			fialReason = "PayType not valid"
			fmt.Println(fialReason)
			continue
		}

		if productType == "Pay" && v.BankCode != orderData.BankCode {
			fialReason = "BankCode not valid"
			fmt.Println(fialReason)
			continue
		}

		if v.Rate > 0 || v.Fixed > 0 {
			if v.RateType == "Rate" {
				serviceCharge = v.Rate * orderAmount
			} else if v.RateType == "FixedValue" {
				serviceCharge = v.Fixed
			} else if v.RateType == "Mixed" {
				serviceCharge = v.Rate*orderAmount + v.Fixed
			} else {
				continue
			}
		}

		if v.MinServiceCharge > 0 && serviceCharge < v.MinServiceCharge {
			serviceCharge = v.MinServiceCharge
		}

		if v.MaxServiceCharge > 0 && serviceCharge > v.MaxServiceCharge {
			serviceCharge = v.MaxServiceCharge
		}

		return serviceCharge, true
	}
	return
}

func (s *merchantRate) RefreshCache() {
	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData := repositories.MerchantRepository.Find(sqls.DB(), cnd)
	//merchantData = repositories.MerchantRateRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		//fmt.Println("%+v", val)
		cnd = sqls.NewCnd().Where("merchantId = ?", val.MerchantID).Desc("merchantId")
		merchantRate := repositories.MerchantRateRepository.Find(sqls.DB(), cnd)
		jsons, err := json.Marshal(merchantRate)
		//fmt.Println(jsons)
		if err != nil {
			logrus.Error(err)
			continue
		}
		_, err = redisServer.Set(ctx, "merchantRate:n:"+val.MerchantNo, jsons, 0).Result()

		if err != nil {
			logrus.Error(err)
			continue
		}
		_, err = redisServer.Set(ctx, "merchantRate:i:"+strconv.FormatInt(val.MerchantID, 10), jsons, 0).Result()
		if err != nil {
			logrus.Error(err)
			continue
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
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-json.Marshal-", err)
		return
	}
	_, err = redisServer.Set(ctx, "merchantRate:n:"+merchantNo, jsons, 0).Result()

	if err != nil {
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-rdb.Set-", err)
		return
	}
	_, err = redisServer.Set(ctx, "merchantRate:i:"+strconv.FormatInt(merchantData.MerchantID, 10), jsons, 0).Result()
	if err != nil {
		logrus.Error(merchantData.MerchantID, "-merchantRate-RefreshOne-rdb.Set-", err.Error())
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
		logrus.Info(err)
		return
	}
	res = true
	return
}

package services

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"luckypay/repositories"
	"strconv"
	"time"

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

func (s *channelMerchantRate) Get(id int64) *model.ChannelMerchantRate {
	return repositories.ChannelMerchantRateRepository.Get(sqls.DB(), id)
}

func (s *channelMerchantRate) Take(where ...interface{}) *model.ChannelMerchantRate {
	return repositories.ChannelMerchantRateRepository.Take(sqls.DB(), where...)
}

func (s *channelMerchantRate) Find(cnd *sqls.Cnd) []model.ChannelMerchantRate {
	return repositories.ChannelMerchantRateRepository.Find(sqls.DB(), cnd)
}

func (s *channelMerchantRate) FindOne(cnd *sqls.Cnd) *model.ChannelMerchantRate {
	return repositories.ChannelMerchantRateRepository.FindOne(sqls.DB(), cnd)
}

func (s *channelMerchantRate) GetServiceCharge(channelRateData []model.ChannelMerchantRate, orderData model.ReqPayOrder, productType string) (serviceCharge float64, res bool) {

	res = false
	if len(channelRateData) == 0 {
		return
	}

	if productType == "Settlement" {
		orderData.PayType = "D0Settlement"
	}
	serviceCharge = 0.00
	fialReason := ""
	for _, v := range channelRateData {

		if v.Status != "Normal" {
			fialReason = "empaty MerchantNo"
			continue
		}

		if v.BeginTime != "" {
			beginTime, _ := time.Parse("2006-01-02 15:04:05", v.BeginTime)
			if beginTime.Unix() > time.Now().Unix() {
				fialReason = "BeginTime not valid"
				continue
			}
		}

		if v.EndTime != "" {
			endTime, _ := time.Parse("2006-01-02 15:04:05", v.EndTime)
			if endTime.Unix() < time.Now().Unix() {
				fialReason = "EndTime not valid"
				continue
			}

		}

		if v.ProductType != productType {
			fialReason = "ProductType not valid"
			continue
		}

		if v.PayType != orderData.PayType {
			fialReason = "PayType not valid"
			continue
		}

		if productType == "Pay" && v.CardType != orderData.CardType {
			fialReason = "CardType not valid"
			continue
		}

		if productType == "Pay" && v.BankCode != orderData.BankCode {
			fialReason = "BankCode not valid"
			continue
		}

		if v.MinAmount > 0 && v.MaxAmount > 0 {
			if orderData.OrderAmount < v.MinAmount || orderData.OrderAmount > v.MaxAmount {
				fialReason = "not netween MinAmount  MaxAmount"
				continue
			}

		}

		serviceCharge = 0
		if v.Rate > 0 || v.Fixed > 0 {
			if v.RateType == "Rate" {
				serviceCharge = v.Rate * orderData.OrderAmount
			} else if v.RateType == "FixedValue" {
				serviceCharge = v.Fixed
			} else if v.RateType == "Mixed" {
				serviceCharge = v.Rate*orderData.OrderAmount + v.Fixed
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
	logrus.Error("ChannelMerchantRate.getServiceCharge", fialReason)
	return
}

// GetServiceChargeSettle
func (s *channelMerchantRate) GetServiceChargeSettle(channelRateData []model.ChannelMerchantRate, orderData model.ReqSettlement, productType string) (serviceCharge float64, res bool) {

	res = false
	if len(channelRateData) == 0 {
		return
	}

	serviceCharge = 0.00
	fialReason := ""
	for _, v := range channelRateData {

		if v.Status != "Normal" {
			fialReason = "empaty MerchantNo"
			//fmt.Println(fialReason)
			continue
		}

		if v.BeginTime != "" {
			beginTime, _ := time.Parse("2006-01-02 15:04:05", v.BeginTime)
			if beginTime.Unix() > time.Now().Unix() {
				fialReason = "BeginTime not valid"
				//fmt.Println(fialReason)
				continue
			}
		}

		if v.EndTime != "" {
			endTime, _ := time.Parse("2006-01-02 15:04:05", v.EndTime)
			if endTime.Unix() < time.Now().Unix() {
				fialReason = "EndTime not valid"
				//fmt.Println(fialReason)
				continue
			}

		}

		if v.ProductType != productType {
			fialReason = "ProductType not valid"
			continue
		}

		if productType == "Pay" && v.BankCode != orderData.BankCode {
			fialReason = "BankCode not valid"
			//fmt.Println(fialReason)
			continue
		}

		if v.MinAmount > 0 && v.MaxAmount > 0 {
			if orderData.OrderAmount < v.MinAmount || orderData.OrderAmount > v.MaxAmount {
				//logrus.Error("ChannelMerchantRate.getServiceCharge", v)
				fialReason = "OrderAmount limit"
				continue
			}

		}

		serviceCharge = 0
		if v.Rate > 0 || v.Fixed > 0 {
			if v.RateType == "Rate" {
				serviceCharge = v.Rate * orderData.OrderAmount
			} else if v.RateType == "FixedValue" {
				serviceCharge = v.Fixed
			} else if v.RateType == "Mixed" {
				serviceCharge = v.Rate*orderData.OrderAmount + v.Fixed
			} else {
				fialReason = "serviceCharge error"
				continue
			}
		}

		if v.MinServiceCharge > 0 && serviceCharge < v.MinServiceCharge {
			serviceCharge = v.MinServiceCharge
		}

		if v.MaxServiceCharge > 0 && serviceCharge > v.MaxServiceCharge {
			serviceCharge = v.MaxServiceCharge
		}
		//log.Println(serviceCharge)
		return serviceCharge, true
	}
	logrus.Error(orderData, fialReason)
	return
}

func (s *channelMerchantRate) RefreshCache() {
	cnd := sqls.NewCnd().Desc("channelMerchantId")
	merchantData := repositories.ChannelMerchantRepository.Find(sqls.DB(), cnd)
	//merchantData = repositories.ChannelMerchantRateRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {

		cnd = sqls.NewCnd().Where("channelMerchantId = ?", val.ChannelMerchantID).Desc("channelMerchantId")
		channelMerchantRates := repositories.ChannelMerchantRateRepository.Find(sqls.DB(), cnd)
		jsons, errs := json.Marshal(channelMerchantRates)
		//fmt.Println(jsons)
		if errs != nil {
			logrus.Error(errs)
			continue
		}
		_, cerr := redisServer.Set(ctx, "channelMerchantRate:n:"+val.ChannelMerchantNo, jsons, 0).Result()

		if cerr != nil {
			logrus.Error(cerr)
			continue
		}
		_, cerr = redisServer.Set(ctx, "channelMerchantRate:i:"+strconv.FormatInt(val.ChannelMerchantID, 10), jsons, 0).Result()
		if cerr != nil {
			logrus.Error(cerr)
			continue
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
		logrus.Error(merchantNo, "-merchantRate-RefreshOne-rdb.Set-", err.Error())
		fmt.Println(err.Error())
		return
	}
	_, err = redisServer.Set(ctx, "channelMerchantRate:i:"+strconv.FormatInt(merchantData.ChannelMerchantID, 10), jsons, 0).Result()
	if err != nil {
		logrus.Error(merchantData.ChannelMerchantID, "-merchantRate-RefreshOne-rdb.Set-", err.Error())
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

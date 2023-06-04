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

var MerchantChannel = newMerchantChannel()

func newMerchantChannel() *merchantChannel {
	return &merchantChannel{}
}

type merchantChannel struct {
	RedisClient *redis.Client
}

func (s *merchantChannel) Get(id int64) *model.MerchantChannel {
	return repositories.MerchantChannelRepository.Get(sqls.DB(), id)
}

func (s *merchantChannel) Take(where ...interface{}) *model.MerchantChannel {
	return repositories.MerchantChannelRepository.Take(sqls.DB(), where...)
}

func (s *merchantChannel) Find(cnd *sqls.Cnd) []model.MerchantChannel {
	return repositories.MerchantChannelRepository.Find(sqls.DB(), cnd)
}

func (s *merchantChannel) FindOne(cnd *sqls.Cnd) *model.MerchantChannel {
	return repositories.MerchantChannelRepository.FindOne(sqls.DB(), cnd)
}

func (s *merchantChannel) RefreshCache() {

	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData := repositories.MerchantRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		cnd = sqls.NewCnd().Where("merchantId = ?", val.MerchantID).Desc("merchantId")
		merchantChannelData := repositories.MerchantChannelRepository.Find(sqls.DB(), cnd)
		jsons, errs := json.Marshal(merchantChannelData)
		//fmt.Println(jsons)
		if errs != nil {
			//fmt.Println()
			logrus.Error(errs)
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
		logrus.Error(errs.Error())
		return
	}
	_, errs = redisServer.Set(ctx, "merchantChannel:n:"+merchantNo, jsons, 0).Result()

	if errs != nil {
		logrus.Error(errs.Error())
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

func (s *merchantChannel) FetchConfig(merchantNo string, merchantChannelData []model.MerchantChannel, payType string, payMoney float64, bankCode string, cardType string) (payChannel model.MerchantChannel, res bool) {
	//$st = intval(date("Hi"));
	now := time.Now()
	// 时分秒
	hour, minute, _ := now.Clock()
	sts := strconv.Itoa(hour) + strconv.Itoa(minute)
	st, _ := strconv.ParseInt(sts, 10, 64)
	fmt.Print(sts)
	fmt.Print(st)
	res = false
	payChannel = model.MerchantChannel{}
	for _, val := range merchantChannelData {
		if val.PayChannelStatus != "Normal" {
			logrus.Error("PayChannelStatus != Normal")
			continue
		}
		if val.Status != "Normal" {
			logrus.Error("Status != Normal")
			continue
		}
		if val.PayType != payType {
			logrus.Error("PayType != " + payType)
			continue
		}
		if bankCode != "" && val.BankCode != bankCode {
			logrus.Error("BankCode != " + bankCode)
			continue
		}
		if val.CardType != cardType {
			logrus.Error("CardType != " + cardType)
			continue
		}
		if val.OpenOneAmountLimit == 1 && val.OneMinAmount > 0 && val.OneMinAmount > payMoney {
			amount := strconv.FormatFloat(payMoney, 'f', 2, 32)
			OneMinAmount := strconv.FormatFloat(val.OneMinAmount, 'f', 2, 32)
			logrus.Error("payMoney" + amount + " 低于最小金额 " + OneMinAmount)
			continue
		}
		if val.OpenOneAmountLimit == 1 && val.OneMaxAmount > 0 && val.OneMaxAmount < payMoney {
			amount := strconv.FormatFloat(payMoney, 'f', 2, 32)
			OneMaxAmount := strconv.FormatFloat(val.OneMaxAmount, 'f', 2, 32)
			logrus.Error("payMoney" + amount + " 高于最小金额 " + OneMaxAmount)
			continue
		}
		if val.OpenTimeLimit > 0 && val.BeginTime > 0 && val.BeginTime > st {
			logrus.Error(" 123 ")
			continue
		}

		if val.OpenTimeLimit > 0 && val.EndTime > 0 && val.BeginTime < st {
			logrus.Error(" 456 ")
			continue
		}
		//if ($v['openDayAmountLimit'] && $this->getCacheByDayAmountLimit($merchantNo, $v['channelMerchantNo'], $payType, $bankCode, $cardType) + $payMoney * 100 > $v['dayAmountLimit'] * 100) {
		//continue;
		//}
		//
		//if ($v['openDayNumLimit'] && $this->getCacheByDayNumLimit($merchantNo, $v['channelMerchantNo'], $payType, $bankCode, $cardType) + 1 > $v['dayNumLimit']) {
		//continue;
		//}

		channelMerchantData, ret := ChannelMerchant.GetCacheByChannelMerchantNo(val.ChannelMerchantNo)
		if !ret {
			logrus.Error(val.ChannelMerchantNo)
			continue
		}
		if channelMerchantData.OpenPay == false {
			logrus.Error(" 10 ")
			continue
		}

		if channelMerchantData.Status == "Close" {
			logrus.Error(" 12 ")
			continue
		}
		payChannel = val
		res = true
		break
	}
	return payChannel, res
}

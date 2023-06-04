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

var MerchantChannelSettlement = newMerchantChannelSettlement()

func newMerchantChannelSettlement() *merchantChannelSettlement {
	return &merchantChannelSettlement{}
}

type merchantChannelSettlement struct {
	RedisClient *redis.Client
}

func (s *merchantChannelSettlement) Get(id int64) *model.MerchantChannelSettlement {
	return repositories.MerchantChannelSettlementRepository.Get(sqls.DB(), id)
}

func (s *merchantChannelSettlement) Take(where ...interface{}) *model.MerchantChannelSettlement {
	return repositories.MerchantChannelSettlementRepository.Take(sqls.DB(), where...)
}

func (s *merchantChannelSettlement) Find(cnd *sqls.Cnd) []model.MerchantChannelSettlement {
	return repositories.MerchantChannelSettlementRepository.Find(sqls.DB(), cnd)
}

func (s *merchantChannelSettlement) FindOne(cnd *sqls.Cnd) *model.MerchantChannelSettlement {
	return repositories.MerchantChannelSettlementRepository.FindOne(sqls.DB(), cnd)
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

func (s *merchantChannelSettlement) FetchConfig(merchantNo string, merchantChannelSettlementData []model.MerchantChannelSettlement, payType string, payMoney float64, bankCode string) (payChannel model.MerchantChannelSettlement, res bool) {
	//$st = intval(date("Hi"));
	now := time.Now()
	// 时分秒
	hour, minute, _ := now.Clock()
	sts := strconv.Itoa(hour) + strconv.Itoa(minute)
	st, _ := strconv.ParseInt(sts, 10, 64)
	res = false
	payChannel = model.MerchantChannelSettlement{}
	for _, val := range merchantChannelSettlementData {
		fmt.Println(val.ChannelMerchantNo)
		if val.Status != "Normal" {
			logrus.Error(val.ChannelMerchantNo + " Status != Normal")
			continue
		}

		if val.OpenOneAmountLimit == 1 && val.OneMinAmount > 0 && val.OneMinAmount > payMoney {
			amount := strconv.FormatFloat(payMoney, 'f', 2, 32)
			OneMinAmount := strconv.FormatFloat(val.OneMinAmount, 'f', 2, 32)
			logrus.Error(val.ChannelMerchantNo + " payMoney" + amount + " 低于最小金额 " + OneMinAmount)
			continue
		}
		if val.OpenOneAmountLimit == 1 && val.OneMaxAmount > 0 && val.OneMaxAmount < payMoney {
			amount := strconv.FormatFloat(payMoney, 'f', 2, 32)
			OneMaxAmount := strconv.FormatFloat(val.OneMaxAmount, 'f', 2, 32)
			logrus.Error(val.ChannelMerchantNo + " payMoney" + amount + " 高于最大金额 " + OneMaxAmount)
			continue
		}
		if val.OpenTimeLimit > 0 && val.BeginTime > 0 && val.BeginTime > st {
			logrus.Error(val.ChannelMerchantNo + " OpenTimeLimit ")
			continue
		}

		if val.OpenTimeLimit > 0 && val.EndTime > 0 && val.BeginTime < st {
			logrus.Error(val.ChannelMerchantNo + " OpenTimeLimit ")
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
			logrus.Error(val.ChannelMerchantNo + " is not exists")
			continue
		}
		if channelMerchantData.OpenPay == false {
			logrus.Error(val.ChannelMerchantNo + " channelMerchant OpenPay is false ")
			continue
		}

		if channelMerchantData.Status == "Close" {
			logrus.Error(val.ChannelMerchantNo + " channelMerchant Status is Close ")
			continue
		}
		payChannel = val
		res = true
		break
	}
	return payChannel, res
}

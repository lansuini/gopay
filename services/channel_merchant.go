package services

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/lgbya/go-dump"
	"github.com/sirupsen/logrus"
	"luckypay/model"
	"luckypay/repositories"
	"strconv"
	"time"

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

func (s *channelMerchant) Get(id int64) *model.ChannelMerchant {
	return repositories.ChannelMerchantRepository.Get(sqls.DB(), id)
}

func (s *channelMerchant) Take(where ...interface{}) *model.ChannelMerchant {
	return repositories.ChannelMerchantRepository.Take(sqls.DB(), where...)
}

func (s *channelMerchant) Find(cnd *sqls.Cnd) []model.ChannelMerchant {
	return repositories.ChannelMerchantRepository.Find(sqls.DB(), cnd)
}

func (s *channelMerchant) FindOne(cnd *sqls.Cnd) *model.ChannelMerchant {
	return repositories.ChannelMerchantRepository.FindOne(sqls.DB(), cnd)
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

// 根据日期获取渠道的支付，代付，充值
func (s *channelMerchant) DayStats(date string) {
	merchants := []model.ChannelMerchant{}
	err := sqls.DB().Table("merchant").Find(&merchants).Error
	if err != nil {
		logrus.Error("DayStats 商户查询失败: ", err)
	}
	startDatetime := date + " 00:00:00"
	endDatetime := date + " 23:59:59"
	for _, merchant := range merchants {
		whereMap := map[string]interface{}{
			"channelMerchantNo": merchant.ChannelMerchantNo,
			"orderStatus":       "Success",
		}
		payData := DayStatSearch{}
		err = sqls.DB().Table("platform_pay_order").Where(whereMap).Where("updated_at >= ?", startDatetime).Where("updated_at <= ?", endDatetime).Select("count(orderId) as counts, sum(realOrderAmount) as orderAmounts, sum(serviceCharge) as serviceCharges, sum(channelServiceCharge) as channelServiceCharges, sum(agentFee) as agentFees").Scan(&payData).Error
		if err != nil {
			logrus.Error("DayStats platform_pay_order: ", err)
		}
		dump.Printf(payData)
		settleData := DayStatSearch{}
		err = sqls.DB().Table("platform_settlement_order").Where(whereMap).Where("updated_at >= ?", startDatetime).Where("updated_at <= ?", endDatetime).Select("count(orderId) as counts, sum(realOrderAmount) as orderAmounts, sum(serviceCharge) as serviceCharges, sum(channelServiceCharge) as channelServiceCharges, sum(agentFee) as agentFees").Scan(&settleData).Error
		if err != nil {
			logrus.Error("DayStats platform_settlement_order: ", err)
		}
		dump.Printf(settleData)

		rechargeData := DayStatSearch{}
		err = sqls.DB().Table("platform_recharge_order").Where(whereMap).Where("updated_at >= ?", startDatetime).Where("updated_at <= ?", endDatetime).Select("count(id) as counts, sum(realOrderAmount) as orderAmounts, sum(serviceCharge) as serviceCharges, sum(channelServiceCharge) as channelServiceCharges, sum(agentFee) as agentFees").Scan(&rechargeData).Error
		if err != nil {
			logrus.Error("DayStats platform_settlement_order: ", err)
		}
		dump.Printf(rechargeData)

		if payData.OrderAmounts == 0 && settleData.OrderAmounts == 0 && rechargeData.OrderAmounts == 0 {
			fmt.Println(merchant.ChannelMerchantNo, "无数据")
			continue
		} else {
			ChannelDailyStats := map[string]interface{}{
				"merchantId":            merchant.ChannelMerchantID,
				"merchantNo":            merchant.ChannelMerchantNo,
				"accountDate":           date,
				"payCount":              payData.Counts,
				"payAmount":             payData.OrderAmounts,
				"payServiceFees":        payData.ServiceCharges,
				"payChannelServiceFees": payData.ChannelServiceCharges,
				"agentPayFees":          payData.AgentFees,

				"settlementCount":              settleData.Counts,
				"settlementAmount":             settleData.OrderAmounts,
				"settlementServiceFees":        settleData.ServiceCharges,
				"settlementChannelServiceFees": settleData.ChannelServiceCharges,
				"agentsettlementFees":          settleData.AgentFees,

				"chargeCount":              rechargeData.Counts,
				"chargeAmount":             rechargeData.OrderAmounts,
				"chargeServiceFees":        rechargeData.ServiceCharges,
				"chargeChannelServiceFees": rechargeData.ChannelServiceCharges,
				"agentchargeFees":          rechargeData.AgentFees,
			}
			err = sqls.DB().Table("channel_daily_stats").Create(ChannelDailyStats).Error
			if err != nil {
				logrus.Error("merchant_daily_stats: ", err)
				break
			}
		}
	}
}

func (s *channelMerchant) RunDayStats() {
	logrus.Info("RunChannelDayStats Start ")
	cacheKey := "reportService:channel:" + time.Now().Format("2006-01-02")

	res, err := redisServer.Get(ctx, cacheKey).Result()
	if err != nil {
		logrus.Error("RunChannelDayStats Error: ", err)
	}
	if res != "" {
		logrus.Error("RunChannelDayStats Error: 重复运行错误")
		return
	}
	redisServer.SetEX(ctx, cacheKey, 1, 24*time.Hour)
	statDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var counts int64
	err = sqls.DB().Table("channel_daily_stats").Where("accountDate = ?", statDate).Count(&counts).Error
	if err != nil {
		logrus.Error("RunChannelDayStats count Error: ", err)
		return
	}
	if counts > 0 {
		logrus.Error("RunChannelDayStats Error: 重复运行错误2 ")
		return
	}
	go s.DayStats(statDate)
}

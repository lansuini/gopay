package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/lgbya/go-dump"
	"github.com/sirupsen/logrus"
	"luckypay/config"
	"luckypay/model"
	yamlConfig "luckypay/pkg/config"
	"luckypay/repositories"
	"luckypay/utils"
	"strconv"
	"time"

	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var MerchantService = newMerchantService()
var ctx context.Context = context.Background()

var redisServer = config.NewRedis()

type DayStatSearch struct {
	Counts                int64   `gorm:"column:counts;" json:"counts"`
	OrderAmounts          float64 `gorm:"column:orderAmounts;" json:"orderAmounts"`
	ServiceCharges        float64 `gorm:"column:serviceCharges;" json:"serviceCharges"`
	ChannelServiceCharges float64 `gorm:"column:channelServiceCharges;" json:"channelServiceCharges"`
	AgentFees             float64 `gorm:"column:agentFees;" json:"agentFees"`
}

func newMerchantService() *merchantService {
	return &merchantService{}
}

type merchantService struct {
	RedisClient *redis.Client
}

func (s *merchantService) Get(id int64) *model.Merchant {
	return repositories.MerchantRepository.Get(sqls.DB(), id)
}

func (s *merchantService) Take(where ...interface{}) *model.Merchant {
	return repositories.MerchantRepository.Take(sqls.DB(), where...)
}

func (s *merchantService) Find(cnd *sqls.Cnd) []model.Merchant {
	return repositories.MerchantRepository.Find(sqls.DB(), cnd)
}

func (s *merchantService) FindOne(cnd *sqls.Cnd) *model.Merchant {
	return repositories.MerchantRepository.FindOne(sqls.DB(), cnd)
}

func (s *merchantService) RefreshCache() (merchantData []model.Merchant) {
	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData = repositories.MerchantRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		jsons, err := json.Marshal(val)
		if err != nil {
			logrus.Error(err)
			continue
		}
		_, err = redisServer.Set(ctx, "merchant:n:"+val.MerchantNo, jsons, 0).Result()

		if err != nil {
			logrus.Error(err)
			continue
		}
		_, err = redisServer.Set(ctx, "merchant:i:"+strconv.FormatInt(val.MerchantID, 10), jsons, 0).Result()
		if err != nil {
			logrus.Error(err)
			continue
		}
	}
	return
}

func (s *merchantService) RefreshOne(merchantNo string) {
	merchantData := model.Merchant{}
	err := sqls.DB().Table("merchant").Where("merchantNo = ?", merchantNo).First(&merchantData).Error
	if err != nil {
		logrus.Error(merchantNo, "merchantService RefreshOne Fail: ", err.Error())
	}

	jsons, err := json.Marshal(merchantData)
	if err != nil {
		logrus.Error(merchantNo, "merchantService RefreshOne Fail: ", err)
	}
	_, err = redisServer.Set(ctx, "merchant:n:"+merchantNo, jsons, 0).Result()
	if err != nil {
		logrus.Error(merchantNo, "merchantService RefreshOne Fail: ", err)
		return
	}
	_, err = redisServer.Set(ctx, "merchant:i:"+strconv.FormatInt(merchantData.MerchantID, 10), jsons, 0).Result()
	if err != nil {
		logrus.Error(merchantNo, "merchantService RefreshOne Fail: ", err)
		return
	}

	return
}

func (s *merchantService) GetCacheByMerchantNo(merchantNo string) (merchant model.Merchant, res bool) {

	key := "merchant:n:" + merchantNo
	jsons, err := redisServer.Get(ctx, key).Result()
	if err != nil {
		logrus.Error(err)
		return merchant, false
	}

	err = json.Unmarshal([]byte(jsons), &merchant)
	if err != nil {
		logrus.Error(err)
		return merchant, false
	}
	//fmt.Println("%+v", merchant)
	return merchant, true
}

func (s *merchantService) GetCacheByDaySettleAmountLimit(merchantNo string) (amount float64, err error) {
	err = nil
	key := "merchant:settle:tc:" + time.Now().Format("20060102") + ":" + merchantNo
	cmd := redisServer.Get(ctx, key)
	if err = cmd.Err(); err != nil {
		if err == redis.Nil {
			logrus.Error("GetCacheByDaySettleAmountLimit key does not exists")
			return 0, nil
		}
		return 0, err
	}
	amount, err = cmd.Float64()
	if err != nil {
		return 0, err
	}
	return amount, nil
}

func (s *merchantService) IsAllowIPAccess(ip string) (res bool) {
	key := "ipaccess:" + ip
	if yamlConfig.Instance.SystemConfig.GATE_IP_PROTECT == "true" {
		return true
	}
	cmd := redisServer.Incr(ctx, key)
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			logrus.Error(err.Error())
		}
		return false
	}
	fmt.Println(cmd.Val())
	if num, _ := cmd.Uint64(); num > 2 {
		logrus.Error(ip + "-GATE_IP_PROTECT")
		return false
	}
	redisServer.Expire(ctx, key, 5*time.Second)

	return true
}

func (s *merchantService) GetPlatformOrderNo(capLetter string) string {

	now := time.Now()
	ymdkey := utils.GetFormatTime(now)

	p1 := utils.GetYMDHISTime(now)

	p2, _ := redisServer.Get(ctx, ymdkey).Result()
	if p2 == "" {
		p2 = utils.RandomString(3)
		_, err := redisServer.Set(ctx, ymdkey, p2, 48*time.Hour).Result()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	p4 := utils.RandomString(5)
	orderNo := capLetter + p1 + p2 + p4
	redisServer.SAdd(ctx, ymdkey+":balanceorder", orderNo).Result()
	redisServer.Expire(ctx, ymdkey+":balanceorder", 48*time.Hour)
	return orderNo

}

// 根据日期获取商户的支付，代付，充值
func (s *merchantService) DayStats(date string) {
	merchants := []model.Merchant{}
	err := sqls.DB().Table("merchant").Find(&merchants).Error
	if err != nil {
		logrus.Error("DayStats 商户查询失败: ", err)
	}
	startDatetime := date + " 00:00:00"
	endDatetime := date + " 23:59:59"
	logrus.Info("商户报表统计-" + startDatetime + "----" + endDatetime)
	for _, merchant := range merchants {
		fmt.Println("商户报表统计-" + merchant.MerchantNo)
		whereMap := map[string]interface{}{
			"merchantNo":  merchant.MerchantNo,
			"orderStatus": "Success",
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
			fmt.Println(merchant.MerchantNo, "无数据")
			continue
		} else {
			MerchantDailyStats := map[string]interface{}{
				"merchantId":            merchant.MerchantID,
				"merchantNo":            merchant.MerchantNo,
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
			err = sqls.DB().Table("merchant_daily_stats").Create(MerchantDailyStats).Error
			if err != nil {
				logrus.Error("merchant_daily_stats: ", err)
				break
			}
			logrus.Info("商户报表统计完成-" + startDatetime + "----" + endDatetime)
		}
	}
}

func (s *merchantService) RunDayStats() {
	logrus.Info("RunMerchantDayStats Start ")
	cacheKey := "reportService:merchant:" + time.Now().Format("2006-01-02")

	res, err := redisServer.Get(ctx, cacheKey).Result()
	if err != nil {
		logrus.Error("RunMerchantDayStats Error: ", err)
	}
	if res != "" {
		logrus.Error("RunMerchantDayStats Error: 重复运行错误")
		return
	}
	redisServer.SetEX(ctx, cacheKey, 1, 24*time.Hour)
	statDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var counts int64
	err = sqls.DB().Table("merchant_daily_stats").Where("accountDate = ?", statDate).Count(&counts).Error
	if err != nil {
		logrus.Error("RunMerchantDayStats count Error: ", err)
		return
	}
	if counts > 0 {
		logrus.Error("RunMerchantDayStats Error: 重复运行错误2 ")
		return
	}
	go s.DayStats(statDate)
}

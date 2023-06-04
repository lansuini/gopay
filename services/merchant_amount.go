package services

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
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

func (s *merchantAmount) Get(id int64) *model.MerchantAmount {
	return repositories.MerchantAmountRepository.Get(sqls.DB(), id)
}

func (s *merchantAmount) Take(where ...interface{}) *model.MerchantAmount {
	return repositories.MerchantAmountRepository.Take(sqls.DB(), where...)
}

func (s *merchantAmount) Find(cnd *sqls.Cnd) []model.MerchantAmount {
	return repositories.MerchantAmountRepository.Find(sqls.DB(), cnd)
}

func (s *merchantAmount) FindOne(cnd *sqls.Cnd) *model.MerchantAmount {
	return repositories.MerchantAmountRepository.FindOne(sqls.DB(), cnd)
}

func (s *merchantAmount) RefreshCache() {
	cnd := sqls.NewCnd().Desc("merchantId")
	merchantData := repositories.MerchantRepository.Find(sqls.DB(), cnd)
	//merchantData = repositories.MerchantAmountRepository.Find(sqls.DB(), cnd)

	for _, val := range merchantData {
		//iris.New().Logger().Info(json.Marshal(val))
		cnd = sqls.NewCnd().Where("merchantId = ?", val.MerchantID).Desc("merchantId")
		merchantAmountData := repositories.MerchantAmountRepository.FindOne(sqls.DB(), cnd)
		jsons, err := json.Marshal(merchantAmountData)
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		_, err = redisServer.Set(ctx, "merchantAmount:n:"+val.MerchantNo, jsons, 0).Result()

		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		_, err = redisServer.Set(ctx, "merchantAmount:i:"+strconv.FormatInt(val.MerchantID, 10), jsons, 0).Result()
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
	}
	return
}

func (s *merchantAmount) RefreshOne(merchantNo string) error {
	merchantAmountData := model.MerchantAmount{}
	err := sqls.DB().Where("merchantNo = ?", merchantNo).First(&merchantAmountData).Error
	if err != nil {
		logrus.Error(merchantNo + "查询merchantAmount失败：" + err.Error())
		return err
	}

	jsons, err := json.Marshal(merchantAmountData)
	if err != nil {
		logrus.Error(err)
		return err
	}
	_, err = redisServer.Set(ctx, "merchantAmount:n:"+merchantNo, jsons, 0).Result()

	if err != nil {
		logrus.Error(err)
		return err
	}
	_, err = redisServer.Set(ctx, "merchantAmount:i:"+strconv.FormatInt(merchantAmountData.MerchantID, 10), jsons, 0).Result()
	if err != nil {
		logrus.Error(err)
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
		logrus.Info(err.Error())
		return
	}
	res = true
	return
}

func (s *merchantAmount) GetAmount(merchantNo string) (amountData map[string]float64) {

	//merchantData := MerchantService.GetCacheByMerchantNo(merchantNo)
	//
	cacheAmountData, _ := s.GetCacheByMerchantNo(merchantNo)
	//now := time.Now()
	//accountDate := utils.GetFormatTime(now)
	//var todaySettlement float64
	//sqls.DB().Table("amount_pay").Where("merchantNo = ?", merchantNo).Where("accountDate = ?", accountDate).Select("sum(amount) - sum(serviceCharge)").Scan(&todaySettlement)
	//fmt.Println(todaySettlement)
	amountData = make(map[string]float64)
	amountData["settlementAmount"] = cacheAmountData.SettlementAmount
	amountData["accountBalance"] = cacheAmountData.SettlementAmount
	amountData["availableBalance"] = cacheAmountData.SettlementAmount
	amountData["freezeAmount"] = cacheAmountData.FreezeAmount
	amountData["settledAmount"] = cacheAmountData.SettlementAmount + cacheAmountData.FreezeAmount
	return

}

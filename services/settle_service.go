package services

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"luckypay/cache"
	"luckypay/channels"
	"reflect"
	"strings"

	"luckypay/model/constants"
	"luckypay/utils"
	"strconv"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"luckypay/model"
	"luckypay/repositories"
)

var SettleService = newSettleService()

func newSettleService() *settleService {
	return &settleService{}
}

type settleService struct {
}

func (s *settleService) Get(id int64) *model.PlatformSettlementOrder {
	return repositories.SettleRepository.Get(sqls.DB(), id)
}

func (s *settleService) Take(where ...interface{}) *model.PlatformSettlementOrder {
	return repositories.SettleRepository.Take(sqls.DB(), where...)
}

func (s *settleService) Find(cnd *sqls.Cnd) []model.PlatformSettlementOrder {
	return repositories.SettleRepository.Find(sqls.DB(), cnd)
}

func (s *settleService) FindOne(cnd *sqls.Cnd) *model.PlatformSettlementOrder {
	return repositories.SettleRepository.FindOne(sqls.DB(), cnd)
}

func (s *settleService) FindPageByParams(params *params.QueryParams) (list []model.PlatformSettlementOrder, paging *sqls.Paging) {
	return repositories.SettleRepository.FindPageByParams(sqls.DB(), params)
}

func (s *settleService) FindPageByCnd(cnd *sqls.Cnd) (list []model.PlatformSettlementOrder, paging *sqls.Paging) {
	return repositories.SettleRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *settleService) Count(cnd *sqls.Cnd) int64 {
	return repositories.SettleRepository.Count(sqls.DB(), cnd)
}

func (s *settleService) Create(t *model.PlatformSettlementOrder) error {
	return repositories.SettleRepository.Create(sqls.DB(), t)
}

func (s *settleService) Update(t *model.PlatformSettlementOrder) error {
	return repositories.SettleRepository.Update(sqls.DB(), t)
}

func (s *settleService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.SettleRepository.Updates(sqls.DB(), id, columns)
}

func (s *settleService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.SettleRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *settleService) Delete(id int64) error {
	return repositories.SettleRepository.UpdateColumn(sqls.DB(), id, "status", constants.StatusDeleted)
}

func (s *settleService) GetPlatformOrderNo(capLetter string) string {

	now := time.Now()
	ymdkey := utils.GetFormatTime(now)

	p1 := utils.GetYMDHISTime(now)
	p2, _ := redisServer.Get(ctx, ymdkey).Result()
	if p2 == "" {
		p2 = utils.RandomString(3)
		_, err := redisServer.Set(ctx, ymdkey, p2, 48*time.Hour).Result()
		if err != nil {
			panic(err)
		}
	}

	p3 := strconv.FormatInt(utils.GetTimeTick64(), 10)
	p4 := utils.RandomString(3)
	orderNo := capLetter + p1 + p2 + p3 + p4
	redisServer.SAdd(ctx, ymdkey+":settlementorder", orderNo).Result()
	redisServer.Expire(ctx, ymdkey+":settlementorder", 48*time.Hour)
	fmt.Print("orderNo:", orderNo, "\n")
	return orderNo

}

func (s *settleService) CreateSettlementOrder(tx *gorm.DB, request model.ReqSettlement, platformOrderNo string, channel model.MerchantChannelSettlement, serviceCharge float64, channelServiceCharge float64) (settleStruct model.PlatformSettlementOrder, err error) {
	merchantData, res := MerchantService.GetCacheByMerchantNo(request.MerchantNo)
	if !res {
		return settleStruct, errors.New("获取商户信息失败")
	}
	//代理手续费
	//agentId = AgentMerchantRelation::where('merchantId',$merchantData['merchantId'])->value('agentId');
	//if(agentId) {
	//	$agentLog = new AgentIncomeLog();
	//	//代付订单类型只有一种
	//	$agentFee = $agentLog->getFee($agentId,$merchantData['merchantId'],platformOrderNo,request.OrderAmount,'pay',$request->getParam('payType'),$request->getParam('bankCode'));
	//
	//	$agentName=Agent::where('id',$agentId)->value('loginName');
	//}else {
	var agentFee float64 = 0.00
	agentName := ""
	//}
	//settleStruct.MerchantID = merchantData.MerchantID
	//settleStruct.AgentFee = agentFee
	//settleStruct.AgentName = agentName
	settleStruct = model.PlatformSettlementOrder{
		MerchantID:      merchantData.MerchantID,
		MerchantNo:      request.MerchantNo,
		MerchantOrderNo: request.MerchantOrderNo,
		PlatformOrderNo: platformOrderNo,
		//ChannelOrderNo:  channelOrderNo,
		OrderAmount:     request.OrderAmount,
		RealOrderAmount: request.OrderAmount,

		MerchantParam:        request.MerchantParam,
		MerchantReqTime:      time.Now().Format("2006-01-02 15:04:05"),
		BankAccountNo:        request.BankAccountNo,
		BankAccountName:      request.BankAccountName,
		Province:             request.Province,
		City:                 request.City,
		BackNoticeURL:        request.BackNoticeUrl,
		TradeSummary:         request.TradeSummary,
		BankCode:             request.BankCode,
		AccountDate:          time.Now().Format("2006-01-02 15:04:05"),
		Channel:              channel.Channel,
		ChannelSetID:         channel.SetID,
		ChannelMerchantNo:    channel.ChannelMerchantNo,
		ChannelMerchantID:    channel.ChannelMerchantID,
		OrderStatus:          "Transfered",
		ProcessType:          "WaitPayment",
		ServiceCharge:        serviceCharge,
		ChannelServiceCharge: channelServiceCharge,
		AgentFee:             agentFee,
		AgentName:            agentName,
	}

	err = tx.Create(&settleStruct).Error

	if err != nil {
		return settleStruct, err
	}
	//s.SetCacheByPlatformOrderNo(platformOrderNo, settleStruct)
	return settleStruct, nil
}

func (s *settleService) SetCacheByPlatformOrderNo(platformOrderNo string, payStruct model.PlatformSettlementOrder) {
	key1 := "settlementorder:" + platformOrderNo
	key2 := "settlementorder:m:" + payStruct.MerchantNo + ":" + payStruct.MerchantOrderNo

	jsons, err := json.Marshal(payStruct)
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo1 : ", err.Error())
	}
	_, err = redisServer.Set(ctx, key1, jsons, 7*24*time.Hour).Result()
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo2 : ", err.Error())
	}
	_, err = redisServer.Set(ctx, key2, jsons, 7*24*time.Hour).Result()
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo3 : ", err.Error())
	}

}

func (s *settleService) GetCacheByPlatformOrderNo(platformOrderNo string) (payOrder *model.PlatformSettlementOrder, res bool) {
	key := "settlementorder:" + platformOrderNo
	res = false
	jsons, _ := redisServer.Get(ctx, key).Result()
	if jsons == "" {
		return
	}
	err := json.Unmarshal([]byte(jsons), &payOrder)
	if err != nil {
		iris.New().Logger().Info(err.Error())
		return
	}
	res = true
	return

}

func (s *settleService) RefreshOne(platformOrderNo string) {
	var order model.PlatformSettlementOrder
	err := sqls.DB().Where("platformOrderNo = ?", platformOrderNo).First(&order).Error
	if err != nil {
		logrus.Error(platformOrderNo + "刷新代付订单失败-" + err.Error())
		return
	}
	key1 := "settlementorder:" + platformOrderNo
	key2 := "settlementorder:m:" + order.MerchantNo + ":" + order.MerchantOrderNo

	jsons, err := json.Marshal(order)
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo1 : ", err.Error())
	}
	_, err = redisServer.Set(ctx, key1, jsons, 7*24*time.Hour).Result()
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo2 : ", err.Error())
	}
	_, err = redisServer.Set(ctx, key2, jsons, 7*24*time.Hour).Result()
	if err != nil {
		logrus.Error("SetCacheByPlatformOrderNo3 : ", err.Error())
	}
}

func (s *settleService) DelCacheByPlatformOrderNo(platformOrderNo string, MerchantNo string, MerchantOrderNo string) {
	key1 := "settlementorder:" + platformOrderNo
	key2 := "settlementorder:m:" + MerchantNo + ":" + MerchantOrderNo
	redisServer.Del(ctx, key1)
	redisServer.Del(ctx, key2)

}

func (s *settleService) Success(order model.PlatformSettlementOrder, orderAmount float64, tx *gorm.DB) error {
	var err error
	//merchantRate := model.MerchantRate{}
	//amountSettlement := model.AmountSettlement{}
	orderData := model.PlatformSettlementOrder{}
	//accountAmount := model.MerchantAmount{}
	//merchantData,res := MerchantService.GetCacheByMerchantNo(order.MerchantNo)
	//	$merchantData = (new Merchant)->getCacheByMerchantId($orderData['merchantId']);
	//	$channelOrderNo = empty($channelOrderNo) ? $orderData['channelOrderNo'] : $channelOrderNo;
	//	$channelNoticeTime = empty($channelNoticeTime) ? date('YmdHis') : $channelNoticeTime;
	//	$accountDate = Tools::getAccountDate($merchantData['settlementTime'], $channelNoticeTime);
	//]
	//	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", order.MerchantNo).First(&accountAmount).Error

	if err != nil {
		logrus.Error(order.MerchantNo, "-查询商户余额失败 : ", err.Error())
		return err
	}
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("orderId = ?", order.OrderID).First(&orderData).Error
	if err != nil {
		logrus.Error(order.OrderID, "-查询订单失败 : ", err.Error())
		return err
	}
	if orderData.OrderStatus != "Transfered" {
		logrus.Error(order.OrderID, "数据已处理，或不存在 : ")
		return errors.New("数据已处理")
	}
	if len(orderData.AccountDate) >= 10 {
		orderData.AccountDate = order.AccountDate[0:10]
	} else {
		orderData.AccountDate = time.Now().Format("2006-01-02")
	}
	amountSettlementData := model.AmountSettlement{
		MerchantID:        order.MerchantID,
		MerchantNo:        order.MerchantNo,
		ChannelMerchantID: order.ChannelMerchantID,
		ChannelMerchantNo: order.ChannelMerchantNo,
		AccountDate:       orderData.AccountDate,
	}
	// 有冲突时什么都不做
	err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&amountSettlementData).Error
	if err != nil {
		logrus.Error(order.MerchantNo, "-创建amountSettlement失败 : ", err.Error())
		return err
	}
	whereMap := map[string]interface{}{"merchantId": order.MerchantID, "channelMerchantId": order.ChannelMerchantID, "accountDate": order.AccountDate}
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereMap).First(&amountSettlementData).Error
	if err != nil {
		logrus.Error(order.MerchantNo, "-查询amountSettlement失败 : ", err.Error())
		return err
	}

	//修改订单
	if order.ChannelNoticeTime == "" {
		order.ChannelNoticeTime = time.Now().Format("2006-01-02 15:04:05")
	}
	ts, err := time.Parse(time.RFC3339, order.ChannelNoticeTime)
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "-time.Parse失败 : ", err.Error())
		ts = time.Now()
	}
	orderData.ChannelNoticeTime = ts.Format("2006-01-02 15:04:05")
	orderData.ChannelOrderNo = order.ChannelOrderNo
	orderData.ProcessType = "Success"
	orderData.FailReason = order.FailReason
	orderData.OrderStatus = "Success"

	//if orderData.PushChannelTime == "" {
	//	orderData.PushChannelTime = time.Now().Format("2006-01-02 15:04:05")
	//}
	if order.OrderAmount != orderData.OrderAmount {
		orderData.RealOrderAmount = order.OrderAmount
	} else {
		orderData.RealOrderAmount = orderData.OrderAmount
	}
	orderData.AuditPerson = order.AuditPerson
	orderData.AuditIP = order.AuditIP
	orderData.AuditTime = order.AuditTime
	orderData.IsLock = 0
	orderData.LockUser = ""

	err = tx.Omit("merchantReqTime", "auditTime").Save(&orderData).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "修改代付订单失败 : ", err.Error())
		return err
	}
	amountSettlementData.AccountDate = order.AccountDate
	amountSettlementData.TransferTimes = amountSettlementData.TransferTimes + 1
	amountSettlementData.Amount = amountSettlementData.Amount + order.OrderAmount
	amountSettlementData.ServiceCharge = amountSettlementData.ServiceCharge + orderData.ServiceCharge
	amountSettlementData.ChannelServiceCharge = amountSettlementData.ChannelServiceCharge + order.ChannelServiceCharge
	err = tx.Save(&amountSettlementData).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "保存amountSettlementData失败 : ", err.Error())
		return err
	}
	//刷新缓存
	s.SetCacheByPlatformOrderNo(order.PlatformOrderNo, orderData)
	//代付成功减存储的余额
	mcsData := model.MerchantChannelSettlement{}
	err = tx.Where("channelMerchantId = ?", order.ChannelMerchantID).First(&mcsData).Error
	if err == nil {
		if mcsData.AccountBalance >= order.OrderAmount {
			err = tx.Where("channelMerchantId = ?", order.ChannelMerchantID).Update("accountBalance", gorm.Expr("accountBalance - ?", order.OrderAmount)).Error
			if err != nil {
				logrus.Error(order.ChannelMerchantID, "修改MerchantChannelSettlement余额失败 : ", err.Error())
			}
		}
	} else {
		logrus.Error(order.ChannelMerchantID, "查询 MerchantChannelSettlement 失败: ", err.Error())
	}

	//	//TODO:代理手续费
	//	$agentId = AgentMerchantRelation::where('merchantId',$orderData['merchantId'])->value('agentId');
	//	if($agentId ||isset($orderData['agentFee']) && $orderData['agentFee'] > 0) {
	//	$agentLog = new AgentIncomeLog();
	//	$agentLog->updateIncomeLog($orderData['merchantId'],$orderData['platformOrderNo'],$orderAmount,'settlement');
	//	}
	return nil
}

func (s *settleService) Fail(order model.PlatformSettlementOrder, orderAmount float64, tx *gorm.DB) error {
	var err error
	//merchantRate := model.MerchantRate{}
	//amountSettlement := model.AmountSettlement{}
	orderData := model.PlatformSettlementOrder{}
	accountAmount := model.MerchantAmount{}
	//merchantData,res := MerchantService.GetCacheByMerchantNo(order.MerchantNo)
	//	$merchantData = (new Merchant)->getCacheByMerchantId($orderData['merchantId']);
	//	$channelOrderNo = empty($channelOrderNo) ? $orderData['channelOrderNo'] : $channelOrderNo;
	//	$channelNoticeTime = empty($channelNoticeTime) ? date('YmdHis') : $channelNoticeTime;
	//	$accountDate = Tools::getAccountDate($merchantData['settlementTime'], $channelNoticeTime);
	//]
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", order.MerchantNo).First(&accountAmount).Error

	if err != nil {
		logrus.Error(order.MerchantNo, "-查询商户余额失败 : ", err.Error())
		return err
	}
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("orderId = ?", order.OrderID).First(&orderData).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "-查询订单失败 : ", err.Error())
		return err
	}
	if orderData.OrderStatus != "Transfered" {
		logrus.Error(order.PlatformOrderNo, "数据已处理，或不存在 : ")
		return errors.New("数据已处理")
	}
	accountAmount.SettlementAmount = accountAmount.SettlementAmount + orderData.OrderAmount + orderData.ServiceCharge
	err = tx.Model(&accountAmount).Update("settlementAmount", accountAmount.SettlementAmount).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "-更新MerchantAmount失败 : ", err.Error())
		return err
	}
	//修改订单
	if len(order.ChannelNoticeTime) < 10 {
		orderData.ChannelNoticeTime = time.Now().Format("2006-01-02 15:04:05")
	} else {
		ts, err := time.Parse(time.RFC3339, order.ChannelNoticeTime)
		if err != nil {
			logrus.Error(order.PlatformOrderNo, "-time.Parse失败 : ", err.Error())
			ts = time.Now()
		}
		orderData.ChannelNoticeTime = ts.Format("2006-01-02 15:04:05")
	}
	orderData.ChannelOrderNo = order.ChannelOrderNo
	orderData.ProcessType = "Success"
	orderData.FailReason = order.FailReason
	orderData.OrderStatus = "Fail"
	if len(orderData.AccountDate) >= 10 {
		orderData.AccountDate = order.AccountDate[0:10]
	}

	//if orderData.PushChannelTime == nil {
	//	nowtime := time.Now().Format("2006-01-02 15:04:05")
	//	orderData.PushChannelTime = &nowtime
	//}

	if order.OrderAmount != orderData.OrderAmount {
		orderData.RealOrderAmount = order.OrderAmount
	} else {
		orderData.RealOrderAmount = orderData.OrderAmount
	}
	orderData.AuditPerson = order.AuditPerson
	orderData.AuditIP = order.AuditIP
	//orderData.AuditTime = time.Now().Format("2006-01-02 15:04:05")
	orderData.IsLock = 0
	orderData.LockUser = ""
	orderData.ChannelServiceCharge = order.ChannelServiceCharge
	//dump.Printf(orderData)
	err = tx.Omit("merchantReqTime", "auditTime").Save(&orderData).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "修改代付订单失败 : ", err.Error())
		return err
	}
	//添加流水
	financeData := []map[string]interface{}{
		{"merchantId": orderData.MerchantID, "merchantNo": orderData.MerchantNo, "platformOrderNo": orderData.PlatformOrderNo, "amount": orderData.OrderAmount, "balance": accountAmount.SettlementAmount - orderData.ServiceCharge, "financeType": "PayIn", "accountDate": orderData.AccountDate, "accountType": "SettledAccount", "sourceId": orderData.OrderID, "sourceDesc": "结算返还服务", "summary": "代付订单退还", "merchantOrderNo": orderData.MerchantOrderNo, "operateSource": "ports"},
		{"merchantId": orderData.MerchantID, "merchantNo": orderData.MerchantNo, "platformOrderNo": orderData.PlatformOrderNo, "amount": orderData.ServiceCharge, "balance": accountAmount.SettlementAmount, "financeType": "PayIn", "accountDate": orderData.AccountDate, "accountType": "ServiceChargeAccount", "sourceId": orderData.OrderID, "sourceDesc": "结算返还手续费", "summary": "代付订单退还", "merchantOrderNo": orderData.MerchantOrderNo, "operateSource": "ports"},
	}

	err = tx.Table("finance").Create(financeData).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "代付金额返还添加流水失败 : ", err.Error())
		return err
	}
	//刷新缓存
	//$ppo->setCacheByPlatformOrderNo($orderData['platformOrderNo'], $orderDataLock->toArray());
	//$merchantAmountData->refreshCache(['merchantId' => $merchantAmountData->merchantId]);
	s.SetCacheByPlatformOrderNo(order.PlatformOrderNo, orderData)
	//代付失败更改渠道代付累计数量和金额
	//if (Tools::isToday($accountDate)) {
	//AmountPay::where('merchantId', $orderData['merchantId'])
	//->where('accountDate', $accountDate)
	//->update(['balance' => $merchantAmountData->settlementAmount]);
	//}
	//
	//if ($this->isRollbackMerchantChannelSettleDayLimit($orderDataLock)) {
	//(new MerchantChannelSettlement)->incrCacheByDayAmountLimit($orderData['merchantNo'], $orderData['channelMerchantNo'], -intval($orderData['orderAmount'] * 100));
	//(new MerchantChannelSettlement)->incrCacheByDayNumLimit($orderData['merchantNo'], $orderData['channelMerchantNo'], -1);
	//(new ChannelSettlementConfig)->incrCacheByDayAmountLimit($orderData['channelMerchantNo'], -intval($orderData['orderAmount'] * 100));
	//(new ChannelSettlementConfig)->incrCacheByDayNumLimit($orderData['channelMerchantNo'], -1);
	//(new ChannelSettlementConfig)->incrCacheByCardDayAmountLimit(Tools::decrypt($orderDataLock->bankAccountNo), $orderDataLock->channelMerchantNo, -intval($orderDataLock->orderAmount * 100));
	//(new ChannelSettlementConfig)->incrCacheByCardDayNumLimit(Tools::decrypt($orderDataLock->bankAccountNo), $orderDataLock->channelMerchantNo, -1);
	//}
	//
	//if ($this->isRollbackMerchantSettleDayLimit($orderDataLock)) {
	//(new Merchant)->incrCacheByDaySettleAmountLimit($orderData['merchantNo'], -intval($orderData['orderAmount'] * 100));
	//}

	return nil
}

func (s *settleService) Settlement(channel channels.Channel, settleParams model.SettleParams, channelMerchant model.ChannelMerchant) {

	logrus.Info("向上游请求代付-", settleParams.PlatformOrderNo)
	settleResult, err := channels.SettlementOrder(channel, settleParams, channelMerchant)
	logrus.Info(settleParams.PlatformOrderNo, settleResult)
	if err != nil {
		logrus.Error(settleParams.PlatformOrderNo, "-请求订单异常 ：", err.Error())
		//err := sqls.DB().Table("settlement_fetch_task").Where("platformOrderNo = ?", settleParams.PlatformOrderNo).Update("status", "Execute").Error
		//if err != nil {
		//	logrus.Error(settleParams.PlatformOrderNo, "-修改 settlement_fetch_task 失败 ：", err.Error())
		//}
		SettlementFetch.Push(0, settleParams.PlatformOrderNo)
		return
	}
	if settleResult.Status != "Success" {
		logrus.Error(settleParams.PlatformOrderNo, "-请求代付订单失败")
		res := model.RspQuerySettle{
			PlatformOrderNo: settleParams.PlatformOrderNo,
			FailReason:      settleResult.FailReason,
			ChannelOrderNo:  settleResult.ChannelOrderNo,
		}
		s.CallFail(res)
		SettlementNotify.Push(0, settleParams.PlatformOrderNo)
		return
	}
	return
}

func (s *settleService) QueryOrder(channel channels.Channel, order model.PlatformSettlementOrder, channelMerchant model.ChannelMerchant, task model.SettlementFetchTask) {
	payResult, err := channels.QuerySettlementOrder(channel, order, channelMerchant)
	logrus.Info(order.PlatformOrderNo, "查询代付请求结果：", payResult)
	if err != nil {
		logrus.Info(order.PlatformOrderNo, "-查询代付失败：", err.Error())
	}
	if payResult.Status != "Success" {
		//SettlementFetchTask处理
		if !reflect.DeepEqual(task, model.SettlementFetchTask{}) {
			updates := map[string]interface{}{
				"status":     "Execute",
				"retryCount": task.RetryCount + 1,
				"failReason": payResult.FailReason,
			}
			go SettlementFetch.UpdateTask(task, updates)
			//go SettlementFetch.Push(task.ID, task.PlatformOrderNo)
		}
		return
	}
	//SettlementFetchTask处理
	if !reflect.DeepEqual(task, model.SettlementFetchTask{}) {
		updates := map[string]interface{}{
			"status":     "Success",
			"retryCount": task.RetryCount + 1,
			"failReason": payResult.FailReason,
		}
		go SettlementFetch.UpdateTask(task, updates)
	}
	if payResult.OrderStatus == "Success" {

		err = sqls.DB().Transaction(func(tx *gorm.DB) error {
			err = s.Success(order, order.OrderAmount, sqls.DB())
			return err
		})
		if err != nil {
			logrus.Info(order.PlatformOrderNo, "-Success代付修改失败：", err.Error())
		}
	} else if payResult.OrderStatus == "Fail" {

		err = sqls.DB().Transaction(func(tx *gorm.DB) error {
			order.FailReason = payResult.FailReason
			err = s.Fail(order, order.OrderAmount, sqls.DB())
			return err
		})
		if err != nil {
			logrus.Info(order.PlatformOrderNo, "-Fail代付修改失败：", err.Error())
		}
	}
	//刷新订单缓存
	go s.RefreshOne(order.PlatformOrderNo)
	//回调商户
	if payResult.OrderStatus != "Execute" {
		go s.CallbackMerchant(order.PlatformOrderNo, model.PlatformSettlementOrder{}, model.SettlementNotifyTask{})
		SettlementNotify.Push(0, order.PlatformOrderNo)
	} else {
		updates := map[string]interface{}{
			"status":     "Execute",
			"retryCount": task.RetryCount + 1,
			"failReason": payResult.FailReason,
		}
		go SettlementFetch.UpdateTask(task, updates)
		//go SettlementFetch.Push(task.ID, task.PlatformOrderNo)
	}
	return
}

func (s *settleService) CallbackMerchant(orderNo string, order model.PlatformSettlementOrder, task model.SettlementNotifyTask) {
	if reflect.DeepEqual(order, model.PlatformSettlementOrder{}) {
		order, res := cache.SettleCache.GetCacheByPlatformOrderNo(orderNo)
		if !res {
			logrus.Error(order.PlatformOrderNo, "-代付回调失败：查询订单失败")
			return
		}
	}
	//if order.BackNoticeURL == ""{
	//	logrus.Error(order.PlatformOrderNo,"-代付回调失败：回调地址为空")
	//}
	res := utils.IsValidUrl(order.BackNoticeURL)
	if !res {
		logrus.Error(order.PlatformOrderNo, "-代付回调失败：回调地址格式错误-", order.BackNoticeURL)
		return
		//order.BackNoticeURL = "http://cb.luckypay.mm:8082/paycallback"
	}
	//fmt.Println(order.BackNoticeURL)
	callbackData := make(map[string]interface{})
	callbackData["code"] = "SUCCESS"
	callbackData["msg"] = "成功"
	biz := make(map[string]interface{})
	biz["platformOrderNo"] = order.PlatformOrderNo
	biz["merchantOrderNo"] = order.MerchantOrderNo
	biz["orderStatus"] = order.OrderStatus
	biz["orderAmount"] = order.OrderAmount
	biz["merchantParam"] = order.MerchantParam
	if order.OrderStatus == "Fail" {
		biz["orderMsg"] = "代付失败"
	} else if order.OrderStatus == "Success" {
		biz["orderMsg"] = "代付成功"
	} else {
		biz["orderMsg"] = "处理中"
	}
	merchantData, res := MerchantService.GetCacheByMerchantNo(order.MerchantNo)
	if !res {
		logrus.Error(order.PlatformOrderNo, "-代付回调失败：获取商户信息失败")
		return
	}
	encrytkey := merchantData.SignKey
	sign := utils.GetSignStr(encrytkey, biz)
	callbackData["biz"] = callbackData
	callbackData["sign"] = sign
	data, err := json.Marshal(callbackData)
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "-代付回调失败：格式化数据失败", callbackData)
		return
	}
	response, err, statusCode := utils.HttpPostJson(order.BackNoticeURL, data)
	logrus.Info(order.PlatformOrderNo, "-代付回调信息：", statusCode, string(response))
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "-代付回调失败：请求失败-", err.Error())
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": task.RetryCount + 1,
			"failReason": "回调异常" + err.Error(),
		}
		go SettlementNotify.UpdateTask(task, updates)
		updateOrderMaps := map[string]interface{}{
			"callbackLimit": order.CallbackLimit + 1,
		}
		go s.UpdateOrderMap(order, updateOrderMaps)
		return
	}
	if strings.ToLower(string(response)) == "success" {
		if !reflect.DeepEqual(order, model.PlatformSettlementOrder{}) {
			updates := map[string]interface{}{
				"status":     "Success",
				"retryCount": task.RetryCount + 1,
				"failReason": "回调成功",
			}
			go SettlementNotify.UpdateTask(task, updates)
		}

		updateOrderMaps := map[string]interface{}{
			"callbackLimit":   order.CallbackLimit + 1,
			"callbackSuccess": true,
		}
		go s.UpdateOrderMap(order, updateOrderMaps)
	} else {
		if !reflect.DeepEqual(order, model.PlatformSettlementOrder{}) {
			updates := map[string]interface{}{
				"retryCount": task.RetryCount + 1,
				"failReason": string(response),
			}
			go SettlementNotify.UpdateTask(task, updates)
		}
		go SettlementNotify.Push(0, orderNo)
		updateOrderMaps := map[string]interface{}{
			"callbackLimit": order.CallbackLimit + 1,
		}
		go s.UpdateOrderMap(order, updateOrderMaps)
	}

	return
}

func (s *settleService) CallSuccess(callback model.RspQuerySettle) error {
	var err error
	orderData := model.PlatformSettlementOrder{}
	//merchantData,res := MerchantService.GetCacheByMerchantNo(order.MerchantNo)
	//	$merchantData = (new Merchant)->getCacheByMerchantId($orderData['merchantId']);
	//	$channelOrderNo = empty($channelOrderNo) ? $orderData['channelOrderNo'] : $channelOrderNo;
	//	$channelNoticeTime = empty($channelNoticeTime) ? date('YmdHis') : $channelNoticeTime;
	//	$accountDate = Tools::getAccountDate($merchantData['settlementTime'], $channelNoticeTime);
	//]
	//	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", order.MerchantNo).First(&accountAmount).Error
	//if err != nil {
	//	logrus.Error(order.MerchantNo, "-查询商户余额失败 : ", err.Error())
	//	return err
	//}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("platformOrderNo = ?", callback.PlatformOrderNo).First(&orderData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "-查询订单失败 : ", err.Error())
			return err
		}
		if orderData.OrderStatus != "Transfered" {
			logrus.Error(callback.PlatformOrderNo, "数据已处理，或不存在 : ")
			return errors.New("数据已处理")
		}
		if len(orderData.AccountDate) >= 10 {
			orderData.AccountDate = orderData.AccountDate[0:10]
		}
		amountSettlementData := model.AmountSettlement{}
		whereAnd := model.AmountSettlement{
			MerchantID:        orderData.MerchantID,
			MerchantNo:        orderData.MerchantNo,
			ChannelMerchantID: orderData.ChannelMerchantID,
			ChannelMerchantNo: orderData.ChannelMerchantNo,
			AccountDate:       orderData.AccountDate,
		}
		// 有冲突时什么都不做
		err = tx.Where(whereAnd).FirstOrCreate(&amountSettlementData).Error
		if err != nil {
			logrus.Error(orderData.MerchantNo, "-创建amountSettlement失败 : ", err.Error())
			return err
		}
		//dump.Printf(amountSettlementData)
		//whereMap := map[string]interface{}{"merchantId": orderData.MerchantID, "channelMerchantId": orderData.ChannelMerchantID, "accountDate": orderData.AccountDate}
		//err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereMap).First(&amountSettlementData).Error
		//if err != nil {
		//	logrus.Error(orderData.MerchantNo, "-查询amountSettlement失败 : ", err.Error())
		//	return err
		//}

		//修改订单
		orderData.OrderStatus = "Success"
		orderData.ProcessType = "Success"
		orderData.ChannelNoticeTime = time.Now().Format("2006-01-02 15:04:05")
		orderData.ChannelOrderNo = callback.ChannelOrderNo
		orderData.FailReason = callback.FailReason

		/*if callback.OrderAmount != orderData.OrderAmount {
			orderData.RealOrderAmount = callback.OrderAmount
		} else {
			orderData.RealOrderAmount = orderData.OrderAmount
		}*/
		orderData.IsLock = 0
		orderData.LockUser = ""

		err = tx.Omit("merchantReqTime", "auditTime").Save(&orderData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "修改代付订单失败 : ", err.Error())
			return err
		}
		amountSettlementData.AccountDate = orderData.AccountDate
		amountSettlementData.TransferTimes = amountSettlementData.TransferTimes + 1
		amountSettlementData.Amount = amountSettlementData.Amount + orderData.OrderAmount
		amountSettlementData.ServiceCharge = amountSettlementData.ServiceCharge + orderData.ServiceCharge
		amountSettlementData.ChannelServiceCharge = amountSettlementData.ChannelServiceCharge + orderData.ChannelServiceCharge
		err = tx.Save(&amountSettlementData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "保存amountSettlementData失败 : ", err.Error())
			return err
		}
		return nil
	})

	if err != nil {
		logrus.Error(callback.PlatformOrderNo, "代付订单回调CallSuccess失败 : ", err.Error())
		return err
	}
	//刷新缓存
	go cache.SettleCache.SetCacheByPlatformOrderNo(orderData.PlatformOrderNo, orderData)
	//代付成功减存储的余额
	go s.DecreaseChannelAmount(orderData.ChannelMerchantNo, orderData.OrderAmount)
	//回调商户地址
	go s.CallbackMerchant(callback.PlatformOrderNo, orderData, model.SettlementNotifyTask{})
	//	//TODO:代理手续费
	//	$agentId = AgentMerchantRelation::where('merchantId',$orderData['merchantId'])->value('agentId');
	//	if($agentId ||isset($orderData['agentFee']) && $orderData['agentFee'] > 0) {
	//	$agentLog = new AgentIncomeLog();
	//	$agentLog->updateIncomeLog($orderData['merchantId'],$orderData['platformOrderNo'],$orderAmount,'settlement');
	//	}
	return nil
}

func (s *settleService) CallFail(callback model.RspQuerySettle) error {
	var err error
	//merchantRate := model.MerchantRate{}
	//amountSettlement := model.AmountSettlement{}
	orderData := model.PlatformSettlementOrder{}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("platformOrderNo = ?", callback.PlatformOrderNo).First(&orderData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "-查询订单失败 : ", err.Error())
			return err
		}
		accountAmount := model.MerchantAmount{}
		//merchantData,res := MerchantService.GetCacheByMerchantNo(order.MerchantNo)
		//	$merchantData = (new Merchant)->getCacheByMerchantId($orderData['merchantId']);
		//	$channelOrderNo = empty($channelOrderNo) ? $orderData['channelOrderNo'] : $channelOrderNo;
		//	$channelNoticeTime = empty($channelNoticeTime) ? date('YmdHis') : $channelNoticeTime;
		//	$accountDate = Tools::getAccountDate($merchantData['settlementTime'], $channelNoticeTime);
		//]
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", orderData.MerchantNo).First(&accountAmount).Error

		if err != nil {
			logrus.Error(orderData.MerchantNo, "-查询商户余额失败 : ", err.Error())
			return err
		}

		if orderData.OrderStatus != "Transfered" {
			logrus.Error(callback.PlatformOrderNo, "数据已处理，或不存在 : ")
			return errors.New("数据已处理")
		}
		accountAmount.SettlementAmount = accountAmount.SettlementAmount + orderData.OrderAmount + orderData.ServiceCharge
		err = tx.Model(&accountAmount).Update("settlementAmount", accountAmount.SettlementAmount).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "-更新MerchantAmount失败 : ", err.Error())
			return err
		}
		//修改订单

		orderData.ChannelNoticeTime = time.Now().Format("2006-01-02 15:04:05")
		orderData.ChannelOrderNo = callback.ChannelOrderNo
		orderData.ProcessType = "Success"
		orderData.FailReason = callback.FailReason
		orderData.OrderStatus = "Fail"

		orderData.IsLock = 0
		orderData.LockUser = ""
		accountDate := time.Now().Format("2006-01-02")
		if len(orderData.AccountDate) >= 10 {
			accountDate = orderData.AccountDate[0:10]
		} else {
			accountDate = orderData.CreatedAt.Format("2006-01-02")
		}
		orderData.AccountDate = accountDate
		//dump.Printf(orderData)
		err = tx.Omit("merchantReqTime", "auditTime").Save(&orderData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "修改代付订单失败 : ", err.Error())
			return err
		}
		//添加流水
		financeData := []map[string]interface{}{
			{"merchantId": orderData.MerchantID, "merchantNo": orderData.MerchantNo, "platformOrderNo": orderData.PlatformOrderNo, "amount": orderData.OrderAmount, "balance": accountAmount.SettlementAmount - orderData.ServiceCharge, "financeType": "PayIn", "accountDate": accountDate, "accountType": "SettledAccount", "sourceId": orderData.OrderID, "sourceDesc": "结算返还服务", "summary": "手动处理", "merchantOrderNo": orderData.MerchantOrderNo, "operateSource": "ports"},
			{"merchantId": orderData.MerchantID, "merchantNo": orderData.MerchantNo, "platformOrderNo": orderData.PlatformOrderNo, "amount": orderData.ServiceCharge, "balance": accountAmount.SettlementAmount, "financeType": "PayIn", "accountDate": accountDate, "accountType": "ServiceChargeAccount", "sourceId": orderData.OrderID, "sourceDesc": "结算返还手续费", "summary": "手动处理", "merchantOrderNo": orderData.MerchantOrderNo, "operateSource": "ports"},
		}

		err = tx.Table("finance").Create(financeData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "代付金额返还添加流水失败 : ", err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		logrus.Error(callback.PlatformOrderNo, "代付订单回调CallFail失败 : ", err.Error())
		return err
	}
	//刷新缓存
	//$ppo->setCacheByPlatformOrderNo($orderData['platformOrderNo'], $orderDataLock->toArray());
	//$merchantAmountData->refreshCache(['merchantId' => $merchantAmountData->merchantId]);
	go s.SetCacheByPlatformOrderNo(callback.PlatformOrderNo, orderData)
	//代付失败更改渠道代付累计数量和金额
	//if (Tools::isToday($accountDate)) {
	//AmountPay::where('merchantId', $orderData['merchantId'])
	//->where('accountDate', $accountDate)
	//->update(['balance' => $merchantAmountData->settlementAmount]);
	//}
	//
	//if ($this->isRollbackMerchantChannelSettleDayLimit($orderDataLock)) {
	//(new MerchantChannelSettlement)->incrCacheByDayAmountLimit($orderData['merchantNo'], $orderData['channelMerchantNo'], -intval($orderData['orderAmount'] * 100));
	//(new MerchantChannelSettlement)->incrCacheByDayNumLimit($orderData['merchantNo'], $orderData['channelMerchantNo'], -1);
	//(new ChannelSettlementConfig)->incrCacheByDayAmountLimit($orderData['channelMerchantNo'], -intval($orderData['orderAmount'] * 100));
	//(new ChannelSettlementConfig)->incrCacheByDayNumLimit($orderData['channelMerchantNo'], -1);
	//(new ChannelSettlementConfig)->incrCacheByCardDayAmountLimit(Tools::decrypt($orderDataLock->bankAccountNo), $orderDataLock->channelMerchantNo, -intval($orderDataLock->orderAmount * 100));
	//(new ChannelSettlementConfig)->incrCacheByCardDayNumLimit(Tools::decrypt($orderDataLock->bankAccountNo), $orderDataLock->channelMerchantNo, -1);
	//}
	//
	//if ($this->isRollbackMerchantSettleDayLimit($orderDataLock)) {
	//(new Merchant)->incrCacheByDaySettleAmountLimit($orderData['merchantNo'], -intval($orderData['orderAmount'] * 100));
	//}
	//回调商户地址
	go s.CallbackMerchant(callback.PlatformOrderNo, orderData, model.SettlementNotifyTask{})
	return nil
}

func (s *settleService) DecreaseChannelAmount(channelMerchantNo string, amount float64) {
	tx := sqls.DB()
	mcsData := model.MerchantChannelSettlement{}
	err := tx.Where("channelMerchantNo = ?", channelMerchantNo).First(&mcsData).Error
	if err == nil {
		if mcsData.AccountBalance >= amount {
			err = tx.Where("channelMerchantNo = ?", channelMerchantNo).Update("accountBalance", gorm.Expr("accountBalance - ?", amount)).Error
			if err != nil {
				logrus.Error(channelMerchantNo, "修改MerchantChannelSettlement余额失败 : ", err.Error())
			}
		}
	} else {
		logrus.Error(channelMerchantNo, "查询 MerchantChannelSettlement 失败: ", err.Error())
	}
}

func (s *settleService) UpdateOrderMap(order model.PlatformSettlementOrder, updates map[string]interface{}) {
	err := sqls.DB().Model(&order).Updates(updates).Error
	if err != nil {
		logrus.Info(order.PlatformOrderNo, "-UpdateOrderMap", err)
	}
	go cache.SettleCache.SetCacheByPlatformOrderNo(order.PlatformOrderNo, order)
	return
}

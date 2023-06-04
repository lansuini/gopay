package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
	"luckypay/cache"
	"luckypay/channels"
	"luckypay/model/constants"
	"luckypay/utils"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
	"luckypay/repositories"
)

var PayService = newPayService()

func newPayService() *payService {
	return &payService{}
}

type payService struct {
}

func (s *payService) Get(id int64) *model.PlatformPayOrder {
	return repositories.PayRepository.Get(sqls.DB(), id)
}

func (s *payService) Take(where ...interface{}) *model.PlatformPayOrder {
	return repositories.PayRepository.Take(sqls.DB(), where...)
}

func (s *payService) Find(cnd *sqls.Cnd) []model.PlatformPayOrder {
	return repositories.PayRepository.Find(sqls.DB(), cnd)
}

func (s *payService) FindOne(cnd *sqls.Cnd) *model.PlatformPayOrder {
	return repositories.PayRepository.FindOne(sqls.DB(), cnd)
}

func (s *payService) FindPageByParams(params *params.QueryParams) (list []model.PlatformPayOrder, paging *sqls.Paging) {
	return repositories.PayRepository.FindPageByParams(sqls.DB(), params)
}

func (s *payService) FindPageByCnd(cnd *sqls.Cnd) (list []model.PlatformPayOrder, paging *sqls.Paging) {
	return repositories.PayRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *payService) Count(cnd *sqls.Cnd) int64 {
	return repositories.PayRepository.Count(sqls.DB(), cnd)
}

func (s *payService) Create(t *model.PlatformPayOrder) error {
	return repositories.PayRepository.Create(sqls.DB(), t)
}

func (s *payService) Update(t *model.PlatformPayOrder) error {
	return repositories.PayRepository.Update(sqls.DB(), t)
}

func (s *payService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.PayRepository.Updates(sqls.DB(), id, columns)
}

func (s *payService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.PayRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *payService) Delete(id int64) error {
	return repositories.PayRepository.UpdateColumn(sqls.DB(), id, "status", constants.StatusDeleted)
}

// 查询支付订单状态
func (s *payService) QueryOrder(channel channels.Channel, order model.PlatformPayOrder, channelMerchant model.ChannelMerchant) {
	payResult, err := channels.QueryPayOrder(channel, order, channelMerchant)
	logrus.Info(order.PlatformOrderNo, "查询支付请求结果：", payResult)
	if err != nil {
		logrus.Info(order.PlatformOrderNo, "-查询支付失败：", err.Error())
		return
	}
	if payResult.Status != "Success" {
		logrus.Info(order.PlatformOrderNo, "-查询支付失败：", payResult.FailReason)
		return
	}
	if payResult.OrderStatus == "Success" {

		err = sqls.DB().Transaction(func(tx *gorm.DB) error {
			err = s.Success(order, order.OrderAmount, sqls.DB())
			return err
		})
		if err != nil {
			logrus.Info(order.PlatformOrderNo, "-Success支付修改失败：", err.Error())
			return
		}
		//刷新订单缓存
		s.RefreshOne(order.PlatformOrderNo)
		go s.CallbackMerchant(order.PlatformOrderNo, model.PlatformPayOrder{}, model.PayNotifyTask{})
		PayNotify.Push(0, order.PlatformOrderNo)
	}

	return
}

func (s *payService) GetCacheByMerchantOrderNo(merchantNo string, merchantOrderNo string) (payOrder model.PlatformPayOrder, res bool) {

	key := "payorder:m:" + merchantNo + ":" + merchantOrderNo
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

func (s *payService) RefreshOne(platformOrderNo string) {
	var order model.PlatformPayOrder
	err := sqls.DB().Where("platformOrderNo = ?", platformOrderNo).First(&order).Error
	if err != nil {
		logrus.Error(platformOrderNo + "刷新支付订单失败-" + err.Error())
		return
	}
	key1 := "payorder:" + platformOrderNo
	key2 := "payorder:m:" + order.MerchantNo + ":" + order.MerchantOrderNo

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

func (s *payService) GetPlatformOrderNo(capLetter string) string {
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
	redisServer.SAdd(ctx, ymdkey+":payorder", orderNo).Result()
	redisServer.Expire(ctx, ymdkey+":payorder", 48*time.Hour)
	fmt.Println("orderNo:", orderNo)
	return orderNo

}

func (s *payService) CreateOrder(request model.ReqPayOrder, platformOrderNo string, channel model.MerchantChannel, channelOrderNo string, serviceCharge float64, channelServiceCharge float64) (payOrder model.PlatformPayOrder, err error) {
	merchantData, res := MerchantService.GetCacheByMerchantNo(request.MerchantNo)
	if !res {
		return payOrder, errors.New("获取商户信息失败")
	}
	//代理手续费
	//agentId = AgentMerchantRelation::where('merchantId',$merchantData['merchantId'])->value('agentId');
	//if(agentId) {
	//	$agentLog = new AgentIncomeLog();
	//	//支付订单类型只有一种
	//	$agentFee = $agentLog->getFee($agentId,$merchantData['merchantId'],platformOrderNo,request.OrderAmount,'pay',$request->getParam('payType'),$request->getParam('bankCode'));
	//
	//	$agentName=Agent::where('id',$agentId)->value('loginName');
	//}else {
	var agentFee float64 = 0.00
	agentName := ""
	//}

	payStruct := model.PlatformPayOrder{
		MerchantId:           merchantData.MerchantID,
		MerchantNo:           request.MerchantNo,
		MerchantOrderNo:      request.MerchantOrderNo,
		PlatformOrderNo:      platformOrderNo,
		ChannelOrderNo:       channelOrderNo,
		OrderAmount:          request.OrderAmount,
		RealOrderAmount:      request.OrderAmount,
		PayType:              request.PayType,
		PayModel:             request.PayModel,
		MerchantParam:        request.MerchantParam,
		MerchantReqTime:      time.Now().Format("2006-01-02 15:04:05"),
		UserIp:               request.UserIp,
		UserTerminal:         request.UserTerminal,
		BackNoticeUrl:        request.BackNoticeUrl,
		TradeSummary:         request.TradeSummary,
		BankCode:             request.BankCode,
		CardType:             request.CardType,
		CardHolderName:       request.CardHolderName,
		CardNum:              request.CardNum,
		Channel:              channel.Channel,
		ChannelSetId:         channel.SetID,
		ChannelMerchantNo:    channel.ChannelMerchantNo,
		ChannelMerchantID:    channel.ChannelMerchantID,
		OrderStatus:          "WaitPayment",
		ProcessType:          "WaitPayment",
		ServiceCharge:        serviceCharge,
		ChannelServiceCharge: channelServiceCharge,
		AgentFee:             agentFee,
		AgentName:            agentName,
		AccountDate:          time.Now().Format("2006-01-02"),
	}

	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := repositories.PayRepository.Create(tx, &payStruct); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return payOrder, err
	}
	s.SetCacheByPlatformOrderNo(platformOrderNo, payStruct)
	return payStruct, nil

	//'userTerminal' => $request->getParam('userTerminal', ''),
	//'userIp' => $request->getParam('userIp'),
	//'thirdUserId' => $request->getParam('thirdUserId', ''),
	//'cardHolderName' => $request->getParam('cardHolderName', ''),
	//'cardNum' => Tools::encrypt($request->getParam('cardNum', '')),
	//'idType' => $request->getParam('idType', ''),
	//'idNum' => Tools::encrypt($request->getParam('idNum', '')),
	//'cardHolderMobile' => Tools::encrypt($request->getParam('cardHolderMobile', '')),
	//'frontNoticeUrl' => $request->getParam('frontNoticeUrl', ''),
	//'pushChannelTime' => date('Y-m-d H:i:s'),

}

func (s *payService) Success(order model.PlatformPayOrder, orderAmount float64, tx *gorm.DB) error {
	var err error
	//merchantRate := model.MerchantRate{}
	//amountSettlement := model.AmountSettlement{}
	orderData := model.PlatformPayOrder{}
	accountAmount := model.MerchantAmount{}

	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", order.MerchantNo).First(&accountAmount).Error

	if err != nil {
		logrus.Error(order.MerchantNo, "-查询商户余额失败 : ", err.Error())
		return err
	}
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("orderId = ?", order.OrderId).First(&orderData).Error
	if err != nil {
		logrus.Error(order.OrderId, "-查询订单失败 : ", err.Error())
		return err
	}
	if orderData.OrderStatus == "Success" || orderData.OrderStatus == "Fail" {
		logrus.Error(order.OrderId, "数据已处理，或不存在 : ")
		return errors.New("数据已处理")
	}
	if len(orderData.AccountDate) >= 10 {
		orderData.AccountDate = orderData.AccountDate[0:10]
	} else {
		orderData.AccountDate = orderData.CreatedAt.Format("2006-01-02")
	}
	whereMap := map[string]interface{}{
		"merchantNo":        orderData.MerchantNo,
		"channelMerchantNo": orderData.ChannelMerchantNo,
		"payType":           orderData.PayType,
		"accountDate":       orderData.AccountDate,
	}
	amountPayData := model.AmountPay{
		MerchantID:        order.MerchantId,
		MerchantNo:        order.MerchantNo,
		ChannelMerchantID: order.ChannelMerchantID,
		ChannelMerchantNo: order.ChannelMerchantNo,
		PayType:           order.PayType,
		AccountDate:       orderData.AccountDate,
	}
	// 查询创建锁定
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereMap).FirstOrCreate(&amountPayData).Error
	if err != nil {
		logrus.Error(order.MerchantNo, "-创建amountPay失败 : ", err.Error())
		return err
	}
	//whereMap := map[string]interface{}{"merchantId": order.MerchantId, "channelMerchantId": order.ChannelMerchantID, "accountDate": orderData.AccountDate}
	//err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereMap).First(&amountPayData).Error
	//if err != nil {
	//	logrus.Error(order.MerchantNo, "-查询amountSettlement失败 : ", err.Error())
	//	return err
	//}

	//修改订单
	orderData.ChannelNoticeTime = order.ChannelNoticeTime
	orderData.ChannelOrderNo = order.ChannelOrderNo
	orderData.ProcessType = "Success"
	orderData.OrderStatus = "Success"

	//if (!empty($orderData['payType']) && $orderDataLock->payType != $orderData['payType']) {
	//	$orderDataLock->payType = $orderData['payType'];
	//}
	if orderAmount != orderData.OrderAmount && orderAmount > 0 {
		orderData.RealOrderAmount = order.OrderAmount
		//TODO:重新计算手续费
		merchantRate, res := MerchantRate.GetCacheByMerchantNo(orderData.MerchantNo)
		if !res {
			logrus.Error(orderData.MerchantNo, "E2026费率未设置")
			return errors.New(orderData.MerchantNo + "-商户未设置费率")
		}
		channelRate, res := ChannelMerchantRate.GetCacheByMerchantNo(orderData.ChannelMerchantNo)
		if !res {
			logrus.Error("merchantChannel 没有设置费率", orderData.ChannelMerchantNo)
			return errors.New(orderData.ChannelMerchantNo + "-merchantChannel未设置费率")
		}
		reqPayOrder := s.GetReqPayOrderParams(orderData)
		orderData.ServiceCharge, res = MerchantRate.GetServiceCharge(merchantRate, reqPayOrder, "Pay")
		if !res {
			logrus.Error("GetServiceCharge 没有对应支付渠道费率", orderData.ChannelMerchantNo)
			return errors.New(orderData.ChannelMerchantNo + "GetServiceCharge 没有对应支付渠道费率")
		}
		orderData.ChannelServiceCharge, res = ChannelMerchantRate.GetServiceCharge(channelRate, reqPayOrder, "Pay")
		if !res {
			logrus.Error("getChannelServiceCharge 没有设置对应上游支付渠道费率: ", orderData.ChannelMerchantNo)
			return errors.New(orderData.ChannelMerchantNo + "getChannelServiceCharge 没有设置对应上游支付渠道费率")
		}
	}

	orderData.TimeoutTime = time.Now().Format("2006-01-02 15:04:05")
	err = tx.Omit("merchantReqTime").Save(&orderData).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "修改支付订单失败 : ", err.Error())
		return err
	}

	amountPayData.AccountDate = orderData.AccountDate
	amountPayData.Amount = amountPayData.Amount + order.RealOrderAmount
	amountPayData.ServiceCharge = amountPayData.ServiceCharge + orderData.ServiceCharge
	amountPayData.ChannelServiceCharge = amountPayData.ChannelServiceCharge + orderData.ChannelServiceCharge
	err = tx.Save(&amountPayData).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "保存amountSettlementData失败 : ", err.Error())
		return err
	}

	//添加流水
	financeData := []map[string]interface{}{
		{"merchantId": orderData.MerchantId, "merchantNo": orderData.MerchantNo, "platformOrderNo": orderData.PlatformOrderNo, "amount": orderData.RealOrderAmount, "balance": accountAmount.SettlementAmount + orderData.RealOrderAmount, "financeType": "PayIn", "accountDate": orderData.AccountDate, "accountType": "SettlementAccount", "sourceId": orderData.OrderId, "sourceDesc": "支付服务", "summary": "支付服务", "merchantOrderNo": orderData.MerchantOrderNo, "operateSource": "ports"},
		{"merchantId": orderData.MerchantId, "merchantNo": orderData.MerchantNo, "platformOrderNo": orderData.PlatformOrderNo, "amount": orderData.ServiceCharge, "balance": accountAmount.SettlementAmount + orderData.RealOrderAmount - orderData.ServiceCharge, "financeType": "PayOut", "accountDate": orderData.AccountDate, "accountType": "ServiceChargeAccount", "sourceId": orderData.OrderId, "sourceDesc": "支付手续费", "summary": "支付手续费", "merchantOrderNo": orderData.MerchantOrderNo, "operateSource": "ports"},
	}

	err = tx.Table("finance").Create(financeData).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "支付订单补单添加流水失败 : ", err.Error())
		return err
	}
	//更新MerchantAmount余额
	accountAmount.SettlementAmount = accountAmount.SettlementAmount + orderData.RealOrderAmount - orderData.ServiceCharge
	err = tx.Model(&accountAmount).Update("settlementAmount", accountAmount.SettlementAmount).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "更新MerchantAmount余额失败 : ", err.Error())
		return err
	}
	//更新AmountPay余额
	err = tx.Table("amount_pay").Where("merchantId = ?", orderData.MerchantId).Where("accountDate = ?", orderData.AccountDate).Update("balance", accountAmount.SettlementAmount).Error
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "更新AmountPay余额失败 : ", err.Error())
		return err
	}

	//$ppo->setCacheByPlatformOrderNo($orderData['platformOrderNo'], $orderDataLock->toArray());
	//支付成功减存储的余额
	/*mcsData := model.MerchantChannelSettlement{}
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
	}*/
	/*AmountPay::where('merchantId', $orderData['merchantId'])
	->where('accountDate', $accountDate)
	->update(['balance' => $merchantAmountData->settlementAmount]);*/
	//	//TODO:代理手续费
	//	$agentId = AgentMerchantRelation::where('merchantId',$orderData['merchantId'])->value('agentId');
	//	if($agentId ||isset($orderData['agentFee']) && $orderData['agentFee'] > 0) {
	//	$agentLog = new AgentIncomeLog();
	//	$agentLog->updateIncomeLog($orderData['merchantId'],$orderData['platformOrderNo'],$orderAmount,'settlement');
	//	}

	//刷新缓存
	s.SetCacheByPlatformOrderNo(order.PlatformOrderNo, orderData)
	//$merchantAmountData->refreshCache(['merchantId' => $merchantAmountData->merchantId]);
	//
	//(new MerchantChannel)->incrCacheByDayAmountLimit($orderData['merchantNo'], $orderData['channelMerchantNo'], $orderData['payType'], $orderData['bankCode'], $orderData['cardType'], intval($orderData['realOrderAmount'] * 100));
	//(new MerchantChannel)->incrCacheByDayNumLimit($orderData['merchantNo'], $orderData['channelMerchantNo'], $orderData['payType'], $orderData['bankCode'], $orderData['cardType']);
	//(new ChannelPayConfig)->incrCacheByDayAmountLimit($orderData['channelMerchantNo'], $orderData['payType'], $orderData['bankCode'], $orderData['cardType'], intval($orderData['realOrderAmount'] * 100));
	//(new ChannelPayConfig)->incrCacheByDayNumLimit($orderData['channelMerchantNo'], $orderData['payType'], $orderData['bankCode'], $orderData['cardType']);
	//$ppo->setCacheByPlatformOrderNo($orderData['platformOrderNo'], $orderDataLock->toArray());
	return nil
}

func (s *payService) CallSuccess(callback model.RspQuerySettle) error {
	var err error
	//merchantRate := model.MerchantRate{}
	//amountSettlement := model.AmountSettlement{}
	orderData := model.PlatformPayOrder{}
	accountAmount := model.MerchantAmount{}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("platformOrderNo = ?", callback.PlatformOrderNo).First(&orderData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "-查询订单失败 : ", err.Error())
			return err
		}

		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", orderData.MerchantNo).First(&accountAmount).Error

		if err != nil {
			logrus.Error(orderData.MerchantNo, "-查询商户余额失败 : ", err.Error())
			return err
		}
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("orderId = ?", orderData.OrderId).First(&orderData).Error
		if err != nil {
			logrus.Error(orderData.OrderId, "-查询订单失败 : ", err.Error())
			return err
		}
		if orderData.OrderStatus == "Success" || orderData.OrderStatus == "Fail" {
			logrus.Error(orderData.OrderId, "数据已处理，或不存在 : ")
			return errors.New("数据已处理")
		}
		if len(orderData.AccountDate) >= 10 {
			orderData.AccountDate = orderData.AccountDate[0:10]
		}
		whereMap := map[string]interface{}{
			"merchantNo":        orderData.MerchantNo,
			"channelMerchantNo": orderData.ChannelMerchantNo,
			"payType":           orderData.PayType,
			"accountDate":       orderData.AccountDate,
		}
		amountPayData := model.AmountPay{
			MerchantID:        orderData.MerchantId,
			MerchantNo:        orderData.MerchantNo,
			ChannelMerchantID: orderData.ChannelMerchantID,
			ChannelMerchantNo: orderData.ChannelMerchantNo,
			PayType:           orderData.PayType,
			AccountDate:       orderData.AccountDate,
		}
		// 查询创建锁定
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereMap).FirstOrCreate(&amountPayData).Error
		if err != nil {
			logrus.Error(orderData.MerchantNo, "-创建amountPay失败 : ", err.Error())
			return err
		}
		//whereMap := map[string]interface{}{"merchantId": orderData.MerchantId, "channelMerchantId": orderData.ChannelMerchantID, "accountDate": orderData.AccountDate}
		//err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereMap).First(&amountPayData).Error
		//if err != nil {
		//	logrus.Error(orderData.MerchantNo, "-查询amountPay失败 : ", err.Error())
		//	return err
		//}

		//修改订单
		orderData.ChannelNoticeTime = time.Now().Format("2006-01-02 15:04:05")
		orderData.ChannelOrderNo = orderData.ChannelOrderNo
		orderData.ProcessType = "Success"
		orderData.OrderStatus = "Success"
		//orderData.AccountDate = orderData.AccountDate

		//if (!empty($orderData['payType']) && $orderDataLock->payType != $orderData['payType']) {
		//	$orderDataLock->payType = $orderData['payType'];
		//}
		if callback.OrderAmount != orderData.OrderAmount && callback.OrderAmount > 0 {
			orderData.RealOrderAmount = callback.OrderAmount
			//TODO:重新计算手续费
			merchantRate, res := MerchantRate.GetCacheByMerchantNo(orderData.MerchantNo)
			if !res {
				logrus.Error(orderData.MerchantNo, "E2026费率未设置")
				return errors.New(orderData.MerchantNo + "-商户未设置费率")
			}
			channelRate, res := ChannelMerchantRate.GetCacheByMerchantNo(orderData.ChannelMerchantNo)
			if !res {
				logrus.Error("merchantChannel 没有设置费率", orderData.ChannelMerchantNo)
				return errors.New(orderData.ChannelMerchantNo + "-merchantChannel未设置费率")
			}
			reqPayOrder := s.GetReqPayOrderParams(orderData)
			orderData.ServiceCharge, res = MerchantRate.GetServiceCharge(merchantRate, reqPayOrder, "Pay")
			if !res {
				logrus.Error("GetServiceCharge 没有对应支付渠道费率", orderData.ChannelMerchantNo)
				return errors.New(orderData.ChannelMerchantNo + "GetServiceCharge 没有对应支付渠道费率")
			}
			orderData.ChannelServiceCharge, res = ChannelMerchantRate.GetServiceCharge(channelRate, reqPayOrder, "Pay")
			if !res {
				logrus.Error("getChannelServiceCharge 没有设置对应上游支付渠道费率: ", orderData.ChannelMerchantNo)
				return errors.New(orderData.ChannelMerchantNo + "getChannelServiceCharge 没有设置对应上游支付渠道费率")
			}
		} else {
			orderData.RealOrderAmount = orderData.OrderAmount
		}

		orderData.TimeoutTime = time.Now().Format("2006-01-02 15:04:05")
		err = tx.Save(&orderData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "修改支付订单失败 : ", err.Error())
			return err
		}

		amountPayData.AccountDate = orderData.AccountDate
		amountPayData.Amount = amountPayData.Amount + callback.OrderAmount
		amountPayData.ServiceCharge = amountPayData.ServiceCharge + orderData.ServiceCharge
		amountPayData.ChannelServiceCharge = amountPayData.ChannelServiceCharge + orderData.ChannelServiceCharge
		err = tx.Save(&amountPayData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "保存amountPay失败 : ", err.Error())
			return err
		}

		//添加流水
		financeData := []map[string]interface{}{
			{"merchantId": orderData.MerchantId, "merchantNo": orderData.MerchantNo, "platformOrderNo": orderData.PlatformOrderNo, "amount": orderData.RealOrderAmount, "balance": accountAmount.SettlementAmount + orderData.RealOrderAmount, "financeType": "PayIn", "accountDate": orderData.AccountDate, "accountType": "SettlementAccount", "sourceId": orderData.OrderId, "sourceDesc": "支付服务", "summary": "支付服务", "merchantOrderNo": orderData.MerchantOrderNo, "operateSource": "ports"},
			{"merchantId": orderData.MerchantId, "merchantNo": orderData.MerchantNo, "platformOrderNo": orderData.PlatformOrderNo, "amount": orderData.ServiceCharge, "balance": accountAmount.SettlementAmount + orderData.RealOrderAmount - orderData.ServiceCharge, "financeType": "PayOut", "accountDate": orderData.AccountDate, "accountType": "ServiceChargeAccount", "sourceId": orderData.OrderId, "sourceDesc": "支付手续费", "summary": "支付手续费", "merchantOrderNo": orderData.MerchantOrderNo, "operateSource": "ports"},
		}

		err = tx.Table("finance").Create(financeData).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "支付订单补单添加流水失败 : ", err.Error())
			return err
		}
		//更新MerchantAmount余额
		accountAmount.SettlementAmount = accountAmount.SettlementAmount + orderData.RealOrderAmount - orderData.ServiceCharge
		err = tx.Model(&accountAmount).Update("settlementAmount", accountAmount.SettlementAmount).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "更新MerchantAmount余额失败 : ", err.Error())
			return err
		}
		//更新AmountPay余额
		err = tx.Table("amount_pay").Where("merchantId = ?", orderData.MerchantId).Where("accountDate = ?", orderData.AccountDate).Update("balance", accountAmount.SettlementAmount).Error
		if err != nil {
			logrus.Error(callback.PlatformOrderNo, "更新AmountPay余额失败 : ", err.Error())
			return err
		}

		//$ppo->setCacheByPlatformOrderNo($orderData['platformOrderNo'], $orderDataLock->toArray());
		//支付成功减存储的余额
		/*mcsData := model.MerchantChannelSettlement{}
		err = tx.Where("channelMerchantId = ?", orderData.ChannelMerchantID).First(&mcsData).Error
		if err == nil {
			if mcsData.AccountBalance >= orderData.RealOrderAmount {
				err = tx.Where("channelMerchantId = ?", orderData.ChannelMerchantID).Update("accountBalance", gorm.Expr("accountBalance + ?", orderData.RealOrderAmount)).Error
				if err != nil {
					logrus.Error(orderData.ChannelMerchantID, "修改MerchantChannelSettlement余额失败 : ", err.Error())
				}
			}
		} else {
			logrus.Error(orderData.ChannelMerchantID, "查询 MerchantChannelSettlement 失败: ", err.Error())
		}*/

		return nil
	})

	/*AmountPay::where('merchantId', $orderData['merchantId'])
	->where('accountDate', $accountDate)
	->update(['balance' => $merchantAmountData->settlementAmount]);*/
	//	//TODO:代理手续费
	//	$agentId = AgentMerchantRelation::where('merchantId',$orderData['merchantId'])->value('agentId');
	//	if($agentId ||isset($orderData['agentFee']) && $orderData['agentFee'] > 0) {
	//	$agentLog = new AgentIncomeLog();
	//	$agentLog->updateIncomeLog($orderData['merchantId'],$orderData['platformOrderNo'],$orderAmount,'settlement');
	//	}
	if err != nil {
		logrus.Error(callback.PlatformOrderNo, "支付订单回调CallSuccess失败 : ", err.Error())
		return err
	}
	//刷新缓存
	go cache.PayCache.SetCacheByPlatformOrderNo(callback.PlatformOrderNo, orderData)
	//回调商户
	go s.CallbackMerchant(callback.PlatformOrderNo, orderData, model.PayNotifyTask{})
	//$merchantAmountData->refreshCache(['merchantId' => $merchantAmountData->merchantId]);
	//
	//(new MerchantChannel)->incrCacheByDayAmountLimit($orderData['merchantNo'], $orderData['channelMerchantNo'], $orderData['payType'], $orderData['bankCode'], $orderData['cardType'], intval($orderData['realOrderAmount'] * 100));
	//(new MerchantChannel)->incrCacheByDayNumLimit($orderData['merchantNo'], $orderData['channelMerchantNo'], $orderData['payType'], $orderData['bankCode'], $orderData['cardType']);
	//(new ChannelPayConfig)->incrCacheByDayAmountLimit($orderData['channelMerchantNo'], $orderData['payType'], $orderData['bankCode'], $orderData['cardType'], intval($orderData['realOrderAmount'] * 100));
	//(new ChannelPayConfig)->incrCacheByDayNumLimit($orderData['channelMerchantNo'], $orderData['payType'], $orderData['bankCode'], $orderData['cardType']);
	//$ppo->setCacheByPlatformOrderNo($orderData['platformOrderNo'], $orderDataLock->toArray());
	return nil
}

func (s *payService) SetCacheByPlatformOrderNo(platformOrderNo string, payStruct model.PlatformPayOrder) {
	key1 := "payorder:" + platformOrderNo
	key2 := "payorder:m:" + payStruct.MerchantNo + ":" + payStruct.MerchantOrderNo

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

func (s *payService) GetReqPayOrderParams(orderData model.PlatformPayOrder) (payOrderParams model.ReqPayOrder) {

	payOrderParams.MerchantNo = orderData.MerchantNo
	payOrderParams.MerchantOrderNo = orderData.MerchantOrderNo
	payOrderParams.OrderAmount = orderData.RealOrderAmount
	payOrderParams.PayType = orderData.PayType
	payOrderParams.PayModel = orderData.PayModel
	payOrderParams.CardType = orderData.CardType
	payOrderParams.BankCode = orderData.BankCode

	return
}

func (s *payService) CallbackMerchant(orderNo string, order model.PlatformPayOrder, task model.PayNotifyTask) {
	if reflect.DeepEqual(order, model.PlatformPayOrder{}) {
		order, res := cache.PayCache.GetCacheByPlatformOrderNo(orderNo)
		if !res {
			logrus.Error(order.PlatformOrderNo, "-支付回调失败：查询订单失败")
			return
		}
	}
	logrus.Info(orderNo + "-支付回调-" + order.BackNoticeUrl)
	//if order.BackNoticeURL == ""{
	//	logrus.Error(order.PlatformOrderNo,"-支付回调失败：回调地址为空")
	//}
	res := utils.IsValidUrl(order.BackNoticeUrl)
	if !res {
		logrus.Error(order.PlatformOrderNo, "-支付回调失败：回调地址格式错误-", order.BackNoticeUrl)
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
		biz["orderMsg"] = "支付失败"
	} else if order.OrderStatus == "Success" {
		biz["orderMsg"] = "支付成功"
	} else {
		biz["orderMsg"] = "等待支付"
	}
	merchantData, res := MerchantService.GetCacheByMerchantNo(order.MerchantNo)
	if !res {
		logrus.Error(order.PlatformOrderNo, "-支付回调失败：获取商户信息失败")
		return
	}
	encrytkey := merchantData.SignKey
	sign := utils.GetSignStr(encrytkey, biz)
	callbackData["biz"] = biz
	callbackData["sign"] = sign
	data, err := json.Marshal(callbackData)
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "-支付回调失败：格式化数据失败", callbackData)
		return
	}
	response, err, statusCode := utils.HttpPostJson(order.BackNoticeUrl, data)
	logrus.Info(order.PlatformOrderNo, "-支付回调信息：", statusCode, string(response))
	if err != nil {
		logrus.Error(order.PlatformOrderNo, "-支付回调失败：请求失败-", err.Error())
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": task.RetryCount + 1,
			"failReason": "回调异常" + err.Error(),
		}
		go PayNotify.UpdateTask(task, updates)
		updateOrderMaps := map[string]interface{}{
			"callbackLimit": order.CallbackLimit + 1,
		}
		go s.UpdateOrderMap(order, updateOrderMaps)
		return
	}
	if strings.ToLower(string(response)) == "success" {
		if !reflect.DeepEqual(order, model.PlatformPayOrder{}) {
			updates := map[string]interface{}{
				"status":     "Success",
				"retryCount": task.RetryCount + 1,
				"failReason": "回调成功",
			}
			go PayNotify.UpdateTask(task, updates)
		}

		updateOrderMaps := map[string]interface{}{
			"callbackLimit":   order.CallbackLimit + 1,
			"callbackSuccess": true,
		}
		go s.UpdateOrderMap(order, updateOrderMaps)
	} else {
		if !reflect.DeepEqual(order, model.PlatformPayOrder{}) {
			updates := map[string]interface{}{
				"retryCount": task.RetryCount + 1,
				"failReason": string(response),
			}
			go PayNotify.UpdateTask(task, updates)
		}
		go PayNotify.Push(0, orderNo)
		updateOrderMaps := map[string]interface{}{
			"callbackLimit": order.CallbackLimit + 1,
		}
		go s.UpdateOrderMap(order, updateOrderMaps)
	}

	return
}

func (s *payService) UpdateOrderMap(order model.PlatformPayOrder, updates map[string]interface{}) {
	err := sqls.DB().Model(&order).Updates(updates).Error
	if err != nil {
		logrus.Info(order.PlatformOrderNo, "-UpdateOrderMap", err)
	}
	go cache.PayCache.SetCacheByPlatformOrderNo(order.PlatformOrderNo, order)
	return
}

package merchant

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"log"
	"luckypay/channels"
	"luckypay/config"
	"luckypay/model"
	"luckypay/mytool"
	"luckypay/services"
	"luckypay/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SettlementOrderController struct {
	Ctx iris.Context
}

func (c *SettlementOrderController) GetSearch() {
	queryParams := model.SearchSettlementOrder{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var builder strings.Builder
	builder.WriteString("merchantNo =  " + LoginMerchantNo)
	//cnd := sqls.DB().Debug()
	cnd := sqls.DB().Table("platform_settlement_order").Where("merchantNo = ?", LoginMerchantNo)
	if queryParams.PlatformOrderNo != "" {
		builder.WriteString(" and platformOrderNo = '" + queryParams.PlatformOrderNo + "'")
		cnd = cnd.Where("platformOrderNo = ?", queryParams.PlatformOrderNo)
	}
	if queryParams.MerchantOrderNo != "" {
		builder.WriteString(" and merchantOrderNo = '" + queryParams.MerchantOrderNo + "'")
		cnd = cnd.Where("merchantOrderNo = ?", queryParams.MerchantOrderNo)
	}
	if queryParams.MerchantNo > 100 {
		builder.WriteString(" and merchantNo = '" + strconv.FormatInt(queryParams.MerchantNo, 10) + "'")
		//fmt.Print(queryParams.MerchantNo)
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.ChannelMerchantNo != "" {
		builder.WriteString(" and channelMerchantNo = '" + queryParams.ChannelMerchantNo + "'")
		cnd = cnd.Where("channelMerchantNo = ?", queryParams.ChannelMerchantNo)
	}
	if queryParams.BankAccountNo != "" {
		builder.WriteString(" and bankAccountNo = '" + queryParams.BankAccountNo + "'")
		cnd = cnd.Where("bankAccountNo = ?", queryParams.BankAccountNo)
	}
	if queryParams.BankAccountName != "" {
		builder.WriteString(" and bankAccountName = '" + queryParams.BankAccountName + "'")
		cnd = cnd.Where("bankAccountName = ?", queryParams.BankAccountName)
	}
	//fmt.Println(queryParams.MinMoney)
	if queryParams.MinMoney > 0 {
		builder.WriteString(" and orderAmount >= '" + strconv.FormatInt(queryParams.MinMoney, 10) + "'")
		cnd = cnd.Where("orderAmount >= ?", queryParams.MinMoney)
	}
	if queryParams.MaxMoney > queryParams.MinMoney {
		builder.WriteString(" and orderAmount >= '" + strconv.FormatInt(queryParams.MaxMoney, 10) + "'")
		cnd = cnd.Where("orderAmount <= ?", queryParams.MaxMoney)
	}
	if queryParams.BankCode != "" {
		builder.WriteString(" and bankCode = '" + queryParams.BankCode + "'")
		cnd = cnd.Where("bankCode = ?", queryParams.BankCode)
	}

	if queryParams.OrderStatus != "" {
		builder.WriteString(" and orderStatus = '" + queryParams.OrderStatus + "'")
		cnd = cnd.Where("orderStatus = ?", queryParams.OrderStatus)
	}
	if queryParams.Channel != "" {
		builder.WriteString(" and channel = '" + queryParams.Channel + "'")
		cnd = cnd.Where("channel = ?", queryParams.Channel)
	}
	if queryParams.PayType != "" {
		builder.WriteString(" and payType = '" + queryParams.PayType + "'")
		cnd = cnd.Where("payType = ?", queryParams.PayType)
	}
	if queryParams.BeginTime != "" {
		builder.WriteString(" and created_at >= '" + queryParams.BeginTime + "'")
		cnd = cnd.Where("created_at >= ?", queryParams.BeginTime)
	}
	if queryParams.EndTime != "" {
		builder.WriteString(" and created_at <= '" + queryParams.EndTime + "'")
		cnd = cnd.Where("created_at <= ?", queryParams.EndTime)
	}

	//payOrders := []model.PlatformSettlementOrder{}
	payOrders := []map[string]interface{}{}
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("orderId desc").Find(&payOrders).Offset(-1).Count(&count).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	for index, payOrder := range payOrders {
		payOrders[index]["bankCodeDesc"] = config.BankCode[fmt.Sprint(payOrder["bankCode"])]
		payOrders[index]["bankAccountNo"] = payOrder["bankAccountNo"]
		payOrders[index]["channelDesc"] = config.Channel[fmt.Sprint(payOrder["channel"])]
		payOrders[index]["orderStatusDesc"] = config.SettlementOrderStatus[fmt.Sprint(payOrder["orderStatus"])]
		payOrders[index]["payTypeDesc"] = config.PayType[fmt.Sprint(payOrder["payType"])]
	}
	if queryParams.Export == 0 {
		whereStr := builder.String()
		payOrderStat := model.MerchantSettlementOrderStat{}
		sql := "select count(orderId) as number,sum(orderAmount) as orderAmount," + "(select sum(orderAmount) from platform_settlement_order where " + whereStr + " and orderStatus = 'Exception') as exceptionAmount," + "(select count(orderId) from platform_settlement_order where " + whereStr + " and orderStatus = 'Exception') as exceptionNumber," + "(select sum(orderAmount) from platform_settlement_order where " + whereStr + " and orderStatus = 'Success') as successAmount," + "(select count(orderId) from platform_settlement_order where " + whereStr + " and orderStatus = 'Success') as successNumber," + "(select sum(orderAmount) from platform_settlement_order where " + whereStr + " and orderStatus = 'Fail') as failAmount," + "(select count(orderId) from platform_settlement_order where " + whereStr + " and orderStatus = 'Fail') as failNumber,(select sum(orderAmount) from platform_settlement_order where " + whereStr + " and orderStatus = 'Transfered') as transferedAmount,(select count(orderId) from platform_settlement_order where " + whereStr + " and orderStatus = 'Transfered') as transferedNumber,(select sum(serviceCharge) from platform_settlement_order where " + whereStr + ") as serviceCharge from platform_settlement_order"
		err = sqls.DB().Raw(sql).Scan(&payOrderStat).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
		//fmt.Println(payOrderStat)
		rateArr := []model.MerchantRate{}
		err = sqls.DB().Where("merchantId", LoginMerchantId).Where("productType", "Settlement").Find(&rateArr).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
		c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "stat": payOrderStat, "rows": payOrders, "rateArr": rateArr})
	}

}

func (c *SettlementOrderController) PostCreate(r *http.Request) {

	unikey := "merchantSettlement-" + LoginMerchantNo
	redisServer := config.NewRedis()
	exists, err := redisServer.Exists(c.Ctx, unikey).Result()
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "redis！" + err.Error()})
		return
	}
	if exists > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "请求过于频繁！稍候再试"})
		return
	} else {
		redisServer.SetEX(c.Ctx, unikey, 1, 10*time.Second)
	}

	queryParams := model.CreateSettlement{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "form")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	paramStr, _ := json.Marshal(queryParams)
	logrus.Info(LoginMerchantNo + "-商户后台请求代付" + string(paramStr))
	googleAuthKey := Session.GetString("googleAuthSecretKey")
	if googleAuthKey == "" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "请先绑定谷歌验证码"})
		return
	}
	secret := utils.AesCBCDecrypt(googleAuthKey)
	//TODO:密码错误次数上限
	res, err := utils.GoogleAuthenticator.VerifyCode(secret, queryParams.GoogleAuth)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "谷歌验证码错误" + err.Error()})
		return
	}
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 2, "result": "谷歌验证码错误，请重新输入！"})
		return
	}

	merchantAccount := model.MerchantAccount{}
	err = sqls.DB().Where("accountId", LoginAccountId).First(&merchantAccount).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 2, "result": "商户查询失败：" + err.Error()})
		return
	}
	if mytool.GetHashPassword(queryParams.ApplyPerson) != merchantAccount.SecurePwd {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "支付密码错误"})
		return
	}
	//是否开启代付
	merchantData, res := services.MerchantService.GetCacheByMerchantNo(LoginMerchantNo)
	//merchantDataStr, _ := json.Marshal(merchantData)
	//logrus.Info(string(merchantDataStr))
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商务信息有误，请联系客服"})
		return
	}
	if !merchantData.OpenSettlement {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "未开通代付业务，请联系客服"})
		return
	}
	if !merchantData.OpenManualSettlement {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "通道错误，请联系客服"})
		return
	}
	dayamount, err := services.MerchantService.GetCacheByDaySettleAmountLimit(merchantData.MerchantNo)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "查询单日金额限制错误，请联系客服-" + err.Error()})
		return
	}
	if (dayamount + queryParams.OrderAmount) > merchantData.OneSettlementMaxAmount {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "单日累计代付金额超出系统限制：" + strconv.FormatFloat(merchantData.OneSettlementMaxAmount, 'f', 2, 32)})
		return
	}
	userIP := utils.GetRealIp(r)
	if !services.MerchantService.IsAllowIPAccess(userIP) {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "E2104"})
		return
	}
	//TODO:代付黑名单
	//if ($code == 'SUCCESS') {
	//	$blackUserSettlement = new BlackUserSettlement();
	//	$isblackUserExists = $blackUserSettlement->checkBlackUser($request->getParam('bankCode'),$request->getParam('bankAccountName'),$request->getParam('bankAccountNo'));
	//	if($isblackUserExists){
	//		$code = 'E2201';
	//		$logger->error("代付请求：代付黑名单用户！");
	//	}
	//}
	//TODO:风控限制

	merchantRate, res := services.MerchantRate.GetCacheByMerchantNo(merchantData.MerchantNo)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "代付请求：商户未配置费率"})
		return
	}
	//余额判断
	//services.MerchantAmount.RefreshOne(merchantData.MerchantNo)
	amountData := services.MerchantAmount.GetAmount(merchantData.MerchantNo)
	buildParams := model.ReqSettlement{}
	buildParams.MerchantNo = merchantData.MerchantNo
	buildParams.BankCode = queryParams.BankCode
	buildParams.BankAccountNo = queryParams.BankAccountNo
	buildParams.BankAccountName = queryParams.BankAccountName
	buildParams.OrderAmount = queryParams.OrderAmount
	buildParams.Province = queryParams.Province
	buildParams.City = queryParams.City
	merchantServiceCharge, res := services.MerchantRate.GetServiceChargeSettle(merchantRate, buildParams, "Settlement")
	if (queryParams.OrderAmount + merchantServiceCharge) > amountData["settlementAmount"] {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "代付余额不足"})
		return
	}
	//获取代付渠道
	settleChannels, res := services.MerchantChannelSettlement.GetCacheByMerchantNo(merchantData.MerchantNo)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "代付请求：未配置商户代付通道"})
		return
	}
	channelStr, _ := json.Marshal(settleChannels)
	channel, res := services.MerchantChannelSettlement.FetchConfig(merchantData.MerchantNo, settleChannels, "D0Settlement", queryParams.OrderAmount, queryParams.BankCode)
	if !res {
		logrus.Error("merchantChannel fetchConfig失败", string(channelStr))
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户配置相应通道"})
		return
	}
	//上游渠道判断
	channelRate, res := services.ChannelMerchantRate.GetCacheByMerchantNo(channel.ChannelMerchantNo)
	if !res {
		logrus.Error("上游渠道费率未设置-", channel.ChannelMerchantNo)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户费率未设置"})
		return
	}

	channelServiceCharge, res := services.ChannelMerchantRate.GetServiceChargeSettle(channelRate, buildParams, "Settlement")
	if !res {
		logrus.Error("GetServiceChargeSettle 没有设置对应上游代付渠道费率: ", channel.ChannelMerchantNo)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户渠道费率未设置"})
		return
	}

	if _, ok := channels.Channels[channel.Channel]; !ok {
		logrus.Error("-渠道配置错误 ：", channel.Channel)
		c.Ctx.JSON(iris.Map{"success": 0, "result": "渠道配置错误"})
		return
	}
	cacheChannelMerchantData, res := services.ChannelMerchant.GetCacheByChannelMerchantNo(channel.ChannelMerchantNo)
	if !res {
		logrus.Error("-代付渠道获取失败 ：", channel.Channel)
		c.Ctx.JSON(iris.Map{"success": 0, "result": "渠道配置错误"})
		return
	}
	platformOrderNo := services.SettleService.GetPlatformOrderNo("S")
	settleParams := model.SettleParams{
		//Channel:           payChannel.Channel,
		ChannelMerchantNo: channel.ChannelMerchantNo,
		PlatformOrderNo:   platformOrderNo,
		OrderAmoumt:       queryParams.OrderAmount,
		BankCode:          queryParams.BankCode,
		BankAccountName:   queryParams.BankAccountName,
		BankAccountNo:     queryParams.BankAccountNo,
		Province:          queryParams.Province,
		City:              queryParams.City,
	}
	channelObj := channels.Channels[channel.Channel]
	//创建订单数据
	var merchantAmount model.MerchantAmount
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		var err error

		settleOrder, err := services.SettleService.CreateSettlementOrder(tx, buildParams, platformOrderNo, channel, merchantServiceCharge, channelServiceCharge)
		if err != nil {
			logrus.Error(platformOrderNo, "-创建订单失败 : ", err.Error())
			return err
		}

		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", buildParams.MerchantNo).First(&merchantAmount).Error

		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, buildParams.MerchantNo, buildParams.MerchantOrderNo)
			logrus.Error(platformOrderNo, "-查询商户余额失败 : ", err.Error())
			log.Println(err)
			return err
		}
		balance := merchantAmount.SettlementAmount - buildParams.OrderAmount - merchantServiceCharge

		err = tx.Model(merchantAmount).Update("settlementAmount", balance).Error
		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, buildParams.MerchantNo, platformOrderNo)
			logrus.Error(platformOrderNo, "-商户余额更新失败 : ", err.Error())
			log.Println(err)
			return err
		}
		now := time.Now()
		//accountDate := utils.GetFormatTime(now)
		//sqls.DB().Save(&merchantAmount)
		settleFinance := model.Finance{
			MerchantID:      merchantData.MerchantID,
			MerchantNo:      merchantData.MerchantNo,
			PlatformOrderNo: platformOrderNo,
			Amount:          buildParams.OrderAmount,
			Balance:         merchantAmount.SettlementAmount - buildParams.OrderAmount,
			FinanceType:     "PayOut",
			AccountDate:     now.Format("2006-01-02"),
			AccountType:     "SettlementAccount",
			SourceID:        settleOrder.OrderID,
			SourceDesc:      "代付",
			MerchantOrderNo: "",
			OperateSource:   "ports",
			Summary:         queryParams.OrderReason,
		}
		feeFinance := model.Finance{
			MerchantID:      merchantData.MerchantID,
			MerchantNo:      merchantData.MerchantNo,
			PlatformOrderNo: platformOrderNo,
			Amount:          merchantServiceCharge,
			Balance:         balance,
			FinanceType:     "PayOut",
			AccountDate:     now.Format("2006-01-02"),
			AccountType:     "SettlementAccount",
			SourceID:        settleOrder.OrderID,
			SourceDesc:      "代付手续费",
			MerchantOrderNo: "",
			OperateSource:   "ports",
			Summary:         queryParams.OrderReason,
		}
		var finances = []model.Finance{settleFinance, feeFinance}

		err = tx.Create(finances).Error
		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, buildParams.MerchantNo, platformOrderNo)
			logrus.Error(platformOrderNo, "-插入金流日志失败 : ", err.Error())
			log.Println(err)
			return err
		}
		//	now := time.Now()
		accountDate := utils.GetFormatTime(now)
		err = tx.Table("amount_pay").Where("merchantNo = ?", buildParams.MerchantNo).Where("accountDate = ?", accountDate).Update("balance", balance).Error
		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, buildParams.MerchantNo, platformOrderNo)
			logrus.Error(platformOrderNo, "-更新amount_pay失败: ", err.Error())
			log.Println(err)
			return err
		}
		services.SettleService.SetCacheByPlatformOrderNo(platformOrderNo, settleOrder)
		return nil
	})

	if err != nil {
		logrus.Error("创建代付订单失败: " + err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "系统忙，请稍后再试"})
		return
	}
	services.MerchantAmount.RefreshOne(buildParams.MerchantNo)
	//defer c.Settlement(channelObj, settleParams, cacheChannelMerchantData)
	defer services.SettleService.Settlement(channelObj, settleParams, cacheChannelMerchantData)

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "代付创建成功"})

	return
}

func (c *SettlementOrderController) Settlement(channel channels.Channel, settleParams model.SettleParams, channelMerchant model.ChannelMerchant) {
	settleResult, err := channels.SettlementOrder(channel, settleParams, channelMerchant)

	if err != nil {
		logrus.Error(settleParams.PlatformOrderNo, "-请求订单异常 ：", err.Error())
		//c.Ctx.JSON(iris.Map{"success": 0, "result": "订单请求失败"})
		return
	}
	if settleResult.Status != "Success" {
		logrus.Error(settleParams.PlatformOrderNo, "-请求代付订单失败")
		//c.Ctx.JSON(iris.Map{"success": 0, "result": "订单请求失败"})
		return
	}
}

func httpPost(reqUrl string, requestBody []byte) ([]byte, error, int) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	//`这里请注意，使用 InsecureSkipVerify: true 来跳过证书验证`
	client = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	//log.Println("request_data:", string(body))
	logrus.Info("request_data:", string(requestBody))
	// 获取 request请求
	request, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("GetHttpSkip Request Error:", err)
		return nil, err, 400
	}
	// 加入 token
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request.WithContext(context.TODO()))

	if err != nil {
		logrus.Info(string(requestBody), "StatusCode:", resp.StatusCode, "GetHttpSkip Response Error:", err)
		log.Println("GetHttpSkip Response Error:", err)
		return nil, err, 400
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	defer client.CloseIdleConnections()
	//fmt.Println("Response: ", string(responseBody))
	logrus.Info(string(requestBody), "StatusCode:", resp.StatusCode, "resp:", string(responseBody))

	return responseBody, nil, resp.StatusCode
}

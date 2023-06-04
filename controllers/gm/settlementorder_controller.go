package gm

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
	"gorm.io/gorm"
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
	builder.WriteString("1 = 1 ")
	//cnd := sqls.DB().Debug()
	cnd := sqls.DB()
	if queryParams.PlatformOrderNo != "" {
		builder.WriteString(" and platformOrderNo = '" + queryParams.PlatformOrderNo + "'")
		cnd = cnd.Where("platformOrderNo = ?", queryParams.PlatformOrderNo)
	}
	if queryParams.MerchantOrderNo != "" {
		builder.WriteString(" and merchantOrderNo = '" + queryParams.MerchantOrderNo + "'")
		cnd = cnd.Where("merchantOrderNo = ?", queryParams.MerchantOrderNo)
	}
	if queryParams.MerchantNo > 100 {
		builder.WriteString(" and merchantNo = '" + string(queryParams.MerchantNo) + "'")
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
	if queryParams.MinMoney > 0 {
		builder.WriteString(" and orderAmount >= '" + string(queryParams.MinMoney) + "'")
		cnd = cnd.Where("orderAmount >= ?", queryParams.MinMoney)
	}
	if queryParams.MaxMoney > queryParams.MinMoney {
		builder.WriteString(" and orderAmount >= '" + string(queryParams.MaxMoney) + "'")
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
	if queryParams.CreateBeginTime != "" {
		fmt.Println(queryParams.CreateBeginTime)
		builder.WriteString(" and created_at >= '" + queryParams.CreateBeginTime + "'")
		cnd = cnd.Where("created_at >= ?", queryParams.CreateBeginTime)
	}
	if queryParams.CreateEndTime != "" {
		builder.WriteString(" and created_at <= '" + queryParams.CreateEndTime + "'")
		cnd = cnd.Where("created_at <= ?", queryParams.CreateEndTime)
	}

	payOrders := []model.PlatformSettlementOrder{}
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("orderId desc").Find(&payOrders).Offset(-1).Count(&count).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var rows []map[string]interface{}
	for _, payOrder := range payOrders {
		row := make(map[string]interface{})
		row["shortName"] = ""
		merchantData, res := services.MerchantService.GetCacheByMerchantNo(payOrder.MerchantNo)
		if res {
			row["shortName"] = merchantData.ShortName
		}
		row["merchantNo"] = payOrder.MerchantNo
		row["merchantOrderNo"] = payOrder.MerchantOrderNo
		row["orderAmount"] = payOrder.OrderAmount
		row["bankCode"] = payOrder.BankCode
		row["bankAccountNo"] = payOrder.BankAccountNo
		row["bankAccountName"] = payOrder.BankAccountName
		row["backNoticeUrl"] = payOrder.BackNoticeURL
		row["serviceCharge"] = payOrder.ServiceCharge
		row["realOrderAmount"] = payOrder.RealOrderAmount
		row["tradeSummary"] = payOrder.TradeSummary
		row["merchantReqTime"] = payOrder.MerchantReqTime
		row["channel"] = payOrder.Channel
		row["channelMerchantNo"] = payOrder.ChannelMerchantNo
		row["channelNoticeTime"] = payOrder.ChannelNoticeTime
		row["channelServiceCharge"] = payOrder.ChannelServiceCharge
		row["callbackSuccess"] = payOrder.CallbackSuccess
		row["orderId"] = payOrder.OrderID
		row["orderStatus"] = payOrder.OrderStatus
		row["orderType"] = payOrder.OrderType
		row["platformOrderNo"] = payOrder.PlatformOrderNo
		row["processType"] = payOrder.ProcessType
		row["province"] = payOrder.Province
		row["city"] = payOrder.City
		row["pushChannelTime"] = payOrder.PushChannelTime
		row["created_at"] = payOrder.CreatedAt
		row["updated_at"] = payOrder.UpdatedAt
		row["userIp"] = payOrder.UserIP
		row["accountDate"] = payOrder.AccountDate
		row["applyIp"] = payOrder.ApplyIP
		row["applyPerson"] = payOrder.ApplyPerson
		row["auditIp"] = payOrder.AuditIP
		row["auditPerson"] = payOrder.AuditPerson
		row["auditTime"] = payOrder.AuditTime
		row["bankLineNo"] = payOrder.BankLineNo
		row["bankName"] = payOrder.BankName
		row["isLock"] = payOrder.IsLock
		row["lockUser"] = payOrder.LockUser
		row["merchantParam"] = payOrder.MerchantParam
		rows = append(rows, row)
	}
	if queryParams.Export == 0 {
		whereStr := builder.String()
		payOrderStat := model.SettlementOrderStat{}
		sql := "select count(orderId) as number,sum(orderAmount) as orderAmount," + "(select sum(orderAmount) from platform_settlement_order where " + whereStr + " and orderStatus = 'Exception') as exceptionAmount," + "(select count(orderId) from platform_settlement_order where " + whereStr + " and orderStatus = 'Exception') as exceptionNumber," + "(select sum(orderAmount) from platform_settlement_order where " + whereStr + " and orderStatus = 'Success') as successAmount," + "(select count(orderId) from platform_settlement_order where " + whereStr + " and orderStatus = 'Success') as successNumber," + "(select sum(orderAmount) from platform_settlement_order where " + whereStr + " and orderStatus = 'Fail') as failAmount," + "(select count(orderId) from platform_settlement_order where " + whereStr + " and orderStatus = 'Fail') as failNumber,(select sum(orderAmount) from platform_settlement_order where " + whereStr + " and orderStatus = 'Transfered') as transferedAmount,(select count(orderId) from platform_settlement_order where " + whereStr + " and orderStatus = 'Transfered') as transferedNumber from platform_settlement_order"
		err = sqls.DB().Raw(sql).Scan(&payOrderStat).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}

		c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "stat": payOrderStat, "rows": rows})
	}

	if queryParams.Export == 1 {
		titles := []string{"商户号", "商户订单号", "平台订单号", "订单金额(元)", "平台手续费", "上游手续费", "收款银行", "上游渠道", "上游渠道号", "订单状态", "订单生成时间", "处理时间"}
		var data = [][]string{}
		for _, order := range payOrders {
			values := []string{
				order.MerchantNo,
				//order.CreateTime.Format("2006-01-02 15:04:05"),
				order.MerchantOrderNo,
				order.PlatformOrderNo,
				strconv.FormatFloat(order.OrderAmount, 'f', 2, 32),
				strconv.FormatFloat(order.ServiceCharge, 'f', 2, 32),
				strconv.FormatFloat(order.ChannelServiceCharge, 'f', 2, 32),
				order.BankCode,
				order.Channel,
				order.OrderStatus,
				order.CreatedAt.Format("2006-01-02 15:04:05"),
				order.ChannelNoticeTime,
			}
			data = append(data, values)
		}
		fileName := "代付订单" + time.Now().Format("20060102150405") + ".xlsx"
		utils.ExportExcel(c.Ctx, fileName, titles, data)
		return
	}
}

func (c *SettlementOrderController) GetDetail() {
	var orderId, _ = c.Ctx.URLParamInt64("orderId")
	if orderId <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "error OrderId"})
		return
	}
	//var orderDetail model.PlatformSettlementOrder
	orderDetail := services.SettleService.Get(orderId)
	//err := sqls.DB().Where("orderId = ?", orderDetail).First(&orderDetail, 1).Error
	//if err != nil {
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
	//	return
	//}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": orderDetail})
}

func (c *SettlementOrderController) GetNotify() {
	var orderId, _ = c.Ctx.URLParamInt64("orderId")
	if orderId <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "error OrderId"})
		return
	}
	order := model.PlatformSettlementOrder{}
	err := sqls.DB().Table("platform_settlement_order").Where("orderId = ?", orderId).First(&order).Error
	if err != nil {
		logrus.Error(orderId, "-查询失败: ", err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "订单查询失败: " + err.Error()})
		return
	}
	if order.OrderStatus == "Transfered" || order.OrderStatus == "Exception" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "订单状态未完成"})
		return
	}
	go services.SettleService.CallbackMerchant(order.PlatformOrderNo, order, model.SettlementNotifyTask{})
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "订单回调成功"})
	return
}

func (c *SettlementOrderController) PostQueryorder() {
	orderId := c.Ctx.URLParamTrim("orderId")
	order := model.PlatformSettlementOrder{}
	err := sqls.DB().Where("orderId = ?", orderId).Where("orderStatus = ?", "Transfered").First(&order).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if _, ok := channels.Channels[order.Channel]; !ok {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "订单代付渠道不存在：" + order.Channel})
		return
	}
	cacheChannelMerchantData, res := services.ChannelMerchant.GetCacheByChannelMerchantNo(order.ChannelMerchantNo)
	if !res {
		logrus.Error(order.PlatformOrderNo, "-代付渠道获取失败 ：", order.Channel)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "渠道配置错误"})
		return
	}
	Channel := channels.Channels[order.Channel]

	go services.SettleService.QueryOrder(Channel, order, cacheChannelMerchantData, model.SettlementFetchTask{})
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "同步成功！"})
	return
}

func (c *SettlementOrderController) GetExport() {
	orders := []model.PlatformSettlementOrder{}
	var count int64
	err := sqls.DB().Limit(20).Offset(0).Order("orderId desc").Find(&orders).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	file := xlsx.NewFile()
	sheet, _ := file.AddSheet("订单信息")

	titles := []string{"商户号", "平台订单号", "商户订单号", "订单金额(元)", "订单状态", "原因"}
	row := sheet.AddRow()

	var cell *xlsx.Cell
	for _, title := range titles {
		cell = row.AddCell()
		cell.Value = title
	}

	for _, order := range orders {
		values := []string{
			order.MerchantNo,
			order.PlatformOrderNo,
			//order.CreateTime.Format("2006-01-02 15:04:05"),
			order.MerchantNo,
			"1",
			order.OrderStatus,
			order.TradeSummary,
		}

		row = sheet.AddRow()
		for _, value := range values {
			cell = row.AddCell()
			cell.Value = value
		}
	}

	filename := "订单信息" + ".xlsx"
	c.Ctx.Header("Content-Type", "application/octet-stream")
	c.Ctx.Header("Content-Disposition", "attachment; filename="+filename)
	c.Ctx.Header("Content-Transfer-Encoding", "binary")

	//回写到web 流媒体 形成下载
	_ = file.Write(c.Ctx.ResponseWriter())

}

func (c *SettlementOrderController) PostConfirm(req *http.Request) {
	var queryParams model.ConfirmSettlement
	valid_err := utils.Validate(c.Ctx, &queryParams, "")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	redisClient := config.NewRedis()
	unicheckKey := "settlementOrderConfirm-" + LoginAdmin
	unicheck, _ := redisClient.Exists(c.Ctx, unicheckKey).Result()
	if unicheck > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "重复提交，请稍后刷新页面重试"})
		return
	}
	redisClient.SetEX(c.Ctx, unicheckKey, 1, 10*time.Second)
	order := model.PlatformSettlementOrder{}
	err := sqls.DB().Where("orderId = ?", queryParams.OrderId).First(&order).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if order.ProcessType == "Success" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "数据已处理"})
		return
	}
	beforeData, _ := json.Marshal(order)

	var existsLog model.SystemCheckLog
	err = sqls.DB().Where("relevance = ?", order.PlatformOrderNo).Order("id desc").First(&existsLog).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if existsLog.Status == "1" || existsLog.Status == "0" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "已经提交过补单"})
		return
	}
	actionData := make(map[string]interface{})
	actionData["platformOrderNo"] = order.PlatformOrderNo
	actionData["orderAmount"] = order.OrderAmount
	actionData["channel"] = order.Channel
	actionData["channelMerchantNo"] = order.ChannelMerchantNo
	actionData["channelOrderNo"] = queryParams.ChannelOrderNo
	actionData["channelNoticeTime"] = queryParams.ChannelNoticeTime
	actionData["orderStatus"] = queryParams.OrderStatus
	actionData["orderId"] = order.OrderID
	actionData["desc"] = queryParams.Desc
	actionData["pic"] = ""
	actionData["failReason"] = queryParams.FailReason
	content, err := json.Marshal(actionData)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "格式化数据失败" + err.Error()})
		return
	}
	//开启事务
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		systemCheckLog := model.SystemCheckLog{}
		systemCheckLog.Content = string(content)
		systemCheckLog.AdminId = 0
		systemCheckLog.CommiterId = LoginAdminId
		systemCheckLog.Status = "0"
		systemCheckLog.Relevance = order.PlatformOrderNo
		systemCheckLog.Desc = queryParams.Desc
		UserIP := utils.GetRealIp(req)
		systemCheckLog.Ip = UserIP
		systemCheckLog.IpDesc = ""
		systemCheckLog.Type = "代付补单"

		err = tx.Create(&systemCheckLog).Error
		if err != nil {
			logrus.Error("保存checklog失败:", err.Error())
			return err
		}
		actionLog := make(map[string]interface{})
		actionLog["action"] = "MANUAL_PLATFORMSETTLEMENTORDER"
		actionLog["actionBeforeData"] = string(beforeData)
		actionLog["actionAfterData"] = string(beforeData)
		actionLog["status"] = "Success"
		actionLog["accountId"] = LoginAdminId
		UserIP = utils.GetRealIp(req)
		actionLog["ip"] = UserIP
		actionLog["ipDesc"] = "" //TODO:ip描述
		err = tx.Table("system_account_action_log").Create(actionLog).Error
		if err != nil {
			logrus.Error("system_account_action_log-保存失败 : ", err.Error())
			return err
		}
		return nil
	})

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "提交成功，请等待审核"})
	return
}

func (c *SettlementOrderController) GetMakeup() {
	var checkId, _ = c.Ctx.URLParamInt64("id")
	if checkId < 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "error checkId"})
		return
	}
	//var orderDetail model.PlatformSettlementOrder
	checkDetail := model.SystemCheckLog{}
	err := sqls.DB().Where("id = ?", checkId).First(&checkDetail).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	//fmt.Println(checkDetail)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": checkDetail})
}

func (c *SettlementOrderController) PostMakeupcheck(req *http.Request) {

	unicheckKey := "balanceAuditCkeck" + LoginAdmin
	redisClient := config.NewRedis()
	unicheck, err := redisClient.Exists(c.Ctx, unicheckKey).Result()
	if unicheck > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "重复提交，请稍后刷新页面重试"})
		return
	}
	redisClient.SetEX(c.Ctx, unicheckKey, 1, 10*time.Second)
	rowId, _ := c.Ctx.URLParamInt64("id")
	checkey := "checkPwd:check:count" + strconv.FormatInt(LoginAdminId, 10)
	result := redisClient.Get(c.Ctx, checkey)

	if result.Err() != nil && result.Err() != redis.Nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": result.Err().Error()})
		return
	}
	checkcount, err := result.Int64()
	if checkcount >= 5 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "密码错误次数过多"})
		return
	}
	desc := c.Ctx.URLParamTrim("desc")
	status := c.Ctx.URLParamTrim("status")
	checkPwd := c.Ctx.PostValueTrim("checkPwd")
	if len(checkPwd) > 10 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "审核密码格式错误"})
		return
	}
	systemAccount := model.SystemAccount{}
	err = sqls.DB().First(&systemAccount, LoginAdminId).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if systemAccount.CheckPwd == "error" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "审核密码错误超过指定次数，已封审核权限，联系技术"})
		return
	}
	checkPwd = mytool.GetHashPassword(checkPwd)
	if checkPwd != systemAccount.CheckPwd {
		checkcount = checkcount + 1
		redisClient.SetEX(c.Ctx, checkey, checkcount, 72*time.Hour)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "审核密码错误"})
		return
	}

	systemCheckLog := model.SystemCheckLog{}
	err = sqls.DB().Where("id = ?", rowId).First(&systemCheckLog).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if systemCheckLog.Status != "0" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "数据已经处理"})
		return
	}

	content := make(map[string]interface{})
	err = json.Unmarshal([]byte(systemCheckLog.Content), &content)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	order := model.PlatformSettlementOrder{}
	err = sqls.DB().Where("orderId = ?", content["orderId"]).First(&order).Error
	if err != nil {
		logrus.Error(content["orderId"], "-查询失败 : ", err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "数据不存在或已经处理"})
		return
	}
	beforeData, err := json.Marshal(order)
	if err != nil {
		logrus.Error("json转换失败 : ", err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "json转换失败"})
		return
	}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		//$content = json_decode($data->content, true);
		//$orderId = $content['orderId'];
		//$channelOrderNo = $content['channelOrderNo'];
		//$channelNoticeTime = $content['channelNoticeTime'];
		//$failReason = $content['failReason'];
		//$orderStatus = $content['orderStatus'];
		systemCheckLog.Desc = desc
		systemCheckLog.Status = status
		systemCheckLog.AdminId = LoginAdminId
		systemCheckLog.CheckTime = time.Now().Format("2006-01-02 15:04:05")
		UserIP := utils.GetRealIp(req)
		systemCheckLog.CheckIp = UserIP
		//更新审核状态
		err = tx.Save(&systemCheckLog).Error
		if err != nil {
			logrus.Error(content["merchantNo"].(string), "-更改资金失败 : ", err.Error())
			return err
		}

		if status == "1" {

			order.ProcessType = "ManualOperation"
			order.ChannelOrderNo = content["channelOrderNo"].(string)
			order.FailReason = content["failReason"].(string)
			order.ChannelNoticeTime = content["channelNoticeTime"].(string)
			order.AuditPerson = LoginAdmin
			order.OrderStatus = "Success"
			order.AuditIP = UserIP
			order.AuditTime = systemCheckLog.CheckTime
			order.RealOrderAmount = order.OrderAmount
			order.AccountDate = order.AccountDate
			fmt.Println(order)
			if content["orderStatus"] == "Success" {
				//TODO:修改代付订单状态,成功
				err = services.SettleService.Success(order, 0, tx)
				if err != nil {
					logrus.Error(content["orderId"], "更新(Success)订单失败 : ", err.Error())
					return err
				}
			} else {
				//TODO:修改代付订单状态,失败
				err = services.SettleService.Fail(order, 0, tx)
				if err != nil {
					logrus.Error(content["orderId"], "更新(Fail)订单失败 : ", err.Error())
					return err
				}
			}
			afterData, err := json.Marshal(order)
			if err != nil {
				logrus.Error("json转换失败 : ", err.Error())
				return err
			}
			actionLog := make(map[string]interface{})
			actionLog["action"] = "MANUAL_PLATFORMSETTLEMENTORDER"
			actionLog["actionBeforeData"] = string(beforeData)
			actionLog["actionAfterData"] = string(afterData)
			actionLog["status"] = "Success"
			actionLog["accountId"] = LoginAdminId
			UserIP = utils.GetRealIp(req)
			actionLog["ip"] = UserIP
			actionLog["ipDesc"] = "" //TODO:ip描述
			err = tx.Table("system_account_action_log").Create(actionLog).Error
			if err != nil {
				logrus.Error("system_account_action_log-保存失败 : ", err.Error())
				return err
			}
		}

		return nil
	})

	if err != nil {
		logrus.Error(rowId, err.Error())
		redisClient.Del(c.Ctx, unicheckKey)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	//刷新订单缓存
	services.SettleService.RefreshOne(order.PlatformOrderNo)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "操作成功"})

}

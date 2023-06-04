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

type PayOrderController struct {
	Ctx iris.Context
}

func (c *PayOrderController) GetSearch() {
	queryParams := model.SearchPayOrder{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	//fmt.Println(queryParams)
	//if queryParams.ChannelMerchantNo != "" {
	//	channelMerchantData, res := services.ChannelMerchant.GetCacheByChannelMerchantNo(queryParams.ChannelMerchantNo)
	//	if res {
	//		channelMerchantId := channelMerchantData.ChannelMerchantID
	//	} else {
	//		channelMerchantId := 0
	//	}
	//} else {
	//	channelMerchantId := 0
	//}
	//if queryParams.MerchantNo != "" {
	//	merchantData, res := services.MerchantService.GetCacheByMerchantNo(queryParams.MerchantNo)
	//	if res {
	//		MerchantId := merchantData.MerchantID
	//	} else {
	//		MerchantId := 0
	//	}
	//}
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
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

	payOrders := []model.PlatformPayOrder{}
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("orderId desc").Find(&payOrders).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if queryParams.Export == 0 {
		whereStr := builder.String()
		payOrderStat := model.PayOrderStat{}
		sql := "select count(orderId) as number,sum(orderAmount) as orderAmount," + "(select sum(orderAmount) from platform_pay_order where " + whereStr + " and orderStatus = 'WaitPayment') as waitPaymentAmount," + "(select count(orderId) from platform_pay_order where " + whereStr + " and orderStatus = 'WaitPayment') as waitPaymentNumber," + "(select sum(orderAmount) from platform_pay_order where " + whereStr + " and orderStatus = 'Success') as successAmount," + "(select count(orderId) from platform_pay_order where " + whereStr + " and orderStatus = 'Success') as successNumber," + "(select sum(orderAmount) from platform_pay_order where " + whereStr + " and orderStatus = 'Expired') as expiredAmount," + "(select count(orderId) from platform_pay_order where " + whereStr + " and orderStatus = 'Expired') as expiredNumber from platform_pay_order"
		//fmt.Println(sql)
		err = sqls.DB().Raw(sql).Scan(&payOrderStat).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
		c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "stat": payOrderStat, "rows": payOrders})
	}

	if queryParams.Export == 1 {
		titles := []string{"商户号", "商户订单号", "平台订单号", "订单金额(元)", "实际支付金额", "平台手续费", "上游手续费", "支付渠道", "支付方式", "订单状态", "订单生成时间", "订单支付时间"}
		var data = [][]string{}
		for _, order := range payOrders {
			values := []string{
				order.MerchantNo,
				//order.CreateTime.Format("2006-01-02 15:04:05"),
				order.MerchantOrderNo,
				order.PlatformOrderNo,
				strconv.FormatFloat(order.OrderAmount, 'f', 2, 32),
				strconv.FormatFloat(order.RealOrderAmount, 'f', 2, 32),
				strconv.FormatFloat(order.ServiceCharge, 'f', 2, 32),
				strconv.FormatFloat(order.ChannelServiceCharge, 'f', 2, 32),
				order.Channel,
				order.PayType,
				order.OrderStatus,
				order.CreatedAt.Format("2006-01-02 15:04:05"),
				order.ChannelNoticeTime,
			}
			data = append(data, values)
		}
		fileName := "支付订单" + time.Now().Format("20060102150405") + ".xlsx"
		utils.ExportExcel(c.Ctx, fileName, titles, data)
		return
	}
}

func (c *PayOrderController) GetDetail() {
	var orderId, _ = c.Ctx.URLParamInt64("orderId")
	if orderId < 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "error OrderId"})
		return
	}
	//var orderDetail model.PlatformPayOrder
	orderDetail := services.PayService.Get(orderId)
	//err := sqls.DB().Where("orderId = ?", orderDetail).First(&orderDetail, 1).Error
	//if err != nil {
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
	//	return
	//}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": orderDetail})
}

func (c *PayOrderController) PostConfirm(req *http.Request) {
	var queryParams model.ConfirmPayOrder
	valid_err := utils.Validate(c.Ctx, &queryParams, "")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	redisClient := config.NewRedis()
	unicheckKey := "payOrderConfirm-" + LoginAdmin
	unicheck, _ := redisClient.Exists(c.Ctx, unicheckKey).Result()
	if unicheck > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "重复提交，请稍后刷新页面重试"})
		return
	}
	redisClient.SetEX(c.Ctx, unicheckKey, 1, 10*time.Second)
	order := model.PlatformPayOrder{}
	err := sqls.DB().Debug().Where("orderId = ?", queryParams.OrderId).First(&order).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "PlatformPayOrder" + err.Error()})
		return
	}
	if order.OrderStatus != "WaitPayment" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "数据已处理"})
		return
	}
	beforeData, _ := json.Marshal(order)

	var existsLog model.SystemCheckLog
	err = sqls.DB().Where("relevance = ?", order.PlatformOrderNo).Order("id desc").First(&existsLog).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "system_check_log" + err.Error()})
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
	actionData["orderStatus"] = "Success"
	actionData["orderId"] = order.OrderId
	actionData["desc"] = queryParams.Desc
	actionData["pic"] = ""
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
		systemCheckLog.Type = "支付补单"

		err = tx.Create(&systemCheckLog).Error
		if err != nil {
			logrus.Error("保存checklog失败:", err.Error())
			return err
		}
		actionLog := make(map[string]interface{})
		actionLog["action"] = "MANUAL_PLATFORMPAYORDER"
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

func (c *PayOrderController) GetMakeup() {
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
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": checkDetail})
}

func (c *PayOrderController) PostMakeupcheck(req *http.Request) {
	var queryParams model.ConfirmPayOrderCheck
	valid_err := utils.Validate(c.Ctx, &queryParams, "form")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
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
	desc := c.Ctx.PostValueTrim("desc")
	status := c.Ctx.URLParamTrim("status")
	if status != "1" && status != "2" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}
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
	order := model.PlatformPayOrder{}
	err = sqls.DB().Where("orderId = ?", content["orderId"]).First(&order).Error
	if err != nil {
		logrus.Error(content["orderId"], "-查询失败 : ", err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "数据不存在或已经处理"})
		return
	}
	beforeData, err := json.Marshal(order)
	//fmt.Println(string(beforeData))
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
			order.ChannelNoticeTime = content["channelNoticeTime"].(string)
			order.OrderStatus = "Success"
			order.AccountDate = order.AccountDate[0:10]
			fmt.Println(order)
			if content["orderStatus"] == "Success" {
				//TODO:修改代付订单状态,成功
				err = services.PayService.Success(order, order.OrderAmount, tx)
				if err != nil {
					logrus.Error(content["orderId"], "更新(Success)订单失败 : ", err.Error())
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
			//TODO:刷新缓存
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
	services.PayService.RefreshOne(order.PlatformOrderNo)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "操作成功"})

}

func (c *PayOrderController) GetExport() {
	orders := []model.PlatformPayOrder{}
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

func (c *PayOrderController) GetNotify() {
	var orderId, _ = c.Ctx.URLParamInt64("orderId")
	if orderId <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "error OrderId"})
		return
	}
	order := model.PlatformPayOrder{}
	err := sqls.DB().Table("platform_pay_order").Where("orderId = ?", orderId).First(&order).Error
	if err != nil {
		logrus.Error(orderId, "-查询失败: ", err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "订单查询失败: " + err.Error()})
		return
	}

	if order.OrderStatus != "Fail" && order.OrderStatus != "Success" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "订单状态未完成"})
		return
	}
	go services.PayService.CallbackMerchant(order.PlatformOrderNo, order, model.PayNotifyTask{})
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "订单回调成功"})
	return
}

func (c *PayOrderController) PostQueryorder() {
	orderId := c.Ctx.URLParamTrim("orderId")
	order := model.PlatformPayOrder{}
	err := sqls.DB().Where("orderId = ?", orderId).Where("orderStatus = ?", "WaitPayment").First(&order).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if _, ok := channels.Channels[order.Channel]; !ok {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "订单支付渠道不存在：" + order.Channel})
		return
	}
	cacheChannelMerchantData, res := services.ChannelMerchant.GetCacheByChannelMerchantNo(order.ChannelMerchantNo)
	if !res {
		logrus.Error(order.PlatformOrderNo, "-代付渠道获取失败 ：", order.Channel)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "渠道配置错误"})
		return
	}
	Channel := channels.Channels[order.Channel]

	go services.PayService.QueryOrder(Channel, order, cacheChannelMerchantData)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "同步成功！"})
	return
}

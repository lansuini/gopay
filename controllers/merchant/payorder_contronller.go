package merchant

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
	"luckypay/cache"
	"luckypay/config"
	"luckypay/model"
	"luckypay/services"
	"luckypay/utils"
	"strconv"
	"strings"
	"time"
)

type PayOrderController struct {
	Ctx iris.Context
}

func (c *PayOrderController) GetSearch() {
	queryParams := model.SearchPayOrder{}
	//fmt.Println(queryParams)
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
	builder.WriteString("merchantNo = " + LoginMerchantNo)
	cnd := sqls.DB().Where("merchantNo = ?", LoginMerchantNo)
	if queryParams.PlatformOrderNo != "" {
		builder.WriteString(" and platformOrderNo = '" + queryParams.PlatformOrderNo + "'")
		cnd = cnd.Where("platformOrderNo = ?", queryParams.PlatformOrderNo)
	}
	if queryParams.MerchantOrderNo != "" {
		builder.WriteString(" and merchantOrderNo = '" + queryParams.MerchantOrderNo + "'")
		cnd = cnd.Where("merchantOrderNo = ?", queryParams.MerchantOrderNo)
	}
	//if queryParams.MerchantNo > 100 {
	//	builder.WriteString(" and merchantNo = '" + string(queryParams.MerchantNo) + "'")
	//	//fmt.Print(queryParams.MerchantNo)
	//	cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	//}
	if queryParams.OrderStatus != "" {
		builder.WriteString(" and orderStatus = '" + queryParams.OrderStatus + "'")
		cnd = cnd.Where("orderStatus = ?", queryParams.OrderStatus)
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
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("orderId desc").Find(&payOrders).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var rows []map[string]interface{}
	for _, payOrder := range payOrders {
		row := make(map[string]interface{})

		row["platformOrderNo"] = payOrder.PlatformOrderNo
		row["merchantOrderNo"] = payOrder.MerchantOrderNo
		row["orderAmount"] = payOrder.OrderAmount
		row["serviceCharge"] = payOrder.ServiceCharge
		row["payType"] = payOrder.PayType
		row["payTypeDesc"] = config.PayType[payOrder.PayType]
		row["orderStatus"] = payOrder.OrderStatus
		row["orderStatusDesc"] = config.PayOrderStatus[payOrder.OrderStatus]
		row["createTime"] = payOrder.CreatedAt

		rows = append(rows, row)
	}
	if queryParams.Export == 0 {
		whereStr := builder.String()
		payOrderStat := model.PayOrderStat{}
		sql := "select count(orderId) as number,sum(orderAmount) as orderAmount," + "(select sum(orderAmount) from platform_pay_order where " + whereStr + " and orderStatus = 'WaitPayment') as waitPaymentAmount," + "(select count(orderId) from platform_pay_order where " + whereStr + " and orderStatus = 'WaitPayment') as waitPaymentNumber," + "(select sum(orderAmount) from platform_pay_order where " + whereStr + " and orderStatus = 'Success') as successAmount," + "(select count(orderId) from platform_pay_order where " + whereStr + " and orderStatus = 'Success') as successNumber," + "(select sum(orderAmount) from platform_pay_order where " + whereStr + " and orderStatus = 'Expired') as expiredAmount," + "(select count(orderId) from platform_pay_order where " + whereStr + " and orderStatus = 'Expired') as expiredNumber " + ", (select sum(serviceCharge) from platform_pay_order where " + whereStr + ") as serviceCharge " + " from platform_pay_order"
		err = sqls.DB().Raw(sql).Scan(&payOrderStat).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
		rates := []model.MerchantRate{}
		err = sqls.DB().Where("merchantNo = ?", LoginMerchantNo).Where("productType = ?", "Pay").Find(&rates).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "MerchantRate" + err.Error()})
			return
		}
		//fmt.Println(payOrderStat)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rateArr": rates, "stat": payOrderStat, "rows": rows})
		return
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
	fmt.Println(orderDetail)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": orderDetail})
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
	platformOrderNo := c.Ctx.URLParamTrim("platformOrderNo")
	order, res := cache.PayCache.GetCacheByPlatformOrderNo(platformOrderNo)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "订单不存在-" + platformOrderNo})
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

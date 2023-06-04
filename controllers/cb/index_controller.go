package cb

import (
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"luckypay/cache"
	"luckypay/channels"
	"luckypay/services"
	"net/http/httputil"
)

var logger *logrus.Entry

type IndexController struct {
	Ctx      iris.Context
	validate *validator.Validate
}

func PayCallBack(ctx iris.Context) {

	rawReq, _ := httputil.DumpRequest(ctx.Request(), true)
	logrus.Info(string(rawReq))
	platformOrderNo := ctx.Params().GetTrim("platformOrderNo")
	payChannel := platformOrderNo
	if _, ok := channels.Channels[platformOrderNo]; ok {
		payChannel = platformOrderNo
	} else {
		order, res := cache.PayCache.GetCacheByPlatformOrderNo(platformOrderNo)
		if !res {
			logrus.Info(platformOrderNo, "-订单不存在")
			ctx.JSON(iris.Map{"success": 0, "result": "订单不存在"})
			return
		}
		payChannel = order.Channel
	}
	if _, ok := channels.Channels[payChannel]; !ok {
		ctx.JSON(iris.Map{"success": 0, "result": "channel exception"})
		return
	}
	channelObj := channels.Channels[payChannel]

	res, err := channels.PayCallBack(channelObj, ctx)
	if err != nil {
		ctx.WriteString(err.Error())
		return
	}
	if res.Status != "Success" {
		ctx.WriteString(res.FailReason)
		return
	}

	if res.OrderStatus == "Success" {
		go services.PayService.CallSuccess(res)
	}

	ctx.WriteString("SUCCESS")
	return

}

func SettlementCallBack(ctx iris.Context) {
	rawReq, _ := httputil.DumpRequest(ctx.Request(), true)
	logrus.Info(string(rawReq))
	platformOrderNo := ctx.Params().GetTrim("platformOrderNo")
	payChannel := platformOrderNo
	if _, ok := channels.Channels[platformOrderNo]; ok {
		payChannel = platformOrderNo
	} else {
		order, res := cache.SettleCache.GetCacheByPlatformOrderNo(platformOrderNo)
		if !res {
			ctx.JSON(iris.Map{"success": 0, "result": "订单不存在"})
			return
		}
		payChannel = order.Channel
	}
	if _, ok := channels.Channels[payChannel]; !ok {
		ctx.JSON(iris.Map{"success": 0, "result": "channel exception"})
		return
	}
	channelObj := channels.Channels[payChannel]

	res, err := channels.SettleCallBack(channelObj, ctx)
	if err != nil {
		ctx.WriteString(err.Error())
		return
	}
	if res.Status != "Success" {
		ctx.WriteString(res.FailReason)
		return
	}

	if res.OrderStatus == "Success" {
		go services.SettleService.CallSuccess(res)
	}

	if res.OrderStatus == "Fail" {
		go services.SettleService.CallFail(res)
	}
	ctx.WriteString("SUCCESS")
	return

}

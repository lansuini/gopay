package channels

import (
	"github.com/kataras/iris/v12"
	"luckypay/model"
	"luckypay/pkg/config"
)

type Channel interface {
	PayOrder(params model.PayParams, channelMerchant model.ChannelMerchant) (res model.RspPay, err error)
	SettlementOrder(params model.SettleParams, channelMerchant model.ChannelMerchant) (res model.RspSettle, err error)
	QuerySettlementOrder(order model.PlatformSettlementOrder, channelMerchant model.ChannelMerchant) (res model.RspQuerySettle, err error)
	QueryPayOrder(order model.PlatformPayOrder, channelMerchant model.ChannelMerchant) (res model.RspQuerySettle, err error)
	CallBack()
	PayCallBack(ctx iris.Context) (res model.RspQuerySettle, err error)
	SettleCallBack(ctx iris.Context) (res model.RspQuerySettle, err error)
	QueryBalance(channelMerchant model.ChannelMerchant) (res model.RspQueryBalance, err error)
}

var Channels = map[string]Channel{
	//"aliPay":    &AliPay{},
	"loroPay":   &loroPay{},
	"tianciPay": &tianciPay{},
}

var ChannelDescs = map[string]interface{}{
	"loroPay":   "loroPay",
	"tianciPay": "tianciPay",
}

var SettleCallbacks = map[string]interface{}{
	//"aliPay":    &AliPay{},
	"tianciPay": model.LoroPayPayCallback{},
}

func PayOrder(channel Channel, params model.PayParams, channelMerchant model.ChannelMerchant) (res model.RspPay, err error) {
	res, err = channel.PayOrder(params, channelMerchant)

	return
}

func SettlementOrder(channel Channel, params model.SettleParams, channelMerchant model.ChannelMerchant) (res model.RspSettle, err error) {

	res, err = channel.SettlementOrder(params, channelMerchant)
	return
}

func PayCallBack(channel Channel, ctx iris.Context) (res model.RspQuerySettle, err error) {

	res, err = channel.PayCallBack(ctx)
	return
}

func SettleCallBack(channel Channel, ctx iris.Context) (res model.RspQuerySettle, err error) {
	res, err = channel.SettleCallBack(ctx)
	return
}

func QueryPayOrder(channel Channel, order model.PlatformPayOrder, channelMerchant model.ChannelMerchant) (res model.RspQuerySettle, err error) {
	res, err = channel.QueryPayOrder(order, channelMerchant)
	return

}

func QueryBalance(channel Channel, channelMerchant model.ChannelMerchant) (res model.RspQueryBalance, err error) {
	res, err = channel.QueryBalance(channelMerchant)
	return
}

func QuerySettlementOrder(channel Channel, order model.PlatformSettlementOrder, channelMerchant model.ChannelMerchant) (res model.RspQuerySettle, err error) {
	res, err = channel.QuerySettlementOrder(order, channelMerchant)

	return

}

func GetPayCallbackUrl(platformOrderNo string) string {
	callbackUrl := "https://cb." + config.Instance.HostSet.DOMAIN + "/pay/callback/" + platformOrderNo
	return callbackUrl
}

func GetRsaPrivateKey(key string) string {
	return "-----BEGIN RSA PRIVATE KEY-----\n" + key + "\n-----END RSA PRIVATE KEY-----"
}

func GetRsaPublicKey(key string) string {
	return "-----BEGIN PUBLIC KEY-----\n" + key + "\n-----END PUBLIC KEY-----"
}

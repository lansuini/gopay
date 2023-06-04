package config

var ChannelParam = map[string]map[string]string{
	"tianciPay": tianciPay,
	"loroPay":   loroPay,
}

var tianciPay = map[string]string{
	"name":           "TianciPay",
	"open":           "true",
	"openQuery":      "true",
	"openPay":        "true",
	"openSettlement": "true",
	"settlementType": `{"bank" : "银行卡代付"}`,
	"paramDesc":      `{"company" : "上游账号","gateway" : "接口网关","description" : "渠道备注","merchantNo" : "商户号","api_token" : "api_token","notify_token" : "callback_token","ipWhite" : "回调ip白名单"}`,
	"param":          `{"company" : "LD99","gateway" : "https://tiancip030905.com","description" : "Tianci","merchantNo" : "LD99","api_token" : "22","notify_token" : "ARSHBruc4u1LoQSiqI2CCSbsuSBACgu7xdhLDD2AjtjOlPv7efg8FxvgSf9d","ipWhite" : ""}`,
}
var loroPay = map[string]string{
	"name":           "loroPay",
	"open":           "true",
	"openQuery":      "true",
	"openPay":        "true",
	"openSettlement": "true",
	"settlementType": `{"bank" : "银行卡代付"}`,
	"paramDesc": `{"company" : "上游账号","gateway":"接口网关",
        "description":"渠道备注",
        "merchantNo":"商户号",
        "platformPublicKey":"第三方平台公钥",
        "merchantPrivateKey":"商户私钥",
        "merchantPublicKey":"商户公钥",
        "ipWhite":"回调ip白名单"}`,
	"param": `{"company" : "LD99","gateway":"https://api.payloro.tech",
        "description":"loropay",
        "merchantNo":"NO120910918674",
        "platformPublicKey":"",
        "merchantPrivateKey":"",
        "merchantPublicKey":"",
        "ipWhite":""}`,
}

package config

import (
	"encoding/json"
	"os"
)

var BaseData = map[string]map[string]string{
	"channel":                   Channel,
	"payOrderStatus":            PayOrderStatus,
	"settlementOrderStatus":     SettlementOrderStatus,
	"payType":                   PayType,
	"bankCode":                  BankCode,
	"productType":               ProductType,
	"rateType":                  RateType,
	"commonStatus":              CommonStatus,
	"commonStatus2":             CommonStatus2,
	"switchType":                SwitchType,
	"settlementType":            SettlementType,
	"systemAccountRoleCode":     SystemAccountRoleCode,
	"blackusersettlementStatus": BlackusersettlementStatus,
	"blackusersettlementType":   BlackUserSettlementType,
	"enableDisabledStatus":      EnableDisabledStatus,
	"checkStatusCode":           CheckStatusCode,
	"bankrollType":              BankrollType,
	"bankrollDirection":         BankrollDirection,
	"merchantUserLevel":         MerchantUserLevel,
	"merchantUserStatus":        MerchantUserStatus,
	"openType":                  OpenType,
	"financeType":               FinanceType,
	"bankrollTypeCode":          BankrollTypeCode,
	"operateSource":             OperateSource,
	"accountType":               AccountType,
	"systemAccountstatusCode":   AccountType,
}

var BankrollTypeCode = map[string]string{
	"AccountBalance": "账户资金",
}

var CommonStatus = map[string]string{
	"Normal":    "正常",
	"Close":     "关闭",
	"Exception": "异常",
	"Deleted":   "删除",
}

var CommonStatus2 = map[string]string{
	"Success": "通过",
	"Fail":    "失败",
	"Freeze":  "冻结",
	"Unaudit": "待审核",
}

var FinanceType = map[string]string{
	"PayIn":  "收入",
	"PayOut": "支出",
}

// ==============================start==============================
var OpenType = map[string]string{
	"1": "开通",
	"0": "关闭",
}

var MerchantUserStatus = map[string]string{
	"Normal":    "正常",
	"Close":     "关闭",
	"Exception": "异常",
}

var MerchantUserLevel = map[string]string{
	"MerchantManager": "商户管理员",
	"PlatformManager": "充值管理员",
}

var SystemAccountRoleCode = map[string]string{
	"1":  "客服",
	"2":  "财务",
	"3":  "运维",
	"4":  "主管",
	"5":  "管理员",
	"12": "财务(精简版)",
}

var BankrollDirection = map[string]string{
	//"Recharge": "充值",
	"Restore":  "返还",
	"Retrieve": "追收",
	"Freeze":   "冻结",
	"Unfreeze": "解冻",
}

var BankrollType = map[string]string{
	"AccountBalance": "账户资金",
	// "ServiceCharge" : "手续费",
}

var CheckStatusCode = map[string]string{
	"0": "待审核",
	"1": "审核通过",
	"2": "审核不通过",
}

var BlackUserSettlementType = map[string]string{
	"ALIPAY": "支付宝",
	"EBANK":  "银行卡",
}

var BlackusersettlementStatus = map[string]string{
	"enable":  "启用",
	"disable": "禁用",
}

var EnableDisabledStatus = map[string]string{
	"enable":   "启用",
	"disabled": "禁用",
}

var SystemAccountstatusCode = map[string]string{
	"Close":  "关闭",
	"Normal": "正常",
}

var OperateSource = map[string]string{
	"ports":    "接口",
	"merchant": "商户后台",
	"admin":    "管理后台",
}

var AccountType = map[string]string{
	"SettledAccount":       "已结算账户",
	"SettlementAccount":    "未结算账户",
	"AdvanceAccount":       "垫资账户",
	"ServiceChargeAccount": "手续费账户",
}

//==============================end==============================

var ProductType = map[string]string{
	"Pay":        "支付",
	"Settlement": "结算",
	"Recharge":   "充值",
}

var RateType = map[string]string{
	"Rate":       "按比例收取",
	"FixedValue": "按固定值收取",
	"Mixed":      "混合收取",
}

var Channel = map[string]string{
	"tianciPay": "tianciPay",
	"loroPay":   "loroPay",
}

var PayOrderStatus = map[string]string{
	"WaitPayment": "未支付",
	"Success":     "成功",
	"Expired":     "已过期",
}

var PayType = map[string]string{

	"gcash":      "gcash",
	"grabpay":    "grabpay",
	"711_direct": "711_direct",
	"da5":        "da5",
	"qr":         "qr",
	"payngo":     "payngo",
	"posible":    "posible",
	"RLNT":       "RLNT",
	"RDS":        "RDS",
	"UBPB":       "UBPB",
	"RCBC":       "RCBC",
	"SMR":        "SMR",
	"LBC":        "LBC",
	"ussc":       "ussc",
	"PLWN":       "PLWN",
	"partnerpay": "partnerpay",
	"RDP":        "RDP",
	"BPIA":       "BPIA",
	"ECPY":       "ECPY",
	"CEBL":       "CEBL",

	"D0Settlement":     "D0结算",
	"EnterpriseEBank":  "企业网银",
	"PersonalEBank":    "个人网银",
	"PersonalEBankDNA": "个人网银DNA",
	"EnterpriseAlipay": "企业支付宝",
	"AlipayEBank":      "支付宝网银",
	// "EntrustSettlement" : "委托代付",
	// "AdvanceSettlement" : "垫资结算",
	// "HolidaySettlement" : "节假日结算",

}

var SettlementType = map[string]string{
	"D0": "D0结算",
	"D1": "D1结算",
	"T0": "T0结算",
	"T1": "T1结算",
}

var SwitchType = map[string]string{
	"1": "启用",
	"0": "关闭",
}

var SettlementOrderStatus = map[string]string{
	"WaitTransfer": "等待划款",
	"Transfered":   "已划款",
	"Success":      "划款成功",
	"Fail":         "划款失败",
	"Exception":    "订单异常",
}

var BankCode = map[string]string{
	"Globe Gcash": "Globe Gcash",
	"Robinsons":   "Robinsons Bank",

	"ALLBANK (A Thrift Bank), Inc.": "ALLBANK (A Thrift Bank), Inc.",
	"AUB":                           "Asia United Bank",
	"Allbank Corp.":                 "Allbank Corp.",
	"Allied Banking Corp":           "Allied Banking Corp",
	//    "BDO Network Bank" : "BDO Network Bank",
	//    "BDO Network Bank, Inc." : "BDO Network Bank, Inc.",
	"BOC": "Bank Of Commerce",
	"BPI": "Bank Of The Philippine Islands",
	//    "BPI Direct Banko, Inc., A Savings Bank" : "BPI Direct Banko, Inc., A Savings Bank",
	"Banco De Oro Unibank, Inc.":          "Banco De Oro Unibank, Inc.",
	"Bangko Mabuhay (A Rural Bank), Inc.": "Bangko Mabuhay (A Rural Bank), Inc.",
	"Binangonan Rural Bank Inc":           "Bangko Mabuhay (A Rural Bank), Inc.",
	"CBC":                                 "Chinatrust Banking Corp",
	//    "CBS" : "Chinabank Savings",
	"CSB":                                "Citystate Savings Bank Inc.",
	"CTBC":                               "CTBC Bank (Philippines) :  Inc.",
	"Camalig":                            "Camalig Bank",
	"Cebuana Lhuillier Rural Bank, Inc.": "Cebuana Lhuillier Rural Bank, Inc.",
	"Chinabank":                          "Chinabank",
	"DBI":                                "Dungganun Bank Inc.",
	"DCPay":                              "DCPay Philippines Inc.",
	"EB":                                 "Eastwest Bank",
	"ERB":                                "EastWest Rural Bank",
	"ESB":                                "Equicom Savings Bank",

	"GrabPay":                              "GrabPay Philippines",
	"ING":                                  "ING Bank N.V.",
	"ISLA Bank (A Thrift Bank) Inc.":       "ISLA Bank (A Thrift Bank) Inc.",
	"Landbank of the Philippines":          "Landbank of the Philippines",
	"MB":                                   "Malayan Bank",
	"Maybank Philippines Inc.":             "Maybank Philippines Inc.",
	"Metropolitan Bank and Trust Co":       "Metropolitan Bank and Trust Co",
	"Omnipay":                              "Omnipay",
	"PB":                                   "Producers Bank",
	"PBB":                                  "Philippine Business Bank",
	"PBC":                                  "Philippine Bank of Communications",
	"PNB":                                  "Philippine National Bank",
	"PSB":                                  "Philippine Savings Bank",
	"PTC":                                  "Philippine Trust Company",
	"PVB":                                  "Philippine Veterans Bank",
	"Queen City Development Bank, Inc.":    "Queen City Development Bank, Inc.",
	"Paymaya Philippines, Inc.":            "Paymaya Philippines, Inc.",
	"RB":                                   "Robinsons Bank",
	"RBG":                                  "Rural Bank of Guinobatan Inc.",
	"Rizal Commercial Banking Corporation": "Rizal Commercial Banking Corporation",
	"SBA":                                  "Sterling Bank Of Asia",
	"SBC":                                  "Security Bank Corporation",
	"SSB":                                  "Sun Savings Bank",
	"Starpay":                              "Starpay",
	"UCPB SAVINGS BANK":                    "UCPB SAVINGS BANK",
	"UnionBank":                            "UnionBank",
	//    "UnionBank EON" : "UnionBank EON",
	"United Coconut Planters Bank":  "United Coconut Planters Bank",
	"Wealth Development Bank, Inc.": "Wealth Development Bank,  Inc.",
	"Yuanta Savings Bank Inc.":      "Yuanta Savings Bank Inc.",
}

/*
*
服务端配置
*/
type AppConfig struct {
	AppName    string `json:"app_name"`
	Port       string `json:"port"`
	StaticPath string `json:"static_path"`
	Mode       string `json:"mode"`
	Redis      Redis  `json:"redis"`
}

/**
 * Redis 配置
 */
type Redis struct {
	NetWork  string `json:"net_work"`
	Addr     string `json:"addr"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Prefix   string `json:"prefix"`
	Database int    `json:"database"`
}

var ServConfig AppConfig

// 初始化服务器配置
func InitConfig() *AppConfig {
	file, err := os.Open("./config/config.json")
	if err != nil {
		panic(err.Error())
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&ServConfig)
	if err != nil {
		panic(err.Error())
	}
	return &ServConfig
}

//func InitConfig() *AppConfig {
//	file, err := os.Open("./config/config.json")
//	if err != nil {
//		panic(err.Error())
//	}
//	decoder := json.NewDecoder(file)
//	err = decoder.Decode(&ServConfig)
//	if err != nil {
//		panic(err.Error())
//	}
//	Viper.WatchConfig()
//	Viper.OnConfigChange(func(e fsnotify.Event) {
//		logrus.Info("Config file changed:", e.Name)
//		file, err = os.Open("./config/config.json")
//		if err != nil {
//			panic(err.Error())
//		}
//		decoder = json.NewDecoder(file)
//		err = decoder.Decode(&ServConfig)
//	})
//	return &ServConfig
//}

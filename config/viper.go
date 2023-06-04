package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var PayTypeMap = map[interface{}]interface{}{
	"gcash":            "gcash",
	"grabpay":          "grabpay",
	"711_direct":       "711_direct",
	"da5":              "da5",
	"qr":               "qr",
	"payngo":           "payngo",
	"posible":          "posible",
	"RLNT":             "RLNT",
	"RDS":              "RDS",
	"UBPB":             "UBPB",
	"RCBC":             "RCBC",
	"SMR":              "SMR",
	"LBC":              "LBC",
	"ussc":             "ussc",
	"PLWN":             "PLWN",
	"partnerpay":       "partnerpay",
	"RDP":              "RDP",
	"BPIA":             "BPIA",
	"ECPY":             "ECPY",
	"CEBL":             "CEBL",
	"D0Settlement":     "D0结算",
	"EnterpriseEBank":  "企业网银",
	"PersonalEBank":    "个人网银",
	"PersonalEBankDNA": "个人网银DNA",
	"EnterpriseAlipay": "企业支付宝",
	"AlipayEBank":      "支付宝网银",
}

var BankCodeMap = []string{
	"Globe Gcash",
	"Robinsons",
	"ALLBANK (A Thrift Bank), Inc.",
	"AUB",
	"Allbank Corp.",
	"Allied Banking Corp",
	"BDO Network Bank",
	"BDO Network Bank, Inc.",
	"BOC",
	"BPI",
	"BPI Direct Banko, Inc., A Savings Bank",
	"Banco De Oro Unibank, Inc.",
	"Bangko Mabuhay (A Rural Bank), Inc.",
	"Binangonan Rural Bank Inc",
	"CBC",
	"CBS",
	"CSB",
	"CTBC",
	"Camalig",
	"Cebuana Lhuillier Rural Bank, Inc.",
	"Chinabank",
	"DBI",
	"DCPay",
	"EB",
	"ERB",
	"ESB",
	"GrabPay",
	"ING",
	"ISLA Bank (A Thrift Bank) Inc.",
	"Landbank of the Philippines",
	"MB",
	"Maybank Philippines Inc.",
	"Metropolitan Bank and Trust Co",
	"Omnipay",
	"PB",
	"PBB",
	"PBC",
	"PNB",
	"PSB",
	"PTC",
	"PVB",
	"Partner Rural Bank (Cotabato)  Inc.",
	"Paymaya Philippines, Inc.",
	"Queen City Development Bank, Inc.",
	"RB",
	"RBG",
	"Rizal Commercial Banking Corporation",
	"SBA",
	"SBC",
	"SSB",
	"Starpay",
	"UCPB SAVINGS BANK",
	"UnionBank",
	"UnionBank EON",
	"United Coconut Planters Bank",
	"Wealth Development Bank Inc.",
	"Yuanta Savings Bank Inc.",
	"ALIPAY",
}

var once sync.Once

// Viper viper global instance
var Viper *viper.Viper

func init() {
	once.Do(func() {
		Viper = viper.New()
		// scan the file named config in the root directory
		Viper.AddConfigPath("./config")
		Viper.SetConfigName("config")
		Viper.SetConfigType("json")

		// read config, if failed, configure by default
		if err := Viper.ReadInConfig(); err == nil {
			log.Println("Read config successfully: ", Viper.ConfigFileUsed())

		} else {
			log.Printf("Read failed: %s \n", err)
			panic(err)
		}

	})
	Viper.WatchConfig()
	Viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.Info("Config file changed:", e.Name)
	})
}

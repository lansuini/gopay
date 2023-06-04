package model

var ChannelsModel = []interface{}{
	&Merchant{}, &MerchantRate{}, &MerchantChannel{}, &ChannelMerchant{},
}

type LoroPayMerchantParams struct {
	MerchantNo         string `json:"merchantNo"`
	PlatformPublicKey  string `json:"platformPublicKey"`
	MerchantPrivateKey string `json:"merchantPrivateKey"`
	MerchantPublicKey  string `json:"merchantPublicKey"`
	IpWhite            string `json:"ipWhite"`
}

type TianciPayMerchantParams struct {
	MerchantNo         string `json:"merchantNo"`
	PlatformPublicKey  string `json:"platformPublicKey"`
	MerchantPrivateKey string `json:"merchantPrivateKey"`
	MerchantPublicKey  string `json:"merchantPublicKey"`
	IpWhite            string `json:"ipWhite"`
}

type LoroPayPayCallback struct {
	Amount          string `json:"amount"`
	PlatOrderNo     string `json:"platOrderNo"`
	MerchantFee     string `json:"merchantFee"`
	Sign            string `json:"sign"`
	OrderStatus     string `json:"orderStatus"`
	AccountNumber   string `json:"accountNumber"`
	MerchantOrderNo string `json:"merchantOrderNo"`
	OrderMessage    string `json:"orderMessage"`
	MerchantNo      string `json:"merchantNo"`
}

type LoroPaySettleCallback struct {
	Amount          string `json:"amount"`
	PlatOrderNo     string `json:"platOrderNo"`
	MerchantFee     string `json:"merchantFee"`
	Sign            string `json:"sign"`
	OrderStatus     string `json:"orderStatus"`
	MerchantOrderNo string `json:"merchantOrderNo"`
	OrderMessage    string `json:"orderMessage"`
	MerchantNo      string `json:"merchantNo"`
}

type LoroPayQueryBalanceRes struct {
	Status      string                  `json:"status"`
	Message     string                  `json:"message"`
	PlatOrderNo string                  `json:"platOrderNo"`
	MerchantFee string                  `json:"merchantFee"`
	Data        LoroPayQueryBalanceData `json:"data"`
}

type LoroPayQueryBalanceData struct {
	AvailableAmount float64 `json:"availableAmount"`
	FreezeAmount    float64 `json:"freezeAmount"`
	TotalAmount     float64 `json:"totalAmount"`
}

type PayParams struct {
	MerchantNo        string
	ChannelMerchantNo string
	PlatformOrderNo   string
	OrderAmoumt       float64
	PayType           string
	ChannelMerchant   string
}

type SettleParams struct {
	//Channel           string
	ChannelMerchantNo string
	PlatformOrderNo   string
	OrderAmoumt       float64
	BankCode          string
	BankAccountName   string
	BankAccountNo     string
	Province          string
	City              string
}

type QuerySettleParams struct {
	ChannelOrderNo  string
	PlatformOrderNo string
}

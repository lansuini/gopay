package model

import "time"

type TcPay struct {
	OutTradeNo  string  `json:"out_trade_no"`
	Amount      float64 `json:"amount"`
	CallbackUrl string  `json:"callback_url"`
}

type LoroPayParams struct {
	MerchantNo      string  `json:"merchantNo"`
	MerchantOrderNo string  `json:"merchantOrderNo"`
	PayAmount       float64 `json:"payAmount"`
	Method          string  `json:"method"`
	Name            string  `json:"name"`
	FeeType         int     `json:"feeType"`
	Mobile          string  `json:"mobile"`
	Email           string  `json:"email"`
	ExpiryPeriod    string  `json:"expiryPeriod"`
}

type LoroSettle struct {
	MerchantNo      string  `json:"merchantNo"`
	MerchantOrderNo string  `json:"merchantOrderNo"`
	PayAmount       float64 `json:"payAmount"`
	Description     string  `json:"description"`
	BankCode        string  `json:"bankCode"`
	BankNumber      string  `json:"bankNumber"`
	AccountHoldName string  `json:"accountHoldName"`
	Address         string  `json:"address"`
	Barangay        string  `json:"barangay"`
	City            string  `json:"city"`
	ZipCode         string  `json:"zipCode"`
	Gender          string  `json:"gender"`
	FirstName       string  `json:"firstName"`
	MiddleName      string  `json:"middleName"`
	LastName        string  `json:"lastName"`
	Mobile          string  `json:"mobile"`
	Sign            string  `json:"sign"`
}

type RspPay struct {
	ChannelOrderNo string
	PayUrl         string
	Status         string
	FailReason     string
}

type RspSettle struct {
	ChannelOrderNo string
	Status         string
	//OrderAmount     float64
	FailReason      string
	PushChannelTime time.Time
}

type RspQuerySettle struct {
	Status          string
	OrderStatus     string
	PlatformOrderNo string
	ChannelOrderNo  string
	OrderAmount     float64
	FailReason      string
}

type RspQueryBalance struct {
	Status     string  `json:"status"`
	Balance    float64 `json:"balance"`
	FailReason string  `json:"failReason"`
}

type AmountData struct {
	SettlementAmount float64
	AccountBalance   float64
	FreezeAmount     float64
	AvailableBalance float64
	SettledAmount    float64
}

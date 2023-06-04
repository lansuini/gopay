package model

type LoginForm struct {
	LoginName   string `json:"loginName" validate:"required,alphanum,min=5,max=20"`
	LoginPwd    string `json:"loginPwd" validate:"required,alphanum,min=5,max=40"`
	CaptchaId   string `json:"captchaId" validate:"required,alphanum,min=4,max=30"`
	CaptchaCode string `json:"captchaCode" validate:"required,alphanum,min=4,max=6"`
}

type Search struct {
	Limit  int `form:"limit" validate:"required,gte=1,lte=50"`
	Offset int `form:"offset" validate:"omitempty,gte=0,lte=100000000"`
}

type ReqPayOrder struct {
	MerchantNo      string  `json:"merchantNo" validate:"required,alphanum,min=6,max=20,notblank"`
	MerchantOrderNo string  `json:"merchantOrderNo" validate:"required,alphanum,min=5,max=40"`
	OrderAmount     float64 `json:"orderAmount" validate:"required,notblank,gte=10,numeric"`
	TradeSummary    string  `json:"tradeSummary" validate:"required,alphanum,min=3,max=20"`
	PayType         string  `json:"payType" validate:"required,validPayType,max=20"`
	PayModel        string  `json:"payModel" validate:"required,alphanum,min=6,max=20"`
	CardType        string  `json:"cardType" validate:"required,alphanum,min=3,max=20"`
	UserTerminal    string  `json:"userTerminal" validate:"required,alphanum,min=3,max=20"`
	UserIp          string  `json:"userIp" validate:"required,ip,min=6,max=20"`
	BackNoticeUrl   string  `json:"backNoticeUrl" validate:"required,url,min=6,max=150"`
	MerchantParam   string  `json:"merchantParam" validate:"omitempty,max=150,alphanum"`
	MerchantReqTime string  `json:"merchantReqTime"`
	ThirdUserId     string  `json:"thirdUserId"`
	CardHolderName  string  `json:"cardHolderName"`
	CardNum         string  `json:"cardNum"`
	BankCode        string  `json:"bankCode"`
	IdType          string  `json:"idType"`
	IdNum           string  `json:"idNum"`
	Sign            string  `json:"sign" validate:"required,alphanum,min=10,max=50"`
}

type QueryPayOrder struct {
	MerchantNo      string `json:"merchantNo" validate:"required,min=6,max=20,notblank"`
	MerchantOrderNo string `json:"merchantOrderNo" validate:"required,min=6,max=40,notblank"`
	Sign            string `json:"sign" validate:"required,min=10,max=50"`
}

type QueryBalance struct {
	MerchantNo string `json:"merchantNo" validate:"required,min=6,max=20,notblank"`
	Sign       string `json:"sign" validate:"required,min=10,max=50"`
}

type ReqSettlement struct {
	MerchantNo      string  `json:"merchantNo" validate:"required,min=6,max=20,alphanum"`
	MerchantOrderNo string  `json:"merchantOrderNo" validate:"required,alphanum,min=6,max=40"`
	OrderAmount     float64 `json:"orderAmount" validate:"required,notblank,gte=10,numeric"`
	TradeSummary    string  `json:"tradeSummary" validate:"required,alphanum,min=3,max=20"`
	MerchantReqTime string  `json:"merchantReqTime" validate:"required,alphanum,min=6,max=50"`
	BankAccountName string  `json:"bankAccountName" validate:"required,min=3,max=50"` // 收款人姓名
	BankAccountNo   string  `json:"bankAccountNo" validate:"required,alphanum,min=3,max=50"`
	Province        string  `json:"province" validate:"required,alphanum,min=3,max=50"`
	City            string  `json:"city" validate:"required,alphanum,min=3,max=50"`
	BankName        string  `json:"bankName" validate:"required,min=2,max=50"`
	BankCode        string  `json:"bankCode" validate:"required,min=2,max=50,validBankCode"`
	OrderReason     string  `json:"orderReason" validate:"min=3,max=100"`
	BackNoticeUrl   string  `json:"backNoticeUrl" validate:"required,url,min=6,max=150"`
	MerchantParam   string  `json:"merchantParam" validate:"omitempty,alphanum,max=150"`
	RequestIp       string  `json:"requestIp" validate:"required,ip,min=9,max=30"`
	Sign            string  `json:"sign" validate:"required,alphanum,min=10,max=50"`
}

type ConfirmSettlement struct {
	OrderId           int64  `json:"orderId" validate:"required,gte=1,lte=10000000000"`
	ChannelOrderNo    string `json:"channelOrderNo" validate:"omitempty,alphanum,min=3,max=40"`
	ChannelNoticeTime string `json:"channelNoticeTime" validate:"required,max=25"`
	Type              string `json:"type" validate:"required,alphanum,max=25"`
	Desc              string `json:"desc" validate:"required,min=2,max=50"`
	OrderStatus       string `json:"orderStatus" validate:"required,alphanum,oneof=Fail Success"`
	FailReason        string `json:"failReason" validate:"required_if=OrderStatus Fail,omitempty,min=2,max=50"` // 收款人姓名
}

type ConfirmPayOrder struct {
	OrderId           int64  `json:"orderId" validate:"required,gte=1,lte=10000000000"`
	ChannelOrderNo    string `json:"channelOrderNo" validate:"required,alphanum,min=3,max=40"`
	ChannelNoticeTime string `json:"channelNoticeTime" validate:"required,max=25"`
	Type              string `json:"type" validate:"required,alphanum,max=25"`
	Desc              string `json:"desc" validate:"required,alphanum,min=2,max=50"`
}

type ConfirmPayOrderCheck struct {
	Id int64 `json:"id" validate:"required,gte=1,lte=10000000000"`
	//ChannelOrderNo    string `json:"channelOrderNo" validate:"required,alphanum,min=3,max=40"`
	//ChannelNoticeTime string `json:"channelNoticeTime" validate:"required,max=25"`
	Type     string `json:"type" validate:"required,alphanum,max=25"`
	Desc     string `json:"desc" validate:"required,alphanum,min=2,max=50"`
	CheckPwd string `json:"checkPwd" validate:"required,alphanum,min=2,max=20"`
}

type RequireItems struct {
	payOrderStatus string
	payType        string
	channel        string
}

// ========================================start=========================================

type SearchCheck struct {
	Search
	Relevance   string `form:"relevance" validate:"omitempty,min=1,max=40"`
	CheckType   string `form:"checkType" validate:"omitempty,min=1,max=40"`
	CheckStatus string `form:"checkStatus" validate:"omitempty,min=2,max=50"`
}

type SearchBank struct {
	Search
	BankName string `form:"bankName" validate:"omitempty,min=1,max=40"`
	Status   string `form:"status" validate:"omitempty,min=2,max=50"`
}

type SearchBlackUserSettlement struct {
	Search
	BlackUserName    string `form:"blackUserName" validate:"omitempty,min=1,max=40,notblank"`
	BlackUserAccount string `form:"blackUserAccount" validate:"omitempty,min=1,max=40,notblank"`
	BlackUserType    string `form:"blackUserType" validate:"omitempty,min=1,max=40,notblank"`
}

type SystemAccountPwdUpdate struct {
	OldPwd string `form:"oldPwd" validate:"required,alphanum,min=6,max=40,notblank"`
	NewPwd string `form:"newPwd" validate:"required,alphanum,min=6,max=40,notblank"`
}

type SystemAccountUpdate struct {
	Role                string `form:"role" validate:"required,min=1,max=50"`
	Status              string `form:"status" validate:"required,alphanum,min=2,max=50"`
	Id                  int64  `form:"id" validate:"required,gte=1,lte=10000000000,notblank"`
	GoogleAuthSecretKey string `form:"googleAuthSecretKey" validate:"required,alphanum,min=6,max=40,notblank"`
}
type SystemAccountDelete struct {
	Id                  int64  `form:"id" validate:"required,gte=1,lte=10000000000,notblank"`
	GoogleAuthSecretKey string `form:"googleAuthSecretKey" validate:"required,min=6,max=40,notblank"`
}

type MerchantResetset struct {
	OpenPay                *int64  `form:"openPay" validate:"required,oneof=0 1"`
	MerchantNo             string  `form:"merchantNo" validate:"required,alphanum,min=6,max=20"`
	OpenSettlement         *int64  `form:"openSettlement" validate:"required,oneof=0 1"`
	OpenAutoSettlement     *int64  `form:"openAutoSettlement" validate:"required,oneof=0 1"`
	OpenAliSettlement      *int64  `form:"openAliSettlement" validate:"required,oneof=0 1"`
	OneSettlementMaxAmount float64 `form:"oneSettlementMaxAmount" validate:"required,gte=0.0,lte=1000000000"`
	//SettlementTime         *int64   `form:"settlementTime" validate:"required,min=1,max=50,notblank"`
	D0SettlementRate float64 `form:"D0SettlementRate" validate:"required,gt=0,lte=2"`
	//SettlementType         string  `form:"settlementType" validate:"required,oneof=D0 D1 T0 T1"`
}

type MerchantUpdate struct {
	MerchantNo int64  `form:"merchantNo" validate:"required,gte=100,lte=10000000000,notblank"`
	ShortName  string `form:"shortName" validate:"required,alphanum,min=2,max=40,notblank"`
	FullName   string `form:"fullName" validate:"required,alphanum,min=2,max=40,notblank"`
	Status     string `form:"status" validate:"required,min=2,max=20"`
	MerchantId int64  `form:"merchantId" validate:"required,gte=1,lte=10000000000,notblank"`
}

type MerchantAdd struct {
	MerchantNo     string `form:"merchantNo" validate:"required,number,gte=5,lte=100,notblank"`
	ShortName      string `form:"shortName" validate:"required,alphanum,min=2,max=40,notblank"`
	FullName       string `form:"fullName" validate:"required,alphanum,min=2,max=40,notblank"`
	LoginName      string `form:"loginName" validate:"required,alphanum,min=2,max=40,notblank"`
	SecurePwd      string `form:"securePwd" validate:"required,alphanum,min=6,max=40,notblank"`
	LoginPwd       string `form:"loginPwd" validate:"required,alphanum,min=6,max=40,notblank"`
	Description    string `form:"description" validate:"required,alphanum,min=2,max=40,notblank"`
	SettlementType string `form:"settlementType" validate:"required,alphanum,min=1,max=50"`
}

type SearchMerchant struct {
	Search
	MerchantNo int64  `form:"merchantNo" validate:"omitempty,alphanum,gte=100,lte=10000000000,notblank"`
	ShortName  string `form:"shortName" validate:"omitempty,alphanum,min=6,max=40,notblank"`
	FullName   string `form:"fullName" validate:"omitempty,alphanum,min=6,max=40,notblank"`
	LoginName  string `form:"loginName" validate:"omitempty,alphanum,min=6,max=40,notblank"`
	PlatformNo string `form:"platformNo" validate:"omitempty,alphanum,min=6,max=40,notblank"`
	Status     string `form:"status" validate:"omitempty,min=2,max=50"`
	BeginTime  string `form:"beginTime" json:"beginTime" form:"beginTime" validate:"omitempty,min=2,max=20"`
	EndTime    string `form:"endTime" json:"endTime" form:"endTime" validate:"omitempty,min=2,max=20"`
}

type SearchUserMerchant struct {
	Search
	MerchantNo int64  `form:"merchantNo" validate:"omitempty,gte=100,lte=10000000000,notblank"`
	LoginName  string `form:"loginName" validate:"omitempty,min=2,max=40,notblank"`
	PlatformNo int64  `form:"platformNo" validate:"omitempty,gte=100,lte=10000000000,notblank"`
	Status     string `form:"status" validate:"omitempty,min=1,max=50"`
	UserLevel  string `form:"userLevel" validate:"omitempty,min=1,max=40,notblank"`
}

type UserMerchantUpdate struct {
	LoginName string `form:"loginName" validate:"required,alphanum,min=2,max=40,notblank"`
	UserId    int64  `form:"userId" validate:"required,number,min=1,max=100000000000"`
	UserName  string `form:"loginName" validate:"required,alphanum,min=2,max=40,notblank"`
	Status    string `form:"status" validate:"required,min=2,max=50"`
	UserLevel string `form:"userLevel" validate:"required,alphanum,min=1,max=40,notblank"`
}

type BalanceAdjustmentSearch struct {
	Search
	MerchantNo        string `form:"merchantNo" validate:"omitempty,number,min=6,max=40,notblank"`
	PlatformOrderNo   string `form:"platformOrderNo" validate:"omitempty,alphanum,notblank"`
	BankrollType      string `form:"bankrollType" validate:"omitempty,notblank,oneof=AccountBalance"`
	BankrollDirection string `form:"bankrollDirection" validate:"omitempty,notblank,oneof=Restore Retrieve Freeze Unfreeze"`
	Status            string `form:"status" validate:"omitempty,notblank,oneof=Success Freeze Fail Unaudit"`
	BeginTime         string `form:"beginTime" validate:"omitempty,alphanum,min=2,max=20"`      // 生效时间
	EndTime           string `form:"endTime" validate:"omitempty,alphanum,min=2,max=20"`        // 失效时间
	AuditBeginTime    string `form:"auditBeginTime" validate:"omitempty,alphanum,min=2,max=20"` // 生效时间
	AuditEndTime      string `form:"auditEndTime" validate:"omitempty,alphanum,min=2,max=20"`   // 失效时间
}

type PlatformSearch struct {
	Search
	FullName    string `form:"fullName" validate:"omitempty,min=6,max=40,notblank"`
	PlatformNo  string `form:"platformNo" validate:"omitempty,gte=100,lte=10000000000,notblank"`
	Status      string `form:"status" validate:"omitempty,min=2,max=50"`
	Description string `form:"description" validate:"omitempty,min=2,max=40,notblank"`
}

type PlatformUpdate struct {
	Domains              string `form:"domains" validate:"omitempty,url,max=250"`
	PlatformNo           int64  `form:"platformNo" validate:"required,gte=100,lte=10000000000,notblank"`
	Description          string `form:"description" validate:"required,min=2,max=255,notblank"`
	Status               string `form:"status" validate:"required,min=2,max=50"`
	OpenCheckAccount     *int64 `form:"openCheckAccount" validate:"required,min=0,max=1"`
	OpenCheckDomain      *int64 `form:"openCheckDomain" validate:"required,min=0,max=1"`
	OpenFrontNotice      *int64 `form:"openFrontNotice" validate:"required,min=0,max=1"`
	OpenBackNotice       *int64 `form:"openBackNotice" validate:"required,min=0,max=1"`
	OpenRepayNotice      *int64 `form:"openRepayNotice" validate:"required,min=0,max=1"`
	OpenManualSettlement *int64 `form:"openManualSettlement" validate:"required,min=0,max=1"`
	LoginIpWhite         string `form:"loginIpWhite" validate:"omitempty,min=1,max=500"`
	IpWhite              string `form:"ipWhite" validate:"omitempty,min=1,max=500"`
}

type MerchantMerchantUpdatePwd struct {
	NewPwd string `form:"newPwd" validate:"required,alphanum,min=6,max=15"`
	OldPwd string `form:"oldPwd" validate:"required,alphanum,min=6,max=15"`
}

type MerchantMerchantFinanceSearch struct {
	Search
	MerchantOrderNo string `form:"merchantOrderNo" validate:"omitempty,alphanum,min=1,max=50,notblank"`
	PlatformOrderNo string `form:"platformOrderNo" validate:"omitempty,alphanum,min=1,max=50,notblank"`
	OperateSource   string `form:"operateSource" validate:"omitempty,alphanum,oneof=ports merchant admin"`
	FinanceType     string `form:"financeType" validate:"omitempty,alphanum,min=1,max=50"`
	PlatformNo      int64  `form:"platformNo" validate:"omitempty,alphanum,gte=100,lte=10000000000,notblank"`
	BeginTime       string `form:"beginTime" validate:"omitempty,alphanum,min=2,max=20"` // 生效时间
	EndTime         string `form:"endTime" validate:"omitempty,alphanum,min=2,max=20"`   // 失效时间
}

type MerchantReport struct {
	Search
	BeginTime string `form:"beginTime" validate:"omitempty,alphanum,min=2,max=20"` // 生效时间
	EndTime   string `form:"endTime" validate:"omitempty,alphanum,min=2,max=20"`   // 失效时间
}

// ========================================end=========================================

type SearchMerchantRate struct {
	Search
	MerchantNo        int64  `form:"merchantNo" validate:"omitempty,gte=100,lte=10000000000,notblank"`
	MerchantOrderNo   string `form:"merchantOrderNo" validate:"omitempty,min=6,max=40,notblank"`
	PlatformOrderNo   string `form:"platformOrderNo" validate:"omitempty,min=6,max=40,notblank"`
	ChannelMerchantNo string `form:"channelMerchantNo" validate:"omitempty,min=6,max=40,notblank"`
	Status            string `form:"status" validate:"omitempty,min=2,max=50"`
	PayType           string `form:"payType" validate:"omitempty,min=2,max=30"`
	ProductType       string `form:"productType" validate:"omitempty,min=2,max=30"`
	RateType          string `form:"rateType" validate:"omitempty,min=2,max=30"`
}

type SearchChannelMerchantRate struct {
	Search
	MerchantNo   string `form:"merchantNo" validate:"omitempty,alphanum,min=6,max=40,isNumberStr"`
	merchantFlag string `form:"merchantFlag" validate:"omitempty,min=2,max=40,alphanum"`
	Status       string `form:"status" validate:"omitempty,min=2,max=50"`
	PayType      string `form:"payType" validate:"omitempty,min=2,max=30,validPayType"`
	ProductType  string `form:"productType" validate:"omitempty,min=2,max=30"`
	RateType     string `form:"rateType" validate:"omitempty,oneof=Rate FixedValue Mixed,min=2,max=30"`
}

type SearchMerchantPayChannel struct {
	Search
	MerchantNo        string `form:"merchantNo" validate:"omitempty,min=6,max=20,isNumberStr"`
	ChannelMerchantNo string `form:"channelMerchantNo" validate:"omitempty,min=6,max=20,isNumberStr"`
	//Status            string `form:"status" validate:"omitempty,min=2,max=50"`
	Channel string `form:"channel" validate:"omitempty,min=2,max=30"`
}

type SearchMerchantSettleChannel struct {
	Search
	MerchantNo        string `form:"merchantNo" validate:"omitempty,gte=100,lte=10000000000,isNumberStr"`
	ChannelMerchantNo string `form:"channelMerchantNo" validate:"omitempty,min=6,max=40,isNumberStr"`
	//Status            string `form:"status" validate:"omitempty,min=2,max=50"`
	Channel string `form:"channel" validate:"omitempty,min=2,max=30"`
}

type MerchantRateImport struct {
	//BankCode         string  `csv:"bankCode" json:"bankCode" form:"bankCode" validate:"required_if=ProductType Pay,min=2,max=20"`
	BankCode         string   `csv:"bankCode" json:"bankCode" form:"bankCode" validate:"omitempty,validBankCode,min=2,max=20"`
	CardType         string   `csv:"cardType" json:"cardType" form:"cardType" validate:"required,contains=DEBIT,min=2,max=20"`
	BeginTime        string   `csv:"beginTime" json:"beginTime" form:"beginTime" validate:"omitempty,min=2,max=20"` // 生效时间
	EndTime          string   `csv:"endTime" json:"endTime" form:"endTime" validate:"omitempty,min=2,max=20"`       // 失效时间
	MaxServiceCharge *float64 `csv:"maxServiceCharge" json:"maxServiceCharge" form:"maxServiceCharge" validate:"required,gte=0.1,lte=10000000"`
	MinServiceCharge *float64 `csv:"minServiceCharge" json:"minServiceCharge" form:"minServiceCharge" validate:"required,gte=0.1,lte=10000000"`
	MinAmount        *float64 `csv:"minAmount" json:"minAmount" form:"minAmount" validate:"required,gte=0.0,lte=10000000"`
	MaxAmount        *float64 `csv:"maxAmount" json:"maxAmount" form:"maxAmount" validate:"required,gte=0.0,lte=100000000"`
	MerchantID       int64    `csv:"merchantId" json:"merchantId" form:"merchantId" validate:"omitempty,gt=0,lte=100000000"`
	MerchantNo       string   `csv:"merchantNo" json:"merchantNo" form:"merchantNo" validate:"required,min=2,max=20"`
	PayType          string   `csv:"payType" json:"payType" form:"payType" validate:"required,min=2,max=20"`
	ProductType      string   `csv:"productType" json:"productType" form:"productType" validate:"required,oneof=Pay Settlement Recharge"`
	Rate             float64  `csv:"rate" json:"rate" form:"rate" validate:"required_without=Fixed,gte=0,lte=1000"`
	Fixed            float64  `csv:"fixed" json:"fixed" form:"fixed" validate:"required_without=Rate,gte=0,lte=1000"`
	RateType         string   `csv:"rateType" json:"rateType" form:"rateType" validate:"required,oneof=Rate FixedValue Mixed"`
	Status           string   `csv:"status" json:"status" form:"status" validate:"required,oneof=Normal Close"`
}

type ChannelMerchantRateImport struct {
	Channel          string   `csv:"channel" json:"channel" form:"channel" validate:"omitempty,alphanum,validPayChannel,min=2,max=20"`
	BankCode         string   `csv:"bankCode" json:"bankCode" form:"bankCode" validate:"omitempty,min=2,max=20"`
	CardType         string   `csv:"cardType" json:"cardType" form:"cardType" validate:"required,contains=DEBIT,min=2,max=20"`
	BeginTime        string   `csv:"beginTime" json:"beginTime" form:"beginTime" validate:"required,datetime=2006-01-02,min=2,max=20"` // 生效时间
	EndTime          string   `csv:"endTime" json:"endTime" form:"endTime" validate:"omitempty,datetime=2006-01-02,min=2,max=20"`      // 失效时间
	MaxServiceCharge *float64 `csv:"maxServiceCharge" json:"maxServiceCharge" form:"maxServiceCharge" validate:"required,gte=0.1,lte=10000000"`
	MinServiceCharge *float64 `csv:"minServiceCharge" json:"minServiceCharge" form:"minServiceCharge" validate:"required,gte=0.1,lte=10000000"`
	MinAmount        *float64 `csv:"minAmount" json:"minAmount" form:"minAmount" validate:"required,gte=0.0,lte=10000000"`
	MaxAmount        *float64 `csv:"maxAmount" json:"maxAmount" form:"maxAmount" validate:"required,gte=0.0,lte=100000000"`
	MerchantID       int64    `csv:"channelMerchantId" json:"channelMerchantId" form:"channelMerchantId" validate:"omitempty,gt=0,lte=100000000"`
	MerchantNo       string   `csv:"channelMerchantNo" json:"channelMerchantNo" form:"channelMerchantNo" validate:"required,min=2,max=20"`
	PayType          string   `csv:"payType" json:"payType" form:"payType" validate:"required,min=2,max=20"`
	ProductType      string   `csv:"productType" json:"productType" form:"productType" validate:"required,oneof=Pay Settlement Recharge"`
	Rate             float64  `csv:"rate" json:"rate" form:"rate" validate:"required_without=Fixed,gte=0,lte=1000"`
	Fixed            float64  `csv:"fixed" json:"fixed" form:"fixed" validate:"required_without=Rate,gte=0,lte=1000"`
	RateType         string   `csv:"rateType" json:"rateType" form:"rateType" validate:"required,oneof=Rate FixedValue Mixed"`
	Status           string   `csv:"status" json:"status" form:"status" validate:"required,oneof=Normal Close"`
}

type MerchantPayChannelImport struct {
	MerchantNo        string `csv:"merchantNo" json:"merchantNo" form:"merchantNo" validate:"required,isNumberStr,min=2,max=20"`
	MerchantID        int64  `csv:"merchantId" json:"merchantId" form:"merchantId" validate:"omitempty,gt=0,lte=100000000"`
	Channel           string `csv:"channel" json:"channel" form:"channel" validate:"required,alphanum,min=2,max=20,validPayChannel"`
	ChannelMerchantNo string `csv:"channelMerchantNo" json:"channelMerchantNo" form:"channelMerchantNo" validate:"required,isNumberStr,min=2,max=20"`
	PayChannelStatus  string `csv:"payChannelStatus" json:"payChannelStatus" form:"payChannelStatus" validate:"required,alphanum,min=2,max=20"`
	PayType           string `csv:"payType" json:"payType" form:"payType" validate:"required,min=2,max=20,validPayType"`
	//BankCode         string  `csv:"bankCode" json:"bankCode" form:"bankCode" validate:"required_if=ProductType Pay,min=2,max=20"`
	BankCode           string   `csv:"bankCode" json:"bankCode" form:"bankCode" validate:"omitempty,alphanum,min=1,max=20"`
	CardType           string   `csv:"cardType" json:"cardType" form:"cardType" validate:"required,contains=DEBIT,min=2,max=20"`
	OpenOneAmountLimit *int64   `csv:"openOneAmountLimit" json:"openOneAmountLimit" form:"openOneAmountLimit" validate:"required,oneof=0 1"`
	OneMinAmount       *float64 `csv:"oneMinAmount" json:"oneMinAmount" form:"oneMinAmount" validate:"required,gte=0.0,lte=10000000"`
	OneMaxAmount       *float64 `csv:"oneMaxAmount" json:"oneMaxAmount" form:"oneMaxAmount" validate:"required,gte=0.0,lte=100000000"`
	OpenDayAmountLimit *int64   `csv:"openDayAmountLimit" json:"openDayAmountLimit" form:"openDayAmountLimit" validate:"required,oneof=0 1"`
	DayAmountLimit     *float64 `csv:"dayAmountLimit" json:"dayAmountLimit" form:"dayAmountLimit" validate:"required,gte=0.0,lte=10000000"`
	OpenDayNumLimit    *int64   `csv:"openDayNumLimit" json:"openDayNumLimit" form:"openDayNumLimit" validate:"required,oneof=0 1"`
	DayNumLimit        *int64   `csv:"dayNumLimit" json:"dayNumLimit" form:"dayNumLimit" validate:"required,gte=0,lte=100000"`
	OpenTimeLimit      *int64   `csv:"openTimeLimit" json:"openTimeLimit" form:"openTimeLimit" validate:"required,oneof=0 1"`
	BeginTime          int64    `csv:"beginTime" json:"beginTime" form:"beginTime" validate:"omitempty,gte=0,max=2359"` // 生效时间
	EndTime            int64    `csv:"endTime" json:"endTime" form:"endTime" validate:"omitempty,gte=0,max=2400"`       // 失效时间
	Status             string   `csv:"status" json:"status" form:"status" validate:"required,oneof=Normal Close"`
}

type MerchantSettleChannelImport struct {
	MerchantNo              string   `csv:"merchantNo" json:"merchantNo" form:"merchantNo" validate:"required,isNumberStr,min=2,max=20"`
	MerchantID              int64    `csv:"merchantId" json:"merchantId" form:"merchantId" validate:"omitempty,gt=0,lte=100000000"`
	Channel                 string   `csv:"channel" json:"channel" form:"channel" validate:"required,alphanum,min=2,max=20,validPayChannel"`
	ChannelMerchantNo       string   `csv:"channelMerchantNo" json:"channelMerchantNo" form:"channelMerchantNo" validate:"required,isNumberStr,min=2,max=20"`
	ChannelMerchantId       int64    `csv:"_" json:"channelMerchantId" form:"channelMerchantId" validate:"omitempty"`
	SettlementChannelStatus string   `csv:"settlementChannelStatus" json:"settlementChannelStatus" form:"settlementChannelStatus" validate:"required,alphanum,min=2,max=20"`
	SettlementAccountType   string   `csv:"settlementAccountType" json:"settlementAccountType" form:"settlementAccountType" validate:"omitempty,alphanum,min=2,max=20"`
	OpenOneAmountLimit      *int64   `csv:"openOneAmountLimit" json:"openOneAmountLimit" form:"openOneAmountLimit" validate:"required,oneof=0 1"`
	OneMinAmount            *float64 `csv:"oneMinAmount" json:"oneMinAmount" form:"oneMinAmount" validate:"required,gte=0.0,lte=10000000"`
	OneMaxAmount            *float64 `csv:"oneMaxAmount" json:"oneMaxAmount" form:"oneMaxAmount" validate:"required,gte=0.0,lte=100000000"`
	OpenDayAmountLimit      *int64   `csv:"openDayAmountLimit" json:"openDayAmountLimit" form:"openDayAmountLimit" validate:"required,oneof=0 1"`
	DayAmountLimit          *float64 `csv:"dayAmountLimit" json:"dayAmountLimit" form:"dayAmountLimit" validate:"required,gte=0.0,lte=10000000"`
	OpenDayNumLimit         *int64   `csv:"openDayNumLimit" json:"openDayNumLimit" form:"openDayNumLimit" validate:"required,oneof=0 1"`
	DayNumLimit             *int64   `csv:"dayNumLimit" json:"dayNumLimit" form:"dayNumLimit" validate:"required,gte=0,lte=100000"`
	OpenTimeLimit           *int64   `csv:"openTimeLimit" json:"openTimeLimit" form:"openTimeLimit" validate:"required,oneof=0 1"`
	BeginTime               int64    `csv:"beginTime" json:"beginTime" form:"beginTime" validate:"omitempty,gte=0,max=2359"` // 生效时间
	EndTime                 int64    `csv:"endTime" json:"endTime" form:"endTime" validate:"omitempty,gte=0,max=2400"`       // 失效时间
	Status                  string   `csv:"status" json:"status" form:"status" validate:"required,oneof=Normal Close"`
}

type SearchPayOrder struct {
	MerchantNo        int64  `form:"merchantNo" validate:"omitempty,gte=100,lte=10000000000,notblank"`
	MerchantOrderNo   string `form:"merchantOrderNo" validate:"omitempty,min=6,max=40,notblank"`
	PlatformOrderNo   string `form:"platformOrderNo" validate:"omitempty,min=6,max=40,notblank"`
	ChannelMerchantNo string `form:"channelMerchantNo" validate:"omitempty,min=6,max=40,notblank"`
	OrderStatus       string `form:"orderStatus" validate:"omitempty,min=2,max=50"`
	BeginTime         string `form:"beginTime" validate:"omitempty,required,min=5,max=50"`
	EndTime           string `form:"endTime" validate:"omitempty,min=5,max=50"`
	Channel           string `form:"channel" validate:"omitempty,min=2,max=50"`
	PayType           string `form:"payType" validate:"omitempty,min=2,max=30"`
	Export            int64  `form:"export" validate:"omitempty,oneof=1 0"`
	Limit             int    `form:"limit" validate:"omitempty,gte=1,lte=50"`
	Offset            int    `form:"offset" validate:"omitempty,gte=0,lte=100000000"`
}

type SearchSettlementOrder struct {
	MerchantNo        int64  `form:"merchantNo" validate:"omitempty,gte=100,lte=10000000000,notblank"`
	MerchantOrderNo   string `form:"merchantOrderNo" validate:"omitempty,alphanum,min=6,max=40,notblank"`
	PlatformOrderNo   string `form:"platformOrderNo" validate:"omitempty,alphanum,min=6,max=40,notblank"`
	ChannelMerchantNo string `form:"channelMerchantNo" validate:"omitempty,alphanum,min=6,max=40,notblank"`
	OrderStatus       string `form:"orderStatus" validate:"omitempty,alphanum,min=2,max=50"`
	BankCode          string `form:"bankCode" validate:"omitempty,min=2,max=50"`
	BankAccountName   string `form:"bankAccountName" validate:"omitempty,min=2,max=50"`
	BankAccountNo     string `form:"bankAccountNo" validate:"omitempty,min=2,max=50"`
	CreateBeginTime   string `form:"createBeginTime" validate:"omitempty,required,min=5,max=50"`
	CreateEndTime     string `form:"createEndTime" validate:"omitempty,min=5,max=50"`
	Channel           string `form:"channel" validate:"omitempty,min=2,max=50"`
	PayType           string `form:"payType" validate:"omitempty,min=2,max=30"`
	MinMoney          int64  `form:"minMoney" validate:"omitempty,gte=0,lte=100000000"`
	MaxMoney          int64  `form:"maxMoney" validate:"omitempty,gte=0,lte=100000000"`
	Export            int64  `form:"export" validate:"omitempty,oneof=1 0"`
	Limit             int    `form:"limit" validate:"omitempty,gte=1,lte=50"`
	Offset            int    `form:"offset" validate:"omitempty,gte=0,lte=100000000"`
	BeginTime         string `form:"beginTime" validate:"omitempty,required,min=5,max=50"`
	EndTime           string `form:"endTime" validate:"omitempty,min=5,max=50"`
}

type SearchChannelMerchant struct {
	Search
	MerchantNo int64  `form:"merchantNo" validate:"omitempty,gte=100,lte=10000000000"`
	Channel    string `form:"channel" validate:"omitempty,alphanum,min=2,max=20"`
	Status     string `form:"status" validate:"omitempty,alphanum,min=2,max=50"`
}

type SearchFinance struct {
	MerchantNo      string `form:"merchantNo" validate:"omitempty,isNumberStr,min=3,max=40"`
	PlatformOrderNo string `form:"platformOrderNo" validate:"omitempty,min=6,max=40,alphanum"`
	SourceDesc      string `form:"sourceDesc" validate:"omitempty,min=2,max=40"`
	BeginTime       string `form:"beginTime" validate:"omitempty,min=5,max=50"`
	EndTime         string `form:"endTime" validate:"omitempty,min=5,max=50"`
	FinanceType     string `form:"financeType" validate:"omitempty,alphanum,oneof=PayOut PayIn,min=2,max=30"`
	Export          int64  `form:"export" validate:"omitempty,oneof=1 0"`
	Limit           int    `form:"limit" validate:"omitempty,gte=1,lte=50"`
	Offset          int    `form:"offset" validate:"omitempty,gte=0,lte=100000000"`
}

type SearchCheckList struct {
	CheckStatus string `form:"checkStatus" validate:"required,oneof=-1 0 1 2"`
	Relevance   string `form:"relevance" validate:"omitempty,min=2,max=40,alphanum"`
	CheckType   string `form:"checkType" validate:"omitempty,min=2,max=40,alphanum"`
	Limit       int    `form:"limit" validate:"required,gte=1,lte=50"`
	Offset      *int   `form:"offset" validate:"required,gte=0,lte=100000000"`
}

type SearchBusinessAmount struct {
	MerchantNo string `form:"merchantNo" validate:"omitempty,isNumberStr,min=3,max=40"`
	PayType    string `form:"payType" validate:"omitempty,alphanum,min=2,max=30"`
	BeginDate  string `form:"beginDate" validate:"omitempty,min=5,max=50"`
	EndDate    string `form:"endDate" validate:"omitempty,min=5,max=50"`
	Export     int64  `form:"export" validate:"omitempty,oneof=1 0"`
	Limit      int    `form:"limit" validate:"omitempty,gte=1,lte=50"`
	Offset     int    `form:"offset" validate:"omitempty,gte=0,lte=100000000"`
}

type SearchPayAmount struct {
	MerchantNo string `form:"merchantNo" validate:"omitempty,isNumberStr,min=3,max=40"`
	BeginDate  string `form:"beginDate" validate:"omitempty,min=5,max=50"`
	EndDate    string `form:"endDate" validate:"omitempty,min=5,max=50"`
	Export     int64  `form:"export" validate:"omitempty,oneof=1 0"`
	Limit      int    `form:"limit" validate:"omitempty,gte=1,lte=50"`
	Offset     int    `form:"offset" validate:"omitempty,gte=0,lte=100000000"`
	PayType    string `form:"payType" validate:"omitempty,min=2,max=40,alphanum"`
}

type SearchBalanceAdjustment struct {
	MerchantNo        string `form:"merchantNo" validate:"omitempty,isNumberStr,min=3,max=40"`
	PlatformOrderNo   string `form:"platformOrderNo" validate:"omitempty,min=6,max=40,alphanum"`
	BankrollDirection string `form:"bankrollDirection" validate:"omitempty,alphanum,oneof=Recharge Restore Retrieve Freeze Unfreeze,min=2,max=40"`
	BankrollType      string `form:"bankrollType" validate:"omitempty,alphanum,oneof=AccountBalance ServiceCharge,min=2,max=40"`
	Status            string `form:"status" validate:"omitempty,alphanum,oneof=Success Fail Freeze Unaudit,min=2,max=50"`
	AuditBeginTime    string `form:"auditBeginTime" validate:"omitempty,min=5,max=50"`
	AuditEndTime      string `form:"auditEndTime" validate:"omitempty,min=5,max=50"`
	BeginTime         string `form:"beginTime" validate:"omitempty,min=5,max=50"`
	EndTime           string `form:"endTime" validate:"omitempty,min=5,max=50"`
	Limit             int    `form:"limit" validate:"omitempty,gte=1,lte=50"`
	Offset            int    `form:"offset" validate:"omitempty,gte=0,lte=100000000"`
}

type BalanceAdj struct {
	MerchantNo        string  `json:"merchantNo" form:"merchantNo" validate:"required,alphanum,isNumberStr,min=3,max=40"`
	Amount            float64 `json:"amount" form:"amount" validate:"required,gte=1,lte=1000000"`
	FactFee           float64 `json:"factFee" form:"factFee" validate:"required_if=BankrollDirection Recharge,omitempty,min=1,max=1000000"`
	SysFee            float64 `json:"sysFee" form:"factFee" validate:"required_if=BankrollDirection Recharge,omitempty,min=1,max=1000000"`
	Summary           string  `json:"summary" form:"summary" validate:"required,alphanum,min=2,max=40"`
	Random            string  `json:"random" form:"random" validate:"required,alphanum,min=2,max=40"`
	BankrollDirection string  `json:"bankrollDirection" form:"bankrollDirection" validate:"required,oneof=Recharge Restore Retrieve Freeze Unfreeze,min=3,max=40"`
}

type ChannelMerchantInsert struct {
	MerchantNo     string            `json:"merchantNo" form:"merchantNo" validate:"required,alphanum,min=2,max=20,isNumberStr"`
	Channel        string            `json:"channel" form:"channel" validate:"required,alphanum,min=2,max=20"`
	DelegateDomain string            `json:"delegateDomain" form:"delegateDomain" validate:"required,url,min=2,max=100"`
	Param          map[string]string `json:"param" form:"param" validate:"required,max=15,dive,min=4,max=10000"`
}

type ChannelMerchantUpdate struct {
	ChannelMerchantInsert
	MerchantId int64  `json:"merchantId" form:"merchantId" validate:"required,min=2,max=100000000"`
	Status     string `json:"status" form:"status" validate:"required,min=2,max=20,oneof=Normal Close Exception Deleted"`
}

type LoroPayMerchantParam struct {
	Company            string `json:"company" form:"company" validate:"omitempty,min=2,max=50"`
	Description        string `json:"description" form:"description" validate:"omitempty,min=2,max=50"`
	IpWhite            string `json:"ipWhite" form:"ipWhite" validate:"omitempty,min=2,max=50"`
	MerchantNo         string `json:"merchantNo" form:"merchantNo" validate:"required,min=2,max=50"`
	MerchantPublicKey  string `json:"merchantPublicKey" form:"merchantPublicKey" validate:"max=3000"`
	MerchantPrivateKey string `json:"merchantPrivateKey" form:"merchantPrivateKey" validate:"max=3000"`
	PlatformPublicKey  string `json:"platformPublicKey" form:"platformPublicKey" validate:"required,min=2,max=3000"`
}

type ChannelAddBalance struct {
	MerchantNo        string `json:"merchantNo" form:"merchantNo" validate:"required,alphanum,min=2,max=20,isNumberStr"`
	Balace            string `json:"balace" form:"balace" validate:"required,alphanum,min=2,max=20,isNumberStr"`
	ChannelAccount    string `json:"channelAccount" form:"channelAccount" validate:"required,alphanum,min=2,max=20,isNumberStr"`
	NotifyOrderNumber string `json:"notifyOrderNumber" form:"notifyOrderNumber" validate:"required,alphanum,min=2,max=20,isNumberStr"`
}

type CreateSettlement struct {
	BankAccountNo   string  `form:"bankAccountNo" validate:"required,alphanum,min=2,max=30"`
	BankCode        string  `form:"bankCode" validate:"required,validBankCode,min=2,max=30"`
	GoogleAuth      string  `form:"googleAuth" validate:"required,alphanum,min=2,max=30"`
	BankAccountName string  `form:"bankAccountName" validate:"required,alphanum,min=2,max=30"`
	Province        string  `form:"province" validate:"required,alphanum,min=2,max=30"`
	City            string  `form:"city" validate:"required,alphanum,min=2,max=20"`
	BankName        string  `form:"bankName" validate:"required,alphanumspace,min=2,max=30"`
	OrderReason     string  `form:"orderReason" validate:"required,alphanum,min=2,max=30"`
	ApplyPerson     string  `form:"applyPerson" validate:"required,alphanum,alphanum,min=2,max=30"`
	OrderAmount     float64 `form:"orderAmount" validate:"required,gte=10,lte=10000000"`
}

type SearchAdminLog struct {
	Search
	LoginName string `form:"loginName" validate:"omitempty,alphanum,min=1,max=20"`
	Ip        string `form:"ip" validate:"omitempty,alphanum,min=2,max=50"`
}

/*type ChannelMerchantInsert struct {
	MerchantNo     int64             `json:"merchantNo" form:"merchantNo" validate:"required|int|withoutWhitespace|gte:100|lte:1000000000"`
	Channel        string            `json:"channel" form:"channel" validate:"required|string|withoutWhitespace|minLen:2|maxLen:20"`
	DelegateDomain string            `json:"delegateDomain" form:"delegateDomain" validate:"required|withoutWhitespace|fullUrl|minLen:2|maxLen:100"`
	Param          map[string]string `json:"param" form:"param" validate:"required|map|withoutWhitespace"`
	//Param          map[string]string `json:"param" form:"param" validate:"required|map|max:15|dive|min:4|max:1000"`
}*/

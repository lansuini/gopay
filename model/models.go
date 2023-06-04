package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

var Models = []interface{}{
	//&Merchant{}, &MerchantRate{}, &MerchantChannel{}, &ChannelMerchant{},
}

type Model struct {
	Id int64 `gorm:"primaryKey;autoIncrement" json:"id" form:"id"`
}

const timeFormat = "2006-01-02 15:04:05"
const timezone = "Asia/Shanghai"

type DateTime time.Time

func (t DateTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormat)
	b = append(b, '"')
	return b, nil
}

func (t *DateTime) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	*t = DateTime(now)
	return
}
func (t DateTime) String() string {
	return time.Time(t).Format(timeFormat)
}

func (t DateTime) Local() time.Time {
	loc, _ := time.LoadLocation(timezone)
	return time.Time(t).In(loc)
}

func (t DateTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	var ti = time.Time(t)
	if ti.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ti, nil
}

func (t *DateTime) Scan(v interface{}) error {
	return nil
	value, ok := v.(time.Time)
	if ok {
		*t = DateTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

type LocalTime time.Time

func (t *LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return []byte(fmt.Sprintf("\"%v\"", tTime.Format("2006-01-02 15:04:05"))), nil
}

func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	//判断给定时间是否和默认零时间的时间戳相同
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

//type DateTime struct {
//	time.Time
//}
//
//// Convert the internal date as CSV string
//func (date *DateTime) MarshalCSV() (string, error) {
//	return date.String(), nil
//}
//
//// You could also use the standard Stringer interface
//func (date DateTime) String() string {
//	return date.Time.Format("2006-02-01 15:04:05")
//}

type SystemCheckLog struct {
	Model
	AdminId    int64     `gorm:"size:55;unique;column:admin_id;" json:"admin_id" form:"admin_id"`
	CommiterId int64     `gorm:"size:55;unique;column:commiter_id;" json:"commiter_id" form:"commiter_id"`
	Status     string    `gorm:"column:status;size:10;" json:"status" form:"status"`
	Content    string    `gorm:"column:content;type:longtext;not null;" json:"content" form:"content"`
	Desc       string    `gorm:"column:desc;type:longtext;not null;" json:"desc" form:"desc"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"` //
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"` //
	Type       string    `gorm:"column:type;type:varchar(255);not null;" json:"type"`
	IpDesc     string    `gorm:"column:ipDesc;type:varchar(50);not null;" json:"ipDesc"`
	Ip         string    `gorm:"column:ip;type:varchar(50);not null;" json:"ip"`
	Relevance  string    `gorm:"column:relevance;type:varchar(200);not null;" json:"relevance"`
	CheckTime  string    `gorm:"column:check_time;type:varchar(200);not null;" json:"check_time"`
	CheckIp    string    `gorm:"column:check_ip;type:varchar(200);not null;" json:"check_ip"`
}

type Banks struct {
	Id        uint      `gorm:"column:id;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"id"` //
	Name      string    `gorm:"column:name;type:varchar(64);not null;" json:"name"`                             // 名称
	Code      string    `gorm:"column:code;type:varchar(15);not null;" json:"code"`                             // 代号
	Status    string    `gorm:"column:status;type:enum('enabled','disabled');not null;" json:"status"`          //
	StartTime string    `gorm:"column:start_time;type:datetime;" json:"start_time"`                             // 开始时间
	EndTime   string    `gorm:"column:end_time;type:datetime;" json:"end_time"`                                 // 结束时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                    //
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                    //
}

type BlackUserSettlement struct {
	BlackUserID      int64     `gorm:"column:blackUserId;primary_key;auto_increment;type:int(11);not null;" json:"blackUserId"` //
	BlackUserType    string    `gorm:"column:blackUserType;type:varchar(50);not null;" json:"blackUserType"`                    //
	BlackUserAccount string    `gorm:"column:blackUserAccount;type:varchar(50);" json:"blackUserAccount"`                       //
	BlackUserName    string    `gorm:"column:blackUserName;type:varchar(50);" json:"blackUserName"`                             //
	BlackUserStatus  string    `gorm:"column:blackUserStatus;type:varchar(50);" json:"blackUserStatus"`                         //
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime;" json:"created_at"`                                      //
	UpdatedAt        time.Time `gorm:"column:updated_at;type:datetime;" json:"updated_at"`                                      //
}

type SystemAccount struct {
	Model
	UserName            string    `gorm:"size:55;unique;column:userName;" json:"userName" form:"userName"`
	LoginName           string    `gorm:"size:55;unique;column:loginName;" json:"loginName" form:"loginName"`
	LoginPwd            string    `gorm:"size:255;column:loginPwd;" json:"loginPwd" form:"loginPwd"`
	Status              string    `gorm:"column:status;size:10;" json:"status" form:"status"`
	Role                int64     `gorm:"column:role;type:tinyint(3);" json:"role" form:"role"`
	LoginFailNum        int64     `gorm:"column:loginFailNum;type:int(50) unsigned;not null;" json:"loginFailNum"`           // 登录失败次数
	LoginPwdAlterTime   time.Time `gorm:"column:loginPwdAlterTime;type:datetime;" json:"loginPwdAlterTime"`                  // 密码修改时间
	GoogleAuthSecretKey string    `gorm:"column:googleAuthSecretKey;type:varchar(500);not null;" json:"googleAuthSecretKey"` // 谷歌auth密钥
	CreatedAt           time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                       //
	UpdatedAt           time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                       //
	CheckPwd            string    `gorm:"column:check_pwd;type:varchar(50);not null;" json:"check_pwd"`                      // 审核密码
	GoogleBind          string    `gorm:"_" json:"googleBind" form:"googleBind"`
}

type PlatformPayOrder struct {
	OrderId              int64     `gorm:"primaryKey;autoIncrement;column:orderId;type:int(11);not null" json:"orderId"`
	MerchantNo           string    `gorm:"column:merchantNo;type:varchar(50)" json:"merchantNo"`
	MerchantId           int64     `gorm:"column:merchantId;type:int(11)" json:"merchantId"`
	MerchantOrderNo      string    `gorm:"column:merchantOrderNo;type:char(50)" json:"merchantOrderNo"`
	MerchantParam        string    `gorm:"column:MerchantParam;type:varchar(150)" json:"merchantParam"`
	MerchantReqTime      string    `gorm:"column:merchantReqTime;type:datetime" json:"merchantReqTime"`
	PlatformOrderNo      string    `gorm:"column:platformOrderNo;type:varchar(50)" json:"platformOrderNo"`
	ChannelOrderNo       string    `gorm:"column:channelOrderNo;type:varchar(50)" json:"channelOrderNo"`
	ChannelMerchantID    int64     `gorm:"column:channelMerchantId;type:int(11);not null;"`
	OrderStatus          string    `gorm:"column:orderStatus;type:text" json:"orderStatus"`
	ProcessType          string    `gorm:"column:processType;type:text" json:"processType"`
	OrderAmount          float64   `gorm:"column:orderAmount;type:decimal(10,2)" json:"orderAmount"`
	RealOrderAmount      float64   `gorm:"column:realOrderAmount;type:decimal(10,2)" json:"realOrderAmount"`
	PayType              string    `gorm:"column:payType;type:varchar(32)" json:"payType"`
	PayModel             string    `gorm:"column:payModel;type:varchar(10)" json:"payModel"`
	UserIp               string    `gorm:"column:userIp;type:varchar(25)" json:"userIp"`
	UserTerminal         string    `gorm:"column:userTerminal;type:varchar(25)" json:"userTerminal"`
	Channel              string    `gorm:"column:channel;type:varchar(25)" json:"channel"`
	ChannelSetId         int64     `gorm:"column:channelSetId;type:varchar(25)" json:"channelSetId"`
	ChannelMerchantNo    string    `gorm:"column:channelMerchantNo;type:varchar(25)" json:"channelMerchantNo"`
	BackNoticeUrl        string    `gorm:"column:backNoticeUrl;type:varchar(250)" json:"backNoticeUrl"`
	TradeSummary         string    `gorm:"column:tradeSummary;type:varchar(250)" json:"tradeSummary"`
	AccountDate          string    `gorm:"column:accountDate;type:date;default:Null" json:"accountDate"`                                  // 财务日期
	TimeoutTime          string    `gorm:"column:timeoutTime;type:datetime;default:Null" json:"timeoutTime"`                              // 订单超时时间
	ServiceCharge        float64   `gorm:"column:serviceCharge;type:decimal(10,2) unsigned;not null;" json:"serviceCharge"`               // 手续费
	ChannelServiceCharge float64   `gorm:"column:channelServiceCharge;type:decimal(10,2) unsigned;not null;" json:"channelServiceCharge"` // 上游手续费
	BankCode             string    `gorm:"column:bankCode;type:varchar(50);not null;" json:"bankCode"`                                    // 支付银行
	CardHolderMobile     string    `gorm:"column:cardHolderMobile;type:varchar(500);not null;" json:"cardHolderMobile"`                   // 付款人手机号码
	CardHolderName       string    `gorm:"column:cardHolderName;type:varchar(50);not null;" json:"cardHolderName"`                        // 付款人姓名
	CardNum              string    `gorm:"column:cardNum;type:varchar(500);not null;" json:"cardNum"`                                     // 银行卡号
	CardType             string    `gorm:"column:cardType;type:enum('','DEBIT','CREDIT');not null;" json:"cardType"`                      // 银行卡类型(DEBIT=借记卡,CREDIT=信用卡)
	AgentFee             float64   `gorm:"column:agentFee;type:decimal(10,2);" json:"agentFee"`                                           //
	AgentName            string    `gorm:"column:agentName;type:varchar(50);" json:"agentName"`                                           //
	ChannelNoticeTime    string    `gorm:"column:channelNoticeTime;type:datetime;default:null;" json:"channelNoticeTime"`
	CallbackLimit        int64     `gorm:"column:callbackLimit;type:int(11) unsigned;not null;" json:"callbackLimit"` // 回调次数
	CallbackSuccess      bool      `gorm:"column:callbackSuccess;type:tinyint(1);not null;" json:"callbackSuccess"`   // 回调成功
	CreatedAt            time.Time `gorm:"column:created_at;type:datetime" json:"createTime"`
	UpdatedAt            time.Time `gorm:"column:updated_at;type:datetime" json:"-"`
}

type PayNotifyTask struct {
	ID              int64     `gorm:"column:id;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"id"` //
	Status          string    `gorm:"column:status;type:enum('Normal','Close');not null;" json:"status"`              // 状态(正常，关闭)
	RetryCount      int64     `gorm:"column:retryCount;type:int(11);not null;" json:"retryCount"`                     // 重试次数
	FailReason      string    `gorm:"column:failReason;type:varchar(150);not null;" json:"failReason"`                // 失败原因
	PlatformOrderNo string    `gorm:"column:platformOrderNo;type:varchar(50);not null;" json:"platformOrderNo"`       // 平台订单号
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                    //
	UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                    //
}

type PlatformSettlementOrder struct {
	OrderID              int64     `gorm:"column:orderId;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"orderId"`                                      //
	PlatformOrderNo      string    `gorm:"column:platformOrderNo;type:varchar(50);not null;" json:"platformOrderNo"`                                                      // 平台订单号
	MerchantID           int64     `gorm:"column:merchantId;type:int(11) unsigned;not null;" json:"merchantId"`                                                           // 商户ID
	MerchantNo           string    `gorm:"column:merchantNo;type:varchar(50);not null;" json:"merchantNo"`                                                                // 商户号
	MerchantOrderNo      string    `gorm:"column:merchantOrderNo;type:varchar(50);not null;" json:"merchantOrderNo"`                                                      // 商户订单号
	MerchantParam        string    `gorm:"column:merchantParam;type:varchar(150);not null;" json:"merchantParam"`                                                         // 回传参数
	MerchantReqTime      string    `gorm:"column:merchantReqTime;type:datetime;not null;" json:"merchantReqTime"`                                                         // 商户请求时间
	OrderAmount          float64   `gorm:"column:orderAmount;type:decimal(10,2) unsigned;not null;" json:"orderAmount"`                                                   // 订单金额
	RealOrderAmount      float64   `gorm:"column:realOrderAmount;type:decimal(10,2) unsigned;not null;" json:"realOrderAmount"`                                           // 真实订单金额
	ServiceCharge        float64   `gorm:"column:serviceCharge;type:decimal(10,2) unsigned;not null;" json:"serviceCharge"`                                               // 商户手续费
	ChannelServiceCharge float64   `gorm:"column:channelServiceCharge;type:decimal(10,2) unsigned;not null;" json:"channelServiceCharge"`                                 // 上游手续费
	ChannelSetID         int64     `gorm:"column:channelSetId;type:int(11) unsigned;not null;" json:"channelSetId"`                                                       // ID
	FailReason           string    `gorm:"column:failReason;type:varchar(150);not null;" json:"failReason"`                                                               // 失败原因
	Channel              string    `gorm:"column:channel;type:varchar(50);not null;" json:"channel"`                                                                      // 上游渠道
	ChannelMerchantID    int64     `gorm:"column:channelMerchantId;type:int(11) unsigned;not null;" json:"channelMerchantId"`                                             // 上游渠道商户号
	ChannelMerchantNo    string    `gorm:"column:channelMerchantNo;type:varchar(50);not null;" json:"channelMerchantNo"`                                                  // 上游渠道商户号
	ChannelOrderNo       string    `gorm:"column:channelOrderNo;type:varchar(50);not null;" json:"channelOrderNo"`                                                        // 上游订单号
	ChannelNoticeTime    string    `gorm:"column:channelNoticeTime;type:datetime;default:null;" json:"channelNoticeTime"`                                                 // 上游通知时间(处理时间)
	OrderReason          string    `gorm:"column:orderReason;type:varchar(55);not null;" json:"orderReason"`                                                              // 用途
	OrderStatus          string    `gorm:"column:orderStatus;type:enum('Transfered','Success','Fail','Exception');not null;" json:"orderStatus"`                          // 订单状态(Transfered=已划款,Success=划款成功,Fail=划款失败,Exception=异常)
	OrderType            string    `gorm:"column:orderType;type:enum('SettlementOrder');not null;default:'SettlementOrder'" json:"orderType"`                             // 支付类型
	PushChannelTime      *string   `gorm:"column:pushChannelTime;type:datetime;default:null;" json:"pushChannelTime"`                                                     // 向上游推送时间
	BackNoticeURL        string    `gorm:"column:backNoticeUrl;type:varchar(255);not null;" json:"backNoticeUrl"`                                                         // 异步通知地址
	BankLineNo           string    `gorm:"column:bankLineNo;type:varchar(50);not null;" json:"bankLineNo"`                                                                // 银行编号
	BankCode             string    `gorm:"column:bankCode;type:varchar(50);not null;" json:"bankCode"`                                                                    // 收款银行
	BankName             string    `gorm:"column:bankName;type:varchar(50);not null;" json:"bankName"`                                                                    // 开户行
	BankAccountName      string    `gorm:"column:bankAccountName;type:varchar(50);not null;" json:"bankAccountName"`                                                      // 收款人姓名
	BankAccountNo        string    `gorm:"column:bankAccountNo;type:varchar(50);not null;" json:"bankAccountNo"`                                                          // 收款卡号
	City                 string    `gorm:"column:city;type:varchar(50);not null;" json:"city"`                                                                            // 开户行所属市
	Province             string    `gorm:"column:province;type:varchar(50);not null;" json:"province"`                                                                    // 开户行所属省
	UserIP               string    `gorm:"column:userIp;type:varchar(255);not null;" json:"userIp"`                                                                       // 用户IP
	ApplyPerson          string    `gorm:"column:applyPerson;type:varchar(50);not null;" json:"applyPerson"`                                                              // 申请人
	ApplyIP              string    `gorm:"column:applyIp;type:varchar(50);not null;" json:"applyIp"`                                                                      // 申请人IP
	AccountDate          string    `gorm:"column:accountDate;type:date;" json:"accountDate"`                                                                              // 财务日期
	AuditPerson          string    `gorm:"column:auditPerson;type:varchar(50);not null;" json:"auditPerson"`                                                              // 审核人
	AuditIP              string    `gorm:"column:auditIp;type:varchar(50);not null;" json:"auditIp"`                                                                      // 审核人IP
	AuditTime            string    `gorm:"column:auditTime;type:datetime;default:null;" json:"auditTime"`                                                                 // 审核时间
	TradeSummary         string    `gorm:"column:tradeSummary;type:varchar(100);not null;" json:"tradeSummary"`                                                           // 交易摘要
	ProcessType          string    `gorm:"column:processType;type:enum('Expired','WaitPayment','Success','ServiceQuery','ManualOperation');not null;" json:"processType"` // 处理标识-暂定(Expired=已过期,WaitPayment=未支付,Success=成功)
	CallbackLimit        int64     `gorm:"column:callbackLimit;type:int(11) unsigned;not null;default:0;" json:"callbackLimit"`                                           // 回调次数
	CallbackSuccess      bool      `gorm:"column:callbackSuccess;type:tinyint(1);not null;default:0;" json:"callbackSuccess"`                                             // 回调成功
	CreatedAt            time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                                                                   //
	UpdatedAt            time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                                                                   //
	AgentFee             float64   `gorm:"column:agentFee;type:decimal(10,2) unsigned;not null;" json:"agentFee"`                                                         // 代理手续费
	AgentName            string    `gorm:"column:agentName;type:varchar(50);default:'';" json:"agentName"`                                                                // 代理账号
	IsLock               int64     `gorm:"column:isLock;type:int(11) unsigned;not null;default:0;" json:"isLock"`                                                         // 代理账号
	LockUser             string    `gorm:"column:lockUser;type:varchar(50);default:'';" json:"lockUser"`                                                                  // 代理账号
}

type SettlementFetchTask struct {
	ID              int64     `gorm:"column:id;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"id"` //
	Status          string    `gorm:"column:status;type:enum('Normal','Close');not null;" json:"status"`              // 状态(正常，关闭)
	RetryCount      int64     `gorm:"column:retryCount;type:int(11);not null;" json:"retryCount"`                     // 重试次数
	FailReason      string    `gorm:"column:failReason;type:varchar(150);not null;" json:"failReason"`                // 失败原因
	PlatformOrderNo string    `gorm:"column:platformOrderNo;type:varchar(50);not null;" json:"platformOrderNo"`       // 平台订单号
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                    //
	UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                    //
}

type SettlementNotifyTask struct {
	ID              int64     `gorm:"column:id;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"id"` //
	Status          string    `gorm:"column:status;type:enum('Normal','Close');not null;" json:"status"`              // 状态(正常，关闭)
	RetryCount      int64     `gorm:"column:retryCount;type:int(11);not null;" json:"retryCount"`                     // 重试次数
	FailReason      string    `gorm:"column:failReason;type:varchar(150);not null;" json:"failReason"`                // 失败原因
	PlatformOrderNo string    `gorm:"column:platformOrderNo;type:varchar(50);not null;" json:"platformOrderNo"`       // 平台订单号
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                    //
	UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                    //
}

type Merchant struct {
	MerchantID             int64     `gorm:"column:merchantId;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"merchantId"`           //
	MerchantNo             string    `gorm:"column:merchantNo;type:varchar(50);not null;" json:"merchantNo"`                                           // 商户号
	FullName               string    `gorm:"column:fullName;type:varchar(50);not null;" json:"fullName"`                                               // 商户全称
	ShortName              string    `gorm:"column:shortName;type:varchar(50);not null;" json:"shortName"`                                             // 商户简称
	PayerName              string    `gorm:"column:payerName;type:varchar(100);not null;" json:"payerName"`                                            // 商户付款人信息
	Status                 string    `gorm:"column:status;type:enum('Normal','Close');not null;" json:"status"`                                        // 状态(正常，关闭)
	PlatformNo             string    `gorm:"column:platformNo;type:varchar(50);not null;" json:"platformNo"`                                           // 平台号码
	D0SettlementRate       float64   `gorm:"column:D0SettlementRate;type:decimal(4,2) unsigned;not null;" json:"D0SettlementRate,string"`              // D0垫资比例
	SettlementTime         int64     `gorm:"column:settlementTime;type:int(11);not null;" json:"settlementTime"`                                       // 结算时间
	SettlementType         string    `gorm:"column:settlementType;type:enum('D0','D1','T0','T1');not null;" json:"settlementType"`                     // 商户结算方式
	OneSettlementMaxAmount float64   `gorm:"column:oneSettlementMaxAmount;type:decimal(10,2) unsigned;not null;" json:"oneSettlementMaxAmount,string"` // 单卡单日最大结算金额
	OpenPay                bool      `gorm:"column:openPay;type:tinyint(1);not null;" json:"openPay"`                                                  // 支付开关
	OpenQuery              bool      `gorm:"column:openQuery;type:tinyint(1);not null;" json:"openQuery"`                                              // 查询开关
	OpenSettlement         bool      `gorm:"column:openSettlement;type:tinyint(1);not null;" json:"openSettlement"`                                    // 结算开关
	OpenAliSettlement      bool      `gorm:"column:openAliSettlement;type:tinyint(1);not null;" json:"openAliSettlement"`                              // 支付宝结算开关
	OpenBackNotice         bool      `gorm:"column:openBackNotice;type:tinyint(1);not null;" json:"openBackNotice"`                                    // 后台通知开关
	OpenCheckAccount       bool      `gorm:"column:openCheckAccount;type:tinyint(1);not null;" json:"openCheckAccount"`                                // 对账开关
	OpenCheckDomain        bool      `gorm:"column:openCheckDomain;type:tinyint(1);not null;" json:"openCheckDomain"`                                  // 域名验证开关
	OpenFrontNotice        bool      `gorm:"column:openFrontNotice;type:tinyint(1);not null;" json:"openFrontNotice"`                                  // 前台通知开关
	OpenRepayNotice        bool      `gorm:"column:openRepayNotice;type:tinyint(1);not null;" json:"openRepayNotice"`                                  //
	SignKey                string    `gorm:"column:signKey;type:varchar(500);not null;" json:"signKey"`                                                // 加密key
	Domain                 string    `gorm:"column:domain;type:varchar(255);not null;" json:"domain"`                                                  // 域名
	Description            string    `gorm:"column:description;type:varchar(255);not null;" json:"description"`                                        // 描述
	BackNoticeMaxNum       int64     `gorm:"column:backNoticeMaxNum;type:int(11);not null;" json:"backNoticeMaxNum"`                                   // 最大回调次数
	PlatformType           string    `gorm:"column:platformType;type:enum('Normal','Proxy');not null;" json:"platformType"`                            // 平台(一般、代理)
	IpWhite                string    `gorm:"column:ipWhite;type:varchar(512);not null;" json:"ipWhite"`                                                // 商户IP白名单
	CreatedAt              time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                                              //
	UpdatedAt              time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                                              //
	LoginIPWhite           string    `gorm:"column:loginIpWhite;type:varchar(512);not null;" json:"loginIpWhite"`                                      // 登录白名单
	OpenManualSettlement   bool      `gorm:"column:openManualSettlement;type:tinyint(1);not null;" json:"openManualSettlement"`                        // 手动代付开关
	OpenAutoSettlement     bool      `gorm:"column:openAutoSettlement;type:tinyint(1);not null;" json:"openAutoSettlement"`                            // 手动代付开关
}

type MerchantJoin struct {
	Merchant
	SettlementAmount float64 `gorm:"column:settlementAmount;type:decimal(10,2) unsigned;not null;" json:"settlementAmount" form:"settlementAmount"` // 商户余额
}

type MerchantAccount struct {
	AccountID           int64     `gorm:"column:accountId;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"accountId"` //
	LoginName           string    `gorm:"column:loginName;type:varchar(50);not null;" json:"loginName"`                                 // 登录名称
	LoginPwd            string    `gorm:"column:loginPwd;type:varchar(50);not null;" json:"loginPwd"`                                   // 密码
	SecurePwd           string    `gorm:"column:securePwd;type:varchar(255);not null;" json:"securePwd"`                                // 支付密码
	UserName            string    `gorm:"column:userName;type:varchar(50);not null;" json:"userName"`                                   // 商户号
	LoginFailNum        int64     `gorm:"column:loginFailNum;type:int(11);not null;" json:"loginFailNum"`                               // 登录失败次数
	LoginPwdAlterTime   string    `gorm:"column:loginPwdAlterTime;type:datetime;default:null;" json:"loginPwdAlterTime"`                // 密码修改时间
	MerchantID          int64     `gorm:"column:merchantId;type:int(11);not null;" json:"merchantId"`                                   // 商户ID
	MerchantNo          string    `gorm:"column:merchantNo;type:varchar(50);not null;" json:"merchantNo"`                               // 商户号
	PlatformNo          string    `gorm:"column:platformNo;type:varchar(50);not null;" json:"platformNo"`                               // 平台号
	PlatformType        string    `gorm:"column:platformType;type:enum('Normal','Proxy');not null;" json:"platformType"`                // 平台(一般、代理)
	Status              string    `gorm:"column:status;type:enum('Normal','Exception','Close');not null;" json:"status"`                // 状态
	UserLevel           string    `gorm:"column:userLevel;type:enum('MerchantManager','PlatformManager');not null;" json:"userLevel"`   // 用户级别
	LatestLoginTime     string    `gorm:"column:latestLoginTime;type:datetime;default:null" json:"latestLoginTime"`                     // 最后登录时间
	GoogleAuthSecretKey string    `gorm:"column:googleAuthSecretKey;type:varchar(500);not null;" json:"googleAuthSecretKey"`            // 谷歌auth密钥
	CreatedAt           time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                                  //
	UpdatedAt           time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                                  //
}

type MerchantRate struct {
	RateID           int64     `gorm:"column:rateId;primary_key;auto_increment;type:int(11) unsigned;not null;"  json:"rateId" form:"rateId"`         //
	BankCode         string    `gorm:"column:bankCode;type:varchar(50);not null;"  json:"bankCode" form:"bankCode"`                                   // 银行
	CardType         string    `gorm:"column:cardType;type:enum('','DEBIT','CREDIT');not null;" json:"cardType" form:"cardType"`                      // 银行卡类型(DEBIT=借记卡,CREDIT=信用卡)
	BeginTime        string    `gorm:"column:beginTime;type:date;default:null;" json:"beginTime" form:"beginTime"`                                    // 生效时间
	EndTime          string    `gorm:"column:endTime;type:date;default:null;" json:"endTime" form:"endTime"`                                          // 失效时间
	MaxServiceCharge float64   `gorm:"column:maxServiceCharge;type:decimal(10,2) unsigned;not null;" json:"maxServiceCharge" form:"maxServiceCharge"` // 最大手续费
	MinServiceCharge float64   `gorm:"column:minServiceCharge;type:decimal(10,2) unsigned;not null;" json:"minServiceCharge" form:"minServiceCharge"`
	MinAmount        float64   `gorm:"column:minAmount;type:decimal(10,2) unsigned;not null;" json:"minAmount" form:"minAmount"` // 最小有效金额
	MaxAmount        float64   `gorm:"column:maxAmount;type:decimal(10,2) unsigned;not null;" json:"maxAmount" form:"maxAmount"` // 最大有效金额
	MerchantID       int64     `gorm:"column:merchantId;type:int(11) unsigned;not null;" json:"merchantId" form:"merchantId"`    // 商户ID
	MerchantNo       string    `gorm:"column:merchantNo;type:varchar(50);not null;" json:"merchantNo" form:"merchantNo"`         // 商户号
	PayType          string    `gorm:"column:payType;type:varchar(50);not null;" json:"payType" form:"payType"`
	PayTypeDesc      string    `gorm:"_" json:"payTypeDesc" form:"payTypeDesc"`
	ProductType      string    `gorm:"column:productType;type:enum('Pay','Settlement','Recharge');not null;" json:"productType" form:"productType"` // 产品类型
	ProductTypeDesc  string    `gorm:"_" json:"productTypeDesc" form:"productTypeDesc"`                                                             // 产品类型
	Rate             float64   `gorm:"column:rate;type:float unsigned;not null;" json:"rate" form:"rate"`                                           // 费率值
	Fixed            float64   `gorm:"column:fixed;type:float unsigned;not null;" json:"fixed" form:"fixed"`                                        //
	RateType         string    `gorm:"column:rateType;type:enum('Rate','FixedValue','Mixed');not null;" json:"rateType" form:"rateType"`            // 费率类型
	RateTypeDesc     string    `gorm:"_" json:"rateTypeDesc" form:"rateTypeDesc"`                                                                   // 费率类型
	Status           string    `gorm:"column:status;type:enum('Normal','Close');not null;" json:"status" form:"status"`                             // 状态
	StatusDesc       string    `gorm:"_" json:"statusDesc" form:"statusDesc"`                                                                       // 状态
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at" form:"created_at"`                               //
	UpdatedAt        time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at" form:"updated_at"`                               //
}

// MerchantAmount -> merchant_amount 商户金额统计表
type MerchantAmount struct {
	Id               int       `gorm:"column:id;primary_key;auto_increment;type:int(11) unsigned;not null;"` //
	MerchantID       int64     `gorm:"column:merchantId;type:int(11);not null;"`                             // 商户ID
	MerchantNo       string    `gorm:"column:merchantNo;type:varchar(50);not null;"`                         // 商户号
	SettledAmount    float64   `gorm:"column:settledAmount;type:decimal(10,2) unsigned;not null;"`           // 已结算金额
	SettlementAmount float64   `gorm:"column:settlementAmount;type:decimal(10,2) unsigned;not null;"`        // 未结算金额
	FreezeAmount     float64   `gorm:"column:freezeAmount;type:decimal(10,2) unsigned;not null;"`            // 冻结金额
	ModTime          time.Time `gorm:"column:modTime;type:datetime;not null;"`                               //
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime;not null;"`                            //
	UpdatedAt        time.Time `gorm:"column:updated_at;type:datetime;not null;"`                            //
}

type MerchantDailyStats struct {
	DailyID                      int64     `gorm:"column:dailyId;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"dailyId"`                      //
	MerchantID                   int64     `gorm:"column:merchantId;type:int(11) unsigned;not null;" json:"merchantId"`                                           // 商户ID
	MerchantNo                   string    `gorm:"column:merchantNo;type:varchar(50);not null;" json:"merchantNo"`                                                // 商户号
	PayCount                     int64     `gorm:"column:payCount;type:int(11) unsigned;not null;" json:"payCount"`                                               // 今日支付笔数
	PayAmount                    float64   `gorm:"column:payAmount;type:decimal(10,2) unsigned;not null;" json:"payAmount"`                                       // 今日支付金额
	PayServiceFees               float64   `gorm:"column:payServiceFees;type:decimal(10,2) unsigned;not null;" json:"payServiceFees"`                             // 今日支付手续费
	PayChannelServiceFees        float64   `gorm:"column:payChannelServiceFees;type:decimal(10,2) unsigned;not null;" json:"payChannelServiceFees"`               // 今日上游支付手续费
	SettlementCount              int64     `gorm:"column:settlementCount;type:int(11) unsigned;not null;" json:"settlementCount"`                                 // 今日代付比数
	SettlementAmount             float64   `gorm:"column:settlementAmount;type:decimal(10,2) unsigned;not null;" json:"settlementAmount"`                         // 今日代付金额
	SettlementServiceFees        float64   `gorm:"column:settlementServiceFees;type:decimal(10,2) unsigned;not null;" json:"settlementServiceFees"`               // 今日代付手续费
	SettlementChannelServiceFees float64   `gorm:"column:settlementChannelServiceFees;type:decimal(10,2) unsigned;not null;" json:"settlementChannelServiceFees"` // 今日上游代付手续费
	ChargeCount                  int64     `gorm:"column:chargeCount;type:int(11) unsigned;not null;" json:"chargeCount"`                                         // 今日充值笔数
	ChargeAmount                 float64   `gorm:"column:chargeAmount;type:decimal(10,2) unsigned;not null;" json:"chargeAmount"`                                 // 今日充值金额
	ChargeServiceFees            float64   `gorm:"column:chargeServiceFees;type:decimal(10,2) unsigned;not null;" json:"chargeServiceFees"`                       // 今日充值手续费
	ChargeChannelServiceFees     float64   `gorm:"column:chargeChannelServiceFees;type:decimal(10,2) unsigned;not null;" json:"chargeChannelServiceFees"`         // 今日上游充值手续费
	AccountDate                  string    `gorm:"column:accountDate;type:date;not null;" json:"accountDate"`                                                     // 财务日期
	CreatedAt                    time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                                                   //
	UpdatedAt                    time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                                                   //
	ChannelMerchantID            int64     `gorm:"column:channelMerchantId;type:int(11) unsigned;not null;" json:"channelMerchantId"`                             // 渠道ID
	ChannelMerchantNo            int64     `gorm:"column:channelMerchantNo;type:int(11) unsigned;not null;" json:"channelMerchantNo"`                             // 渠道号
	AgentPayFees                 float64   `gorm:"column:agentPayFees;type:decimal(10,2) unsigned;not null;" json:"agentPayFees"`                                 // 代理支付手续费
	AgentsettlementFees          float64   `gorm:"column:agentsettlementFees;type:decimal(10,2) unsigned;not null;" json:"agentsettlementFees"`                   // 代理代付手续费
	AgentchargeFees              float64   `gorm:"column:agentchargeFees;type:decimal(10,2) unsigned;not null;" json:"agentchargeFees"`                           // 代理充值手续费
}

// BalanceAdjustment -> balance_adjustment 商户余额调整
type BalanceAdjustment struct {
	AdjustmentID      int64     `gorm:"column:adjustmentId;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"adjustmentId"`                         //
	Amount            float64   `gorm:"column:amount;type:decimal(10,2) unsigned;not null;" json:"amount"`                                                          // 金额
	ApplyPerson       string    `gorm:"column:applyPerson;type:varchar(50);not null;" json:"applyPerson"`                                                           // 申请人
	AuditPerson       string    `gorm:"column:auditPerson;type:varchar(50);" json:"auditPerson"`                                                                    // 审核人
	AuditTime         string    `gorm:"column:auditTime;type:datetime;" json:"auditTime"`                                                                           // 审核时间
	BankrollDirection string    `gorm:"column:bankrollDirection;type:enum('Restore','Retrieve','Freeze','Unfreeze','Recharge');not null;" json:"bankrollDirection"` // 资金方向(Restore=返还, Retrieve=追收, Freeze=冻结， Unfreeze=解冻)
	BankrollType      string    `gorm:"column:bankrollType;type:enum('AccountBalance','ServiceCharge');not null;" json:"bankrollType"`                              // 资金类型(AccountBalance=账户资金, ServiceCharge=手续费)
	Status            string    `gorm:"column:status;type:enum('Success','Fail','Freeze','Unaudit');not null;" json:"status"`                                       // 审核状态(通过，不通过，待解冻, 待审核)
	Summary           string    `gorm:"column:summary;type:varchar(50);not null;" json:"summary"`                                                                   // 摘要
	MerchantID        int64     `gorm:"column:merchantId;type:int(11) unsigned;not null;" json:"merchantId"`                                                        // 商户ID
	MerchantNo        string    `gorm:"column:merchantNo;type:varchar(50);not null;" json:"merchantNo"`                                                             // 商户号
	PlatformOrderNo   string    `gorm:"column:platformOrderNo;type:varchar(255);not null;" json:"platformOrderNo"`                                                  // 平台订单号
	CreatedAt         time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                                                                //
	UpdatedAt         time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                                                                //
}

// AmountPay -> amount_pay 支付订单金额统计
type AmountPay struct {
	Id                   int       `gorm:"column:id;primary_key;auto_increment;type:int(11) unsigned;not null;"`                                                                                                                                                                                 //
	Amount               float64   `gorm:"column:amount;type:decimal(10,2) unsigned;not null;"`                                                                                                                                                                                                  // 金额
	MerchantID           int64     `gorm:"column:merchantId;type:int(11) unsigned;not null;"`                                                                                                                                                                                                    // 商户ID
	MerchantNo           string    `gorm:"column:merchantNo;type:varchar(50);not null;"`                                                                                                                                                                                                         // 商户号
	ChannelMerchantID    int64     `gorm:"column:channelMerchantId;type:int(11) unsigned;not null;"`                                                                                                                                                                                             // 商户ID
	ChannelMerchantNo    string    `gorm:"column:channelMerchantNo;type:varchar(50);not null;"`                                                                                                                                                                                                  // 商户号
	PayType              string    `gorm:"column:payType;type:enum('EBank','Quick','OfflineWechatQR','OfflineAlipayQR','OnlineWechatQR','OnlineAlipayQR','OnlineWechatH5','OnlineAlipayH5','UnionPayQR','D0Settlement','UnionPay','OnlineAlipayOriginalH5','OnlineWechatOriginalH5');not null;"` // 支付方式(EBank=网银,Quick=快捷,OfflineWechatQR=线下微信扫码,OfflineAlipayQR=线下支付宝扫码, OnlineWechatQR=线上微信扫码,OnlineAlipayQR=线上支付宝扫码,OnlineWechatH5=线上微信H5,OnlineAlipayH5=线上支付宝H5, OnlineAlipayOriginalH5=线上支付宝原生H5,OnlineWechatOriginalH5=线上微信原生H5,UnionPay=银联快捷/云闪付,UnionPayQR=银联扫码,D0Settlement=D0结算)
	AccountDate          string    `gorm:"column:accountDate;type:date;not null;"`                                                                                                                                                                                                               // 财务日期
	Balance              float64   `gorm:"column:balance;type:decimal(10,2) unsigned;not null;"`                                                                                                                                                                                                 // 余额
	ServiceCharge        float64   `gorm:"column:serviceCharge;type:decimal(10,2) unsigned;not null;"`                                                                                                                                                                                           // 手续费
	ChannelServiceCharge float64   `gorm:"column:channelServiceCharge;type:decimal(10,2) unsigned;not null;"`                                                                                                                                                                                    // 上游手续费
	CreatedAt            time.Time `gorm:"column:created_at;type:datetime;not null;"`                                                                                                                                                                                                            //
	UpdatedAt            time.Time `gorm:"column:updated_at;type:datetime;not null;"`                                                                                                                                                                                                            //
}

// AmountSettlement -> amount_settlement 代付订单金额统计
type AmountSettlement struct {
	Id                   int64     `gorm:"column:id;primary_key;auto_increment;type:int(11) unsigned;not null;"` //
	Amount               float64   `gorm:"column:amount;type:decimal(10,2) unsigned;not null;"`                  // 金额
	MerchantID           int64     `gorm:"column:merchantId;type:int(11) unsigned;not null;"`                    // 商户ID
	MerchantNo           string    `gorm:"column:merchantNo;type:varchar(50);not null;"`                         // 商户号
	ChannelMerchantID    int64     `gorm:"column:channelMerchantId;type:int(11) unsigned;not null;"`             // 商户ID
	ChannelMerchantNo    string    `gorm:"column:channelMerchantNo;type:varchar(50);not null;"`                  // 商户号
	AccountDate          string    `gorm:"column:accountDate;type:date;not null;"`                               // 财务日期
	TransferTimes        int64     `gorm:"column:transferTimes;type:int(11);not null;"`                          // 代付次数
	ServiceCharge        float64   `gorm:"column:serviceCharge;type:decimal(10,2) unsigned;not null;"`           // 手续费
	ChannelServiceCharge float64   `gorm:"column:channelServiceCharge;type:decimal(10,2) unsigned;not null;"`    // 上游手续费
	CreatedAt            time.Time `gorm:"column:created_at;type:datetime;not null;"`                            //
	UpdatedAt            time.Time `gorm:"column:updated_at;type:datetime;not null;"`                            //
}

// Finance -> finance 财务明细
type Finance struct {
	Id              int64     `gorm:"column:id;primary_key;auto_increment;type:int(11) unsigned;not null;" json:"id"`                                                          //
	Amount          float64   `gorm:"column:amount;type:decimal(10,2) unsigned;not null;" json:"amount"`                                                                       // 金额
	Balance         float64   `gorm:"column:balance;type:decimal(10,2) unsigned;not null;" json:"balance"`                                                                     // 余额
	MerchantID      int64     `gorm:"column:merchantId;type:int(11) unsigned;not null;" json:"merchantId"`                                                                     // 商户号
	MerchantNo      string    `gorm:"column:merchantNo;type:varchar(50);not null;" json:"merchantNo"`                                                                          // 商户号
	Summary         string    `gorm:"column:summary;type:varchar(500);not null;" json:"summary"`                                                                               // 描述
	SourceDesc      string    `gorm:"column:sourceDesc;type:varchar(50);not null;" json:"sourceDesc"`                                                                          // 来源描述
	SourceID        int64     `gorm:"column:sourceId;type:int(11) unsigned;not null;" json:"sourceId"`                                                                         // ID
	FinanceType     string    `gorm:"column:financeType;type:enum('PayIn','PayOut');not null;" json:"financeType"`                                                             // 收支类型(PayIn=收入,PayOut=支出)
	PlatformOrderNo string    `gorm:"column:platformOrderNo;type:varchar(50);not null;" json:"platformOrderNo"`                                                                // 平台订单号
	AccountDate     string    `gorm:"column:accountDate;type:date;not null;" json:"accountDate"`                                                                               // 账务日期
	AccountType     string    `gorm:"column:accountType;type:enum('SettledAccount','SettlementAccount','AdvanceAccount','ServiceChargeAccount');not null;" json:"accountType"` // 账户类型(SettledAccount=已结算账户,SettlementAccount=未结算账户,AdvanceAccount=垫资账户,ServiceChargeAccount=手续费账户)
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at"`                                                                             //
	UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at"`                                                                             //
	MerchantOrderNo string    `gorm:"column:merchantOrderNo;type:varchar(50);" json:"merchantOrderNo"`                                                                         // 商户订单号
	OperateSource   string    `gorm:"column:operateSource;type:enum('ports','merchant','admin');not null;" json:"operateSource"`                                               // 操作来源(ports:接口,merchant:商户后台,admin:管理后台)
}

type MerchantChannel struct {
	SetID              int64     `json:"setId" gorm:"column:setId;primary_key;auto_increment;type:int(11) unsigned;not null;"`
	MerchantID         int64     `json:"merchantId" gorm:"column:merchantId;type:int(11) unsigned;not null;"`
	MerchantNo         string    `json:"merchantNo" gorm:"column:merchantNo;type:varchar(50);not null;"`
	BankCode           string    `json:"bankCode" gorm:"column:bankCode;type:varchar(50);"`
	CardType           string    `json:"cardType" gorm:"column:cardType;type:enum('','DEBIT','CREDIT');not null;"`
	Channel            string    `json:"channel" gorm:"column:channel;type:varchar(50);not null;"`
	ChannelMerchantID  int64     `json:"channelMerchantId" gorm:"column:channelMerchantId;type:int(11);not null;"`
	ChannelMerchantNo  string    `json:"channelMerchantNo" gorm:"column:channelMerchantNo;type:varchar(50);not null;"`
	PayChannelStatus   string    `json:"payChannelStatus" gorm:"column:payChannelStatus;type:enum('Normal','Close');not null;"`
	PayType            string    `json:"payType" gorm:"column:payType;type:enum('EBank','Quick','OfflineWechatQR','OfflineAlipayQR','OnlineWechatQR','OnlineAlipayQR','OnlineWechatH5','OnlineAlipayH5','UnionPayQR');not null;"`
	OpenTimeLimit      int64     `json:"openTimeLimit" gorm:"column:openTimeLimit;type:tinyint(1);not null;"`
	BeginTime          int64     `json:"beginTime" gorm:"column:beginTime;type:int(11) unsigned;not null;"`
	EndTime            int64     `json:"endTime" gorm:"column:endTime;type:int(11) unsigned;not null;"`
	OpenOneAmountLimit int64     `json:"openOneAmountLimit" gorm:"column:openOneAmountLimit;type:tinyint(1);not null;"`
	OneMaxAmount       float64   `json:"oneMaxAmount" gorm:"column:oneMaxAmount;type:decimal(10,2) unsigned;not null;"`
	OneMinAmount       float64   `json:"oneMinAmount" gorm:"column:oneMinAmount;type:decimal(10,2) unsigned;not null;"`
	OpenDayAmountLimit int64     `json:"openDayAmountLimit" gorm:"column:openDayAmountLimit;type:tinyint(1);not null;"`
	DayAmountLimit     float64   `json:"dayAmountLimit" gorm:"column:dayAmountLimit;type:decimal(10,2) unsigned;not null;"`
	OpenDayNumLimit    int64     `json:"openDayNumLimit" gorm:"column:openDayNumLimit;type:tinyint(1);not null;"`
	DayNumLimit        int64     `json:"dayNumLimit" gorm:"column:dayNumLimit;type:int(11) unsigned;not null;"`
	Status             string    `json:"status" gorm:"column:status;type:enum('Normal','Close');not null;"`
	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at;type:datetime;not null;"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"column:updated_at;type:datetime;not null;"`
}

type MerchantChannelSettlement struct {
	SetID                   int64     `json:"setId" gorm:"column:setId;primary_key;auto_increment;type:int(11) unsigned;not null;"`                       //
	MerchantID              int64     `json:"merchantId" gorm:"column:merchantId;type:int(11);not null;"`                                                 // 下游商户ID
	MerchantNo              string    `json:"merchantNo" gorm:"column:merchantNo;type:varchar(50);not null;"`                                             // 下游商户号
	Channel                 string    `json:"channel" gorm:"column:channel;type:varchar(50);not null;"`                                                   // 渠道名称
	ChannelMerchantID       int64     `json:"channelMerchantId" gorm:"column:channelMerchantId;type:int(11);not null;"`                                   // 渠道商户
	ChannelMerchantNo       string    `json:"channelMerchantNo" gorm:"column:channelMerchantNo;type:varchar(50);not null;"`                               // 渠道商户号
	SettlementChannelStatus string    `json:"settlementChannelStatus" gorm:"column:settlementChannelStatus;type:enum('Normal','Close');not null;"`        // 代付渠道状态
	SettlementAccountType   string    `json:"settlementAccountType" gorm:"column:settlementAccountType;type:enum('UsableAccount','T1Account');not null;"` // 代付账户类型
	AccountBalance          float64   `json:"accountBalance" gorm:"column:accountBalance;type:decimal(10,2) unsigned;not null;"`                          // 账户余额
	AccountReservedBalance  float64   `json:"accountReservedBalance" gorm:"column:accountReservedBalance;type:decimal(10,2) unsigned;not null;"`          // 账户最少保留金额
	OpenTimeLimit           int64     `json:"openTimeLimit" gorm:"column:openTimeLimit;type:tinyint(1);not null;"`                                        // 是否开启控制交易时间
	BeginTime               int64     `json:"beginTime" gorm:"column:beginTime;type:int(11);not null;"`                                                   // 开始时间(00:00格式转整形)
	EndTime                 int64     `json:"endTime" gorm:"column:endTime;type:int(11);not null;"`                                                       // 结束时间(00:00格式转整形)
	OpenOneAmountLimit      int64     `json:"openOneAmountLimit" gorm:"column:openOneAmountLimit;type:tinyint(1);not null;"`                              // 是否开启控制单笔金额控制
	OneMaxAmount            float64   `json:"oneMaxAmount" gorm:"column:oneMaxAmount;type:decimal(10,2) unsigned;not null;"`                              // 单笔最大金额
	OneMinAmount            float64   `json:"oneMinAmount" gorm:"column:oneMinAmount;type:decimal(10,2) unsigned;not null;"`                              // 单笔最小金额
	OpenDayAmountLimit      int64     `json:"openDayAmountLimit" gorm:"column:openDayAmountLimit;type:tinyint(1);not null;"`                              // 是否开启单日累计金额控制
	DayAmountLimit          float64   `json:"dayAmountLimit" gorm:"column:dayAmountLimit;type:decimal(10,2) unsigned;not null;"`                          // 累计金额限制
	OpenDayNumLimit         int64     `json:"openDayNumLimit" gorm:"column:openDayNumLimit;type:tinyint(1);not null;"`                                    // 是否开启单日累计金额控制
	DayNumLimit             int64     `json:"dayNumLimit" gorm:"column:dayNumLimit;type:int(11) unsigned;not null;"`                                      // 累计次数限制
	Status                  string    `json:"status" gorm:"column:status;type:enum('Normal','Close');not null;"`                                          // 配置状态
	CreatedAt               time.Time `json:"created_at" gorm:"column:created_at;type:datetime;not null;"`                                                //
	UpdatedAt               time.Time `json:"updated_at" gorm:"column:updated_at;type:datetime;not null;"`                                                //
}

type ChannelMerchant struct {
	ChannelMerchantID      int64     `gorm:"column:channelMerchantId;primary_key;auto_increment;type:int(11) unsigned;not null;"  json:"merchantId" form:"merchantId"`
	ChannelMerchantNo      string    `gorm:"column:channelMerchantNo;type:varchar(50);not null;"  json:"merchantNo" form:"merchantNo"`
	Channel                string    `gorm:"column:channel;type:varchar(50);not null;"  json:"channel" form:"channel"`
	ChannelAccount         string    `gorm:"column:channelAccount;type:varchar(50);not null;"  json:"channelAccount" form:"channelAccount"`
	Status                 string    `gorm:"column:status;type:enum('Normal','Close','Exception','Deleted');not null;" json:"status" form:"status"`
	StatusDesc             string    `gorm:"_" json:"statusDesc" form:"statusDesc"`
	PlatformNo             string    `gorm:"column:platformNo;type:varchar(50);not null;" json:"platformNo" form:"platformNo"`
	SettlementTime         int64     `gorm:"column:settlementTime;type:int(11) unsigned;not null;" json:"settlementTime" form:"settlementTime"`
	D0SettlementRate       string    `gorm:"column:D0SettlementRate;type:float unsigned;not null;" json:"D0SettlementRate" form:"D0SettlementRate"`
	OneSettlementMaxAmount float64   `gorm:"column:oneSettlementMaxAmount;type:decimal(10,2) unsigned;not null;" json:"oneSettlementMaxAmount" form:"oneSettlementMaxAmount"`
	OpenPay                bool      `gorm:"column:openPay;type:tinyint(1);not null;" json:"openPay" form:"openPay"`
	OpenQuery              bool      `gorm:"column:openQuery;type:tinyint(1);not null;" json:"openQuery" form:"openQuery"`
	OpenSettlement         bool      `gorm:"column:openSettlement;type:tinyint(1);not null;" json:"openSettlement" form:"openSettlement"`
	DelegateDomain         string    `gorm:"column:delegateDomain;type:varchar(255);not null;" json:"delegateDomain" form:"delegateDomain"`
	Param                  string    `gorm:"column:param;type:longtext;not null;" json:"param" form:"param"`
	CreatedAt              time.Time `gorm:"column:created_at;type:datetime;not null;" json:"created_at" form:"created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at;type:datetime;not null;" json:"updated_at" form:"updated_at"`
	Config                 string    `gorm:"column:config;type:text;" json:"config" form:"config"`
	Balance                float64   `gorm:"column:balance;type:decimal(10,2) unsigned;" json:"balance" form:"balance"`
}

// ChannelMerchantRate -> channel_merchant_rate 渠道商户费率表
type ChannelMerchantRate struct {
	RateID            int64     `json:"rateId" gorm:"column:rateId;primary_key;auto_increment;type:int(11) unsigned;not null;"`   //
	BankCode          string    `json:"bankCode" gorm:"column:bankCode;type:varchar(50);not null;"`                               // 银行
	CardType          string    `json:"cardType" gorm:"column:cardType;type:enum('','DEBIT','CREDIT');not null;"`                 // 银行卡类型(DEBIT=借记卡,CREDIT=信用卡)
	BeginTime         string    `json:"beginTime" gorm:"column:beginTime;type:datetime;default:null"`                             // 生效时间
	EndTime           string    `json:"endTime" gorm:"column:endTime;type:datetime;default:null"`                                 // 失效时间
	MinAmount         float64   `json:"minAmount" gorm:"column:minAmount;type:decimal(10,2) unsigned;not null;"`                  // 最小有效金额
	MaxAmount         float64   `json:"maxAmount" gorm:"column:maxAmount;type:decimal(10,2) unsigned;not null;"`                  // 最大有效金额
	MaxServiceCharge  float64   `json:"maxServiceCharge" gorm:"column:maxServiceCharge;type:decimal(10,2) unsigned;not null;"`    // 最大手续费
	MinServiceCharge  float64   `json:"minServiceCharge" gorm:"column:minServiceCharge;type:decimal(10,2) unsigned;not null;"`    // 最小手续费
	ChannelMerchantID int64     `json:"channelMerchantId" gorm:"column:channelMerchantId;type:int(11) unsigned;not null;"`        // 商户ID
	ChannelMerchantNo string    `json:"channelMerchantNo" gorm:"column:channelMerchantNo;type:varchar(50);not null;"`             // 商户号
	PayType           string    `json:"payType" gorm:"column:payType;type:varchar(100);not null;"`                                // 支付方式(EBank=网银,Quick=快捷,OfflineWechatQR=线下微信扫码,OfflineAlipayQR=线下支付宝扫码,\n            OnlineWechatQR=线上微信扫码,OnlineAlipayQR=线上支付宝扫码,OnlineWechatH5=线上微信H5,OnlineAlipayH5=线上支付宝H5,\n            UnionPayQR=银联扫码,D0Settlement=D0结算,EnterpriseEBank=企业网银,PersonalEBank=个人网银)
	ProductType       string    `json:"productType" gorm:"column:productType;type:enum('Pay','Settlement','Recharge');not null;"` // 产品类型
	Rate              float64   `json:"rate" gorm:"column:rate;type:float unsigned;not null;"`                                    // 费率值
	Fixed             float64   `json:"fixed" gorm:"column:fixed;type:float unsigned;not null;"`                                  // 固定收费，基本收费
	RateType          string    `json:"rateType" gorm:"column:rateType;type:enum('Rate','FixedValue','Mixed');not null;"`         // 费率类型
	Status            string    `json:"status" gorm:"column:status;type:enum('Normal','Close');not null;"`                        // 状态
	Channel           string    `json:"channel" gorm:"column:channel;type:varchar(50);not null;"`                                 // 渠道名称
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at;type:datetime;not null;"`                              //
	UpdatedAt         time.Time `json:"updated_at" gorm:"column:updated_at;type:datetime;not null;"`                              //
}

// ChannelPayConfig -> channel_pay_config 上游支付渠道配置表
type ChannelPayConfig struct {
	SetID              int       `gorm:"column:setId;primary_key;auto_increment;type:int(11) unsigned;not null;"`                                                                                                                                                                              //
	BankCode           string    `gorm:"column:bankCode;type:varchar(50);"`                                                                                                                                                                                                                    // 银行
	CardType           string    `gorm:"column:cardType;type:enum('','DEBIT','CREDIT');not null;"`                                                                                                                                                                                             // 银行卡类型(DEBIT=借记卡,CREDIT=信用卡)
	Channel            string    `gorm:"column:channel;type:varchar(50);not null;"`                                                                                                                                                                                                            // 渠道名称
	ChannelMerchantID  int       `gorm:"column:channelMerchantId;type:int(11);not null;"`                                                                                                                                                                                                      // 渠道商户
	ChannelMerchantNo  string    `gorm:"column:channelMerchantNo;type:varchar(50);not null;"`                                                                                                                                                                                                  // 渠道商户号
	PayChannelStatus   string    `gorm:"column:payChannelStatus;type:enum('Normal','Close');not null;"`                                                                                                                                                                                        // 支付渠道状态
	PayType            string    `gorm:"column:payType;type:enum('EBank','Quick','OfflineWechatQR','OfflineAlipayQR','OnlineWechatQR','OnlineAlipayQR','OnlineWechatH5','OnlineAlipayH5','UnionPayQR','D0Settlement','UnionPay','OnlineAlipayOriginalH5','OnlineWechatOriginalH5');not null;"` // 支付方式(EBank=网银,Quick=快捷,OfflineWechatQR=线下微信扫码,OfflineAlipayQR=线下支付宝扫码,OnlineWechatQR=线上微信扫码,OnlineAlipayQR=线上支付宝扫码,OnlineWechatH5=线上微信H5,OnlineAlipayH5=线上支付宝H5,OnlineAlipayOriginalH5=线上支付宝原生H5,OnlineWechatOriginalH5=线上微信原生H5,UnionPay=银联快捷/云闪付,UnionPayQR=银联扫码,D0Settlement=D0结算)
	OpenTimeLimit      bool      `gorm:"column:openTimeLimit;type:tinyint(1);not null;"`                                                                                                                                                                                                       // 是否开启控制交易时间
	BeginTime          int       `gorm:"column:beginTime;type:int(11) unsigned;not null;"`                                                                                                                                                                                                     // 开始时间(00:00格式转整形)
	EndTime            int       `gorm:"column:endTime;type:int(11) unsigned;not null;"`                                                                                                                                                                                                       // 结束时间(00:00格式转整形)
	OpenOneAmountLimit bool      `gorm:"column:openOneAmountLimit;type:tinyint(1);not null;"`                                                                                                                                                                                                  // 是否开启控制单笔金额控制
	OneMaxAmount       float64   `gorm:"column:oneMaxAmount;type:decimal(10,2) unsigned;not null;"`                                                                                                                                                                                            // 单笔最大金额
	OneMinAmount       float64   `gorm:"column:oneMinAmount;type:decimal(10,2) unsigned;not null;"`                                                                                                                                                                                            // 单笔最小金额
	OpenDayAmountLimit bool      `gorm:"column:openDayAmountLimit;type:tinyint(1);not null;"`                                                                                                                                                                                                  // 是否开启单日累计金额控制
	DayAmountLimit     float64   `gorm:"column:dayAmountLimit;type:decimal(10,2) unsigned;not null;"`                                                                                                                                                                                          // 累计金额限制
	OpenDayNumLimit    bool      `gorm:"column:openDayNumLimit;type:tinyint(1);not null;"`                                                                                                                                                                                                     // 是否开启单日累计笔数控制
	DayNumLimit        int       `gorm:"column:dayNumLimit;type:int(11) unsigned;not null;"`                                                                                                                                                                                                   // 累计次数限制
	Status             string    `gorm:"column:status;type:enum('Normal','Close');not null;"`                                                                                                                                                                                                  // 配置状态
	CreatedAt          time.Time `gorm:"column:created_at;type:datetime;not null;"`                                                                                                                                                                                                            //
	UpdatedAt          time.Time `gorm:"column:updated_at;type:datetime;not null;"`                                                                                                                                                                                                            //
}

//// 系统配置
//type SysConfig struct {
//	Model
//	Key         string `gorm:"not null;size:128;unique" json:"key" form:"key"` // 配置key
//	Value       string `gorm:"type:text" json:"value" form:"value"`            // 配置值
//	Name        string `gorm:"not null;size:32" json:"name" form:"name"`       // 配置名称
//	Description string `gorm:"size:128" json:"description" form:"description"` // 配置描述
//	CreateTime  int64  `gorm:"not null" json:"createTime" form:"createTime"`   // 创建时间
//	UpdateTime  int64  `gorm:"not null" json:"updateTime" form:"updateTime"`   // 更新时间
//}

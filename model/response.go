package model

import (
	"luckypay/model/constants"
	"time"

	"github.com/mlogclub/simple/web"
)

// UserInfo 用户简单信息
type UserInfo struct {
	Id           int64            `json:"id"`
	Nickname     string           `json:"nickname"`
	Avatar       string           `json:"avatar"`
	SmallAvatar  string           `json:"smallAvatar"`
	Gender       constants.Gender `json:"gender"`
	Birthday     *time.Time       `json:"birthday"`
	TopicCount   int              `json:"topicCount"`   // 话题数量
	CommentCount int              `json:"commentCount"` // 跟帖数量
	FansCount    int              `json:"fansCount"`    // 粉丝数量
	FollowCount  int              `json:"followCount"`  // 关注数量
	Score        int              `json:"score"`        // 积分
	Description  string           `json:"description"`
	CreateTime   int64            `json:"createTime"`

	Followed bool `json:"followed"`
}

// UserDetail 用户详细信息
type UserDetail struct {
	UserInfo
	Username             string `json:"username"`
	BackgroundImage      string `json:"backgroundImage"`
	SmallBackgroundImage string `json:"smallBackgroundImage"`
	HomePage             string `json:"homePage"`
	Forbidden            bool   `json:"forbidden"` // 是否禁言
	Status               int    `json:"status"`
}

// UserProfile 用户个人信息
type UserProfile struct {
	UserDetail
	Roles         []string `json:"roles"`
	PasswordSet   bool     `json:"passwordSet"` // 密码已设置
	Email         string   `json:"email"`
	EmailVerified bool     `json:"emailVerified"`
}

// AdminProfile 用户个人信息
type AdminProfile struct {
	Role      int64  `json:"role"`
	UserName  string `json:"userName"`
	LoginName string `json:"loginName"`
	LoginPwd  string `json:"loginPwd"`
}

type TagResponse struct {
	TagId   int64  `json:"tagId"`
	TagName string `json:"tagName"`
}

type ArticleSimpleResponse struct {
	ArticleId    int64          `json:"articleId"`
	User         *UserInfo      `json:"user"`
	Tags         *[]TagResponse `json:"tags"`
	Title        string         `json:"title"`
	Summary      string         `json:"summary"`
	Cover        *ImageInfo     `json:"cover"`
	SourceUrl    string         `json:"sourceUrl"`
	ViewCount    int64          `json:"viewCount"`
	CommentCount int64          `json:"commentCount"`
	LikeCount    int64          `json:"likeCount"`
	CreateTime   int64          `json:"createTime"`
	Status       int            `json:"status"`
}

type ArticleResponse struct {
	ArticleSimpleResponse
	Content string `json:"content"`
}

type NodeResponse struct {
	NodeId      int64  `json:"nodeId"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
}

type SearchTopicResponse struct {
	TopicId    int64          `json:"topicId"`
	User       *UserInfo      `json:"user"`
	Node       *NodeResponse  `json:"node"`
	Tags       *[]TagResponse `json:"tags"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	CreateTime int64          `json:"createTime"`
}

// 帖子列表返回实体
type TopicResponse struct {
	TopicId         int64               `json:"topicId"`
	Type            constants.TopicType `json:"type"`
	User            *UserInfo           `json:"user"`
	Node            *NodeResponse       `json:"node"`
	Tags            *[]TagResponse      `json:"tags"`
	Title           string              `json:"title"`
	Summary         string              `json:"summary"`
	Content         string              `json:"content"`
	ImageList       []ImageInfo         `json:"imageList"`
	LastCommentTime int64               `json:"lastCommentTime"`
	ViewCount       int64               `json:"viewCount"`
	CommentCount    int64               `json:"commentCount"`
	LikeCount       int64               `json:"likeCount"`
	Liked           bool                `json:"liked"`
	CreateTime      int64               `json:"createTime"`
	Recommend       bool                `json:"recommend"`
	RecommendTime   int64               `json:"recommendTime"`
	Sticky          bool                `json:"sticky"`
	StickyTime      int64               `json:"stickyTime"`
	Status          int                 `json:"status"`
}

// CommentResponse 评论返回数据
type CommentResponse struct {
	CommentId    int64             `json:"commentId"`
	User         *UserInfo         `json:"user"`
	EntityType   string            `json:"entityType"`
	EntityId     int64             `json:"entityId"`
	ContentType  string            `json:"contentType"`
	Content      string            `json:"content"`
	ImageList    []ImageInfo       `json:"imageList"`
	LikeCount    int64             `json:"likeCount"`
	CommentCount int64             `json:"commentCount"`
	Liked        bool              `json:"liked"`
	QuoteId      int64             `json:"quoteId"`
	Quote        *CommentResponse  `json:"quote"`
	Replies      *web.CursorResult `json:"replies"`
	Status       int               `json:"status"`
	CreateTime   int64             `json:"createTime"`
}

// 收藏返回数据
type FavoriteResponse struct {
	FavoriteId int64     `json:"favoriteId"`
	EntityType string    `json:"entityType"`
	EntityId   int64     `json:"entityId"`
	Deleted    bool      `json:"deleted"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	User       *UserInfo `json:"user"`
	Url        string    `json:"url"`
	CreateTime int64     `json:"createTime"`
}

// 消息
type MessageResponse struct {
	MessageId    int64     `json:"messageId"`
	From         *UserInfo `json:"from"`    // 消息发送人
	UserId       int64     `json:"userId"`  // 消息接收人编号
	Title        string    `json:"title"`   // 标题
	Content      string    `json:"content"` // 消息内容
	QuoteContent string    `json:"quoteContent"`
	Type         int       `json:"type"`
	DetailUrl    string    `json:"detailUrl"` // 消息详情url
	ExtraData    string    `json:"extraData"`
	Status       int       `json:"status"`
	CreateTime   int64     `json:"createTime"`
}

// 图片
type ImageInfo struct {
	Url     string `json:"url"`
	Preview string `json:"preview"`
}

type PayOrderStat struct {
	OrderAmount       float64 `gorm:"column:orderAmount;type:decimal(10,2);not null;default:0.00;" json:"orderAmount" form:"orderAmount"`
	ServiceCharge     float64 `gorm:"column:serviceCharge;type:decimal(10,2);not null;default:0.00;" json:"serviceCharge" form:"serviceCharge"`
	WaitPaymentAmount float64 `gorm:"column:waitPaymentAmount;type:decimal(10,2);not null;default:0.00;" json:"waitPaymentAmount" form:"waitPaymentAmount"`
	SuccessAmount     float64 `gorm:"column:successAmount;type:decimal(10,2);not null;default:0.00;" json:"successAmount" form:"successAmount"`
	ExpiredAmount     float64 `gorm:"column:expiredAmount;type:decimal(10,2);not null;default:0.00;" json:"expiredAmount" form:"expiredAmount"`
	WaitPaymentNumber int64   `gorm:"column:waitPaymentNumber;type:int(11);not null;default:0;" json:"waitPaymentNumber" form:"waitPaymentNumber"`
	SuccessNumber     int64   `gorm:"column:successNumber;type:int(11);not null;default:0;" json:"successNumber" form:"successNumber"`
	ExpiredNumber     int64   `gorm:"column:expiredNumber;type:int(11);not null;default:0;" json:"expiredNumber" form:"expiredNumber"`
	Number            int64   `gorm:"column:number;type:int(11);not null;default:0;" json:"number" form:"number"`
}

type MerchantStat struct {
	CurrentAmount float64 `gorm:"column:currentAmount;type:decimal(10,2);not null;default:0.00;" json:"currentAmount" form:"currentAmount"`
}
type MerchantStat2 struct {
	TotalAmount float64 `gorm:"column:totalAmount;type:decimal(10,2);not null;default:0.00;" json:"totalAmount" form:"totalAmount"`
}

type SettlementOrderStat struct {
	Number           int64   `json:"number" form:"number"`
	FailAmount       float64 `gorm:"column:failAmount;type:decimal(10,2);not null;default:0.00;" json:"failAmount" form:"failAmount"`
	OrderAmount      float64 `gorm:"column:orderAmount;type:decimal(10,2);not null;default:0.00;" json:"orderAmount" form:"orderAmount"`
	ExceptionAmount  float64 `gorm:"column:exceptionAmount;type:decimal(10,2);not null;default:0.00;" json:"exceptionAmount" form:"exceptionAmount"`
	SuccessAmount    float64 `gorm:"column:successAmount;type:decimal(10,2);not null;default:0.00;" json:"successAmount" form:"successAmount"`
	TransferedAmount float64 `gorm:"column:transferedAmount;type:decimal(10,2);not null;default:0.00;" json:"transferedAmount" form:"transferedAmount"`
	FailNumber       int64   `gorm:"column:failNumber;type:int(11);not null;default:0;" json:"failNumber" form:"failNumber"`
	TransferedNumber int64   `gorm:"column:transferedNumber;type:int(11);not null;default:0;" json:"transferedNumber" form:"transferedNumber"`
	SuccessNumber    int64   `gorm:"column:successNumber;type:int(11);not null;default:0;" json:"successNumber" form:"successNumber"`
	ExceptionNumber  int64   `gorm:"column:exceptionNumber;type:int(11);not null;default:0;" json:"exceptionNumber" form:"exceptionNumber"`
}

type MerchantSettlementOrderStat struct {
	Number           int64   `json:"number" form:"number"`
	FailAmount       float64 `gorm:"column:failAmount;type:decimal(10,2);not null;default:0.00;" json:"failAmount" form:"failAmount"`
	OrderAmount      float64 `gorm:"column:orderAmount;type:decimal(10,2);not null;default:0.00;" json:"orderAmount" form:"orderAmount"`
	ServiceCharge    float64 `gorm:"column:serviceCharge;type:decimal(10,2);not null;default:0.00;" json:"serviceCharge" form:"serviceCharge"`
	ExceptionAmount  float64 `gorm:"column:exceptionAmount;type:decimal(10,2);not null;default:0.00;" json:"exceptionAmount" form:"exceptionAmount"`
	SuccessAmount    float64 `gorm:"column:successAmount;type:decimal(10,2);not null;default:0.00;" json:"successAmount" form:"successAmount"`
	TransferedAmount float64 `gorm:"column:transferedAmount;type:decimal(10,2);not null;default:0.00;" json:"transferedAmount" form:"transferedAmount"`
	FailNumber       int64   `gorm:"column:failNumber;type:int(11);not null;default:0;" json:"failNumber" form:"failNumber"`
	TransferedNumber int64   `gorm:"column:transferedNumber;type:int(11);not null;default:0;" json:"transferedNumber" form:"transferedNumber"`
	SuccessNumber    int64   `gorm:"column:successNumber;type:int(11);not null;default:0;" json:"successNumber" form:"successNumber"`
	ExceptionNumber  int64   `gorm:"column:exceptionNumber;type:int(11);not null;default:0;" json:"exceptionNumber" form:"exceptionNumber"`
}

type BusinessAmount struct {
	AccountDate             string  `gorm:"column:accountDate;type:varchar(30);default:null;" json:"accountDate"`
	ChannelServiceCharge    float64 `gorm:"column:channelServiceCharge;type:decimal(10,2);default:0.00;" json:"channelServiceCharge"`
	MerchantBalance         float64 `gorm:"column:merchantBalance;type:decimal(10,2);default:0.00;" json:"merchantBalance"`
	MerchantId              int64   `gorm:"column:merchantId;type:int(10);default:0;" json:"merchantId"`
	MerchantNo              string  `gorm:"column:merchantNo;type:varchar(30);default:0.00;" json:"merchantNo"`
	NewDate                 string  `gorm:"column:newDate;type:varchar(30);default:null;" json:"newDate"`
	PayAmount               float64 `gorm:"column:payAmount;type:decimal(10,2);default:0.00;" json:"payAmount"`
	PayCSC                  float64 `gorm:"column:payCSC;type:decimal(10,2);default:0.00;" json:"payCSC"`
	PayServiceCharge        float64 `gorm:"column:payServiceCharge;type:decimal(10,2);default:0.00;" json:"payServiceCharge"`
	SettlementAmount        float64 `gorm:"column:settlementAmount;type:decimal(10,2);default:0.00;" json:"settlementAmount"`
	SettlementMerchantId    int64   `gorm:"column:settlementMerchantId;type:int(10);default:0;" json:"settlementMerchantId"`
	SettlementServiceCharge float64 `gorm:"column:settlementServiceCharge;type:decimal(10,2);default:0.00;" json:"settlementServiceCharge"`
	SettlementTimes         int64   `gorm:"column:settlementTimes;type:int(10);default:0;" json:"settlementTimes"`
}

type PayAmount struct {
	AccountDate string  `gorm:"column:channelServiceCharge;type:varchar(30);default:null;" json:"accountDate"`
	ShortName   string  `gorm:"column:shortName;type:varchar(30);default:null;" json:"shortName"`
	Balance     float64 `gorm:"column:balance;type:decimal(10,2);default:0.00;" json:"balance"`
	MerchantId  int64   `gorm:"column:merchantId;type:int(10);default:0;" json:"merchantId"`
	MerchantNo  string  `gorm:"column:merchantNo;type:varchar(30);default:0.00;" json:"merchantNo"`
	Amount      float64 `gorm:"column:amount;type:decimal(10,2);default:0.00;" json:"amount"`
	PayType     string  `gorm:"column:payType;type:varchar(30);default:null;" json:"payType"`
	PayTypeDesc string  `gorm:"_" json:"payTypeDesc"`
}

//type SettleAmount struct {
//	AccountDate string `gorm:"column:channelServiceCharge;type:varchar(30);default:null;" json:"accountDate"`
//	//agentPayFees             float64 `gorm:"column:agentPayFees;type:decimal(10,2);default:0.00;" json:"agentPayFees"`
//	//agentchargeFees          float64 `gorm:"column:agentchargeFees;type:decimal(10,2);default:0.00;" json:"agentchargeFees"`
//	//agentsettlementFees      float64 `gorm:"column:agentsettlementFees;type:decimal(10,2);default:0.00;" json:"agentsettlementFees"`
//	ChannelMerchantId        int64   `gorm:"column:channelMerchantId;type:int(10);default:0;" json:"channelMerchantId"`
//	ChannelMerchantNo        string  `gorm:"column:channelMerchantNo;type:varchar(30);default:0.00;" json:"channelMerchantNo"`
//	ChargeAmount             float64 `gorm:"column:chargeAmount;type:decimal(10,2);default:0.00;" json:"chargeAmount"`
//	ChargeChannelServiceFees float64 `gorm:"column:chargeChannelServiceFees;type:decimal(10,2);default:0.00;" json:"chargeChannelServiceFees"`
//	ChargeCount              int64   `gorm:"column:chargeCount;type:int(10);default:0;" json:"chargeCount"`
//	ChargeServiceFees        float64 `gorm:"column:chargeServiceFees;type:decimal(10,2);default:0.00;" json:"chargeServiceFees"`
//	Created_at               string  `gorm:"column:created_at;type:varchar(30);default:null;" json:"created_at"`
//	DailyId                  int64   `gorm:"column:dailyId;type:int(10);default:0;" json:"dailyId"`
//	MerchantId               int64   `gorm:"column:merchantId;type:int(10);default:0;" json:"merchantId"`
//	MerchantNo               string  `gorm:"column:merchantNo;type:varchar(30);default:null;" json:"merchantNo"`
//	payAmount                float64 `gorm:"column:payAmount;type:decimal(10,2);default:0.00;" json:"payAmount"`
//	payChannelServiceFees
//	payCount
//	payServiceFees
//	settlementAmount
//	settlementChannelServiceFees
//	settlementCount
//	settlementServiceFees
//	shortName
//	updated_at
//}

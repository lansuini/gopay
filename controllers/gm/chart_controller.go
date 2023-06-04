package gm

import (
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"luckypay/config"
	"luckypay/model"
	"luckypay/utils"
	"strings"
	//"encoding/csv"
)

type ChartController struct {
	Ctx iris.Context
}

// 营业报表
func (c *ChartController) GetBusinessamount() {
	queryParams := model.SearchBusinessAmount{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()

	if queryParams.MerchantNo != "" {
		builder.WriteString(" and merchantNo = '" + queryParams.MerchantNo + "'")
		//fmt.Print(queryParams.MerchantNo)
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}

	if queryParams.BeginDate != "" {
		builder.WriteString(" and accountDate >= '" + queryParams.BeginDate + "'")
		cnd = cnd.Where("accountDate >= ?", queryParams.BeginDate)
	}
	if queryParams.EndDate != "" {
		builder.WriteString(" and accountDate <= '" + queryParams.EndDate + "'")
		cnd = cnd.Where("accountDate <= ?", queryParams.EndDate)
	}
	whereStr := builder.String()
	var businessAmounts []model.BusinessAmount
	var count int64
	countSql := `select count(*) from (select *  from ((
                          select merchantId as settlementMerchantId, accountDate
                          from amount_settlement where {$whereStr}
                          group by merchantId, accountDate order by accountDate desc) a
                          left join (
                          select merchantId as payMerchantId, accountDate as pAD
                          from amount_pay
                          where {$whereStr}
                          group by merchantId, accountDate order by accountDate desc
                          ) b on a.settlementMerchantId = b.payMerchantId and a.accountDate = b.pAD)
                          union
                          select * from ((
                          select merchantId as settlementMerchantId, accountDate
                          from amount_settlement where {$whereStr}
                          group by merchantId, accountDate order by accountDate desc) a
                          right join (
                          select merchantId as payMerchantId, accountDate as pAD
                          from amount_pay
                          where {$whereStr}
                          group by merchantId, accountDate order by accountDate desc
                          ) b on a.settlementMerchantId = b.payMerchantId and a.accountDate = b.pAD) ) c`
	countSql = strings.Replace(countSql, `\n`, "", -1)
	countSql = strings.Replace(countSql, `{$whereStr}`, whereStr, -1)
	err := sqls.DB().Raw(countSql).Scan(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	settlementSql := `select c.*, merchant.merchantNo, merchant.shortName,merchant_amount.settlementAmount as merchantBalance from (select *, ifnull(b.pAD,a.accountDate) as newDate , ifnull(b.payMerchantId, a.settlementMerchantId) as merchantId  from ((
                          select merchantId as settlementMerchantId, accountDate, sum(amount) as settlementAmount, sum(serviceCharge) as settlementServiceCharge, sum(transferTimes) as settlementTimes, sum(channelServiceCharge) as channelServiceCharge
                          from amount_settlement where {$whereStr}
                          group by merchantId, accountDate order by accountDate desc) a
                          left join (
                          select merchantId as payMerchantId, accountDate as pAD, sum(amount) as payAmount, sum(serviceCharge) as payServiceCharge, sum(channelServiceCharge) as payCSC
                          from amount_pay
                          where {$whereStr}
                          group by merchantId, accountDate order by accountDate desc
                          ) b on a.settlementMerchantId = b.payMerchantId and a.accountDate = b.pAD)
                          union
                          select *, ifnull(a.accountDate,b.pAD) as newDate, ifnull(a.settlementMerchantId,b.payMerchantId) as merchantId from ((
                          select merchantId as settlementMerchantId, accountDate, sum(amount) as settlementAmount, sum(serviceCharge) as settlementServiceCharge, sum(transferTimes) as settlementTimes, sum(channelServiceCharge)  as channelServiceCharge
                          from amount_settlement where {$whereStr}
                          group by merchantId, accountDate order by accountDate desc) a
                          right join (
                          select merchantId as payMerchantId, accountDate as pAD, sum(amount) as payAmount, sum(serviceCharge) as payServiceCharge, sum(channelServiceCharge) as payCSC
                          from amount_pay
                          where {$whereStr}
                          group by merchantId, accountDate order by accountDate desc
                          ) b on a.settlementMerchantId = b.payMerchantId and a.accountDate = b.pAD) ) c
                          left join
                          merchant on c.merchantId = merchant.merchantId left Join merchant_amount on c.merchantId = merchant_amount.merchantId order by c.newDate desc`
	settlementSql = strings.Replace(settlementSql, `\n`, "", -1)
	settlementSql = strings.Replace(settlementSql, `{$whereStr}`, whereStr, -1)
	err = sqls.DB().Raw(settlementSql).Scan(&businessAmounts).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": businessAmounts})
	return
	//var count int64
	//err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Find(&businessAmounts).Offset(-1).Count(&count).Error
	//if err != nil {
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
	//	return
	//}

	//var rows []map[string]interface{}
	//for _, finance := range finances {
	//	row := make(map[string]interface{})
	//	row["merchantNo"] = finance.MerchantNo
	//	row["platformOrderNo"] = finance.PlatformOrderNo
	//	row["accountDate"] = finance.AccountDate
	//	row["accountType"] = finance.AccountType
	//	row["amount"] = finance.Amount
	//	row["balance"] = finance.Balance
	//	row["financeType"] = finance.FinanceType
	//	row["financeTypeDesc"] = config.FinanceType[finance.FinanceType]
	//	row["insDate"] = finance.CreatedAt
	//	row["insTime"] = finance.CreatedAt
	//	row["sourceDesc"] = finance.SourceDesc
	//	row["summary"] = finance.Summary
	//	rows = append(rows, row)
	//}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": 10, "rows": businessAmounts})
}

// 支付订单金额统计
func (c *ChartController) GetPayamount() {
	queryParams := model.SearchPayAmount{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	//fmt.Println(queryParams.Offset)
	var builder strings.Builder
	builder.WriteString("1 = 1 ")

	if queryParams.MerchantNo != "" {
		builder.WriteString(" and merchantNo = '" + queryParams.MerchantNo + "'")
	}
	if queryParams.PayType != "" {
		builder.WriteString(" and payType = '" + queryParams.PayType + "'")
	}
	if queryParams.BeginDate != "" {
		builder.WriteString(" and accountDate >= '" + queryParams.BeginDate + "'")
	}
	if queryParams.EndDate != "" {
		builder.WriteString(" and accountDate <= '" + queryParams.EndDate + "'")
	}
	whereStr := builder.String()
	var payAmounts []model.PayAmount
	var count int64
	countSql := `select count(*) from (
            select merchantId, accountDate, payType, balance, sum(amount) as amount
            from amount_pay
            where {$whereStr}
            group by merchantId, accountDate, payType order by accountDate desc, merchantId, payType
            ) a`
	countSql = strings.Replace(countSql, `\n`, "", -1)
	countSql = strings.Replace(countSql, `{$whereStr}`, whereStr, -1)
	err := sqls.DB().Raw(countSql).Scan(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var stat struct {
		Amount float64 `gorm:"column:amount;type:decimal(10,2);default:0.00;" json:"amount"`
		Num    int64   `gorm:"column:num;type:int(10);default:0;" json:"num"`
	}
	statSql := `select count(accountDate) as num , sum(amount) as amount from(
            select merchantId, accountDate, payType, balance, sum(amount) as amount
            from amount_pay
            where {$whereStr}
            group by merchantId, accountDate, payType order by accountDate desc, merchantId, payType
            ) a`
	statSql = strings.Replace(statSql, `\n`, "", -1)
	statSql = strings.Replace(statSql, `{$whereStr}`, whereStr, -1)
	err = sqls.DB().Raw(statSql).Scan(&stat).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	paySql := `select a.*, merchant.merchantNo, merchant.shortName from (
            select merchantId, accountDate, payType, balance, sum(amount) as amount
            from amount_pay
            where {$whereStr}
            group by merchantId, accountDate, payType order by accountDate desc, merchantId, payType
            ) a left join merchant on a.merchantId = merchant.merchantId`
	paySql = strings.Replace(paySql, `\n`, "", -1)
	paySql = strings.Replace(paySql, `{$whereStr}`, whereStr, -1)
	err = sqls.DB().Raw(paySql).Scan(&payAmounts).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	for index, value := range payAmounts {
		payAmounts[index].PayTypeDesc = config.PayType[value.PayType]
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": payAmounts, "stat": stat})
	return
}

// 代付订单金额统计
func (c *ChartController) GetSettleamount() {
	queryParams := model.SearchPayAmount{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	//fmt.Println(queryParams.Offset)
	cnd := sqls.DB()

	if queryParams.MerchantNo != "" {
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.BeginDate != "" {
		cnd = cnd.Where("accountDate >= ?", queryParams.BeginDate)
	}
	if queryParams.EndDate != "" {
		cnd = cnd.Where("accountDate <= ?", queryParams.EndDate)
	}
	var settleAmounts []model.MerchantDailyStats
	var count int64

	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Find(&settleAmounts).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var stat struct {
		PAmount    float64 `gorm:"column:pAmount;type:decimal(10,2);default:0.00;" json:"payAmount"`
		PFees      float64 `gorm:"column:pFees;type:decimal(10,2);default:0.00;" json:"payServiceFees"`
		PChanFees  float64 `gorm:"column:pChanFees;type:decimal(10,2);default:0.00;" json:"payChanServiceFees"`
		PAgentFees float64 `gorm:"column:pAgentFees;type:decimal(10,2);default:0.00;" json:"pAgentFees"`
		SAmount    float64 `gorm:"column:sAmount;type:decimal(10,2);default:0.00;" json:"settlementAmount"`
		SFees      float64 `gorm:"column:sFees;type:decimal(10,2);default:0.00;" json:"settlementServiceFees"`
		CFees      float64 `gorm:"column:cFees;type:decimal(10,2);default:0.00;" json:"chargeServiceFees"`
		SChanFees  float64 `gorm:"column:sChanFees;type:decimal(10,2);default:0.00;" json:"settlementChanServiceFees"`
		SAgentFees float64 `gorm:"column:sAgentFees;type:decimal(10,2);default:0.00;" json:"sAgentFees"`
		CAmount    float64 `gorm:"column:cAmount;type:decimal(10,2);default:0.00;" json:"chargeAmount"`
		CChanFees  float64 `gorm:"column:cChanFees;type:decimal(10,2);default:0.00;" json:"chargeChanServiceFees"`
		CAgentFees float64 `gorm:"column:cAgentFees;type:decimal(10,2);default:0.00;" json:"cAgentFees"`
		SCount     int64   `gorm:"column:sCount;type:int(10);default:0;" json:"settlementCount"`
		CCount     int64   `gorm:"column:cCount;type:int(10);default:0;" json:"chargeCount"`
		Num        int64   `gorm:"_" json:"num"`
	}
	statSql := `select sum(payAmount) as pAmount, sum(payServiceFees) as pFees, sum(payChannelServiceFees) as pChanFees, sum(agentPayFees) as pAgentFees, sum(settlementCount) as sCount, sum(settlementAmount) as sAmount, sum(settlementServiceFees) as sFees, sum(settlementChannelServiceFees) as sChanFees, sum(agentsettlementFees) as sAgentFees, sum(chargeCount) as cCount, sum(chargeAmount) as cAmount, sum(chargeServiceFees) as cFees, sum(chargeChannelServiceFees) as cChanFees, sum(agentchargeFees) as cAgentFees from merchant_daily_stats`
	statSql = strings.Replace(statSql, `\n`, "", -1)
	err = sqls.DB().Raw(statSql).Scan(&stat).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	stat.Num = count
	//for index, value := range settleAmounts {
	//	settleAmounts[index].PayTypeDesc = config.PayType[value.PayType]
	//}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": settleAmounts, "stat": stat})
	return
}

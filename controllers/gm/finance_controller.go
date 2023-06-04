package gm

import (
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"luckypay/config"
	"luckypay/model"
	"luckypay/utils"
	//"encoding/csv"
)

type FinanceController struct {
	Ctx iris.Context
}

func (c *FinanceController) GetSearch() {
	queryParams := model.SearchFinance{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	cnd := sqls.DB()

	if queryParams.MerchantNo != "" {
		//fmt.Print(queryParams.MerchantNo)
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}

	if queryParams.SourceDesc != "" {
		cnd = cnd.Where("sourceDesc = ?", queryParams.SourceDesc)
	}
	if queryParams.PlatformOrderNo != "" {
		cnd = cnd.Where("platformOrderNo = ?", queryParams.PlatformOrderNo)
	}
	if queryParams.FinanceType != "" {
		cnd = cnd.Where("financeType = ?", queryParams.FinanceType)
	}

	if queryParams.BeginTime != "" {
		cnd = cnd.Where("created_at >= ?", queryParams.BeginTime)
	}
	if queryParams.EndTime != "" {
		cnd = cnd.Where("created_at <= ?", queryParams.EndTime)
	}

	var finances []model.Finance
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("id desc").Find(&finances).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	var rows []map[string]interface{}
	for _, finance := range finances {
		row := make(map[string]interface{})
		row["merchantNo"] = finance.MerchantNo
		row["platformOrderNo"] = finance.PlatformOrderNo
		row["accountDate"] = finance.AccountDate
		row["accountType"] = finance.AccountType
		row["amount"] = finance.Amount
		row["balance"] = finance.Balance
		row["financeType"] = finance.FinanceType
		row["financeTypeDesc"] = config.FinanceType[finance.FinanceType]
		row["insDate"] = finance.CreatedAt
		row["insTime"] = finance.CreatedAt
		row["sourceDesc"] = finance.SourceDesc
		row["summary"] = finance.Summary
		rows = append(rows, row)
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": rows})
}

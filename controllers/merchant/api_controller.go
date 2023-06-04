package merchant

import (
	"github.com/dchest/captcha"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"luckypay/cache"
	"luckypay/config"
	"luckypay/model"
	"luckypay/mytool"
	"luckypay/services"
	"luckypay/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ApiController struct {
	BaseController
	Ctx iris.Context
}

func (c *ApiController) GetBasedata() {
	//items := $request->getParam('requireItems');
	requireItems := c.Ctx.FormValue("requireItems")
	items := strings.Split(requireItems, ",")
	dataMap := make(map[string]interface{})
	for _, item := range items {
		//childMap = append(dataMap)

		childs := []map[string]string{}
		for key, value := range config.BaseData[item] {
			child := make(map[string]string)
			child["key"] = key
			child["value"] = value
			childs = append(childs, child)
		}
		dataMap[item] = childs
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": dataMap})
	//fmt.Print(items)
}

func (c *ApiController) PostLogin(request *http.Request) {
	var loginParams model.LoginForm
	valid_err := utils.Validate(c.Ctx, &loginParams, "")
	if valid_err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	unikey := "merchantLogin-" + Session.ID()
	redisServer := config.NewRedis()
	exists, err := redisServer.Exists(c.Ctx, unikey).Result()
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "redis！" + err.Error()})
		return
	}
	if exists > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "请求过于频繁！稍候再试"})
		return
	} else {
		redisServer.SetEX(c.Ctx, unikey, 1, 5*time.Second)
	}
	if !captcha.VerifyString(loginParams.CaptchaId, loginParams.CaptchaCode) {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "验证码错误！"})
		return
	}
	merchantAccount, err := services.MerchantAccountService.GetCacheByLoginName(loginParams.LoginName)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "用户名或者密码错误"})
		return
	}
	merchantData, res := services.MerchantService.GetCacheByMerchantNo(merchantAccount.MerchantNo)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "用户名或者密码错误！"})
		return
	}
	loginIP := utils.GetRealIp(request)
	if !utils.IsIpWhite(loginIP, merchantData.LoginIPWhite) {
		logrus.Error(merchantData.LoginIPWhite + "-ip限制，不允许登录，当前登录IP：" + loginIP)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "ip限制，不允许登录，当前登录IP：" + loginIP})
		return
	}
	if merchantAccount.Status != "Normal" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "用户名或者密码错误."})
		return
	}
	//限制密码错误次数
	if merchantAccount.LoginFailNum >= 5 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "用户名或者密码错误，"})
		return
	}

	loginPwd := mytool.GetHashPassword(loginParams.LoginPwd)
	if loginPwd != merchantAccount.LoginPwd {
		if merchantAccount.LoginFailNum+1 >= 5 {
			sqls.DB().Table("merchant_account").Where("accountId = ?", merchantAccount.AccountID).Update("status", "Close")
		} else {
			sqls.DB().Table("merchant_account").Where("accountId = ?", merchantAccount.AccountID).Update("loginFailNum", gorm.Expr("loginFailNum + ?", 1))
		}
		err = sqls.DB().Table("merchant_account").Where("accountId = ?", merchantAccount.AccountID).UpdateColumn("LoginFailNum", gorm.Expr("LoginFailNum + ?", 1)).Error
		if err != nil {
			logrus.Error(err.Error())
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
		//services.MerchantAccountService.RefreshCache()
		cache.MerchantAccount.RefreshOne(merchantAccount.AccountID)
		//登录失败日志
		loginlog := map[string]interface{}{
			"ip":        loginIP,
			"status":    "Fail",
			"accountId": LoginAccountId,
		}
		sqls.DB().Table("merchant_account_login_log").Create(loginlog)

		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "用户名或者密码错误"})
		return
	}

	rdb := config.NewRedis()
	_, err = rdb.Set(c.Ctx, "merchant_login_accountId_"+strconv.FormatInt(merchantAccount.AccountID, 10), 1, 0).Result()
	if err != nil {
		logrus.Error(err.Error())
	}
	err = sqls.DB().Table("merchant_account").Where("accountId = ?", merchantAccount.AccountID).Updates(map[string]interface{}{"latestLoginTime": time.Now(), "loginFailNum": 0}).Error
	if err != nil {
		logrus.Error(err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	//登录日志

	loginlog := map[string]interface{}{
		"ip":        loginIP,
		"status":    "Success",
		"accountId": LoginAccountId,
	}
	err = sqls.DB().Table("merchant_account_login_log").Create(loginlog).Error
	if err != nil {
		logrus.Error("merchant_account_login_log: ", err.Error())
	}
	c.SaveLoginSession(c.Ctx, merchantAccount)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "登录成功！"})
	return
}

// 商户后台修改登录密码
func (c *ApiController) GetChangeloginpwd() {
	//TODO:密码错误次数上限
	updateParams := model.MerchantMerchantUpdatePwd{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	merchantAccount := model.MerchantAccount{}
	err := sqls.DB().Where("accountId = ?", LoginAccountId).First(&merchantAccount).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	if merchantAccount.LoginPwd != mytool.GetHashPassword(updateParams.OldPwd) {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "旧密码不正确"})
		return
	}
	cnd := sqls.DB()
	updates := map[string]interface{}{
		"loginPwd":          mytool.GetHashPassword(updateParams.NewPwd),
		"loginPwdAlterTime": time.Now().Format("2006-01-02 15:04:05"),
		"loginFailNum":      0,
	}
	err = cnd.Model(&merchantAccount).Updates(updates).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "修改失败" + err.Error()})
		return
	}
	// TODO:商户修改登录密码日志
	cache.MerchantAccount.RefreshOne(merchantAccount.AccountID)
	Session.Destroy()
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功,请重新登录"})
}

// 商户后台修改支付密码
func (c *ApiController) GetChangesecurepwd() {
	//TODO:密码错误次数上限
	updateParams := model.MerchantMerchantUpdatePwd{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	merchantAccount := model.MerchantAccount{}
	err := sqls.DB().Where("accountId = ?", LoginAccountId).First(&merchantAccount).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if merchantAccount.SecurePwd != mytool.GetHashPassword(updateParams.OldPwd) {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "旧密码不正确"})
		return
	}
	cnd := sqls.DB()
	updates := map[string]interface{}{
		"securePwd": mytool.GetHashPassword(updateParams.NewPwd),
	}
	err = cnd.Model(&merchantAccount).Updates(updates).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "修改失败" + err.Error()})
		return
	}
	// TODO:商户修改支付密码日志
	//cache.MerchantAccount.RefreshCache()
	cache.MerchantAccount.RefreshOne(merchantAccount.AccountID)
	Session.Destroy()
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功,请重新登录"})
}

// 收支明细
func (c *ApiController) GetFinancesearch() {
	queryParams := model.MerchantMerchantFinanceSearch{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}

	cnd := sqls.DB()
	cnd = cnd.Where("merchantId = ?", LoginMerchantId)
	if queryParams.PlatformOrderNo != "" {
		cnd = cnd.Where("platformOrderNo = ?", queryParams.PlatformOrderNo)
	}
	if queryParams.MerchantOrderNo != "" {
		cnd = cnd.Where("merchantOrderNo = ?", queryParams.MerchantOrderNo)
	}
	if queryParams.FinanceType != "" {
		cnd = cnd.Where("financeType = ?", queryParams.FinanceType)
	}
	if queryParams.OperateSource != "" {
		cnd = cnd.Where("operateSource = ?", queryParams.OperateSource)
	}
	if queryParams.BeginTime != "" {
		cnd = cnd.Where("m.created_at >= ?", queryParams.BeginTime)
	}
	if queryParams.EndTime != "" {
		cnd = cnd.Where("m.created_at <= ?", queryParams.EndTime)
	}

	var finances []model.Finance
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("id desc").Find(&finances).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	result := make(map[string]interface{})
	var results []map[string]interface{}
	for _, financesVal := range finances {
		result["accountDate"] = financesVal.AccountDate
		result["accountType"] = financesVal.AccountType
		result["accountTypeDesc"] = config.AccountType[financesVal.AccountType]
		result["financeType"] = financesVal.FinanceType
		result["financeTypeDesc"] = config.FinanceType[financesVal.FinanceType]
		result["amount"] = financesVal.Amount
		result["balance"] = financesVal.Balance
		result["insTime"] = financesVal.CreatedAt.Format("2006-01-02 15:04:05")
		result["id"] = financesVal.Id
		result["merchantNo"] = financesVal.MerchantNo
		result["merchantOrderNo"] = financesVal.MerchantOrderNo
		result["operateSource"] = config.OperateSource[financesVal.OperateSource]
		result["platformOrderNo"] = financesVal.PlatformOrderNo
		result["sourceDesc"] = financesVal.SourceDesc
		result["sourceId"] = financesVal.SourceID
		result["summary"] = financesVal.Summary
		result["updated_at"] = financesVal.UpdatedAt
		results = append(results, result)
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": results})
}

// 每日报表
func (c *ApiController) GetReport() {
	queryParams := model.MerchantReport{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}

	cnd := sqls.DB()
	cnd = cnd.Where("merchantId = ?", LoginMerchantId)
	if queryParams.BeginTime != "" {
		cnd = cnd.Where("m.accountDate >= ?", queryParams.BeginTime)
	}
	if queryParams.EndTime != "" {
		cnd = cnd.Where("m.accountDate <= ?", queryParams.EndTime)
	}

	var balances []model.BalanceAdjustment
	var merchantDailyStatss []model.MerchantDailyStats
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("dailyId desc").Find(&merchantDailyStatss).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var (
		sum_payAmount              = 0.00
		sum_payCount         int64 = 0
		sum_payFees                = 0.00
		sum_settlementAmount       = 0.00
		sum_settlementCount  int64 = 0
		sum_settlementFees         = 0.00
	)

	cnd2 := sqls.DB()
	var results []map[string]interface{}
	for _, value := range merchantDailyStatss {
		cnd2 = cnd2.Where("created_at >= ?", value.CreatedAt.String()+" 00:00:00").Where("created_at <=?", value.CreatedAt.String()+" 23:59:59")
		cnd2 = cnd2.Where("merchantId =?", LoginMerchantId).Where("status =?", "Success").Raw("bankrollDirection, count(adjustmentId) as pcount, sum(amount) as amount")
		cnd2 = cnd2.Group("bankrollDirection")
		err2 := cnd2.Find(&balances).Error
		if err2 != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err2.Error()})
			return
		}

		nv := make(map[string]interface{})
		nv["account_date"] = value.AccountDate
		nv["pay_amount"] = value.PayAmount
		nv["pay_count"] = value.PayCount
		nv["pay_fee"] = value.PayServiceFees
		nv["settlement_amount"] = value.SettlementAmount
		nv["settlement_count"] = value.SettlementCount
		nv["settlement_fee"] = value.SettlementServiceFees

		sum_payAmount += value.PayAmount
		sum_payCount += value.PayCount
		sum_payFees += value.PayServiceFees
		sum_settlementAmount += value.SettlementAmount
		sum_settlementCount += value.SettlementCount
		sum_settlementFees += value.SettlementServiceFees
		results = append(results, nv)
	}

	stat := make(map[string]interface{})
	stat["sum_payAmount"] = sum_payAmount
	stat["sum_payCount"] = sum_payCount
	stat["sum_payFees"] = sum_payFees
	stat["sum_settlementAmount"] = sum_settlementAmount
	stat["sum_settlementCount"] = sum_settlementCount
	stat["sum_settlementFees"] = sum_settlementFees

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "stat": stat, "rows": results})
}

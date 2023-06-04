package merchant

import (
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"luckypay/config"
	"luckypay/model"
	"luckypay/utils"
	"net/http"
	"strings"
	"time"
	//"encoding/csv"
)

type ManagerController struct {
	BaseController
	Ctx iris.Context
}

func (c *ManagerController) GetAdminlist() {

	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()

	var systemAccounts []model.SystemAccount
	var count int64
	err := cnd.Order("id desc").Find(&systemAccounts).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var rows []map[string]interface{}
	for _, value := range systemAccounts {
		row := make(map[string]interface{})
		row["userName"] = value.UserName
		row["loginName"] = value.LoginName
		row["loginPwdAlterTime"] = value.LoginPwdAlterTime
		row["status"] = value.Status
		row["role"] = value.Role
		row["created_at"] = value.CreatedAt
		if value.GoogleAuthSecretKey != "" {
			row["googleBind"] = "已绑定"
		} else {
			row["googleBind"] = "未绑定"
		}
		rows = append(rows, row)
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": rows})
}

func (c *ManagerController) GetAdminupdate() {
	updateParams := model.SystemAccountUpdate{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	if updateParams.Id <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "id信息错误"})
		return
	}
	vailRes := c.googleAuthVail(updateParams.GoogleAuthSecretKey, updateParams.Id)
	if vailRes["Msg"] == "" {
		vailRes["Msg"] = "操作有风险"
	}
	if vailRes["Success"] == "0" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": vailRes["Msg"]})
		return
	}
	var systemAccounts []model.SystemAccount
	cnd := sqls.DB()
	updates := map[string]interface{}{
		"Role":   updateParams.Role,
		"Status": updateParams.Status,
	}
	err := cnd.Where("id = ?", updateParams.Id).Model(&systemAccounts).Updates(updates).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "更新失败"})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

func (c *ManagerController) GetAdmindelete() {
	deleteParams := model.SystemAccountDelete{}
	valid_err := utils.Validate(c.Ctx, &deleteParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	if deleteParams.Id <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "id信息错误"})
		return
	}
	var systemAccounts []model.SystemAccount
	cnd := sqls.DB()
	err := cnd.Model(&systemAccounts).Delete("id = ?", deleteParams.Id).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "删除失败"})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "删除成功"})
}

// 修改登录密码
func (c *ManagerController) GetAdminpwdupdate() {
	updateParams := model.SystemAccountPwdUpdate{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}

	var systemAccounts []model.SystemAccount
	cnd := sqls.DB()
	updates := map[string]interface{}{
		"LoginPwd": updateParams.NewPwd,
		//"LoginPwdAlterTime" => date('YmdHis'),
		//"loginFailNum" => 0,
	}
	err := cnd.Where("id = ?", 1).Model(&systemAccounts).Updates(updates).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "更新失败"})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

func (c *ManagerController) GetBlackusersettlement() {
	queryParams := model.SearchBlackUserSettlement{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()
	if queryParams.BlackUserName != "" {
		builder.WriteString(" and blackUserName = '" + string(queryParams.BlackUserName) + "'")
		cnd = cnd.Where("blackUserName = ?", queryParams.BlackUserName)
	}
	if queryParams.BlackUserAccount != "" {
		builder.WriteString(" and blackUserAccount = '" + string(queryParams.BlackUserAccount) + "'")
		cnd = cnd.Where("blackUserAccount = ?", queryParams.BlackUserAccount)
	}
	if queryParams.BlackUserType != "" {
		builder.WriteString(" and blackUserType = '" + string(queryParams.BlackUserType) + "'")
		cnd = cnd.Where("blackUserType = ?", queryParams.BlackUserType)
	}

	var blackUserSettlementS []model.BlackUserSettlement
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("blackUserId desc").Find(&blackUserSettlementS).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": blackUserSettlementS})
}

func (c *ManagerController) GetBlackusersettlementupdate() {
	updateParams := model.BlackUserSettlement{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	if updateParams.BlackUserID <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "id信息错误"})
		return
	}
	var blackUserSettlementS []model.BlackUserSettlement
	cnd := sqls.DB()
	updates := map[string]interface{}{
		"blackUserName":    updateParams.BlackUserName,
		"blackUserAccount": updateParams.BlackUserAccount,
		"blackUserType":    updateParams.BlackUserType,
		"blackUserStatus":  updateParams.BlackUserStatus,
	}
	err := cnd.Where("blackUserID = ?", updateParams.BlackUserID).Model(&blackUserSettlementS).Updates(updates).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "更新失败"})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

func (c *ManagerController) GetBlackusersettlementadd() {
	addParams := model.BlackUserSettlement{}
	valid_err := utils.Validate(c.Ctx, &addParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var blackUserSettlementS []model.BlackUserSettlement
	cnd := sqls.DB()
	adds := map[string]interface{}{
		"blackUserName":    addParams.BlackUserName,
		"blackUserAccount": addParams.BlackUserAccount,
		"blackUserType":    addParams.BlackUserType,
		"blackUserStatus":  addParams.BlackUserStatus,
	}
	err := cnd.Model(&blackUserSettlementS).Create(adds).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "更新失败"})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

func (c *ManagerController) GetBlackusersettlementdelete() {
	deleteParams := model.BlackUserSettlement{}
	valid_err := utils.Validate(c.Ctx, &deleteParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	if deleteParams.BlackUserID <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "id信息错误"})
		return
	}
	var blackUserSettlementS []model.BlackUserSettlement
	cnd := sqls.DB()
	err := cnd.Model(&blackUserSettlementS).Delete("id = ?", deleteParams.BlackUserID).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "更新失败"})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

/*func (c *ManagerController) GetBanklist() {
	queryParams := model.SearchBank{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()
	if queryParams.BankName != "" {
		builder.WriteString(" and name = '" + string(queryParams.BankName) + "'")
		cnd = cnd.Where("name = ?", queryParams.BankName)
	}
	if queryParams.Status != "" {
		builder.WriteString(" and status = '" + string(queryParams.Status) + "'")
		cnd = cnd.Where("status = ?", queryParams.Status)
	}

	var bankss []model.Banks
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("id desc").Find(&bankss).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": bankss})
}*/

func (c *ManagerController) GetBankedit() {
	//updateParams := model.Banks{}
	//valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	//if valid_err != nil {
	//	logrus.Info(valid_err.Error())
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
	//	return
	//}
	//if updateParams.Status == "disabled" {
	//	if updateParams.StartTime == "" || updateParams.EndTime == "" {
	//		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "请选择开始时间和结束时间"})
	//		return
	//	}
	//} else if updateParams.StartTime != "" || updateParams.EndTime != "" && updateParams.StartTime > updateParams.EndTime {
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "开始时间不能大于结束时间"})
	//	return
	//} else {
	//	updateParams.Status = "enabled"
	//}
	//
	//var bankss []model.Banks
	//cnd := sqls.DB()
	//updates := map[string]interface{}{
	//	"code":       updateParams.Code,
	//	"status":     updateParams.Status,
	//	"start_time": null,
	//	"end_time":   updateParams.EndTime,
	//}
	//err := cnd.Where("code = ?", updateParams.Code).Model(&bankss).Updates(updates).Error
	//
	//if err != nil {
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "更新失败"})
	//	return
	//}
	//c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

func (c *ManagerController) PostGoogleauth() {
	code := c.Ctx.PostValueTrim("code")
	if len(code) == 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}
	//TODO:限制密码错误次数
	unikey := "merchantLoginGoogoleAuth-" + Session.ID()
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
	secret := utils.AesCBCDecrypt(Session.GetString("googleAuthSecretKey"))
	res, err := utils.GoogleAuthenticator.VerifyCode(secret, code)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "验证失败，请刷新页面重试"})
		return
	}
	Session.Set("googleAuthCheck", true)

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "验证成功"})
}

func (c *ManagerController) PostBindgoogleauth(r *http.Request) {
	code := c.Ctx.PostValueTrim("code")
	if len(code) == 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}
	secret := Session.GetString("googleNewSecret")
	//fmt.Printf(secret)
	res, err := utils.GoogleAuthenticator.VerifyCode(secret, code)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "验证失败，请刷新页面重试"})
		return
	}
	Session.Set("googleAuthSecretKey", utils.AesCBCEncrypt(secret))
	Session.Set("googleAuthCheck", true)
	err = sqls.DB().Table("merchant_account").Where("accountId = ?", LoginAccountId).Update("googleAuthSecretKey", utils.AesCBCEncrypt(secret)).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "绑定失败，请重试！"})
		return
	}
	//TODO:添加绑定谷歌验证码日志
	/*MerchantAccountActionLog::insert([
	    'action' => 'BIND_GOOGLE_AUTH',
	    'actionBeforeData' => '',
	    'actionAfterData' => '',
	    'status' => 'Success',
	    'ip' => Tools::getIp(),
	    'ipDesc' => Tools::getIpDesc(),
	    'accountId' => $_SESSION['accountId'],
	]);*/
	actionLog := make(map[string]interface{})
	actionLog["action"] = "BIND_GOOGLE_AUTH"
	actionLog["status"] = "Success"
	userIP := utils.GetRealIp(r)
	actionLog["ip"] = userIP
	//actionLog["ipDesc"] = userIP TODO:ip归属地
	actionLog["accountId"] = LoginAccountId
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "绑定成功"})
}

func (c *ManagerController) googleAuthVail(googleAuthSecretKey string, accountId int64) map[string]string {

	var vailRes = map[string]string{
		"Success": "1",
		"Msg":     "",
	}
	if googleAuthSecretKey == "" {
		vailRes["Success"] = "0"
		vailRes["Msg"] = "请输入谷歌验证码！"
		return vailRes
	}
	var systemAccounts []model.SystemAccount
	cnd := sqls.DB()
	err := cnd.Where("id = ?", accountId).First(&systemAccounts).Error
	if err != nil {
		vailRes["Msg"] = "账户不存在"
		return vailRes
	}
	for _, value := range systemAccounts {
		if value.Role != 5 {
			vailRes["Success"] = "0"
			vailRes["Msg"] = "没有权限操作此功能"
			return vailRes
		}
		if value.GoogleAuthSecretKey == "" {
			vailRes["Success"] = "0"
			vailRes["Msg"] = "请绑定谷歌验证码后再操作！"
			return vailRes
		}
	}

	return vailRes
}

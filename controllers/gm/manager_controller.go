package gm

import (
	"github.com/dchest/captcha"
	"github.com/kataras/iris/v12"
	"github.com/lgbya/go-dump"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"luckypay/config"
	"luckypay/model"
	"luckypay/services"
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

func (c *ManagerController) PostLogin(req *http.Request) {
	var loginParams model.LoginForm
	valid_err := utils.Validate(c.Ctx, &loginParams, "")
	if valid_err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}

	//TODO:限制密码错误次数
	unikey := "gmLogin-" + Session.ID()
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

	cacheKey := "system_account:" + loginParams.LoginName
	cacheData, _ := redisServer.Get(c.Ctx, cacheKey).Result()
	if err != nil {
		logrus.Error("PostLogin Error", err)
	} else {
		if len(cacheData) == 0 {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "用户名或密码错误."})
			return
		}
	}

	user, err := services.AdminService.SignIn(loginParams.LoginName, loginParams.LoginPwd)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	//没有绑定谷歌验证码的自动绑定，由技术线下提供
	if len(user.GoogleAuthSecretKey) == 0 {
		secret := utils.GoogleAuthenticator.GetSecret()
		Session.Set("googleNewSecret", secret)
		sqls.DB().Table("system_account").Where("loginName = ?", loginParams.LoginName).Update("googleAuthSecretKey", utils.AesCBCEncrypt(secret))
		user.GoogleAuthSecretKey = utils.AesCBCEncrypt(secret)
	}
	c.SaveLoginSession(c.Ctx, user)
	//TODO:ip
	loginlog := make(map[string]interface{})
	loginlog["ip"] = utils.GetRealIp(req)
	loginlog["ipDesc"] = ""
	loginlog["accountId"] = user.Id
	loginlog["status"] = "Success"
	err = sqls.DB().Table("system_account_login_log").Create(loginlog).Error
	if err != nil {
		logrus.Error("插入登录日志失败", err)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "插入登录日志失败！"})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "登录成功！"})
	return
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

func (c *ManagerController) GetAdminloginlog() {
	queryParams := model.SearchAdminLog{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	cnd := sqls.DB().Table("system_account_login_log as s").Joins("left join system_account as sa on s.accountId=sa.id").Select("s.*,sa.loginName,sa.userName")
	if queryParams.LoginName != "" {
		cnd = cnd.Where("sa.loginName = ?", queryParams.LoginName)
	}
	if queryParams.Ip != "" {
		cnd = cnd.Where("s.ip = ?", queryParams.Ip)
	}
	var rows []map[string]interface{}
	var count int64
	err := cnd.Order("id desc").Find(&rows).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": rows})
}

func (c *ManagerController) GetAdminactionlog() {
	queryParams := model.SearchAdminLog{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	cnd := sqls.DB().Table("system_account_action_log as s").Joins("left join system_account as sa on s.accountId=sa.id").Select("s.*,sa.loginName,sa.userName")
	if queryParams.LoginName != "" {
		cnd = cnd.Where("sa.loginName = ?", queryParams.LoginName)
	}
	if queryParams.Ip != "" {
		cnd = cnd.Where("s.ip = ?", queryParams.Ip)
	}
	var rows []map[string]interface{}
	var count int64
	err := cnd.Order("id desc").Find(&rows).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": rows})
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

func (c *ManagerController) GetBlackusersettlementUpdate() {
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

func (c *ManagerController) GetBlackusersettlementCreate() {
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

func (c *ManagerController) GetBlackusersettlementDelete() {
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
	unikey := "gmLoginGoogoleAuth-" + Session.ID()
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
	//fmt.Println(secret)
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

func (c *ManagerController) PostBindgoogleauth() {
	code := c.Ctx.PostValueTrim("code")
	if len(code) == 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}
	secret := Session.GetString("googleNewSecret")
	if len(secret) == 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "验证失败，请刷新页面重试1"})
		return
	}
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
	err = sqls.DB().Table("system_account").Where("id = ?", LoginAdminId).Update("googleAuthSecretKey", utils.AesCBCEncrypt(secret)).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "绑定失败，请重试！"})
		return
	}
	//TODO:添加绑定谷歌验证码日志
	/*SystemAccountActionLog::insert([
			'action' => 'UPDATE_LOGINNAME',
			'actionBeforeData' => $_SESSION['loginName'],
			'actionAfterData' => $loginName,
			'status' => 'Success',
			'accountId' => $_SESSION['accountId'],
			'ip' => Tools::getIp(),
			'ipDesc' => Tools::getIpDesc(),
	]);*/
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

func (c *ManagerController) GetTest(request *http.Request) {
	dump.Printf(request.Header.Get("X-Forwarded-For"))
	dump.Printf(request.Header.Get("X-Real-Ip"))
	ip := utils.GetRealIp(request)
	c.Ctx.WriteString(ip)
	return
}

package merchant

import (
	"github.com/kataras/iris/v12"
	"luckypay/cache"
	"luckypay/menu"
	"luckypay/model"
	yamlConfig "luckypay/pkg/config"
	"luckypay/utils"
	"net/http"
	"strings"
	// "github.com/mlogclub/simple/web"
	// "luckypay/pkg/errs"
	// "luckypay/services"
)

type views struct {
	Host string `form:"host" json:"host"`
}

type Layout struct {
	Menu        menu.MerchantMenu
	AppName     string
	Host        string
	GlobalJsVer string
}

type IndexController struct {
	BaseController
	Ctx    iris.Context
	Layout Layout
}

func (c *IndexController) BeginRequest(ctx iris.Context) {
	if yamlConfig.Instance.Env == "dev" {
		c.Layout.Host = yamlConfig.Instance.BaseUrl + ":" + yamlConfig.Instance.Port
	} else {
		c.Layout.Host = yamlConfig.Instance.BaseUrl
	}
	c.Layout.AppName = yamlConfig.Instance.AppName
	c.Layout.GlobalJsVer = yamlConfig.Instance.GlobalJsVer
}

func (c *IndexController) EndRequest(ctx iris.Context) {}

func (c *IndexController) GetLogin() {

	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)
	c.Ctx.View("/merchant/login.twig")
}

// TODO:这里的查询没有写完（商户后台首页数据）
func (c *IndexController) GetHead(r *http.Request) {
	//date := time.Now().Format("2006-01-02")
	merchantAmount := model.MerchantAmount{}
	merchantAmount, _ = cache.MerchantAmount.GetCacheByMerchantNo(LoginMerchantNo)
	//amountData := services.MerchantAmount.GetAmount(LoginMerchantNo)
	//var merchantAmountRes model.MerchantAmountJoin
	//sql := `select merchant.merchantId,
	//    merchant.merchantNo,
	//    merchant.shortName,
	//    merchant_amount.updated_at,
	//    merchant.settlementType,merchant_amount.accountBalance,merchant_amount.availableBalance,merchant_amount.settlementAmount,
	//    merchant_amount.freezeAmount
	//    (select sum(amount) from amount_pay where accountDate='{$date}' and amount_pay.merchantId = merchant.merchantId) as todayPayAmount,
	//    (select sum(serviceCharge) from amount_pay where accountDate='{$date}' and amount_pay.merchantId = merchant.merchantId) as todayPayServiceCharge,
	//    (select sum(amount) from amount_settlement where accountDate='{$date}' and amount_settlement.merchantId = merchant.merchantId) as todaySettlementAmount,
	//    (select sum(serviceCharge) from amount_settlement where accountDate='{$date}' and amount_settlement.merchantId = merchant.merchantId) as todaySettlementServiceCharge
	//    from merchant left join merchant_amount on merchant.merchantId = merchant_amount.merchantId where merhcant.merchantId = {$merchantId}
	//    `
	//sql = strings.Replace(sql, `{$merchantId}`, Session.GetString("merchantId"), -1)
	//sql = strings.Replace(sql, `{$date}`, date, -1)
	//err := sqls.DB().Raw(sql).Scan(&merchantAmountRes).Error
	//if err != nil {
	//	logrus.Error(err)
	//	//return
	//}

	//charge = 0.00
	//if (!empty($data)) {
	//$charge = $data->todayPayServiceCharge + $data->todaySettlementServiceCharge;
	//}

	c.Ctx.ViewData("menus", Menus)
	c.Ctx.ViewData("userName", Session.GetString("userName"))
	c.Ctx.ViewData("settlementAmount", merchantAmount.SettlementAmount)
	c.Ctx.ViewData("accountBalance", merchantAmount.SettlementAmount)
	c.Ctx.ViewData("availableBalance", merchantAmount.SettlementAmount)
	c.Ctx.ViewData("freezeAmount", merchantAmount.FreezeAmount)
	//c.Ctx.ViewData("shortName", merchantAmountRes)
	//c.Ctx.ViewData("settlementType", merchantAmountRes)
	//c.Ctx.ViewData("merchantNo", merchantAmountRes)
	//c.Ctx.ViewData("todayPayAmount", merchantAmountRes)
	//c.Ctx.ViewData("todaySettlementAmount", merchantAmountRes)
	//c.Ctx.ViewData("charge", merchantAmountRes)

	//logrus.Error(utils.GetUserIP(r))
	//fmt.Fprintln(os.Stdout,utils.GetUserIP(r))

	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/merchant/head.twig")
	//模板数据渲染
	//var test = make(map[string]interface{})
	//test["name"] = "this is name"
	////模板路径说明：基于ConfTwigPath，“/”表示根目录，请使用绝对路径，相对路径会出错
	////--使用模板嵌套时也一样{% extends "/main.twig" %}，请使用绝对路径
	//rst := twig.Render("/gm/head.twig", test)
	//println(rst)
}

func (c *IndexController) GetPayorder() {

	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)
	//c.Ctx.ViewData("menus", c.Layout.Menu)

	c.Ctx.View("/merchant/payorder.twig")
}

func (c *IndexController) GetSettlementorder() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/merchant/settlementorder.twig")
}

func (c *IndexController) GetSettlementorderCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)
	c.Ctx.View("/merchant/settlementorder_create.twig")
}

func (c *IndexController) GetChangeloginpwd() {

	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/merchant/changeloginpwd.twig")
}

func (c *IndexController) GetChangesecurepwd() {

	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/merchant/changesecurepwd.twig")
}

func (c *IndexController) GetGoogleauth() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)
	c.Ctx.View("/merchant/googleauth.twig")
}

func (c *IndexController) GetBindgoogleauth() {
	user := LoginUserName + "@" + c.Ctx.Domain()
	secret := utils.GoogleAuthenticator.GetSecret()
	Session.Set("googleNewSecret", secret)
	qrCodeUrl := utils.GoogleAuthenticator.GetQrcodeUrl(user, secret)
	c.Ctx.ViewData("googleCaptcha", qrCodeUrl)
	googleAuthSecretKey := Session.GetString("googleAuthSecretKey")
	c.Ctx.ViewData("isHaveGoogleCaptcha", len(googleAuthSecretKey) != 0)

	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/merchant/bindgoogleauth.twig")
}

func (c *IndexController) GetFinance() {

	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/merchant/finance.twig")
}

func (c *IndexController) GetReport() {

	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("userName", LoginUserName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/merchant/report.twig")
}

func (c *IndexController) GetLogout() {
	Session.Destroy()
	//Session.Clear()
	//fmt.Println(Session.ID())
	c.Ctx.Redirect("/login")
}

func (c *IndexController) GetCheckipwhite(httpRequest *http.Request) {
	LoginIpWhite = strings.TrimSpace(LoginIpWhite)
	if len(LoginIpWhite) <= 0 {
		Session.Destroy()
		c.Ctx.StopWithText(401, "请先设置白名单再尝试重新登录!")
		return
	}
	check := false

	userIP := utils.GetRealIp(httpRequest)
	whiteIpList := strings.Split(strings.TrimSpace(LoginIpWhite), ",")
	for _, whiteIp := range whiteIpList {
		if whiteIp == userIP {
			check = true
			break
		}
		continue
	}
	if !check {
		c.Ctx.WriteString("ip验证不通过，请先添加ip白名单：" + userIP)
		return
	} else {
		c.Ctx.Redirect("/head")
	}
}

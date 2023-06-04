package gm

import (
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"luckypay/menu"
	"luckypay/utils"

	// "github.com/dchest/captcha"
	"encoding/json"
	yamlConfig "luckypay/pkg/config"
	"os"
	// "github.com/mlogclub/simple/web"
	// "luckypay/pkg/errs"
	// "luckypay/services"
)

type Layout struct {
	Menu        menu.GmMenu
	AppName     string
	Host        string
	GlobalJsVer string
}

type IndexController struct {
	BaseController
	Ctx    iris.Context
	Layout Layout
}

//	func (c *IndexController) init() {
//		osGwd, _ := os.Getwd()
//		confPath := osGwd + "/menu/gm.json"
//
//		jsonFile, err := os.Open(confPath)
//		if err != nil {
//			fmt.Println("error opening json file")
//			return
//		}
//		defer jsonFile.Close()
//
//		jsonData, err := ioutil.ReadAll(jsonFile)
//		if err != nil {
//			fmt.Println("error reading json file")
//			return
//		}
//
//		json.Unmarshal(jsonData, &c.menu)
//		fmt.Println(c.menu)
//	}
func (c *IndexController) BeginRequest(ctx iris.Context) {
	if yamlConfig.Instance.Env == "dev" {
		c.Layout.Host = yamlConfig.Instance.BaseUrl + ":" + yamlConfig.Instance.Port
	} else {
		c.Layout.Host = yamlConfig.Instance.BaseUrl
	}
	c.Layout.AppName = yamlConfig.Instance.AppName
	c.Layout.GlobalJsVer = yamlConfig.Instance.GlobalJsVer
	osGwd, _ := os.Getwd()
	confPath := osGwd + "/menu/gm.json"

	jsonFile, err := os.Open(confPath)
	if err != nil {
		logrus.Error("error opening json file")
		return
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		logrus.Error("error reading json file")
		return
	}

	json.Unmarshal(jsonData, &c.Layout.Menu)
}

func (c *IndexController) EndRequest(ctx iris.Context) {}

func (c *IndexController) GetIndex() {
	//c.Ctx.Writef("From: %s", c.Ctx.Application().ConfigurationReadOnly().GetVHost())
	//c.Ctx.Writef("From: %s", c.Ctx.Application().GetRoutesReadOnly())
	//c.Ctx.Writef("From: %s", c.Ctx.Path(), c.Ctx.Domain())
	//c.Ctx.ViewData("appName", "大东")
	//c.Ctx.ViewData("globalJsVer", "202211301344")
	//
	//c.Ctx.View("/gm/login.html")
}

func (c *IndexController) GetLogin() {

	c.Ctx.ViewData("host", c.Layout.Host)
	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)
	c.Ctx.View("/gm/login.twig")
}

func (c *IndexController) GetGoogleauth() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)
	c.Ctx.View("/merchant/googleauth.twig")
}

func (c *IndexController) GetManagerBindgoogleauth() {

	user := LoginName + "@" + c.Ctx.Domain()
	secret := utils.GoogleAuthenticator.GetSecret()
	Session.Set("googleNewSecret", secret)
	//fmt.Println("Secret:", secret)
	// 动态码(每隔30s会动态生成一个6位数的数字)
	//code, _ := utils.GoogleAuthenticator.GetCode(secret)
	//fmt.Println("Code:", code, err)

	// 用户名
	//qrCode := utils.GoogleAuthenticator.GetQrcode(user, code)
	//fmt.Println("Qrcode", qrCode)

	// 打印二维码地址
	qrCodeUrl := utils.GoogleAuthenticator.GetQrcodeUrl(user, secret)

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.ViewData("googleCaptcha", qrCodeUrl)
	googleAuthSecretKey := Session.GetString("googleAuthSecretKey")

	c.Ctx.ViewData("isHaveGoogleCaptcha", len(googleAuthSecretKey) != 0)

	c.Ctx.View("/gm/manager/bindgoogleauth.twig")
}

func (c *IndexController) GetLogout() {
	Session.Destroy()
	c.Ctx.Redirect("/login")
}

func (c *IndexController) GetMenu() {
	var Menus menu.GmMenu
	// 打开文件
	// 获取配置文件路径
	osGwd, _ := os.Getwd()
	confPath := osGwd + "/menu/gm.json"
	file, err := os.Open(confPath)
	if err != nil {
		logrus.Error("error opening json file")
		return
	}
	// 关闭文件
	defer file.Close()

	// NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
	decoder := json.NewDecoder(file)
	//Decode从输入流读取下一个json编码值并保存在v指向的值里
	errs := decoder.Decode(&Menus)
	if errs != nil {
		panic(errs)
	}
	c.Ctx.ViewData("menus", Menus)
	//
	c.Ctx.View("/gm/menu.twig")

}

func (c *IndexController) GetHead() {

	c.Ctx.ViewData("host", c.Layout.Host)
	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)
	c.Ctx.ViewData("menus", c.Layout.Menu)
	//
	c.Ctx.View("/gm/head.twig")
	//模板数据渲染
	//var test = make(map[string]interface{})
	//test["name"] = "this is name"
	////模板路径说明：基于ConfTwigPath，“/”表示根目录，请使用绝对路径，相对路径会出错
	////--使用模板嵌套时也一样{% extends "/main.twig" %}，请使用绝对路径
	//rst := twig.Render("/gm/head.twig", test)
	//println(rst)
}

func (c *IndexController) GetPayorder() {

	c.Ctx.ViewData("host", c.Layout.Host)
	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/gm/payorder/index.twig")
}

func (c *IndexController) GetPayorderDetail() {
	c.Ctx.ViewData("host", c.Layout.Host)
	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/gm/payorder/detail.twig")
}

func (c *IndexController) GetSettlementorder() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/settlementorder/index.twig")
}

func (c *IndexController) GetSettlementorderDetail() {
	c.Ctx.ViewData("host", c.Layout.Host)
	c.Ctx.ViewData("appName", c.Layout.AppName)
	c.Ctx.ViewData("globalJsVer", c.Layout.GlobalJsVer)

	c.Ctx.View("/gm/settlementorder/detail.twig")
}

func (c *IndexController) GetMerchantRate() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)
	c.Ctx.ViewData("downTmplUrl", "/resource/merchantRateTmpl.csv")

	c.Ctx.View("/gm/merchant/rate.twig")
}

func (c *IndexController) GetMerchantPaychannel() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)
	c.Ctx.ViewData("downTmplUrl", "/resource/merchantPayTmpl.csv")

	c.Ctx.View("/gm/merchant/paychannel.twig")
}

func (c *IndexController) GetMerchantSettlementchannel() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)
	c.Ctx.ViewData("downTmplUrl", "/resource/merchantPayTmpl.csv")

	c.Ctx.View("/gm/merchant/settlementchannel.twig")
}

func (c *IndexController) GetChannelMerchant() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/channel/merchant.twig")
}

func (c *IndexController) GetChannelRate() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)
	c.Ctx.ViewData("downTmplUrl", "")

	c.Ctx.View("/gm/channel/rate.twig")
}

func (c *IndexController) GetFinance() {
	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/finance/index.twig")
}

func (c *IndexController) GetChartBusinessamount() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/chart/businessamount.twig")
}

func (c *IndexController) GetChartPayorderamount() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/chart/payorderamount.twig")
}

func (c *IndexController) GetChartSettleamount() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/chart/settlementorderamount.twig")
}

func (c *IndexController) GetBalanceadjustment() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/balanceadjustment/index.twig")
}

/*func (c *IndexController) GetRechargeorder() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/rechargeorder/index.twig")
}*/

// ==========================start==================================
func (c *IndexController) GetMerchant() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/merchant/index.twig")
}

func (c *IndexController) GetMerchantUser() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/merchant/user.twig")
}

func (c *IndexController) GetManagerAdminlist() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/manager/adminList.twig")
}

func (c *IndexController) GetManagerChangeloginname() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/manager/changeloginname.twig")
}

func (c *IndexController) GetManagerAdminpwd() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/manager/changeloginpwd.twig")
}

func (c *IndexController) GetManagerBlackusersettlement() {
	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/manager/blackUserSettlement.twig")
}

func (c *IndexController) GetManagerBank() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/manager/bank.twig")
}

func (c *IndexController) GetCheck() {
	var types []string
	err := sqls.DB().Table("system_check_log").Distinct("type").Pluck("type", &types).Error

	if err != nil {
		logrus.Error(err.Error())
	}
	c.Ctx.ViewData("types", types)

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/check/index.twig")
}

func (c *IndexController) GetSettlementorderMakeupcheck() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/settlementorder/makeupcheck.twig")
}

func (c *IndexController) GetPayorderMakeupcheck() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/payorder/makeupcheck.twig")
}

func (c *IndexController) GetMerchantPlatform() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/merchant/platform.twig")
}

func (c *IndexController) GetManagerAdminloginlog() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/manager/adminloginlog.twig")
}

func (c *IndexController) GetManagerAdminactionlog() {

	c.Ctx.ViewData("appName", yamlConfig.Instance.AppName)
	c.Ctx.ViewData("globalJsVer", yamlConfig.Instance.GlobalJsVer)

	c.Ctx.View("/gm/manager/adminactionlog.twig")
}

//==========================end==================================

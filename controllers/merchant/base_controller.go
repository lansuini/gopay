package merchant

import (
	"encoding/json"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/sirupsen/logrus"
	"io"
	"luckypay/config"
	"luckypay/menu"
	"luckypay/model"
	yamlConfig "luckypay/pkg/config"
	"luckypay/utils"
	"os"
	"strings"
	"sync"
	"time"
)

type BaseController struct {
}

/*Session会话管理*/
type SessionMgr struct {
	mCookieName  string       //客户端cookie名称
	mLock        sync.RWMutex //互斥(保证线程安全)
	mMaxLifeTime int64        //垃圾回收时间

	mSessions map[string]*sessions.Session //保存session的指针[sessionID] = session
}

var (
	Menus             menu.MerchantMenu
	LoginMerchantId   int64  //登录者Id
	LoginMerchantNo   string //登陆者昵称
	LoginName         string //登陆者昵称
	LoginUserName     string //登陆者昵称
	LoginAccountId    int64  //登陆者昵称
	LoginIpWhite      string //登陆ip白名单
	LoginMerchantRole int64  //权限
	IsAdministrator   bool   //是否是超级管理员
	//Sess            = sessions.New(sessions.Config{Cookie: "seesionId"})
	//Sessdb, _              = boltdb.New("./sessions.db", os.FileMode(0750))
	cookieNameForSessionID = "luckypay_merchant"
	Sess                   = sessions.New(sessions.Config{Cookie: cookieNameForSessionID, CookieSecureTLS: true, AllowReclaim: true, Expires: 120 * time.Minute, DisableSubdomainPersistence: true})
	Session                *sessions.Session
)

/*
*
初始化
*/
func initialize(ctx iris.Context) {
	osGwd, _ := os.Getwd()
	confPath := osGwd + "/menu/merchant.json"
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
	Host := yamlConfig.Instance.BaseUrl
	if yamlConfig.Instance.Env == "dev" {
		Host = Host + ":" + yamlConfig.Instance.Port
	}
	ctx.ViewData("host", Host)
	ctx.ViewData("menus", Menus)
}

func (bc *BaseController) MerchantMenu() (Menus menu.MerchantMenu) {
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
	return Menus
}

func Auth(ctx iris.Context) {
	initialize(ctx)
	dateTimeFormat := utils.GetFormatTime(time.Now())
	if file, err := os.OpenFile("logs/merchant-"+dateTimeFormat+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		logrus.SetOutput(io.MultiWriter(os.Stdout, file))
		logrus.SetReportCaller(true)
	}
	Sess.UseDatabase(config.NewSessStorage()) //开启session的本地储存，防止长期服务时session丢失
	Session = Sess.Start(ctx)

	if authId, err := Session.GetInt64("accountId"); err == nil && authId != -1 {
		LoginAccountId = authId
		LoginMerchantNo = Session.GetString("merchantNo")
		LoginMerchantId, _ = Session.GetInt64("merchantId")
		LoginName = Session.GetString("loginName")
		LoginUserName = Session.GetString("userName")
		LoginIpWhite = Session.GetString("LoginIpWhite")
		if len(LoginIpWhite) < 6 {
			ctx.Redirect("/checkipwhite")
			return
		}
		//whiteIpList := strings.Split(strings.TrimSpace(LoginIpWhite), ",")
		//userIp := utils.GetRealIp(ctx.Request())
		//fmt.Println(userIp)
		//err = utils.IsInSlice(userIp, whiteIpList)
		//if err != nil {
		//	ctx.Redirect("/checkipwhite")
		//	return
		//}
		if ctx.Path() == "/login" {
			ctx.Redirect("/head")
		}

		//fmt.Println(ctx.Path())
		if Session.GetString("googleAuthSecretKey") != "" && !Session.GetBooleanDefault("googleAuthCheck", false) && (ctx.Path() != "/googleauth") && ctx.Path() != "/api/manager/googleauth" {
			ctx.Redirect("/googleauth")
		}
		if len(Session.GetString("googleAuthSecretKey")) == 0 && ctx.Path() != "/bindgoogleauth" && ctx.Path() != "/logout" && ctx.Path() != "/api/manager/bindgoogleauth" {
			ctx.Redirect("/bindgoogleauth")
		}

		ctx.Next()
	} else { //没有session中的authId
		if (ctx.Path() == "/api/login") || strings.Contains(ctx.Path(), "api/captcha") || (ctx.Path() == "/login") || (ctx.Path() == "/logout") { //登录操作不做权限验证
			initialize(ctx)
			ctx.Next() //继续向下运行
		} else {
			path := ctx.Path()
			if index := strings.Index(path, "api"); index >= 0 {
				ctx.StopWithJSON(200, iris.Map{"success": -1, "result": "登录过期"})
			}
			//跳转到登录
			ctx.Redirect("/login")
		}
	}
}

/**
 * 设置后台的SESSION
 */
func (bc *BaseController) SaveLoginSession(ctx iris.Context, userInfo model.MerchantAccount) {
	//Sess.UseDatabase(Sessdb) //开启session的本地储存，防止长期服务时session丢失
	//session := Sess.Start(ctx)

	Session.Set("merchantNo", userInfo.MerchantNo)
	Session.Set("merchantId", userInfo.MerchantID)
	Session.Set("accountId", userInfo.AccountID)
	Session.Set("loginName", userInfo.LoginName)
	Session.Set("userName", userInfo.UserName)
	Session.Set("loginPwdAlterTime", userInfo.LoginPwdAlterTime)
	Session.Set("googleAuthSecretKey", userInfo.GoogleAuthSecretKey)
	Session.Set("googleAuthCheck", len(userInfo.GoogleAuthSecretKey) == 0)
	//session.Set("merchant_role", userInfo.Role)
	Session.Set("last_login_time", time.Now().Format("2006-01-02 15:04:05"))
	Session.Set("LoginIpWhite", "52.220.205.178,52.220.205.178")
}

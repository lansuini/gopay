package gm

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
	"time"
)

type BaseController struct {
}

var (
	Menus           menu.GmMenu
	LoginAdminId    int64  //登录者Id
	LoginAdmin      string //登陆者昵称
	LoginName       string //登陆者昵称
	LoginAdminRole  int64  //权限
	IsAdministrator bool   //是否是超级管理员
	//Sess      = sessions.New(sessions.Config{Cookie: "seesionId"})
	//Sessdb, _              = boltdb.New("./sessions.db", os.FileMode(0750))
	cookieNameForSessionID = "gm_session"
	Sess                   = sessions.New(sessions.Config{Cookie: cookieNameForSessionID, Expires: 120 * time.Minute, CookieSecureTLS: true, AllowReclaim: true, DisableSubdomainPersistence: true})
	Session                *sessions.Session
)

/*
*
初始化
*/
func initialize(ctx iris.Context) {
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
	Host := yamlConfig.Instance.BaseUrl
	if yamlConfig.Instance.Env == "dev" {
		Host = Host + ":" + yamlConfig.Instance.Port
	}
	//fmt.Println(Host)
	ctx.ViewData("host", Host)
	ctx.ViewData("menus", Menus)
	ctx.ViewData("userName", LoginName)

	//Rbac(ctx)
	//SearchTids = DataAccess(ctx)
}

func (bc *BaseController) GmMenu() (Menus menu.GmMenu) {
	osGwd, _ := os.Getwd()
	confPath := osGwd + "/menu/gm.json"
	file, err := os.Open(confPath)
	if err != nil {
		logrus.Error("error opening json file-", confPath)
		return
	}
	// 关闭文件
	defer file.Close()
	// NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
	decoder := json.NewDecoder(file)
	//Decode从输入流读取下一个json编码值并保存在v指向的值里
	err = decoder.Decode(&Menus)
	if err != nil {
		logrus.Error("decoder.Decode-", confPath)
		return
	}

	return Menus
}

func Auth(ctx iris.Context) {
	initialize(ctx)

	dateTimeFormat := utils.GetFormatTime(time.Now())
	if file, err := os.OpenFile("logs/gm-"+dateTimeFormat+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		logrus.SetOutput(io.MultiWriter(os.Stdout, file))
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetReportCaller(true)
	}
	//initialize(ctx)
	//pathArr := strings.Split(ctx.Path(), "/")
	//RequestApp = pathArr[1]        //app
	//RequestController = pathArr[2] //controller
	//RequestAction = pathArr[3]     //action
	//Sess.UseDatabase(Sessdb) //开启session的本地储存，防止长期服务时session丢失
	Sess.UseDatabase(config.NewSessStorage()) //开启session的本地储存，防止长期服务时session丢失
	Session = Sess.Start(ctx)
	if authId, err := Session.GetInt64("login_admin_id"); err == nil && authId != -1 {
		LoginAdminId = authId
		LoginName = Session.GetString("login_name")
		LoginAdminRole, _ = Session.GetInt64("admin_role")
		if LoginAdminRole == 1 { //超级管理员
			LoginAdmin = Session.GetString("login_admin") + "(超级管理员)"
			IsAdministrator = true
		} else {
			LoginAdmin = Session.GetString("login_admin")
			IsAdministrator = false
		}
		if ctx.Path() == "/login" {
			ctx.Redirect("/head")
		}
		if Session.GetString("googleAuthSecretKey") != "" && !Session.GetBooleanDefault("googleAuthCheck", false) && (ctx.Path() != "/googleauth") && ctx.Path() != "/api/manager/googleauth" {
			ctx.Redirect("/googleauth")
		}
		if len(Session.GetString("googleAuthSecretKey")) == 0 && ctx.Path() != "/manager/bindgoogleauth" && ctx.Path() != "/api/manager/bindgoogleauth" {
			ctx.Redirect("/manager/bindgoogleauth")
		}
		ctx.Next()

		if ctx.Path() == "/login" {
			ctx.Redirect("/head")
		}
	} else { //没有session中的authId
		if (ctx.Path() == "/api/manager/login") || strings.Contains(ctx.Path(), "api/captcha") || (ctx.Path() == "/login") { //登录操作不做权限验证
			initialize(ctx)
			ctx.Next() //继续向下运行
		} else {
			//跳转到登录
			if index := strings.Index(ctx.Path(), "api"); index >= 0 {
				ctx.StopWithJSON(200, iris.Map{"success": -1, "result": "登录过期"})
			}
			ctx.Redirect("/login")
		}
	}
}

/**
 * 设置后台的SESSION
 */
func (bc *BaseController) SaveLoginSession(ctx iris.Context, userInfo *model.SystemAccount) {
	//session := Sess.Start(ctx)
	Session.Set("login_admin_id", int(userInfo.Id))
	Session.Set("login_admin", userInfo.UserName)
	Session.Set("login_name", userInfo.LoginName)
	Session.Set("admin_role", userInfo.Role)
	Session.Set("googleAuthSecretKey", userInfo.GoogleAuthSecretKey)
	Session.Set("googleAuthCheck", len(userInfo.GoogleAuthSecretKey) == 0)

	Session.Set("last_login_time", time.Now().Format("2006-01-02 15:04:05"))
	if userInfo.Role == 1 { //1是超级管理员
		//session.Set(appconf.Rbac.AdminAuthKey, true)
	}
}

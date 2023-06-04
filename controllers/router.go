package controllers

import (
	"context"
	"fmt"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/mvc"
	"github.com/mlogclub/simple/web"
	"github.com/sirupsen/logrus"
	"io"
	"luckypay/controllers/cb"
	"luckypay/controllers/ctrl"
	"luckypay/controllers/gate"
	"luckypay/controllers/gm"
	"luckypay/controllers/merchant"
	"luckypay/controllers/public"
	"luckypay/pkg/config"
	"luckypay/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Router() {
	app := iris.New()
	//app.Logger().Install(logrus.StandardLogger())
	app.Logger().SetLevel("warn")
	app.Use(recover.New())
	app.Use(logger.New())
	/*app.Use(logger.New(logger.Config{
		// 是否记录状态码,默认false
		Status: true,
		// 是否记录远程IP地址,默认false
		IP: true,
		// 是否呈现HTTP谓词,默认false
		Method: true,
		// 是否记录请求路径,默认true
		Path: true,
		// 是否开启查询追加,默认false
		Query: true,
	}))*/
	//app.Validator = validator.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)

	app.OnAnyErrorCode(func(ctx iris.Context) {
		//path := ctx.Path()
		//ctx.Writef("From: %s", ctx.Path(), ctx.Domain())
		var err error
		//if strings.Contains(path, "/api/admin/") {
		err = ctx.JSON(web.JsonErrorCode(ctx.GetStatusCode(), "Http error"))
		//}
		if err != nil {
			//logrus.SetReportCaller(true)
			//logrus.WithFields(logrus.Fields{
			//	"uri": ctx.Path(),
			//	"ip":  ctx.Request().Header.Get("X-Real-IP"),
			//})
			logrus.Error(err)
		}
	})

	app.HandleDir("/static", iris.Dir("./static"))
	app.HandleDir("/resource", iris.Dir("./resource"))
	//tmpl := iris.HTML("./views", ".html")
	tmpl := iris.Django("./views", ".twig")
	// Set custom delimeters.
	//tmpl.Delims("{{", "}}")
	// Enable re-build on local template files changes.
	tmpl.Reload(true)

	app.RegisterView(tmpl)

	//app.Any("/", func(i iris.Context) {
	//	_, _ = i.HTML("<h1>Powered by luckypay</h1>")
	//})

	//appSubdomainGM := app.Subdomain("gm")
	//mvcAppGM := mvc.New(appSubdomainGM)
	//mvcAppGM.Handle(new(gm.IndexController))
	//app.SubdomainRedirect(app, gmd)

	//gmPart := app.Party("/", gm.Auth)
	mvc.Configure(app.Party("gm.", gm.Auth), func(m *mvc.Application) {

		m.Party("/api").Handle(new(gm.ApiController))
		m.Party("/api/payorder/").Handle(new(gm.PayOrderController))
		m.Party("/api/settlementorder/").Handle(new(gm.SettlementOrderController))
		m.Party("/api/merchant/").Handle(new(gm.MerchantController))
		m.Party("/api/channel/").Handle(new(gm.ChannelController))
		m.Party("/api/manager/").Handle(new(gm.ManagerController))
		m.Party("/api/finance/").Handle(new(gm.FinanceController))
		m.Party("/api/chart/").Handle(new(gm.ChartController))
		m.Party("/api/check/").Handle(new(gm.CheckController))
		m.Party("/api/captcha").Handle(new(gm.CaptchaController))
		m.Handle(new(gm.IndexController))

	})
	mvc.Configure(app.Party("/public"), func(m *mvc.Application) {

		m.Party("/captcha").Handle(new(public.CaptchaController))

	})

	mvc.Configure(app.Party("merchant.", merchant.Auth), func(m *mvc.Application) {
		m.Party("/api/captcha").Handle(new(merchant.CaptchaController))
		m.Party("/api").Handle(new(merchant.ApiController))
		m.Party("/api/manager").Handle(new(merchant.ManagerController))
		m.Party("/api/payorder/").Handle(new(merchant.PayOrderController))
		m.Party("/api/settlementorder/").Handle(new(merchant.SettlementOrderController))
		m.Handle(new(merchant.IndexController))

	})
	//path := "gate." + config.Instance.HostSet.DOMAIN + "/paygateway"
	path := "gate."
	mvc.Configure(app.Party(path), func(m *mvc.Application) {
		dateTimeFormat := utils.GetFormatTime(time.Now())
		if file, err := os.OpenFile("logs/gate-"+dateTimeFormat+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logrus.SetOutput(io.MultiWriter(os.Stdout, file))
			logrus.SetReportCaller(true)
			logrus.SetFormatter(&logrus.JSONFormatter{})
		} else {
			logrus.SetOutput(os.Stdout)
			logrus.Error(err)
		}
		//m.Party("/index").Handle(new(gate.IndexController))
		m.Party("/paygateway").Handle(new(gate.PayController))

	})
	//mvc.Configure(app.Party("test."), func(m *mvc.Application) {
	//
	//	/*callback.Use(myAuthMiddlewareHandler)
	//	http://localhost:8080/users/42/profile
	//	callback.Get("/{id:int}/profile", userProfileHandler)
	//	http://localhost:8080/users/messages/1
	//	users.Get("/messages/{id:int}", userMessageHandler)*/
	//	m.Router.Any()
	//	m.Any("/pay/callback/{platformOrderNo:string}",cb.PayCallBack)
	//	m.Any("/check",cb.GetCheck)
	//
	//})
	app.PartyFunc("cb.", func(callback iris.Party) {

		/*callback.Use(myAuthMiddlewareHandler)
		http://localhost:8080/users/42/profile
		callback.Get("/{id:int}/profile", userProfileHandler)
		http://localhost:8080/users/messages/1
		users.Get("/messages/{id:int}", userMessageHandler)*/
		//callback.Party("/d").Handle("","test",cb.GetCheck)
		callback.Any("/pay/callback/{platformOrderNo:string}", cb.PayCallBack)
		//callback.Any("/paycallback", new(cb.IndexController).GetTestpaycallback)
		//callback.Any("/test", new(cb.IndexController).GetTest)

	})
	//mvc.Configure(app.Party("cb."), func(m *mvc.Application) {
	//
	//	m.Handle(new(cb.IndexController))
	//
	//})

	/*app.Get("/api/img/proxy", func(i iris.Context) {
		url := i.FormValue("url")
		resp, err := resty.New().R().Get(url)
		i.Header("Content-Type", "image/jpg")
		if err == nil {
			_, _ = i.Write(resp.Body())
		} else {
			logrus.Error(err)
		}
	})*/
	host := "127.0.0.1" + ":" + config.Instance.HostSet.PORT

	if err := app.Listen(host,
		iris.WithConfiguration(iris.Configuration{
			DisableStartupLog:                 false,
			DisableInterruptHandler:           false,
			DisablePathCorrection:             false,
			EnablePathEscape:                  false,
			FireMethodNotAllowed:              false,
			DisableBodyConsumptionOnUnmarshal: false,
			DisableAutoFireStatusCode:         false,
			EnableOptimizations:               true,
			TimeFormat:                        "2016-01-02 15:04:05",
			Charset:                           "UTF-8",
		}),
	); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}

	//app.Run(iris.Addr(host))
}

func RouteGM() {
	app := iris.New()

	//app.Logger().Install(logrus.StandardLogger())
	app.Logger().SetLevel("warn")
	app.Use(recover.New())
	app.Use(logger.New())
	/*app.Use(logger.New(logger.Config{
		// 是否记录状态码,默认false
		Status: true,
		// 是否记录远程IP地址,默认false
		IP: true,
		// 是否呈现HTTP谓词,默认false
		Method: true,
		// 是否记录请求路径,默认true
		Path: true,
		// 是否开启查询追加,默认false
		Query: true,
	}))*/
	//app.Validator = validator.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)

	app.OnAnyErrorCode(func(ctx iris.Context) {
		//path := ctx.Path()
		//ctx.Writef("From: %s", ctx.Path(), ctx.Domain())
		var err error
		//if strings.Contains(path, "/api/admin/") {
		err = ctx.JSON(web.JsonErrorCode(ctx.GetStatusCode(), "Http error"))
		//}
		if err != nil {
			//logrus.SetReportCaller(true)
			//logrus.WithFields(logrus.Fields{
			//	"uri": ctx.Path(),
			//	"ip":  ctx.Request().Header.Get("X-Real-IP"),
			//})
			logrus.Error(err)
		}
	})

	app.HandleDir("/static", iris.Dir("./static"))
	app.HandleDir("/resource", iris.Dir("./resource"))
	//tmpl := iris.HTML("./views", ".html")
	tmpl := iris.Django("./views", ".twig")
	// Set custom delimeters.
	//tmpl.Delims("{{", "}}")
	// Enable re-build on local template files changes.
	tmpl.Reload(true)

	app.RegisterView(tmpl)

	//app.Any("/", func(i iris.Context) {
	//	_, _ = i.HTML("<h1>Powered by luckypay</h1>")
	//})

	mvc.Configure(app.Party("gm.", gm.Auth), func(m *mvc.Application) {

		m.Party("/api").Handle(new(gm.ApiController))
		m.Party("/api/payorder/").Handle(new(gm.PayOrderController))
		m.Party("/api/settlementorder/").Handle(new(gm.SettlementOrderController))
		m.Party("/api/merchant/").Handle(new(gm.MerchantController))
		m.Party("/api/channel/").Handle(new(gm.ChannelController))
		m.Party("/api/manager/").Handle(new(gm.ManagerController))
		m.Party("/api/finance/").Handle(new(gm.FinanceController))
		m.Party("/api/chart/").Handle(new(gm.ChartController))
		m.Party("/api/check/").Handle(new(gm.CheckController))
		m.Party("/api/captcha").Handle(new(gm.CaptchaController))
		m.Handle(new(gm.IndexController))

	})
	mvc.Configure(app.Party("/public"), func(m *mvc.Application) {

		m.Party("/captcha").Handle(new(public.CaptchaController))

	})

	host := "127.0.0.1" + ":" + config.Instance.HostSet.GM_PORT

	//app.Run(iris.Addr(host), iris.WithoutInterruptHandler)
	server := &http.Server{
		Addr:           host,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	runner := iris.Server(server)
	app.Run(runner, iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		EnableOptimizations:               true,
		TimeFormat:                        "2016-01-02 15:04:05",
		Charset:                           "UTF-8",
	}))
	gracefulExitServer(server)
}

func RouteMerchant() {
	app := iris.New()

	//app.Logger().Install(logrus.StandardLogger())
	app.Logger().SetLevel("warn")
	app.Use(recover.New())
	app.Use(logger.New())
	/*app.Use(logger.New(logger.Config{
		// 是否记录状态码,默认false
		Status: true,
		// 是否记录远程IP地址,默认false
		IP: true,
		// 是否呈现HTTP谓词,默认false
		Method: true,
		// 是否记录请求路径,默认true
		Path: true,
		// 是否开启查询追加,默认false
		Query: true,
	}))*/
	//app.Validator = validator.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)

	app.OnAnyErrorCode(func(ctx iris.Context) {
		//path := ctx.Path()
		//ctx.Writef("From: %s", ctx.Path(), ctx.Domain())
		var err error
		//if strings.Contains(path, "/api/admin/") {
		err = ctx.JSON(web.JsonErrorCode(ctx.GetStatusCode(), "Http error"))
		//}
		if err != nil {
			//logrus.SetReportCaller(true)
			//logrus.WithFields(logrus.Fields{
			//	"uri": ctx.Path(),
			//	"ip":  ctx.Request().Header.Get("X-Real-IP"),
			//})
			logrus.Error(err)
		}
	})

	app.HandleDir("/static", iris.Dir("./static"))
	app.HandleDir("/resource", iris.Dir("./resource"))
	//tmpl := iris.HTML("./views", ".html")
	tmpl := iris.Django("./views", ".twig")
	// Set custom delimeters.
	//tmpl.Delims("{{", "}}")
	// Enable re-build on local template files changes.
	tmpl.Reload(true)

	app.RegisterView(tmpl)

	app.Any("merchant//", func(i iris.Context) {
		_, _ = i.HTML("<h1>Powered by luckypay</h1>")
	})

	mvc.Configure(app.Party("merchant.", merchant.Auth), func(m *mvc.Application) {
		m.Party("/api/captcha").Handle(new(merchant.CaptchaController))
		m.Party("/api").Handle(new(merchant.ApiController))
		m.Party("/api/manager").Handle(new(merchant.ManagerController))
		m.Party("/api/payorder/").Handle(new(merchant.PayOrderController))
		m.Party("/api/settlementorder/").Handle(new(merchant.SettlementOrderController))
		m.Handle(new(merchant.IndexController))

	})
	mvc.Configure(app.Party("/public"), func(m *mvc.Application) {

		m.Party("/captcha").Handle(new(public.CaptchaController))

	})

	host := "127.0.0.1" + ":" + config.Instance.HostSet.MERCHANT_PORT

	//if err := app.Listen(host,
	//	iris.WithConfiguration(iris.Configuration{
	//		DisableStartupLog:                 false,
	//		DisableInterruptHandler:           false,
	//		DisablePathCorrection:             false,
	//		EnablePathEscape:                  false,
	//		FireMethodNotAllowed:              false,
	//		DisableBodyConsumptionOnUnmarshal: false,
	//		DisableAutoFireStatusCode:         false,
	//		EnableOptimizations:               true,
	//		TimeFormat:                        "2016-01-02 15:04:05",
	//		Charset:                           "UTF-8",
	//	}),
	//); err != nil {
	//	logrus.Error(err)
	//	os.Exit(-1)
	//}

	server := &http.Server{
		Addr:           host,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	runner := iris.Server(server)
	if err := app.Run(runner, iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		EnableOptimizations:               true,
		TimeFormat:                        "2016-01-02 15:04:05",
		Charset:                           "UTF-8",
	})); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	gracefulExitServer(server)
}

func RouteCB() {
	app := iris.New()

	//app.Logger().Install(logrus.StandardLogger())
	app.Logger().SetLevel("warn")
	app.Use(recover.New())
	app.Use(logger.New())
	/*app.Use(logger.New(logger.Config{
		// 是否记录状态码,默认false
		Status: true,
		// 是否记录远程IP地址,默认false
		IP: true,
		// 是否呈现HTTP谓词,默认false
		Method: true,
		// 是否记录请求路径,默认true
		Path: true,
		// 是否开启查询追加,默认false
		Query: true,
	}))*/
	//app.Validator = validator.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)

	app.OnAnyErrorCode(func(ctx iris.Context) {
		//path := ctx.Path()
		//ctx.Writef("From: %s", ctx.Path(), ctx.Domain())
		var err error
		//if strings.Contains(path, "/api/admin/") {
		err = ctx.JSON(web.JsonErrorCode(ctx.GetStatusCode(), "Http error"))
		//}
		if err != nil {
			//logrus.SetReportCaller(true)
			//logrus.WithFields(logrus.Fields{
			//	"uri": ctx.Path(),
			//	"ip":  ctx.Request().Header.Get("X-Real-IP"),
			//})
			logrus.Error(err)
		}
	})

	app.HandleDir("/static", iris.Dir("./static"))
	app.HandleDir("/resource", iris.Dir("./resource"))
	//tmpl := iris.HTML("./views", ".html")
	tmpl := iris.Django("./views", ".twig")
	// Set custom delimeters.
	//tmpl.Delims("{{", "}}")
	// Enable re-build on local template files changes.
	tmpl.Reload(true)

	app.RegisterView(tmpl)

	app.Any("/", func(i iris.Context) {
		_, _ = i.HTML("<h1>Powered by luckypay</h1>")
	})
	app.PartyFunc("cb.", func(callback iris.Party) {
		dateTimeFormat := utils.GetFormatTime(time.Now())
		if file, err := os.OpenFile("logs/cb-"+dateTimeFormat+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logrus.SetOutput(io.MultiWriter(os.Stdout, file))
			logrus.SetReportCaller(true)
			logrus.SetFormatter(&logrus.JSONFormatter{})
		} else {
			logrus.SetOutput(os.Stdout)
			logrus.Error(err)
		}
		/*callback.Use(myAuthMiddlewareHandler)
		http://localhost:8080/users/42/profile
		callback.Get("/{id:int}/profile", userProfileHandler)
		http://localhost:8080/users/messages/1
		users.Get("/messages/{id:int}", userMessageHandler)*/
		callback.Any("/settlement/callback/{platformOrderNo:string}", cb.SettlementCallBack)
		callback.Any("/pay/callback/{platformOrderNo:string}", cb.PayCallBack)

	})
	host := "127.0.0.1" + ":" + config.Instance.HostSet.CB_PORT

	server := &http.Server{
		Addr:           host,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	runner := iris.Server(server)
	if err := app.Run(runner, iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		EnableOptimizations:               true,
		TimeFormat:                        "2016-01-02 15:04:05",
		Charset:                           "UTF-8",
	})); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	gracefulExitServer(server)
}

func RouteGate() {
	app := iris.New()

	//app.Logger().Install(logrus.StandardLogger())
	app.Logger().SetLevel("warn")
	app.Use(recover.New())
	app.Use(logger.New())
	/*app.Use(logger.New(logger.Config{
		// 是否记录状态码,默认false
		Status: true,
		// 是否记录远程IP地址,默认false
		IP: true,
		// 是否呈现HTTP谓词,默认false
		Method: true,
		// 是否记录请求路径,默认true
		Path: true,
		// 是否开启查询追加,默认false
		Query: true,
	}))*/
	//app.Validator = validator.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)

	app.OnAnyErrorCode(func(ctx iris.Context) {
		//path := ctx.Path()
		//ctx.Writef("From: %s", ctx.Path(), ctx.Domain())
		var err error
		//if strings.Contains(path, "/api/admin/") {
		err = ctx.JSON(web.JsonErrorCode(ctx.GetStatusCode(), "Http error"))
		//}
		if err != nil {
			//logrus.SetReportCaller(true)
			//logrus.WithFields(logrus.Fields{
			//	"uri": ctx.Path(),
			//	"ip":  ctx.Request().Header.Get("X-Real-IP"),
			//})
			logrus.Error(err)
		}
	})

	app.HandleDir("/static", iris.Dir("./static"))
	app.HandleDir("/resource", iris.Dir("./resource"))
	//tmpl := iris.HTML("./views", ".html")
	tmpl := iris.Django("./views", ".twig")
	// Set custom delimeters.
	//tmpl.Delims("{{", "}}")
	// Enable re-build on local template files changes.
	tmpl.Reload(true)

	app.RegisterView(tmpl)

	app.Any("/", func(i iris.Context) {
		_, _ = i.HTML("<h1>Powered by luckypay</h1>")
	})

	path := "gate."
	mvc.Configure(app.Party(path), func(m *mvc.Application) {
		dateTimeFormat := utils.GetFormatTime(time.Now())
		if file, err := os.OpenFile("logs/gate-"+dateTimeFormat+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logrus.SetOutput(io.MultiWriter(os.Stdout, file))
			logrus.SetReportCaller(true)
			logrus.SetFormatter(&logrus.JSONFormatter{})
		} else {
			logrus.SetOutput(os.Stdout)
			logrus.Error(err)
		}
		//m.Party("/index").Handle(new(gate.IndexController))
		m.Party("/paygateway").Handle(new(gate.PayController))

	})

	host := "127.0.0.1" + ":" + config.Instance.HostSet.GATE_PORT
	server := &http.Server{
		Addr:           host,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	runner := iris.Server(server)
	if err := app.Run(runner, iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		EnableOptimizations:               true,
		TimeFormat:                        "2016-01-02 15:04:05",
		Charset:                           "UTF-8",
	}), iris.WithoutBodyConsumptionOnUnmarshal); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	//gracefulExitServer(server)
}

func RouteMain() {
	app := iris.New()

	//app.Logger().Install(logrus.StandardLogger())
	app.Logger().SetLevel("warn")
	app.Use(recover.New())
	app.Use(logger.New())
	/*app.Use(logger.New(logger.Config{
		// 是否记录状态码,默认false
		Status: true,
		// 是否记录远程IP地址,默认false
		IP: true,
		// 是否呈现HTTP谓词,默认false
		Method: true,
		// 是否记录请求路径,默认true
		Path: true,
		// 是否开启查询追加,默认false
		Query: true,
	}))*/
	//app.Validator = validator.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)

	app.OnAnyErrorCode(func(ctx iris.Context) {
		//path := ctx.Path()
		//ctx.Writef("From: %s", ctx.Path(), ctx.Domain())
		var err error
		//if strings.Contains(path, "/api/admin/") {
		err = ctx.JSON(web.JsonErrorCode(ctx.GetStatusCode(), "Http error"))
		//}
		if err != nil {
			//logrus.SetReportCaller(true)
			//logrus.WithFields(logrus.Fields{
			//	"uri": ctx.Path(),
			//	"ip":  ctx.Request().Header.Get("X-Real-IP"),
			//})
			logrus.Error(err)
		}
	})

	app.PartyFunc("/", func(callback iris.Party) {
		dateTimeFormat := utils.GetFormatTime(time.Now())
		if file, err := os.OpenFile("logs/main-"+dateTimeFormat+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logrus.SetOutput(io.MultiWriter(os.Stdout, file))
			logrus.SetReportCaller(true)
			logrus.SetFormatter(&logrus.JSONFormatter{})
		} else {
			logrus.SetOutput(os.Stdout)
			logrus.Error(err)
		}

		callback.Any("/refreshCache", ctrl.RefreshCache)
		callback.Any("/refreshCacheOne", ctrl.RefreshCacheOne)
		//代付查询
		callback.Any("/cacheSettlementFetch", ctrl.CacheSettlementFetch)
		callback.Any("/pushSettlementFetch", ctrl.PushSettlementFetch)
		callback.Any("/popSettlementFetch", ctrl.PopSettlementFetch)
		//代付回调
		callback.Any("/pushSettlementNotify", ctrl.PushSettlementNotify)
		callback.Any("/popSettlementNotify", ctrl.PopSettlementNotify)
		callback.Any("/cacheSettlementNotify", ctrl.CacheSettlementNotify)
		//支付回调
		callback.Any("/cachePayNotify", ctrl.CachePayNotify)
		callback.Any("/pushPayNotify", ctrl.PushPayNotify)
		callback.Any("/popPayNotify", ctrl.PopPayNotify)

	})

	host := "127.0.0.1" + ":" + config.Instance.HostSet.PORT
	server := &http.Server{
		Addr:           host,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	runner := iris.Server(server)
	if err := app.Run(runner, iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		EnableOptimizations:               true,
		TimeFormat:                        "2016-01-02 15:04:05",
		Charset:                           "UTF-8",
	})); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	gracefulExitServer(server)
}

func RouteNotify() {
	app := iris.New()

	//app.Logger().Install(logrus.StandardLogger())
	app.Logger().SetLevel("warn")
	app.Use(recover.New())
	app.Use(logger.New())
	/*app.Use(logger.New(logger.Config{
		// 是否记录状态码,默认false
		Status: true,
		// 是否记录远程IP地址,默认false
		IP: true,
		// 是否呈现HTTP谓词,默认false
		Method: true,
		// 是否记录请求路径,默认true
		Path: true,
		// 是否开启查询追加,默认false
		Query: true,
	}))*/
	//app.Validator = validator.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)

	app.OnAnyErrorCode(func(ctx iris.Context) {
		//path := ctx.Path()
		//ctx.Writef("From: %s", ctx.Path(), ctx.Domain())
		var err error
		//if strings.Contains(path, "/api/admin/") {
		err = ctx.JSON(web.JsonErrorCode(ctx.GetStatusCode(), "Http error"))
		//}
		if err != nil {
			//logrus.SetReportCaller(true)
			//logrus.WithFields(logrus.Fields{
			//	"uri": ctx.Path(),
			//	"ip":  ctx.Request().Header.Get("X-Real-IP"),
			//})
			logrus.Error(err)
		}
	})

	app.PartyFunc("/", func(callback iris.Party) {
		dateTimeFormat := utils.GetFormatTime(time.Now())
		if file, err := os.OpenFile("logs/notify-"+dateTimeFormat+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logrus.SetOutput(io.MultiWriter(os.Stdout, file))
			logrus.SetReportCaller(true)
			logrus.SetFormatter(&logrus.JSONFormatter{})
		} else {
			logrus.SetOutput(os.Stdout)
			logrus.Error(err)
		}

		callback.Any("/test", ctrl.Test)
		callback.Any("/readMemStats", ctrl.ReadMemStats)
		callback.Any("/reEncrypt", ctrl.ReEncrypt)
		callback.Any("/refreshCache", ctrl.RefreshCache)
		callback.Any("/refreshCacheOne", ctrl.RefreshCacheOne)
		//代付查询
		callback.Any("/cacheSettlementFetch", ctrl.CacheSettlementFetch)
		callback.Any("/pushSettlementFetch", ctrl.PushSettlementFetch)
		callback.Any("/popSettlementFetch", ctrl.PopSettlementFetch)
		//代付回调
		callback.Any("/pushSettlementNotify", ctrl.PushSettlementNotify)
		callback.Any("/popSettlementNotify", ctrl.PopSettlementNotify)
		callback.Any("/cacheSettlementNotify", ctrl.CacheSettlementNotify)
		//支付回调
		callback.Any("/cachePayNotify", ctrl.CachePayNotify)
		callback.Any("/pushPayNotify", ctrl.PushPayNotify)
		callback.Any("/popPayNotify", ctrl.PopPayNotify)

	})

	host := "127.0.0.1" + ":" + config.Instance.HostSet.NOTIFY_PORT
	server := &http.Server{
		Addr:           host,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	runner := iris.Server(server)
	if err := app.Run(runner, iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		EnableOptimizations:               true,
		TimeFormat:                        "2016-01-02 15:04:05",
		Charset:                           "UTF-8",
	})); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	gracefulExitServer(server)
}

func gracefulExitServer(server *http.Server) {
	// 使用缓存的channel；建议用1；详情看Uber Go style；其他情况酌情使用缓冲大小
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	// 接收信号
	sig := <-ch
	fmt.Println("获取一个系统信号", sig)
	// 设置当前时间
	nowTime := time.Now()
	// 设置为5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// 最后关闭
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("-----exited-----", time.Since(nowTime))
}

package main

import (
	"luckypay/controllers"
	"luckypay/model"
	"luckypay/pkg/common"
	"luckypay/pkg/config"
	"luckypay/scheduler"
	"luckypay/utils"

	//_ "luckypay/services/eventhandler"
	"flag"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"io"
	"os"
	"time"
)

var ConfigFile = flag.String("config", "./config.yaml", "配置文件路径")

func init() {
	flag.Parse()

	// 初始化配置
	conf := config.Init(*ConfigFile)
	dateTimeFormat := utils.GetFormatTime(time.Now())
	logFile := conf.LogPath + "notify-" + dateTimeFormat + ".log"
	// 初始化日志
	if file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		logrus.SetOutput(io.MultiWriter(os.Stdout, file))
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetReportCaller(true)
	} else {
		logrus.SetOutput(os.Stdout)
		logrus.Error(err)
	}

	// 连接数据库
	gormConf := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",   // table name prefix, table for `User` would be `t_users`
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
			//NoLowerCase:   true,                              // skip the snake_casing of names
			//NameReplacer:  strings.NewReplacer("CID", "Cid"), // use name replacer to change struct/field name before convert it to db name
		},
		Logger: logger.New(logrus.StandardLogger(), logger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  true,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,

			//}).LogMode(logger.Info),
		}).LogMode(logger.Error),
	}
	if err := sqls.Open(conf.DB.Url, gormConf, conf.DB.MaxIdleConns, conf.DB.MaxOpenConns, model.Models...); err != nil {
		logrus.Error(err)
	}
}

func main() {

	if common.IsProd() {
		// 开启定时任务
		scheduler.NotifyStart()
		//scheduler.Test()
	}
	controllers.RouteNotify()
}

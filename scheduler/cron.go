package scheduler

import (
	"luckypay/pkg/config"
	"luckypay/services"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	//"luckypay/services"
)

var maxTasks = 20

func MainStart() {
	c := cron.New()

	// 刷新系统缓存RefreshCache
	/*addCronFunc(c, "@every 30m", func() {
		logrus.Info("Refresh System Cache")
		go cache.AdminCache.RefreshCache()
		go cache.MerchantCache.RefreshCache()
		go cache.MerchantCache.SetCacheMerchantData()
		go cache.MerchantAmount.RefreshCache()
		go cache.MerchantAccount.RefreshCache()
		go cache.MerchantRate.RefreshCache()
		go cache.MerchantChannel.RefreshCache()
		go cache.MerchantChannelSettlement.RefreshCache()
		go cache.ChannelMerchant.RefreshCache()
		go cache.ChannelMerchantRate.RefreshCache()
	})*/

	//每5s跑一次，代付结果查询
	addCronFunc(c, "*/5 * * * * *", func() {
		tasks := config.Instance.Crontab.SETTLE_FETCH_TASK
		if tasks > maxTasks {
			tasks = maxTasks
		}
		for i := 0; i < tasks; i++ {
			go services.SettlementFetch.Pop()
		}
	})
	//3分钟跑一次，未完成代付订单推送队列
	addCronFunc(c, "@every 2m", func() {
		isAutoPush := config.Instance.Crontab.SETTLE_FETCH_AUTO_PUSH
		if isAutoPush {
			go services.SettlementFetch.AutoPushTask()
		}
	})

	//商户报表凌晨1点
	addCronFunc(c, "0 0 1 ? * *", func() {
		go services.MerchantService.RunDayStats()
		go services.ChannelMerchant.RunDayStats()
	})
	// Generate sitemap
	//addCronFunc(c, "0 0 4 ? * *", func() {
	//	sitemap.Generate()
	//})

	c.Start()
}

func NotifyStart() {
	c := cron.New()
	logrus.Info("-----------------------NotifyStart-----------------------------")
	addCronFunc(c, "*/5 * * * * *", func() {
		tasks := config.Instance.Crontab.SETTLE_NOTIFY_TASK
		if tasks > maxTasks {
			tasks = maxTasks
		}
		for i := 0; i < tasks; i++ {
			go services.SettlementNotify.Pop()
		}
		tasks = config.Instance.Crontab.PAY_NOTIFY_TASK
		if tasks > maxTasks {
			tasks = maxTasks
		}
		for i := 0; i < tasks; i++ {
			go services.PayNotify.Pop()
		}

	})
	addCronFunc(c, "@every 3m", func() {
		isAutoPush := config.Instance.Crontab.SETTLE_NOTIFY_AUTO_PUSH
		if isAutoPush {
			logrus.Info("-----------------------SettlementNotify.AutoPushTask-----------------------------")
			go services.SettlementNotify.AutoPushTask()
		}
		isAutoPush = config.Instance.Crontab.SETTLE_NOTIFY_AUTO_PUSH
		if isAutoPush {
			logrus.Info("-----------------------PayNotify.AutoPushTask-----------------------------")
			go services.PayNotify.AutoPushTask()
		}
	})
	c.Start()
}

//func Test() {
//	c := cron.New()
//
//	addCronFunc(c, "*/20 * * * * *", func() {
//		logrus.Info("-----------------------TesT-----------------------------")
//		tasks := config.Instance.Crontab.PAY_NOTIFY_TASK
//		if tasks > maxTasks {
//			tasks = maxTasks
//		}
//		for i := 0; i < tasks; i++ {
//			logrus.Info(time.Now().Format("2006-01-02 15:04:05"))
//		}
//	})
//	c.Start()
//
//}

func addCronFunc(c *cron.Cron, sepc string, cmd func()) {
	err := c.AddFunc(sepc, cmd)
	if err != nil {
		logrus.Error(err)
	}
}

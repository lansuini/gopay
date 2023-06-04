package ctrl

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"log"
	"luckypay/cache"
	"luckypay/config"
	config2 "luckypay/pkg/config"
	"luckypay/services"
	"luckypay/utils"
	"runtime"
)

type IndexController struct {
	Ctx      iris.Context
	validate *validator.Validate
}

func Test(ctx iris.Context) {
	viper := config.Viper
	pay_notify_task := viper.GetString("crontab.pay_notify_task")
	ctx.WriteString(pay_notify_task)
	return
	fmt.Println(config2.Instance.StaticPath)
	tasks := config.Viper.GetString("crontab.pay_notify_task")
	ctx.WriteString(tasks)
	return
}

func RefreshCache(ctx iris.Context) {
	ctx.WriteString("hell world")
	cache.SystemCache.RefreshCache()
	return
}

func RefreshCacheOne(ctx iris.Context) {
	cacheService := ctx.URLParamTrim("cache_service")
	ctx.WriteString("running " + cacheService + "...")
	switch cacheService {
	case "AdminCache":
		cache.AdminCache.RefreshCache()
	case "MerchantCache":
		cache.MerchantCache.RefreshCache()
	case "MerchantCaches":
		cache.MerchantCache.SetCacheMerchantData()
	case "MerchantAmount":
		cache.MerchantAmount.RefreshCache()
	case "MerchantAccount":
		cache.MerchantAccount.RefreshCache()
	case "MerchantRate":
		cache.MerchantRate.RefreshCache()
	case "MerchantChannel":
		cache.MerchantChannel.RefreshCache()
	case "ChannelMerchant":
		cache.ChannelMerchant.RefreshCache()
	case "ChannelMerchantRate":
		cache.ChannelMerchantRate.RefreshCache()
	}
	ctx.WriteString("finnish.... ")
	return
}

func CacheSettlementFetch(ctx iris.Context) {
	ctx.WriteString("CacheSettlementFetch")
	services.SettlementFetch.AutoCacheOrder()
	return
}

func PushSettlementFetch(ctx iris.Context) {
	ctx.WriteString("PushSettlementFetch")
	services.SettlementFetch.AutoPushTask()
	return
}

func PopSettlementFetch(ctx iris.Context) {
	ctx.WriteString("PopSettlementFetch")
	services.SettlementFetch.Pop()
	return
}

func PushSettlementNotify(ctx iris.Context) {
	ctx.WriteString("PushSettlementNotify")
	services.SettlementNotify.AutoPushTask()
	return
}

func PopSettlementNotify(ctx iris.Context) {
	ctx.WriteString("PopSettlementNotify")
	services.SettlementNotify.Pop()
	return
}

func CacheSettlementNotify(ctx iris.Context) {
	ctx.WriteString("CacheSettlementNotify")
	services.SettlementNotify.AutoCacheOrder()
	return
}

func PushPayNotify(ctx iris.Context) {
	ctx.WriteString("PushPayNotify")
	services.PayNotify.AutoPushTask()
	return
}

func PopPayNotify(ctx iris.Context) {
	ctx.WriteString("PopPayNotify")
	services.PayNotify.Pop()
	return
}

func CachePayNotify(ctx iris.Context) {
	ctx.WriteString("CachePayNotify")
	go services.PayNotify.AutoCacheOrder()
	return
}

func ReEncrypt(ctx iris.Context) {
	ctx.WriteString("ReEncrypt starting ....")
	services.PayNotify.AutoCacheOrder()
	//decrypt := utils.AesCBCDecrypt("BdgqUe5KoU7mU1ry9WwXaG2Q5VJ8STL9QuvNnanB3ImDDrJeQGSbSowCdRBxhNcs")
	decrypt, err := utils.AesCbc.Decrypt("BdgqUe5KoU7mU1ry9WwXaG2Q5VJ8STL9QuvNnanB3ImDDrJeQGSbSowCdRBxhNcs")
	if err != nil {
		logrus.Error("utils.AesCbc.Decrypt Error:", err)
		return
	}
	ctx.WriteString(decrypt)
	ctx.WriteString("ReEncrypt starting ....")
	return
}

func ReadMemStats(ctx iris.Context) {
	// MemStats 描述内存信息的静态变量
	ctx.WriteString("readMemStats")
	var ms runtime.MemStats
	// 读取某一时刻内存情况的快照
	runtime.ReadMemStats(&ms)
	// alloc占用内存情况、堆空闲情况、堆释放情况
	log.Printf("========> Alloc:%d(bytes) HeapIdle:%d(bytes) HeapReleased:%d(bytes)", ms.Alloc, ms.HeapIdle, ms.HeapReleased)
	runtime.GC()
	log.Printf("========> Alloc:%d(bytes) HeapIdle:%d(bytes) HeapReleased:%d(bytes)", ms.Alloc, ms.HeapIdle, ms.HeapReleased)
	return
}

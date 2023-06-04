package cache

import (
	//"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"luckypay/config"

	//"luckypay/channels"
	"context"
)

var SystemCache = newSystemCache()
var redisServer = config.NewRedis()
var ctx context.Context = context.Background()
var Caches struct {
	AdminCache      *adminCache
	MerchantCache   *merchantCache
	MerchantAmount  *merchantAmount
	MerchantChannel *merchantChannel
}

func newSystemCache() *systemCache {
	return &systemCache{}
}

type systemCache struct {
}

func (s *systemCache) RefreshCache() {
	logrus.Info("刷新缓存")
	go AdminCache.RefreshCache()
	go MerchantCache.RefreshCache()
	go MerchantCache.SetCacheMerchantData()
	go MerchantAmount.RefreshCache()
	go MerchantAccount.RefreshCache()
	go MerchantRate.RefreshCache()
	go MerchantChannel.RefreshCache()
	go ChannelMerchant.RefreshCache()
	go MerchantChannelSettlement.RefreshCache()
	go ChannelMerchantRate.RefreshCache()
}

func (s *systemCache) RefreshCacheOne() {

}

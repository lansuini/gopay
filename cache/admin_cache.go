package cache

import (
	"encoding/json"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"luckypay/model"
)

var AdminCache = newAdminCache()

func newAdminCache() *adminCache {
	return &adminCache{}
}

type adminCache struct {
}

// SignIn 登录

func (s *adminCache) RefreshCache() {
	admins := []model.SystemAccount{}
	err := sqls.DB().Table("system_account").Limit(100).Find(&admins).Error
	if err != nil {
		logrus.Error("adminCache FreshCache Find Error", err)
		return
	}

	for _, admin := range admins {

		jsonStr, errs := json.Marshal(admins)
		if errs != nil {
			//fmt.Println()
			logrus.Error("adminCache FreshCache Marshal Error-", admin.LoginName, err)
			continue
		}
		cacheKey := "system_account:" + admin.LoginName
		_, err = redisServer.Set(ctx, cacheKey, jsonStr, 0).Result()
		if errs != nil {
			logrus.Error("adminCache RefreshCache SetCache Error-", admin.LoginName, err)
			continue
		}
	}
	return

}

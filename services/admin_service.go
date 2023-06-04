package services

import (
	"encoding/json"
	"errors"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"luckypay/model"
	"luckypay/mytool"
	"luckypay/repositories"
)

var AdminService = newAdminService()

func newAdminService() *adminService {
	return &adminService{}
}

type adminService struct {
}

// SignIn 登录
func (s *adminService) SignIn(username string, password string) (*model.SystemAccount, error) {
	if len(username) == 0 {
		return nil, errors.New("用户名不能为空")
	}
	if len(password) == 0 {
		return nil, errors.New("密码不能为空")
	}
	var user *model.SystemAccount

	user = s.GetByUsername(username)
	//fmt.Println(user.LoginPwd)
	if user == nil || user.Status != "Normal" {
		return nil, errors.New("用户名或密码错误！")
	}

	if mytool.GetHashPassword(password) != user.LoginPwd {
		if user.LoginFailNum+1 >= 5 {
			sqls.DB().Table("system_account").Where("loginName = ?", username).Update("status", "Close")
		} else {
			sqls.DB().Table("system_account").Where("loginName = ?", username).Update("loginFailNum", gorm.Expr("loginFailNum + ?", 1))
		}
		return nil, errors.New("用户名或密码错误")
	}

	return user, nil
}

func (s *adminService) RefreshCache() {
	admins := []model.SystemAccount{}
	err := sqls.DB().Table("system_account").Limit(100).Find(&admins).Error
	if err != nil {
		logrus.Error("adminService FreshCache Find Error", err)
		return
	}

	for _, admin := range admins {

		jsonStr, errs := json.Marshal(admins)
		if errs != nil {
			//fmt.Println()
			logrus.Error("adminService FreshCache Marshal Error-", admin.LoginName, err)
			continue
		}
		cacheKey := "system_account:" + admin.LoginName
		_, err = redisServer.Set(ctx, cacheKey, jsonStr, 0).Result()
		if errs != nil {
			logrus.Error("adminService RefreshCache SetCache Error-", admin.LoginName, err)
			continue
		}
	}
	return

}

// GetByUsername 根据用户名查找
func (s *adminService) GetByUsername(loginName string) *model.SystemAccount {
	return repositories.AdminRepository.GetByUsername(sqls.DB(), loginName)
}

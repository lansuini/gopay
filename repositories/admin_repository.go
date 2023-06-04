package repositories

import (
	"gorm.io/gorm"

	"luckypay/model"
)

var AdminRepository = newAdminRepository()

func newAdminRepository() *adminRepository {
	return &adminRepository{}
}

type adminRepository struct {
}

func (r *adminRepository) Take(db *gorm.DB, where ...interface{}) *model.SystemAccount {
	ret := &model.SystemAccount{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *adminRepository) GetByUsername(db *gorm.DB, loginName string) *model.SystemAccount {
	return r.Take(db, "loginName = ?", loginName)
}

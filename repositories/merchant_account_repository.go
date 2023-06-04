package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var MerchantAccountRepository = newMerchantAccountRepository()

func newMerchantAccountRepository() *merchantAccountRepository {
	return &merchantAccountRepository{}
}

type merchantAccountRepository struct {
}

func (r *merchantAccountRepository) Get(db *gorm.DB, id int64) *model.MerchantAccount {
	ret := &model.MerchantAccount{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantAccountRepository) Take(db *gorm.DB, where ...interface{}) *model.MerchantAccount {
	ret := &model.MerchantAccount{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantAccountRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantAccount) {
	cnd.Find(db, &list)
	return
}

func (r *merchantAccountRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.MerchantAccount {
	ret := &model.MerchantAccount{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *merchantAccountRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.MerchantAccount, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *merchantAccountRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantAccount, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.MerchantAccount{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *merchantAccountRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.MerchantAccount{})
}

func (r *merchantAccountRepository) Create(db *gorm.DB, t *model.MerchantAccount) (err error) {
	err = db.Create(t).Error
	return
}

func (r *merchantAccountRepository) Update(db *gorm.DB, t *model.MerchantAccount) (err error) {
	err = db.Save(t).Error
	return
}

func (r *merchantAccountRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.MerchantAccount{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *merchantAccountRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.MerchantAccount{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *merchantAccountRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.MerchantAccount{}, "id = ?", id)
}

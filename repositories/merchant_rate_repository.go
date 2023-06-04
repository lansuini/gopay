package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var MerchantRateRepository = newMerchantRateRepository()

func newMerchantRateRepository() *merchantRateRepository {
	return &merchantRateRepository{}
}

type merchantRateRepository struct {
}

func (r *merchantRateRepository) Get(db *gorm.DB, id int64) *model.MerchantRate {
	ret := &model.MerchantRate{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantRateRepository) Take(db *gorm.DB, where ...interface{}) *model.MerchantRate {
	ret := &model.MerchantRate{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantRateRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantRate) {
	cnd.Find(db, &list)
	return
}

func (r *merchantRateRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.MerchantRate {
	ret := &model.MerchantRate{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *merchantRateRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.MerchantRate, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *merchantRateRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantRate, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.MerchantRate{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *merchantRateRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.MerchantRate{})
}

func (r *merchantRateRepository) Create(db *gorm.DB, t *model.MerchantRate) (err error) {
	err = db.Create(t).Error
	return
}

func (r *merchantRateRepository) Update(db *gorm.DB, t *model.MerchantRate) (err error) {
	err = db.Save(t).Error
	return
}

func (r *merchantRateRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.MerchantRate{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *merchantRateRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.MerchantRate{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *merchantRateRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.MerchantRate{}, "id = ?", id)
}

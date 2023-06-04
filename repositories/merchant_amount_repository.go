package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var MerchantAmountRepository = newMerchantAmountRepository()

func newMerchantAmountRepository() *merchantAmountRepository {
	return &merchantAmountRepository{}
}

type merchantAmountRepository struct {
}

func (r *merchantAmountRepository) Get(db *gorm.DB, id int64) *model.MerchantAmount {
	ret := &model.MerchantAmount{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantAmountRepository) Take(db *gorm.DB, where ...interface{}) *model.MerchantAmount {
	ret := &model.MerchantAmount{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantAmountRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantAmount) {
	cnd.Find(db, &list)
	return
}

func (r *merchantAmountRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.MerchantAmount {
	ret := &model.MerchantAmount{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *merchantAmountRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.MerchantAmount, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *merchantAmountRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantAmount, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.MerchantAmount{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *merchantAmountRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.MerchantAmount{})
}

func (r *merchantAmountRepository) Create(db *gorm.DB, t *model.MerchantAmount) (err error) {
	err = db.Create(t).Error
	return
}

func (r *merchantAmountRepository) Update(db *gorm.DB, t *model.MerchantAmount) (err error) {
	err = db.Save(t).Error
	return
}

func (r *merchantAmountRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.MerchantAmount{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *merchantAmountRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.MerchantAmount{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *merchantAmountRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.MerchantAmount{}, "id = ?", id)
}

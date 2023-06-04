package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var MerchantRepository = newMerchantRepository()

func newMerchantRepository() *merchantRepository {
	return &merchantRepository{}
}

type merchantRepository struct {
}

func (r *merchantRepository) Get(db *gorm.DB, id int64) *model.Merchant {
	ret := &model.Merchant{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantRepository) Take(db *gorm.DB, where ...interface{}) *model.Merchant {
	ret := &model.Merchant{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.Merchant) {
	cnd.Find(db, &list)
	return
}

func (r *merchantRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.Merchant {
	ret := &model.Merchant{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *merchantRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.Merchant, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *merchantRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.Merchant, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.Merchant{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *merchantRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.Merchant{})
}

func (r *merchantRepository) Create(db *gorm.DB, t *model.Merchant) (err error) {
	err = db.Create(t).Error
	return
}

func (r *merchantRepository) Update(db *gorm.DB, t *model.Merchant) (err error) {
	err = db.Save(t).Error
	return
}

func (r *merchantRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.Merchant{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *merchantRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.Merchant{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *merchantRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.Merchant{}, "id = ?", id)
}

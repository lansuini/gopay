package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var PayRepository = newPayRepository()

func newPayRepository() *payRepository {
	return &payRepository{}
}

type payRepository struct {
}

func (r *payRepository) Get(db *gorm.DB, id int64) *model.PlatformPayOrder {
	ret := &model.PlatformPayOrder{}
	if err := db.First(ret, "orderId = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *payRepository) Take(db *gorm.DB, where ...interface{}) *model.PlatformPayOrder {
	ret := &model.PlatformPayOrder{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *payRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.PlatformPayOrder) {
	cnd.Find(db, &list)
	return
}

func (r *payRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.PlatformPayOrder {
	ret := &model.PlatformPayOrder{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *payRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.PlatformPayOrder, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *payRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.PlatformPayOrder, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.PlatformPayOrder{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *payRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.PlatformPayOrder{})
}

func (r *payRepository) Create(db *gorm.DB, t *model.PlatformPayOrder) (err error) {
	err = db.Create(t).Error
	return
}

func (r *payRepository) Update(db *gorm.DB, t *model.PlatformPayOrder) (err error) {
	err = db.Save(t).Error
	return
}

func (r *payRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.PlatformPayOrder{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *payRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.PlatformPayOrder{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *payRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.PlatformPayOrder{}, "id = ?", id)
}

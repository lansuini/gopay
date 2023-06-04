package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var SettleRepository = newSettleRepository()

func newSettleRepository() *settleRepository {
	return &settleRepository{}
}

type settleRepository struct {
}

func (r *settleRepository) Get(db *gorm.DB, id int64) *model.PlatformSettlementOrder {
	ret := &model.PlatformSettlementOrder{}
	if err := db.First(ret, "orderId = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *settleRepository) Take(db *gorm.DB, where ...interface{}) *model.PlatformSettlementOrder {
	ret := &model.PlatformSettlementOrder{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *settleRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.PlatformSettlementOrder) {
	cnd.Find(db, &list)
	return
}

func (r *settleRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.PlatformSettlementOrder {
	ret := &model.PlatformSettlementOrder{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *settleRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.PlatformSettlementOrder, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *settleRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.PlatformSettlementOrder, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.PlatformSettlementOrder{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *settleRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.PlatformSettlementOrder{})
}

func (r *settleRepository) Create(db *gorm.DB, t *model.PlatformSettlementOrder) (err error) {
	err = db.Create(t).Error
	return
}

func (r *settleRepository) Update(db *gorm.DB, t *model.PlatformSettlementOrder) (err error) {
	err = db.Save(t).Error
	return
}

func (r *settleRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.PlatformSettlementOrder{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *settleRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.PlatformSettlementOrder{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *settleRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.PlatformSettlementOrder{}, "id = ?", id)
}

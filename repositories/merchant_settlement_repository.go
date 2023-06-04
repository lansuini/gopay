package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var MerchantChannelSettlementRepository = newMerchantChannelSettlementRepository()

func newMerchantChannelSettlementRepository() *merchantSettlementRepository {
	return &merchantSettlementRepository{}
}

type merchantSettlementRepository struct {
}

func (r *merchantSettlementRepository) Get(db *gorm.DB, id int64) *model.MerchantChannelSettlement {
	ret := &model.MerchantChannelSettlement{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantSettlementRepository) Take(db *gorm.DB, where ...interface{}) *model.MerchantChannelSettlement {
	ret := &model.MerchantChannelSettlement{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantSettlementRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantChannelSettlement) {
	cnd.Find(db, &list)
	return
}

func (r *merchantSettlementRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.MerchantChannelSettlement {
	ret := &model.MerchantChannelSettlement{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *merchantSettlementRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.MerchantChannelSettlement, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *merchantSettlementRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantChannelSettlement, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.MerchantChannelSettlement{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *merchantSettlementRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.MerchantChannelSettlement{})
}

func (r *merchantSettlementRepository) Create(db *gorm.DB, t *model.MerchantChannelSettlement) (err error) {
	err = db.Create(t).Error
	return
}

func (r *merchantSettlementRepository) Update(db *gorm.DB, t *model.MerchantChannelSettlement) (err error) {
	err = db.Save(t).Error
	return
}

func (r *merchantSettlementRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.MerchantChannelSettlement{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *merchantSettlementRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.MerchantChannelSettlement{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *merchantSettlementRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.MerchantChannelSettlement{}, "id = ?", id)
}

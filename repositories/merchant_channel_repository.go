package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var MerchantChannelRepository = newMerchantChannelRepository()

func newMerchantChannelRepository() *merchantChannelRepository {
	return &merchantChannelRepository{}
}

type merchantChannelRepository struct {
}

func (r *merchantChannelRepository) Get(db *gorm.DB, id int64) *model.MerchantChannel {
	ret := &model.MerchantChannel{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantChannelRepository) Take(db *gorm.DB, where ...interface{}) *model.MerchantChannel {
	ret := &model.MerchantChannel{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *merchantChannelRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantChannel) {
	cnd.Find(db, &list)
	return
}

func (r *merchantChannelRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.MerchantChannel {
	ret := &model.MerchantChannel{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *merchantChannelRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.MerchantChannel, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *merchantChannelRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.MerchantChannel, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.MerchantChannel{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *merchantChannelRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.MerchantChannel{})
}

func (r *merchantChannelRepository) Create(db *gorm.DB, t *model.MerchantChannel) (err error) {
	err = db.Create(t).Error
	return
}

func (r *merchantChannelRepository) Update(db *gorm.DB, t *model.MerchantChannel) (err error) {
	err = db.Save(t).Error
	return
}

func (r *merchantChannelRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.MerchantChannel{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *merchantChannelRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.MerchantChannel{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *merchantChannelRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.MerchantChannel{}, "id = ?", id)
}

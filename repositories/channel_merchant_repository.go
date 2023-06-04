package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var ChannelMerchantRepository = newChannelMerchantRepository()

func newChannelMerchantRepository() *channelMerchantRepository {
	return &channelMerchantRepository{}
}

type channelMerchantRepository struct {
}

func (r *channelMerchantRepository) Get(db *gorm.DB, id int64) *model.ChannelMerchant {
	ret := &model.ChannelMerchant{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *channelMerchantRepository) Take(db *gorm.DB, where ...interface{}) *model.ChannelMerchant {
	ret := &model.ChannelMerchant{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *channelMerchantRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.ChannelMerchant) {
	cnd.Find(db, &list)
	return
}

func (r *channelMerchantRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.ChannelMerchant {
	ret := &model.ChannelMerchant{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *channelMerchantRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.ChannelMerchant, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *channelMerchantRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.ChannelMerchant, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.ChannelMerchant{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *channelMerchantRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.ChannelMerchant{})
}

func (r *channelMerchantRepository) Create(db *gorm.DB, t *model.ChannelMerchant) (err error) {
	err = db.Create(t).Error
	return
}

func (r *channelMerchantRepository) Update(db *gorm.DB, t *model.ChannelMerchant) (err error) {
	err = db.Save(t).Error
	return
}

func (r *channelMerchantRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.ChannelMerchant{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *channelMerchantRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.ChannelMerchant{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *channelMerchantRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.ChannelMerchant{}, "id = ?", id)
}

package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var ChannelMerchantRateRepository = newChannelMerchantRateRepository()

func newChannelMerchantRateRepository() *channelMerchantRateRepository {
	return &channelMerchantRateRepository{}
}

type channelMerchantRateRepository struct {
}

func (r *channelMerchantRateRepository) Get(db *gorm.DB, id int64) *model.ChannelMerchantRate {
	ret := &model.ChannelMerchantRate{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *channelMerchantRateRepository) Take(db *gorm.DB, where ...interface{}) *model.ChannelMerchantRate {
	ret := &model.ChannelMerchantRate{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *channelMerchantRateRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.ChannelMerchantRate) {
	cnd.Find(db, &list)
	return
}

func (r *channelMerchantRateRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.ChannelMerchantRate {
	ret := &model.ChannelMerchantRate{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *channelMerchantRateRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.ChannelMerchantRate, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *channelMerchantRateRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.ChannelMerchantRate, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.ChannelMerchantRate{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *channelMerchantRateRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.ChannelMerchantRate{})
}

func (r *channelMerchantRateRepository) Create(db *gorm.DB, t *model.ChannelMerchantRate) (err error) {
	err = db.Create(t).Error
	return
}

func (r *channelMerchantRateRepository) Update(db *gorm.DB, t *model.ChannelMerchantRate) (err error) {
	err = db.Save(t).Error
	return
}

func (r *channelMerchantRateRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.ChannelMerchantRate{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *channelMerchantRateRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.ChannelMerchantRate{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *channelMerchantRateRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.ChannelMerchantRate{}, "id = ?", id)
}

package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"luckypay/model"
)

var AmountPayRepository = newAmountPayRepository()

func newAmountPayRepository() *amountPayRepository {
	return &amountPayRepository{}
}

type amountPayRepository struct {
}

func (r *amountPayRepository) Get(db *gorm.DB, id int64) *model.AmountPay {
	ret := &model.AmountPay{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *amountPayRepository) Take(db *gorm.DB, where ...interface{}) *model.AmountPay {
	ret := &model.AmountPay{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *amountPayRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.AmountPay) {
	cnd.Find(db, &list)
	return
}

func (r *amountPayRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.AmountPay {
	ret := &model.AmountPay{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *amountPayRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.AmountPay, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *amountPayRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.AmountPay, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.AmountPay{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *amountPayRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.AmountPay{})
}

func (r *amountPayRepository) Create(db *gorm.DB, t *model.AmountPay) (err error) {
	err = db.Create(t).Error
	return
}

func (r *amountPayRepository) Update(db *gorm.DB, t *model.AmountPay) (err error) {
	err = db.Save(t).Error
	return
}

func (r *amountPayRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.AmountPay{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *amountPayRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.AmountPay{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *amountPayRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.AmountPay{}, "id = ?", id)
}

func (r *amountPayRepository) Sum(db *gorm.DB, cnd *sqls.Cnd) (test *string) {

	if err := cnd.FindOne(db, &test); err != nil {
		return nil
	}
	return test
}

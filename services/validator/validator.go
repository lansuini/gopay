package validator

import (
	"fmt"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/syyongx/php2go"
	"luckypay/config"
)

var Validator *validator.Validate

func isInPayTypeMap(fl validator.FieldLevel) bool {
	PayTypeKeys := php2go.ArrayKeys(config.PayTypeMap)
	return php2go.InArray(fl.Field().String(), PayTypeKeys)
}

func init() {
	fmt.Println(456)
	//v, ok := binding.Validator.Engine().(*validator.Validate)
	//if !ok {
	//	panic("验证引擎初始化失败")
	//}
	// 中文翻译器
	uni := ut.New(zh.New())
	trans, _ := uni.GetTranslator("zh")
	var Validate = validator.New()
	err := zh_translations.RegisterDefaultTranslations(Validate, trans)
	if err != nil {
		fmt.Println(err)
		return
	}
	//Validate.SetTagName("binding")
	errs := Validate.RegisterValidation("ValidPayType", isInPayTypeMap)
	if errs != nil {
		fmt.Println("validator RegisterValidation error: " + err.Error())
	}
	//binding.Validator = new(ginValidator)
	//
	//mgoModel.Validator = new(modelValidator)

}

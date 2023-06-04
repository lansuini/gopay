package utils

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"luckypay/config"
	"luckypay/response"
	"reflect"
	"strconv"
	"strings"

	"github.com/syyongx/php2go"
	"regexp"
	"sync"
)

const (
	alphaNumericRegexSpaceString = "^[a-zA-Z0-99\\s]+$"
	//alphaNumericRegexSpaceString = `\s|[\r\n]`
	//alphaNumericRegexSpaceString = `( )+|(\n)+`
)

var (
	defaultValidator       *validator.Validate
	validatorOnce          sync.Once
	alphaNumericSpaceRegex = regexp.MustCompile(alphaNumericRegexSpaceString)
)

func GetValidator() *validator.Validate {
	validatorOnce.Do(func() {
		//uni := ut.New(zh.New())
		//trans, _ := uni.GetTranslator("zh")
		defaultValidator = validator.New()
		//err := zh_translations.RegisterDefaultTranslations(defaultValidator, trans)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		defaultValidator.RegisterValidation("hh:mm", ValidateDateTime)
		err := defaultValidator.RegisterValidation("validPayType", IsInPayTypeMap)
		err = defaultValidator.RegisterValidation("validPayChannel", IsInPayChannelMap)

		defaultValidator.RegisterValidation("validBankCode", IsInBankCodeMap)
		err = defaultValidator.RegisterValidation("alphanumspace", IsAlphanumSpace)
		err = defaultValidator.RegisterValidation("notblank", NotBlank)
		err = defaultValidator.RegisterValidation("isNumberStr", isNumberStr)
		_ = defaultValidator.RegisterValidation("ISO8601date", IsISO8601Date)
		if err != nil {
			fmt.Println("validator RegisterValidation error: " + err.Error())
		}

	})
	return defaultValidator
}

func isNumberStr(fl validator.FieldLevel) bool {
	n, _ := strconv.Atoi(fl.Field().String())
	return n >= 0
}

func IsAlphanumSpace(fl validator.FieldLevel) bool {
	return alphaNumericSpaceRegex.MatchString(fl.Field().String())
}

func ValidateParam(ctx iris.Context, param interface{}, method string) error {
	var err error
	switch method {
	case "get":
		err = ctx.ReadForm(param)
	case "post":
		err = ctx.ReadJSON(param)
	default:
		err = ctx.ReadBody(param)
	}
	if err != nil {
		response.With(ctx, "E1001", err.Error(), nil)
		return err
	}
	err = GetValidator().Struct(param)
	if err != nil {
		//response.Fail(ctx, response.Error, err.Error(), nil)
		response.With(ctx, "E1002", err.Error(), nil)
		return err
	}
	//if err != nil {
	//	for _, err := range err.(validator.ValidationErrors) {
	//
	//	}
	//}

	return nil
}

func Validate(ctx iris.Context, param interface{}, method string) error {
	var err error
	switch method {
	case "get":
		//err = ctx.ReadForm(param)
		err = ctx.ReadQuery(param)
	case "post":
		err = ctx.ReadJSON(param)
	case "form":
		err = ctx.ReadForm(param)
	default:
		err = ctx.ReadBody(param)
	}
	if err != nil {
		if !iris.IsErrPath(err) || err == iris.ErrEmptyForm {
			logrus.Info(err.Error())
			return err
		}
		//response.With(ctx, "E1001", err.Error(), nil)
	}
	err = GetValidator().Struct(param)
	if err != nil {
		//response.Fail(ctx, response.Error, err.Error(), nil)
		//response.With(ctx, "E1002", err.Error(), nil)
		return err
	}
	//if err != nil {
	//	for _, err := range err.(validator.ValidationErrors) {
	//		fmt.Println(err)
	//	}
	//}

	return nil
}

func ValidateDateTime(f validator.FieldLevel) bool {
	dateTime := f.Field().String()
	return CheckDateTime(dateTime)
}

func CheckDateTime(dateTime string) bool {
	regx, err := regexp.Compile("^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$")
	if err != nil {
		logrus.Error(err)
		return false
	}
	isValid := regx.MatchString(dateTime)
	return isValid
}

func IsInPayTypeMap(fl validator.FieldLevel) bool {
	PayTypeKeys := php2go.ArrayKeys(config.PayTypeMap)
	res := php2go.InArray(fl.Field().String(), PayTypeKeys)
	return res
}

func IsInBankCodeMap(fl validator.FieldLevel) bool {
	res := php2go.InArray(fl.Field().String(), config.BankCodeMap)
	return res
}

func IsInPayChannelMap(fl validator.FieldLevel) bool {
	channel := fl.Field().String()
	if _, ok := config.Channel[channel]; !ok {
		return false
	}
	return true
}

func IsISO8601Date(fl validator.FieldLevel) bool {
	ISO8601DateRegexString := "^(?:[1-9]\\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\\d|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[1-9]\\d(?:0[48]|[2468][048]|[13579][26])|(?:[2468][048]|[13579][26])00)-02-29)T(?:[01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(?:\\.\\d{1,9})?(?:Z|[+-][01]\\d:[0-5]\\d)$"
	ISO8601DateRegex := regexp.MustCompile(ISO8601DateRegexString)
	return ISO8601DateRegex.MatchString(fl.Field().String())
}

func NotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return len(strings.TrimSpace(field.String())) > 0
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

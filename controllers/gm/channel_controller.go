package gm

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"luckypay/cache"
	"luckypay/channels"
	"luckypay/config"
	"luckypay/model"
	"luckypay/services"
	"luckypay/utils"
	//"github.com/go-playground/validator/v10"
	"github.com/gocarina/gocsv"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"path"
	"strconv"
	"strings"
	//"encoding/csv"
)

type ChannelController struct {
	Ctx iris.Context
}

func (c *ChannelController) GetMerchantSearch() {
	queryParams := model.SearchChannelMerchant{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()

	if queryParams.MerchantNo > 100 {
		builder.WriteString(" and channelMerchantNo = '" + string(queryParams.MerchantNo) + "'")
		cnd = cnd.Where("channelMerchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.Status != "" {
		builder.WriteString(" and orderStatus = '" + queryParams.Status + "'")
		cnd = cnd.Where("orderStatus = ?", queryParams.Status)
	}
	if queryParams.Channel != "" {
		builder.WriteString(" and channel = '" + queryParams.Channel + "'")
		cnd = cnd.Where("channel = ?", queryParams.Channel)
	}

	var ChannelMerchants []model.ChannelMerchant
	var count int64
	err := cnd.Where("status != ?", "Deleted").Limit(queryParams.Limit).Offset(queryParams.Offset).Find(&ChannelMerchants).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	for index, merchant := range ChannelMerchants {

		ChannelMerchants[index].StatusDesc = config.CommonStatus[merchant.Status]
		//ChannelMerchants[index].Param = utils.AesCBCDecrypt(merchant.Param)
		ChannelMerchants[index].Param = ""

	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": ChannelMerchants})
}

func (c *ChannelController) GetMerchantParameter() {
	channel := c.Ctx.URLParam("name")
	if channel == "" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}
	if _, ok := config.ChannelParam[channel]; !ok {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}
	res, _ := utils.JsonToMap(config.ChannelParam[channel]["param"])

	desc, _ := utils.JsonToMap(config.ChannelParam[channel]["paramDesc"])

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "desc": desc, "result": res})
	return

}

func (c *ChannelController) GetMerchantDetail() {
	merchantNo := c.Ctx.URLParamTrim("merchantNo")
	merchantNoInt, _ := strconv.Atoi(merchantNo)

	if merchantNoInt < 0 || merchantNoInt > 100000000 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}
	merchantData, res := services.ChannelMerchant.GetCacheByChannelMerchantNo(merchantNo)
	if !res {
		err := sqls.DB().Where("channelMerchantNo = ?", merchantNo).First(&merchantData).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
	}
	if _, ok := config.ChannelParam[merchantData.Channel]; !ok {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "渠道参数错误"})
		return
	}

	desc, _ := utils.JsonToMap(config.ChannelParam[merchantData.Channel]["paramDesc"])
	merchantData.Param = utils.AesCBCDecrypt(merchantData.Param)
	merchantParams := make(map[string]interface{})
	json.Unmarshal([]byte(merchantData.Param), &merchantParams)
	for key, _ := range merchantParams {
		if key == "merchantPrivateKey" || key == "merchantPublicKey" || key == "api_token" {
			merchantParams[key] = ""
		}
	}
	filterParam, _ := json.Marshal(merchantParams)
	merchantData.Param = string(filterParam)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": merchantData, "desc": desc})
	return
}

func (c *ChannelController) PostMerchantUpdate(w http.ResponseWriter, r *http.Request) {
	data := model.ChannelMerchantUpdate{}
	_ = params.ReadForm(c.Ctx, &data)
	//if err != nil {
	//	logrus.Info(err.Error())
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
	//	return
	//}
	param, err := ParseFormToMap(w, r)
	if err != nil {
		logrus.Info(err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	data.Param = param
	err = utils.GetValidator().Struct(data)
	if err != nil {
		logrus.Info(err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	merchantData, res := services.ChannelMerchant.GetCacheByChannelMerchantNo(data.MerchantNo)
	if !res {
		err := sqls.DB().Where("channelMerchantNo = ?", data.MerchantNo).First(&merchantData).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
	}
	newParamStr := ""
	switch data.Channel {
	case "loroPay":

		newParam := model.LoroPayMerchantParam{}
		paramStr, _ := json.Marshal(data.Param)
		json.Unmarshal(paramStr, &newParam)
		err = utils.GetValidator().Struct(newParam)
		if err != nil {
			logrus.Info(err.Error())
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
		merchantData.ChannelAccount = newParam.Company
		sourceParamStr := utils.AesCBCDecrypt(merchantData.Param)
		byteParam := []byte(sourceParamStr)
		sourceParam := model.LoroPayMerchantParam{}
		json.Unmarshal(byteParam, &sourceParam)
		if len(newParam.MerchantPublicKey) == 0 {
			newParam.MerchantPublicKey = sourceParam.MerchantPublicKey
		}

		if len(newParam.MerchantPrivateKey) == 0 {
			newParam.MerchantPrivateKey = sourceParam.MerchantPrivateKey
		}
		byteParam, _ = json.Marshal(newParam)
		newParamStr = string(byteParam)

	default:
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误: " + data.Channel})
		return
	}
	if len(newParamStr) == 0 {
		byteParam, _ := json.Marshal(data.Param)
		newParamStr = string(byteParam)
	}
	merchantData.Status = data.Status
	merchantData.DelegateDomain = data.DelegateDomain
	merchantData.Param = utils.AesCBCEncrypt(newParamStr)
	err = sqls.DB().Where("channelMerchantNo = ?", data.MerchantNo).Updates(&merchantData).Error
	if err != nil {
		logrus.Info(err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	go cache.ChannelMerchant.RefreshOne(data.MerchantNo)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
	return

}

func (c *ChannelController) PostMerchantInsert(w http.ResponseWriter, r *http.Request) {
	data := &model.ChannelMerchantInsert{}
	_ = params.ReadForm(c.Ctx, data)
	//if err != nil {
	//	logrus.Info(err.Error())
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
	//	return
	//}
	param, err := ParseFormToMap(w, r)
	if err != nil {
		logrus.Info(err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	data.Param = param
	err = utils.GetValidator().Struct(data)
	//v := validate.Struct(data)
	//if !v.Validate() {

	//} // 调用验证

	if err != nil {
		logrus.Info(err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	var count int64
	err = sqls.DB().Table("channel_merchant").Where("channelMerchantNo = ?", data.MerchantNo).Count(&count).Error
	if err != nil {
		logrus.Info(err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if count > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号已存在！"})
		return
	}
	insertData := model.ChannelMerchant{}

	switch data.Channel {
	case "loroPay":

		newParam := model.LoroPayMerchantParam{}
		paramStr, _ := json.Marshal(data.Param)
		json.Unmarshal(paramStr, &newParam)
		err = utils.GetValidator().Struct(newParam)
		if err != nil {
			logrus.Info(err.Error())
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
		insertData.ChannelAccount = newParam.Company

	default:
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误: " + data.Channel})
		return
	}
	insertData.ChannelMerchantNo = data.MerchantNo
	insertData.DelegateDomain = data.DelegateDomain
	insertData.PlatformNo = data.MerchantNo
	insertData.Channel = data.Channel
	paramStr, _ := json.Marshal(data.Param)
	insertData.Param = utils.AesCBCEncrypt(string(paramStr))
	err = sqls.DB().Select("ChannelMerchantNo", "PlatformNo", "Channel", "DelegateDomain", "Param").Create(&insertData).Error
	if err != nil {
		logrus.Info(err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	services.ChannelMerchant.RefreshOne(data.MerchantNo)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "创建成功"})
	return

}

func (c *ChannelController) GetRateSearch() {
	queryParams := model.SearchChannelMerchantRate{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()

	if queryParams.MerchantNo != "" {
		builder.WriteString(" and channelMerchantNo = '" + string(queryParams.MerchantNo) + "'")
		cnd = cnd.Where("channelMerchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.Status != "" {
		builder.WriteString(" and status = '" + queryParams.Status + "'")
		cnd = cnd.Where("status = ?", queryParams.Status)
	}
	if queryParams.PayType != "" {
		builder.WriteString(" and payType = '" + queryParams.PayType + "'")
		cnd = cnd.Where("payType = ?", queryParams.PayType)
	}
	if queryParams.ProductType != "" {
		builder.WriteString(" and productType = '" + queryParams.ProductType + "'")
		cnd = cnd.Where("productType = ?", queryParams.ProductType)
	}
	if queryParams.RateType != "" {
		builder.WriteString(" and rateType = '" + queryParams.RateType + "'")
		cnd = cnd.Where("rateType = ?", queryParams.RateType)
	}

	var Rates []model.ChannelMerchantRate
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("rateId desc").Find(&Rates).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": Rates})

}

func (c *ChannelController) GetRateExport() {
	merchantNo, _ := strconv.Atoi(c.Ctx.URLParamTrim("merchantNo"))
	if merchantNo <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号错误"})
		return
	}
	cnd := sqls.DB()
	var merchantRates []model.ChannelMerchantRate
	var count int64
	err := cnd.Where("ChannelMerchantNo = ?", merchantNo).Find(&merchantRates).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": merchantRates, "result": ""})
}

func (c *ChannelController) PostRateImport(r *http.Request) {

	if err := r.ParseMultipartForm(1 * MB); err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	merchantNo := c.Ctx.PostValueInt64Default("merchantNo", 0)
	if merchantNo <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号错误"})
		return
	}
	merchantNoStr := fmt.Sprintf("%d", merchantNo)
	merchantData, res := services.ChannelMerchant.GetCacheByChannelMerchantNo(merchantNoStr)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "渠道商户不存在" + merchantNoStr})
		return
	}
	r.ParseMultipartForm(32 << 20)
	//获取上传文件
	file, multipartFileHeader, err := r.FormFile("file")

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	defer file.Close()

	fileType := path.Ext(multipartFileHeader.Filename)
	if fileType != ".csv" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "上传文件类型必须为csv格式"})
		return
	}

	merchantRates := []*model.ChannelMerchantRateImport{}
	//gocsv.Un
	if err = gocsv.Unmarshal(file, &merchantRates); err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var insertData []model.ChannelMerchantRate
	for index, merchantRate := range merchantRates {
		merchantRate.MerchantID = merchantData.ChannelMerchantID
		merchantRate.MerchantNo = merchantData.ChannelMerchantNo
		merchantRate.Channel = merchantData.Channel
		var singleRow model.ChannelMerchantRate
		jsonByte, _ := json.Marshal(merchantRate)
		merchantRateStr := string(jsonByte)
		err = json.Unmarshal([]byte(merchantRateStr), &singleRow)
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error(), "index": index})
			return
		}
		//if merchantRate.EndTime == ""{
		//	merchantRates[index].EndTime = sql.NullTime
		//}
		errs := utils.GetValidator().Struct(merchantRate)
		if errs != nil {
			logrus.Info(merchantRateStr)
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": errs.Error(), "index": index})
			return
		}

		insertData = append(insertData, singleRow)
	}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		var err error
		rateM := model.ChannelMerchantRate{}
		err = tx.Table("channel_merchant_rate").Where("channelMerchantNo = ?", merchantNo).Delete(&rateM).Error
		if err != nil {
			logrus.Error("-删除渠道商户费率失败 : ", err.Error())
			return err
		}
		err = tx.Create(insertData).Error
		if err != nil {
			logrus.Error("-导入渠道商户费率失败 : ", err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	services.ChannelMerchantRate.RefreshOne(merchantNoStr)
	//services.MerchantRate.RefreshCache()
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "上传成功"})
	return
}

/*
*调整渠道商户余额，暂时不需要
 */
func (c *ChannelController) GetMerchantBalance() {

	merchantNo := c.Ctx.URLParamTrim("merchantNo")
	merchantData, res := cache.ChannelMerchant.GetCacheByChannelMerchantNo(merchantNo)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户信息获取失败"})
		return
	}
	channel := merchantData.Channel
	if _, ok := channels.Channels[channel]; !ok {
		c.Ctx.JSON(iris.Map{"success": 0, "result": "channel exception"})
		return
	}
	channelObj := channels.Channels[channel]
	resp, err := channels.QueryBalance(channelObj, merchantData)
	if err != nil {
		c.Ctx.JSON(resp)
		return
	}
	c.Ctx.JSON(resp)
	return
}

func (c *ChannelController) PostMerchantBatchupdatestatus() {
	status := c.Ctx.URLParamTrim("status")
	if status != "Close" && status != "Normal" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "status参数错误"})
		return
	}
	ids := c.Ctx.URLParamTrim("ids")
	if len(ids) == 0 || len(ids) > 300 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "status参数错误"})
		return
	}
	str_arr := strings.Split(ids, `,`)
	err := sqls.DB().Table("channel_merchant").Where("channelMerchantId in (?)", str_arr).Update("status", status).Error
	if err != nil {
		logrus.Error("批量更新渠状态失败：", err)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "批量更新渠状态失败"})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "批量更新渠状态成功"})
	return
}

func ParseFormToMap(w http.ResponseWriter, r *http.Request) (param map[string]string, err error) {
	param = make(map[string]string)
	err = r.ParseForm()

	if err != nil {

		w.Write([]byte(err.Error()))

		return param, err

	}

	for i := range r.Form {

		if strings.HasPrefix(i, "param[") {

			rp := strings.NewReplacer("param[", "", "]", "")

			param[rp.Replace(i)] = r.Form.Get(i)

		}

	}
	return param, nil

}

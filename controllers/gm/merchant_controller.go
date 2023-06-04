package gm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gocarina/gocsv"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"luckypay/cache"
	"luckypay/config"
	"luckypay/model"
	"luckypay/mytool"
	"luckypay/services"
	"luckypay/utils"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
	//"encoding/csv"
)

const (
	MB = 1 << 20
)

type MerchantController struct {
	BaseController
	Ctx iris.Context
}

func (c *MerchantController) GetRateSearch() {
	queryParams := model.SearchMerchantRate{}
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
		builder.WriteString(" and merchantNo = '" + string(queryParams.MerchantNo) + "'")
		//fmt.Print(queryParams.MerchantNo)
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.Status != "" {
		builder.WriteString(" and orderStatus = '" + queryParams.Status + "'")
		cnd = cnd.Where("orderStatus = ?", queryParams.Status)
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

	var merchantRates []model.MerchantRate
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("rateId desc").Find(&merchantRates).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	for index, rate := range merchantRates {
		merchantRates[index].PayTypeDesc = config.PayType[rate.PayType]
		merchantRates[index].StatusDesc = config.CommonStatus[rate.Status]
		merchantRates[index].RateTypeDesc = config.RateType[rate.RateType]
		merchantRates[index].ProductTypeDesc = config.ProductType[rate.ProductType]
		//merchantRates[index].CardTypeDesc = config.CardType[rate.CardType]
	}
	//fmt.Print(count)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": merchantRates})

}

func (c *MerchantController) GetPaychannelSearch() {
	queryParams := model.SearchMerchantPayChannel{}
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
		builder.WriteString(" and merchantNo = '" + string(queryParams.MerchantNo) + "'")
		//fmt.Print(queryParams.MerchantNo)
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.Channel != "" {
		builder.WriteString(" and channel = '" + queryParams.Channel + "'")
		cnd = cnd.Where("channel = ?", queryParams.Channel)
	}
	if queryParams.ChannelMerchantNo != "" {
		builder.WriteString(" and channelMerchantNo = '" + queryParams.ChannelMerchantNo + "'")
		cnd = cnd.Where("channelMerchantNo = ?", queryParams.ChannelMerchantNo)
	}

	var merchantChannels []model.MerchantChannel
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("setId desc").Find(&merchantChannels).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	//fmt.Print(count)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": merchantChannels})

}

func (c *MerchantController) GetSettlementchannelSearch() {
	queryParams := model.SearchMerchantSettleChannel{}
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
		builder.WriteString(" and merchantNo = '" + string(queryParams.MerchantNo) + "'")
		//fmt.Print(queryParams.MerchantNo)
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.Channel != "" {
		builder.WriteString(" and channel = '" + queryParams.Channel + "'")
		cnd = cnd.Where("channel = ?", queryParams.Channel)
	}
	if queryParams.ChannelMerchantNo != "" {
		builder.WriteString(" and channelMerchantNo = '" + queryParams.ChannelMerchantNo + "'")
		cnd = cnd.Where("channelMerchantNo = ?", queryParams.ChannelMerchantNo)
	}

	var merchantChannels []model.MerchantChannelSettlement
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("setId desc").Find(&merchantChannels).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	//fmt.Print(count)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": merchantChannels})

}

func (c *MerchantController) PostRateImport(r *http.Request) {

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
	merchantData, res := services.MerchantService.GetCacheByMerchantNo(merchantNoStr)
	//services.MerchantRate.RefreshOne(merchantData.MerchantID)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户不存在" + merchantNoStr})
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
	merchantRates := []*model.MerchantRateImport{}
	if err = gocsv.Unmarshal(file, &merchantRates); err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var insertData []model.MerchantRate

	for index, merchantRate := range merchantRates {
		var singleRow model.MerchantRate
		merchantRate.MerchantID = merchantData.MerchantID
		merchantRate.MerchantNo = merchantData.MerchantNo
		errs := utils.GetValidator().Struct(merchantRate)
		if errs != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": errs.Error(), "index": index})
			return
		}
		jsonByte, _ := json.Marshal(merchantRate)
		merchantRateStr := string(jsonByte)
		err = json.Unmarshal([]byte(merchantRateStr), &singleRow)
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error(), "index": index})
			return
		}
		//dump.Printf(singleRow)
		insertData = append(insertData, singleRow)
		//fmt.Println(merchantRateStr)
	}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		var err error
		rateM := model.MerchantRate{}
		err = tx.Table("merchant_rate").Where("merchantNo = ?", merchantNo).Delete(&rateM).Error
		if err != nil {
			logrus.Error("-删除商户费率失败 : ", err.Error())
			return err
		}
		err = tx.Omit("StatusDesc", "ProductTypeDesc", "RateTypeDesc", "PayTypeDesc").Create(insertData).Error
		if err != nil {
			logrus.Error("-导入商户费率失败 : ", err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	defer cache.MerchantRate.RefreshOne(merchantNoStr)
	//services.MerchantRate.RefreshCache()
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "上传成功"})
	return
}

func (c *MerchantController) GetRateExport() {
	merchantNo, _ := strconv.Atoi(c.Ctx.URLParam("merchantNo"))
	if merchantNo <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号错误"})
		return
	}
	cnd := sqls.DB()
	var merchantRates []model.MerchantRate
	var count int64
	err := cnd.Where("merchantNo = ?", merchantNo).Find(&merchantRates).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	for index, rate := range merchantRates {
		merchantRates[index].PayTypeDesc = config.PayType[rate.PayType]
		merchantRates[index].StatusDesc = config.CommonStatus[rate.Status]
		merchantRates[index].RateTypeDesc = config.RateType[rate.RateType]
		merchantRates[index].ProductTypeDesc = config.ProductType[rate.ProductType]
		//merchantRates[index].CardTypeDesc = config.CardType[rate.CardType]
	}
	//fmt.Print(count)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": merchantRates, "result": ""})
}

func (c *MerchantController) PostPaychannelImport(r *http.Request) {

	if err := r.ParseMultipartForm(1 * MB); err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	merchantNo := c.Ctx.PostValueTrim("merchantNo")
	merchantData, res := services.MerchantService.GetCacheByMerchantNo(merchantNo)
	//services.MerchantRate.RefreshOne(merchantData.MerchantID)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户不存在" + merchantNo})
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
	merchantPayChannels := []model.MerchantPayChannelImport{}
	if err = gocsv.Unmarshal(file, &merchantPayChannels); err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var insertData []model.MerchantChannel
	for index, merchantPayChannel := range merchantPayChannels {

		if merchantPayChannel.MerchantNo != merchantNo {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "文件配置商户与修改商户号不一致", "index": index})
			return
		}
		if merchantPayChannel.BankCode == "'" {
			merchantPayChannels[index].BankCode = ""
			merchantPayChannel.BankCode = ""
		}

		errs := utils.GetValidator().Struct(merchantPayChannel)
		if errs != nil {
			logrus.Info("导入商户支付渠道配置错误：" + errs.Error())
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": errs.Error(), "index": index})
			return
		}
		channelData, res := services.ChannelMerchant.GetCacheByChannelMerchantNo(merchantPayChannel.ChannelMerchantNo)
		if !res {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "文件配置渠道商户不存在", "index": index})
			return
		}
		var singleRow model.MerchantChannel

		jsonByte, _ := json.Marshal(merchantPayChannel)
		merchantChannelStr := string(jsonByte)
		err = json.Unmarshal([]byte(merchantChannelStr), &singleRow)
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error(), "index": index})
			return
		}
		singleRow.ChannelMerchantID = channelData.ChannelMerchantID
		singleRow.MerchantID = merchantData.MerchantID
		insertData = append(insertData, singleRow)
	}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		var err error
		MerchantChannelM := model.MerchantChannel{}
		err = tx.Table("merchant_channel").Where("merchantNo = ?", merchantNo).Delete(&MerchantChannelM).Error
		if err != nil {
			logrus.Error("-删除商户支付渠道失败 : ", err.Error())
			return err
		}
		err = tx.Create(insertData).Error
		if err != nil {
			logrus.Error("-导入商户支付渠道失败 : ", err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	cache.MerchantChannel.RefreshOne(merchantNo, merchantData.MerchantID)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "上传成功"})
	return
}

func (c *MerchantController) GetPaychannelExport() {
	merchantNo := c.Ctx.URLParamTrim("merchantNo")
	if merchantNo == "" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号错误"})
		return
	}
	_, res := services.MerchantService.GetCacheByMerchantNo(merchantNo)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号不存在"})
		return
	}
	cnd := sqls.DB()
	var merchantChannels []model.MerchantChannel
	var count int64
	err := cnd.Where("merchantNo = ?", merchantNo).Find(&merchantChannels).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var rows []map[string]interface{}
	for _, merchantChannel := range merchantChannels {
		row := make(map[string]interface{})
		row["merchantNo"] = merchantChannel.MerchantNo
		row["channel"] = merchantChannel.Channel
		row["channelMerchantNo"] = merchantChannel.ChannelMerchantNo
		row["payChannelStatus"] = merchantChannel.PayChannelStatus
		row["payType"] = merchantChannel.PayType
		row["bankCode"] = merchantChannel.BankCode
		row["cardType"] = merchantChannel.CardType
		row["openOneAmountLimit"] = merchantChannel.OpenOneAmountLimit
		row["oneMinAmount"] = merchantChannel.OneMinAmount
		row["oneMaxAmount"] = merchantChannel.OneMaxAmount
		row["openDayAmountLimit"] = merchantChannel.OpenDayAmountLimit
		row["dayAmountLimit"] = merchantChannel.DayAmountLimit
		row["openDayNumLimit"] = merchantChannel.OpenDayNumLimit
		row["dayNumLimit"] = merchantChannel.DayNumLimit
		row["openTimeLimit"] = merchantChannel.OpenTimeLimit
		row["beginTime"] = merchantChannel.BeginTime
		row["endTime"] = merchantChannel.EndTime
		row["status"] = merchantChannel.Status
		rows = append(rows, row)
	}
	//fmt.Print(count)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": rows, "result": ""})
}

func (c *MerchantController) PostSettlementchannelImport(r *http.Request) {

	if err := r.ParseMultipartForm(1 * MB); err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	merchantNo := c.Ctx.PostValueTrim("merchantNo")
	merchantData, res := services.MerchantService.GetCacheByMerchantNo(merchantNo)
	//services.MerchantRate.RefreshOne(merchantData.MerchantID)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户不存在" + merchantNo})
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
	merchantPayChannels := []model.MerchantSettleChannelImport{}
	if err = gocsv.Unmarshal(file, &merchantPayChannels); err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var insertData []model.MerchantChannelSettlement
	for index, merchantPayChannel := range merchantPayChannels {
		channelData := model.ChannelMerchant{}
		err = sqls.DB().Where("channelMerchantNo = ?", merchantPayChannel.ChannelMerchantNo).Where("channel = ?", merchantPayChannel.Channel).First(&channelData).Error
		if err != nil {
			logrus.Error(err)
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error(), "index": index})
			return
		}
		merchantPayChannel.MerchantID = merchantData.MerchantID
		merchantPayChannel.ChannelMerchantId = channelData.ChannelMerchantID
		merchantPayChannel.SettlementAccountType = "UsableAccount"
		jsonByte, _ := json.Marshal(merchantPayChannel)
		merchantChannelStr := string(jsonByte)
		//fmt.Println(merchantChannelStr)

		errs := utils.GetValidator().Struct(merchantPayChannel)
		if errs != nil {
			logrus.Error("导入商户代付渠道配置错误：" + errs.Error())
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": errs.Error(), "index": index})
			return
		}
		var singleRow model.MerchantChannelSettlement

		err = json.Unmarshal([]byte(merchantChannelStr), &singleRow)
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error(), "index": index})
			return
		}
		insertData = append(insertData, singleRow)
	}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		var err error
		MerchantChannelM := model.MerchantChannelSettlement{}
		err = tx.Table("merchant_channel_settlement").Where("merchantNo = ?", merchantNo).Delete(&MerchantChannelM).Error
		if err != nil {
			logrus.Error("-删除商户代付渠道失败 : ", err.Error())
			return err
		}
		err = tx.Create(insertData).Error
		if err != nil {
			logrus.Error("-导入商户代付渠道失败 : ", err.Error())
			return err
		}
		//TODO：插入操作记录
		return nil
	})
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	go cache.MerchantChannelSettlement.RefreshOne(merchantNo, merchantData.MerchantID)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "上传成功"})
	return
}

func (c *MerchantController) GetSettlementchannelExport() {
	merchantNo := c.Ctx.URLParamTrim("merchantNo")
	if merchantNo == "" {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号错误"})
		return
	}
	_, res := services.MerchantService.GetCacheByMerchantNo(merchantNo)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号不存在"})
		return
	}
	cnd := sqls.DB()
	var merchantChannels []model.MerchantChannelSettlement
	var count int64
	err := cnd.Where("merchantNo = ?", merchantNo).Find(&merchantChannels).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var rows []map[string]interface{}
	for _, merchantChannel := range merchantChannels {
		row := make(map[string]interface{})
		row["merchantNo"] = merchantChannel.MerchantNo
		row["channel"] = merchantChannel.Channel
		row["channelMerchantNo"] = merchantChannel.ChannelMerchantNo
		row["settlementChannelStatus"] = merchantChannel.SettlementChannelStatus
		row["openOneAmountLimit"] = merchantChannel.OpenOneAmountLimit
		row["oneMinAmount"] = merchantChannel.OneMinAmount
		row["oneMaxAmount"] = merchantChannel.OneMaxAmount
		row["openDayAmountLimit"] = merchantChannel.OpenDayAmountLimit
		row["dayAmountLimit"] = merchantChannel.DayAmountLimit
		row["openDayNumLimit"] = merchantChannel.OpenDayNumLimit
		row["dayNumLimit"] = merchantChannel.DayNumLimit
		row["openTimeLimit"] = merchantChannel.OpenTimeLimit
		row["beginTime"] = merchantChannel.BeginTime
		row["endTime"] = merchantChannel.EndTime
		row["status"] = merchantChannel.Status
		rows = append(rows, row)
	}
	//fmt.Print(count)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": rows, "result": ""})
}

// 查询商户信息
func (c *MerchantController) GetSearch() {
	queryParams := model.SearchMerchant{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	//fmt.Println(queryParams.Offset)
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()
	cnd = cnd.Table("merchant as m").Joins("left join merchant_amount as ma on m.merchantId=ma.merchantId").Select("m.*,ma.settlementAmount")

	if queryParams.MerchantNo > 100 {
		builder.WriteString(" and m.merchantNo = '" + string(queryParams.MerchantNo) + "'")
		//fmt.Print(queryParams.MerchantNo)
		cnd = cnd.Where("m.merchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.Status != "" {
		builder.WriteString(" and status = '" + queryParams.Status + "'")
		cnd = cnd.Where("status = ?", queryParams.Status)
	}
	if queryParams.ShortName != "" {
		builder.WriteString(" and shortName = '" + queryParams.ShortName + "'")
		cnd = cnd.Where("shortName = ?", queryParams.ShortName)
	}
	if queryParams.FullName != "" {
		builder.WriteString(" and fullName = '" + queryParams.FullName + "'")
		cnd = cnd.Where("fullName = ?", queryParams.FullName)
	}
	if queryParams.PlatformNo != "" {
		builder.WriteString(" and platformNo = '" + queryParams.PlatformNo + "'")
		cnd = cnd.Where("platformNo = ?", queryParams.PlatformNo)
	}
	if queryParams.BeginTime != "" {
		builder.WriteString(" and m.created_at >= '" + queryParams.BeginTime + "'")
		cnd = cnd.Where("m.created_at >= ?", queryParams.BeginTime)
	}
	if queryParams.EndTime != "" {
		builder.WriteString(" and m.created_at <= '" + queryParams.EndTime + "'")
		cnd = cnd.Where("m.created_at <= ?", queryParams.EndTime)
	}

	var merchants []model.MerchantJoin
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("merchantId desc").Find(&merchants).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	whereStr := builder.String()
	merchantStats := model.MerchantStat{}
	sql := "select sum(ma.settlementAmount) as currentAmount from merchant as m left join merchant_amount  as ma on m.merchantId=ma.merchantId where " + whereStr
	err = sqls.DB().Raw(sql).Scan(&merchantStats).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	merchantStats2 := model.MerchantStat2{}
	sql2 := "select sum(ma.settlementAmount) as totalAmount from merchant as m left join merchant_amount  as ma on m.merchantId=ma.merchantId "
	errTwo := sqls.DB().Raw(sql2).Scan(&merchantStats2).Error
	if errTwo != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": errTwo.Error()})
		return
	}

	stat := map[string]interface{}{
		"currentAmount": merchantStats.CurrentAmount,
		"totalAmount":   merchantStats2.TotalAmount,
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "stat": stat, "rows": merchants})

}

// 修改商户信息
func (c *MerchantController) GetMerchantupdate() {
	updateParams := model.MerchantUpdate{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	if updateParams.MerchantId <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}

	var merchants []model.Merchant
	cnd := sqls.DB()
	updates := map[string]interface{}{
		"ShortName": updateParams.ShortName,
		"FullName":  updateParams.FullName,
		"Status":    updateParams.Status,
	}
	err := cnd.Where("merchantId = ?", updateParams.MerchantId).Model(&merchants).Updates(updates).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "修改失败"})
		return
	}
	services.MerchantService.RefreshCache()
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

// 商户信息管理权限配置
func (c *MerchantController) GetMerchantdetail() {
	var merchantNo, _ = c.Ctx.URLParamInt64("merchantNo")

	cnd := sqls.DB()
	if merchantNo < 100 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}

	var merchants model.Merchant
	err := cnd.Where("merchantNo = ?", merchantNo).First(&merchants).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": merchants})

}

// 商户信息管理权限配置修改
func (c *MerchantController) GetResetset() {
	//var settlementType = c.Ctx.URLParam("settlementType")
	//fmt.Println(settlementType)
	updateParams := model.MerchantResetset{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}

	var merchants []model.Merchant
	cnd := sqls.DB()
	updates := map[string]interface{}{
		"OpenPay":                updateParams.OpenPay,
		"OneSettlementMaxAmount": updateParams.OneSettlementMaxAmount,
		//"SettlementTime":         updateParams.SettlementTime,
		"D0SettlementRate":   updateParams.D0SettlementRate,
		"OpenAliSettlement":  updateParams.OpenAliSettlement,
		"OpenAutoSettlement": updateParams.OpenAutoSettlement,
		"OpenSettlement":     updateParams.OpenSettlement,
		//"SettlementType":         updateParams.SettlementType,
	}
	err := cnd.Where("merchantNo = ?", updateParams.MerchantNo).Model(&merchants).Updates(updates).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "修改失败"})
		return
	}
	services.MerchantService.RefreshOne(updateParams.MerchantNo)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

// 获取下一位商户号
func (c *MerchantController) GetMerchantnonext() {
	var merchant model.Merchant
	cnd := sqls.DB()
	err := cnd.Order("merchantId desc").First(&merchant).Error
	if err != nil {
		if err.Error() != "record not found" {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
			return
		}
		merchant.MerchantNo = "10000000"
	}
	var merchantNo int64
	merchantNo, err = strconv.ParseInt(merchant.MerchantNo, 10, 64)
	merchantNo = merchantNo + 1

	var loginNameCode string
	loginNameCode = utils.RandomString(4)

	results := map[string]interface{}{
		"loginNameCode": loginNameCode,
		"merchantNo":    merchantNo,
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": results})

}

// 新增商户信息
func (c *MerchantController) GetMerchantadd() {
	addParams := model.MerchantAdd{}
	valid_err := utils.Validate(c.Ctx, &addParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var merchants []model.Merchant

	err := sqls.DB().Where("merchantNo = ?", addParams.MerchantNo).Find(&merchants).Error
	//if err != nil {
	//	c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
	//	return
	//}
	if len(merchants) > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号已存在"})
		return
	}

	var merchant model.Merchant
	var merchantAccount model.MerchantAccount
	var merchantAmount model.MerchantAmount
	merchant.FullName = addParams.FullName
	merchant.PayerName = addParams.ShortName
	merchant.Status = "Normal"
	merchant.PlatformType = "Normal"
	merchant.SettlementType = addParams.SettlementType
	merchant.MerchantNo = addParams.MerchantNo
	merchant.ShortName = addParams.ShortName
	merchant.Description = addParams.Description
	merchant.PlatformNo = addParams.MerchantNo
	merchant.PlatformNo = addParams.MerchantNo
	merchant.PlatformNo = addParams.MerchantNo
	merchant.SignKey = utils.AesCBCEncrypt(utils.Md5(string(time.Now().Unix())))
	merchant.OpenPay = true
	merchant.OpenSettlement = true
	merchant.D0SettlementRate = 1
	merchant.SettlementTime = 0
	merchant.OneSettlementMaxAmount = 200000
	merchant.OpenQuery = true
	merchant.OpenSettlement = true
	merchant.OpenAliSettlement = true
	merchant.OpenBackNotice = true
	merchant.OpenCheckAccount = true
	merchant.OpenCheckDomain = true
	merchant.OpenFrontNotice = true
	merchant.OpenRepayNotice = true
	merchant.OpenAutoSettlement = true
	merchant.OpenManualSettlement = true
	//merchantData := map[string]interface{}{
	//	"FullName":               addParams.FullName,
	//	"PayerName":              addParams.ShortName,
	//	"SettlementType":         addParams.SettlementType,
	//	"MerchantNo":             addParams.MerchantNo,
	//	"ShortName":              addParams.ShortName,
	//	"Description":            addParams.Description,
	//	"PlatformNo":             addParams.MerchantNo,
	//	"SignKey":                utils.AesCBCEncrypt(utils.Md5(string(time.Now().Unix()))),
	//	"OpenPay":                false,
	//	"OpenSettlement":         false,
	//	"D0SettlementRate":       1,
	//	"settlementTime":         0,
	//	"oneSettlementMaxAmount": 200000,
	//	"openQuery":              1,
	//	"openAliSettlement":      1,
	//	"openBackNotice":         1,
	//	"openCheckAccount":       1,
	//	"openCheckDomain":        1,
	//	"openFrontNotice":        1,
	//	"openRepayNotice":        1,
	//	"openAutoSettlement":     1,
	//	"openManualSettlement":   1,
	//}
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		err = tx.Create(&merchant).Error
		if err != nil {
			return err
		}

		merchantAccount.MerchantNo = addParams.MerchantNo
		merchantAccount.MerchantID = merchant.MerchantID
		merchantAccount.PlatformNo = addParams.MerchantNo
		merchantAccount.PlatformType = "Normal"
		merchantAccount.Status = "Normal"
		merchantAccount.UserLevel = "PlatformManager"
		merchantAccount.LoginName = addParams.LoginName
		merchantAccount.UserName = addParams.LoginName
		merchantAccount.LoginPwd = mytool.GetHashPassword(addParams.LoginPwd)
		merchantAccount.SecurePwd = mytool.GetHashPassword(addParams.SecurePwd)
		err = tx.Create(&merchantAccount).Error
		if err != nil {
			logrus.Error(err)
			return err
		}

		merchantAmountData := map[string]interface{}{
			"MerchantId": merchant.MerchantID,
			"MerchantNo": addParams.MerchantNo,
		}
		err = tx.Omit("loginPwdAlterTime").Model(&merchantAmount).Create(merchantAmountData).Error
		if err != nil {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "添加失败"})
			return err
		}
		return nil
	})

	if err != nil {
		logrus.Error("添加商户失败", err)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "添加失败" + err.Error()})
		return
	}
	go cache.MerchantCache.RefreshOne(merchant.MerchantNo)
	go cache.MerchantAccount.RefreshOne(merchantAccount.AccountID)
	go cache.MerchantAmount.RefreshOne(merchant.MerchantNo)
	go cache.MerchantCache.SetCacheMerchantData()
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "添加成功"})
}

// 商户用户管理
func (c *MerchantController) GetUsersearch() {
	queryParams := model.SearchUserMerchant{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	//fmt.Println(queryParams.Offset)
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()

	if queryParams.MerchantNo > 100 {
		builder.WriteString(" and merchantNo = '" + string(queryParams.MerchantNo) + "'")
		//fmt.Print(queryParams.MerchantNo)
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.Status != "" {
		builder.WriteString(" and status = '" + queryParams.Status + "'")
		cnd = cnd.Where("status = ?", queryParams.Status)
	}
	if queryParams.LoginName != "" {
		builder.WriteString(" and loginName = '" + queryParams.LoginName + "'")
		cnd = cnd.Where("loginName = ?", queryParams.LoginName)
	}
	if queryParams.UserLevel != "" {
		builder.WriteString(" and userLevel = '" + queryParams.UserLevel + "'")
		cnd = cnd.Where("userLevel = ?", queryParams.UserLevel)
	}
	if queryParams.PlatformNo > 100 {
		builder.WriteString(" and platformNo = '" + string(queryParams.PlatformNo) + "'")
		cnd = cnd.Where("platformNo = ?", queryParams.PlatformNo)
	}

	var merchantAccounts []model.MerchantAccount
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("accountId desc").Find(&merchantAccounts).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": merchantAccounts})
}

// 修改商户用户信息
func (c *MerchantController) GetUserupdate() {
	updateParams := model.UserMerchantUpdate{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	if updateParams.UserId <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}

	var merchantAccount model.MerchantAccount
	cnd := sqls.DB()
	err := cnd.Where("accountId = ?", updateParams.UserId).First(&merchantAccount).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	if merchantAccount.LoginName != updateParams.LoginName {
		err = cnd.Where("loginName = ?", updateParams.LoginName).First(&merchantAccount).Error
		if merchantAccount.LoginName != "" {
			c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "用户名称已存在"})
			return
		}
	}

	updates := map[string]interface{}{
		"LoginName": updateParams.LoginName,
		"UserName":  updateParams.UserName,
		"Status":    updateParams.Status,
		"UserLevel": updateParams.UserLevel,
	}
	var merchantAccountRes []model.MerchantAccount
	err = cnd.Where("accountId = ?", updateParams.UserId).Model(&merchantAccountRes).Updates(updates).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "修改失败"})
		return
	}
	cache.MerchantAccount.RefreshOne(merchantAccount.AccountID)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

// 商户用户管理-》重置登录密码
func (c *MerchantController) GetResetloginpwd(httpWriter http.ResponseWriter, r *http.Request) {
	var accountId, _ = c.Ctx.URLParamInt64("userId")
	if accountId <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}

	var merchantAccounts model.MerchantAccount
	cnd := sqls.DB()
	err := cnd.Where("accountId = ?", accountId).First(&merchantAccounts).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var systemCheckLogRes []model.SystemCheckLog
	err = cnd.Where("status=0").Where("type", "登录密码修改").Where("relevance = ?", merchantAccounts.MerchantNo).First(&systemCheckLogRes).Error
	if len(systemCheckLogRes) > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "还有待审核数据"})
		return
	}
	loginPwd := utils.RandomOrderTail(6)
	//$content = ['accountId' => $model->accountId, "password" => $loginPwd];

	content := map[string]interface{}{
		"accountId": accountId,
		"password":  loginPwd,
	}

	jsonByte, _ := json.Marshal(content)
	contentStr := string(jsonByte)
	UserIP := utils.GetRealIp(r)
	insertParams := map[string]interface{}{
		"content":     contentStr,
		"commiter_id": LoginAdminId,
		"relevance":   merchantAccounts.MerchantNo,
		"ip":          UserIP,
		"ipDesc":      UserIP,
		"type":        "登录密码修改",
	}

	var SystemCheckLog []model.SystemCheckLog
	err = cnd.Model(&SystemCheckLog).Create(insertParams).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "修改失败"})
		return
	}
	go cache.MerchantAccount.RefreshOne(accountId)
	newPwd := map[string]string{
		"newPwd": loginPwd,
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": newPwd})

}

// 商户用户管理-》重置支付密码
func (c *MerchantController) GetResetsecurepwd(httpWriter http.ResponseWriter, r *http.Request) {
	var accountId, _ = c.Ctx.URLParamInt64("userId")
	if accountId <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}

	var merchantAccounts model.MerchantAccount
	cnd := sqls.DB()
	err := cnd.Where("accountId = ?", accountId).First(&merchantAccounts).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var systemCheckLogRes []model.SystemCheckLog
	err = cnd.Where("status=0").Where("type", "支付密码修改").Where("relevance = ?", merchantAccounts.MerchantNo).First(&systemCheckLogRes).Error
	if len(systemCheckLogRes) > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "还有待审核数据"})
		return
	}
	securePwd := utils.RandomOrderTail(6)
	//$content = ['accountId' => $model->accountId, "password" => $loginPwd];

	content := map[string]interface{}{
		"accountId": accountId,
		"password":  securePwd,
	}

	jsonByte, _ := json.Marshal(content)
	contentStr := string(jsonByte)
	UserIP := utils.GetRealIp(r)
	insertParams := map[string]interface{}{
		"content":     contentStr,
		"commiter_id": LoginAdminId, //$_SESSION['accountId']
		"relevance":   merchantAccounts.MerchantNo,
		"ip":          UserIP,
		"ipDesc":      UserIP,
		"type":        "支付密码修改",
	}

	var SystemCheckLog []model.SystemCheckLog
	err = cnd.Model(&SystemCheckLog).Create(insertParams).Error

	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "修改失败"})
		return
	}
	go cache.MerchantAccount.RefreshOne(accountId)
	newPwd := map[string]string{
		"newPwd": securePwd,
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": newPwd})

}

// 商户用户信息列表
func (c *MerchantController) GetPlatformsearch() {
	queryParams := model.PlatformSearch{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	cnd := sqls.DB()
	if queryParams.Status != "" {
		cnd = cnd.Where("status = ?", queryParams.Status)
	}
	if queryParams.Description != "" {
		cnd = cnd.Where("fullName = ?", queryParams.Description)
	}
	if queryParams.PlatformNo != "" {
		cnd = cnd.Where("platformNo = ?", queryParams.PlatformNo)
	}

	var merchants []model.Merchant
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("merchantId desc").Find(&merchants).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": merchants})

}

// 根据platformNo获取电子商务平台信息
func (c *MerchantController) GetPlatformdetail() {
	var platformNo, _ = c.Ctx.URLParamInt64("platformNo")

	cnd := sqls.DB()
	if platformNo < 100 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}

	var merchantRes model.Merchant
	err := cnd.Where("platformNo = ?", platformNo).First(&merchantRes).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	result := make(map[string]interface{})
	result["backNoticeMaxNum"] = merchantRes.BackNoticeMaxNum
	result["description"] = merchantRes.Description
	result["ipWhite"] = merchantRes.IpWhite
	result["loginIpWhite"] = merchantRes.LoginIPWhite
	result["openBackNotice"] = merchantRes.OpenBackNotice
	result["openCheckAccount"] = merchantRes.OpenCheckAccount
	result["openCheckDomain"] = merchantRes.OpenCheckDomain
	result["openFrontNotice"] = merchantRes.OpenFrontNotice
	result["openManualSettlement"] = merchantRes.OpenManualSettlement
	result["openRepayNotice"] = merchantRes.OpenRepayNotice
	result["platformNo"] = merchantRes.PlatformNo
	result["status"] = merchantRes.Status
	result["signKey"] = merchantRes.SignKey
	domains := strings.Split(merchantRes.Domain, ",")
	result["domains"] = domains
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": result})

}

// 修改电子商务平台信息
func (c *MerchantController) GetPlatformupdate() {
	updateParams := model.PlatformUpdate{}
	valid_err := utils.Validate(c.Ctx, &updateParams, "get")
	if valid_err != nil {
		logrus.Info("GetPlatformupdate", valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var merchantRes model.Merchant
	cnd := sqls.DB()
	err := cnd.Where("platformNo = ?", updateParams.PlatformNo).First(&merchantRes).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"domain":               updateParams.Domains,
		"description":          updateParams.Description,
		"status":               updateParams.Status,
		"openCheckAccount":     updateParams.OpenCheckAccount,
		"openCheckDomain":      updateParams.OpenCheckDomain,
		"openFrontNotice":      updateParams.OpenFrontNotice,
		"openBackNotice":       updateParams.OpenBackNotice,
		"openRepayNotice":      updateParams.OpenRepayNotice,
		"openManualSettlement": updateParams.OpenManualSettlement,
		"loginIpWhite":         updateParams.LoginIpWhite,
		"ipWhite":              updateParams.IpWhite,
	}
	if updateParams.IpWhite != "" {
		ipWhite := strings.Trim(updateParams.IpWhite, " ")
		//ipWhiteArr := make(map[string]string)
		ipWhiteArr := strings.Split(ipWhite, ",")
		for _, value := range ipWhiteArr {
			ip := net.ParseIP(value)
			if ip == nil {
				c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "回调ip白名单验证不通过"})
				return
			}
		}
	}

	if updateParams.LoginIpWhite != "" {
		loginIpWhite := strings.Trim(updateParams.LoginIpWhite, " ")
		//ipWhiteArr := make(map[string]string)
		loginLpWhiteArr := strings.Split(loginIpWhite, ",")
		for _, value := range loginLpWhiteArr {
			ip := net.ParseIP(value)
			if ip == nil {
				c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "登录ip白名单验证不通过"})
				return
			}
		}
	}
	var merchantUp []model.Merchant
	err2 := cnd.Where("platformNo = ?", updateParams.PlatformNo).Model(&merchantUp).Updates(updates).Error

	if err2 != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "修改失败"})
		return
	}
	services.MerchantService.RefreshCache()
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "修改成功"})
}

// 获取signKey
func (c *MerchantController) GetSignkey() {
	var platformNo, _ = c.Ctx.URLParamInt64("platformNo")
	if platformNo < 100 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}

	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()
	var merchantRes model.Merchant
	err := cnd.Select("signKey").Where("platformNo = ?", platformNo).First(&merchantRes).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	result := make(map[string]interface{})
	result["signKey"] = utils.AesCBCDecrypt(merchantRes.SignKey)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": result})
}

// 重置signKey
func (c *MerchantController) GetResetsignkey() {
	var platformNo, _ = c.Ctx.URLParamInt64("platformNo")
	if platformNo <= 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "参数错误"})
		return
	}
	var merchantRes model.Merchant
	cnd := sqls.DB()
	err := cnd.Where("platformNo = ?", platformNo).First(&merchantRes).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	var merchantUpdate []model.Merchant
	signKey := utils.Md5(strconv.FormatInt(time.Now().Unix(), 10))
	updates := map[string]interface{}{
		"signKey": utils.AesCBCEncrypt(signKey),
	}
	err2 := cnd.Where("platformNo = ?", platformNo).Model(&merchantUpdate).Updates(updates).Error

	if err2 != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "重置失败"})
		return
	}
	services.MerchantService.RefreshCache()
	result := make(map[string]interface{})
	result["signKey"] = signKey
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": result})
}

func (c *MerchantController) GetRandom() {
	redis := config.NewRedis()
	random := strconv.FormatInt(time.Now().UnixNano(), 10)
	login_admin_id, _ := Session.GetInt64("login_admin_id")
	_, err := redis.SetEX(c.Ctx, "balanceRandom-"+strconv.FormatInt(login_admin_id, 10)+random, 1, 5*time.Minute).Result()
	defer redis.Close()
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "", "random": random})
	return
}

// 调整余额
func (c *MerchantController) PostBalanceadjustment(req *http.Request) {
	postParams := model.BalanceAdj{}
	valid_err := utils.Validate(c.Ctx, &postParams, "")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	redis := config.NewRedis()
	defer redis.Close()

	balanceOperation, err := redis.Exists(c.Ctx, "balanceOperation"+postParams.MerchantNo).Result()
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	if balanceOperation == 1 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "此商户正在申请余额调整"})
		return
	}
	_, err = redis.SetEX(c.Ctx, "balanceOperation"+postParams.MerchantNo, 1, 30*time.Second).Result()
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	login_admin_id, _ := Session.GetInt64("login_admin_id")
	balanceKey := "balanceRandom-" + strconv.FormatInt(login_admin_id, 10) + postParams.Random
	exRandom, err := redis.Get(c.Ctx, balanceKey).Result()
	if len(exRandom) != 1 {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "请通过正确的渠道申请"})
		return
	}

	merchantData, res := services.MerchantService.GetCacheByMerchantNo(postParams.MerchantNo)
	if !res {
		redis.Del(c.Ctx, "balanceOperation"+postParams.MerchantNo)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户不存在"})
		return
	}
	var merchantAmount model.MerchantAmount
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", postParams.MerchantNo).First(&merchantAmount).Error

		if err != nil {
			logrus.Error(postParams.MerchantNo, "-查询商户余额失败 : ", err.Error())
			return err
		}
		if (merchantAmount.SettlementAmount-postParams.Amount) < 0 && postParams.BankrollDirection != "Restore" && postParams.BankrollDirection != "Recharge" {
			logrus.Error(postParams.MerchantNo, "-可用金额不足")
			return errors.New("可用金额不足")
		}
		platformOrderNo := ""
		if postParams.BankrollDirection == "Recharge" {
			platformOrderNo = services.MerchantService.GetPlatformOrderNo("R")
		} else {
			platformOrderNo = services.MerchantService.GetPlatformOrderNo("B")
		}
		dataModel := model.BalanceAdjustment{}
		dataModel.MerchantNo = postParams.MerchantNo
		dataModel.MerchantID = merchantData.MerchantID
		dataModel.PlatformOrderNo = platformOrderNo
		dataModel.BankrollDirection = postParams.BankrollDirection
		dataModel.BankrollType = "AccountBalance"
		dataModel.Amount = postParams.Amount
		dataModel.Summary = postParams.Summary
		dataModel.ApplyPerson = Session.GetString("login_admin")
		dataModel.Status = "Unaudit"

		if postParams.BankrollDirection == "Freeze" {
			dataModel.Status = "Freeze"
		}
		err = tx.Select("MerchantNo", "MerchantID", "PlatformOrderNo", "BankrollDirection", "BankrollType", "Amount", "Summary", "ApplyPerson", "Status").Create(&dataModel).Error
		if err != nil {
			logrus.Error(postParams.MerchantNo, "-插入余额调整记录失败 : ", err.Error())
			return err
		}
		//TODO:插入管理员操作日志
		/*SystemAccountActionLog::insert([
				'action' => 'CREATE_BALANCE_ADJUSTMENT',
				'actionBeforeData' => '',
				'actionAfterData' => $model->toJson(),
				'status' => 'Success',
				'accountId' => $_SESSION['accountId'],
				'ip' => Tools::getIp(),
				'ipDesc' => Tools::getIpDesc(),
		]);*/
		//插入审核日志
		if postParams.BankrollDirection != "Freeze" {
			content := make(map[string]interface{})
			content["sourceId"] = dataModel.AdjustmentID
			content["bankrollType"] = dataModel.BankrollType
			content["bankrollDirection"] = dataModel.BankrollDirection
			content["amount"] = dataModel.Amount
			content["summary"] = dataModel.Summary
			content["merchantId"] = dataModel.MerchantID
			content["merchantNo"] = dataModel.MerchantNo
			content["platformOrderNo"] = dataModel.PlatformOrderNo
			content["sysFee"] = postParams.SysFee
			content["factFee"] = postParams.FactFee
			content["type"] = "余额调整-" + postParams.BankrollDirection
			jsonStr, _ := json.Marshal(content)
			systemlog := model.SystemCheckLog{}
			systemlog.AdminId = 0
			systemlog.Status = "0"
			systemlog.CommiterId = LoginAdminId
			systemlog.Content = string(jsonStr)
			systemlog.Relevance = postParams.MerchantNo
			systemlog.Desc = ""
			userIp := utils.GetRealIp(req)
			systemlog.Ip = userIp
			systemlog.IpDesc = ""
			systemlog.Type = "余额调整"
			err = tx.Select("AdminId", "Status", "CommiterId", "Content", "Relevance", "Desc", "Ip", "IpDesc", "Type").Create(&systemlog).Error
			if err != nil {
				logrus.Error(postParams.MerchantNo, "-插入审核记录失败 : ", err.Error())
				return err
			}
		}
		//TODO:充值
		//if($bankrollDirection == 'Recharge'){}//生成充值订单

		//冻结跟追收奖金相应发生变化，返还审核后才发生
		if postParams.BankrollDirection == "Retrieve" || postParams.BankrollDirection == "Freeze" {
			merchantAmount.SettlementAmount = merchantAmount.SettlementAmount - postParams.Amount
		}
		if postParams.BankrollDirection == "Freeze" {
			merchantAmount.FreezeAmount = merchantAmount.FreezeAmount + postParams.Amount
		}
		err = tx.Save(&merchantAmount).Error
		if err != nil {
			logrus.Error(postParams.MerchantNo, "-更改资金失败 : ", err.Error())
			return err
		}
		if postParams.BankrollDirection != "Restore" && postParams.BankrollDirection != "Recharge" {
			sourceDesc := ""
			if postParams.BankrollDirection == "Retrieve" {
				sourceDesc = "追收"
			} else {
				sourceDesc = "冻结"
			}
			finance := model.Finance{}
			finance.MerchantID = dataModel.MerchantID
			finance.MerchantNo = dataModel.MerchantNo
			finance.PlatformOrderNo = platformOrderNo
			finance.Amount = postParams.Amount
			finance.Balance = merchantAmount.SettlementAmount
			finance.FinanceType = "PayOut"
			finance.AccountDate = time.Now().Format("2006-01-02")
			finance.AccountType = "SettlementAccount"
			finance.SourceID = dataModel.AdjustmentID
			finance.SourceDesc = "余额调整-" + sourceDesc + "-账户资金"
			finance.OperateSource = "admin"
			finance.Summary = postParams.Summary
			err = tx.Create(&finance).Error
			if err != nil {
				logrus.Error(postParams.MerchantNo, "-插入财务记录失败 : ", err.Error())
				return err
			}
			err = tx.Table("amount_pay").Where("accountDate = ?", time.Now().Format("2006-01-02")).Update("balance", merchantAmount.SettlementAmount).Error
			if err != nil {
				logrus.Error(postParams.MerchantNo, "-更新amount_pay失败 : ", err.Error())
				return err
			}
		}

		return nil
	})
	if err != nil {
		logrus.Error(postParams.MerchantNo, err.Error())
		redis.Del(c.Ctx, "balanceOperation"+postParams.MerchantNo)
		redis.Del(c.Ctx, balanceKey)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "操作成功"})
	return
}

func (c *MerchantController) GetBalanceadjustmentSearch() {
	queryParams := model.BalanceAdjustmentSearch{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	cnd := sqls.DB()
	if queryParams.Status != "" {
		cnd = cnd.Where("status = ?", queryParams.Status)
	}
	if queryParams.MerchantNo != "" {
		cnd = cnd.Where("merchantNo = ?", queryParams.MerchantNo)
	}
	if queryParams.PlatformOrderNo != "" {
		cnd = cnd.Where("platformOrderNo = ?", queryParams.PlatformOrderNo)
	}

	if queryParams.BankrollType != "" {
		cnd = cnd.Where("bankrollType = ?", queryParams.BankrollType)
	}

	if queryParams.BankrollDirection != "" {
		cnd = cnd.Where("bankrollDirection = ?", queryParams.BankrollDirection)
	}

	if queryParams.BankrollDirection != "" {
		cnd = cnd.Where("bankrollDirection = ?", queryParams.BankrollDirection)
	}

	if queryParams.BeginTime != "" {
		cnd = cnd.Where("m.created_at >= ?", queryParams.BeginTime)
	}
	if queryParams.EndTime != "" {
		cnd = cnd.Where("m.created_at <= ?", queryParams.EndTime)
	}

	if queryParams.AuditBeginTime != "" {
		cnd = cnd.Where("m.auditBeginTime >= ?", queryParams.AuditBeginTime)
	}
	if queryParams.AuditEndTime != "" {
		cnd = cnd.Where("m.auditEndTime <= ?", queryParams.AuditEndTime)
	}

	var merchants []model.BalanceAdjustment
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(queryParams.Offset).Order("adjustmentId desc").Find(&merchants).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": merchants})
	return
}

func (c MerchantController) GetAuditResetpassword(request *http.Request) {
	rowId, err := c.Ctx.URLParamInt64("id")
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	checkPwd := c.Ctx.URLParamTrim("checkPwd")
	auditType := c.Ctx.URLParamTrim("passwordtype")
	if len(checkPwd) > 10 || len(auditType) > 20 {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "参数错误"})
		return
	}
	unicheckKey := "passwordAuditCkeck" + LoginAdmin
	redisClient := config.NewRedis()
	unicheck, err := redisClient.Exists(c.Ctx, unicheckKey).Result()
	if unicheck > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "重复提交"})
		return
	}
	redisClient.SetEX(c.Ctx, unicheckKey, 1, 10*time.Second)

	result := redisClient.Get(c.Ctx, "checkPwd:check:count:"+strconv.FormatInt(LoginAdminId, 10))
	if result.Err() != nil && result.Err() != redis.Nil {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": result.Err().Error()})
		return
	}
	checkcount, err := result.Int64()
	if checkcount >= 5 {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "密码错误次数过多"})
		return
	}
	systemaccount := model.SystemAccount{}
	err = sqls.DB().First(&systemaccount, LoginAdminId).Error
	if err != nil {
		logrus.Error("查询系统账号失败：", err)
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "查询系统账号失败-" + err.Error()})
		return
	}
	checkPwd = mytool.GetHashPassword(checkPwd)
	if checkPwd != systemaccount.CheckPwd {
		checkcount = checkcount + 1
		redisClient.SetEX(c.Ctx, "checkPwd:check:count:"+strconv.FormatInt(LoginAdminId, 10), checkcount, 72*time.Hour)
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "审核密码错误"})
		return
	}

	systemCheckLog := model.SystemCheckLog{}
	err = sqls.DB().Where("id = ?", rowId).Where("type = ? ", auditType).First(&systemCheckLog).Error
	if err != nil {
		logrus.Error("查询数据失败：", err)
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "查询数据失败-" + err.Error()})
		return
	}
	content := model.LoginPassContent{}
	err = json.Unmarshal([]byte(systemCheckLog.Content), &content)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "json.Unmarshal :" + err.Error()})
		return
	}
	if content.AccountId <= 0 || content.Password == "" {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "content格式错误 :" + systemCheckLog.Content})
		return
	}
	merchantData := model.MerchantAccount{}
	err = sqls.DB().Where("accountId = ?", content.AccountId).First(&merchantData).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "查询商户失败 :" + err.Error()})
		return
	}

	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		merchantAccount := model.MerchantAccount{}
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("accountId = ?", content.AccountId).First(&merchantAccount).Error
		if err != nil {
			logrus.Error(content.AccountId, "-查询商户账号失败 : ", err.Error())
			return err
		}
		beforeMerchantAccount := merchantAccount
		if auditType == "支付密码修改" {
			merchantAccount.SecurePwd = mytool.GetHashPassword(content.Password)
		} else {
			merchantAccount.LoginPwd = mytool.GetHashPassword(content.Password)
		}
		merchantAccount.LoginFailNum = 0
		err = tx.Omit("loginPwdAlterTime", "updated_at", "latestLoginTime", "googleAuthSecretKey").Save(&merchantAccount).Error
		if err != nil {
			logrus.Error(content.AccountId, "-修改商户账号失败 : ", err.Error())
			return err
		}

		UserIP := utils.GetRealIp(request)
		actionLog := make(map[string]interface{})
		actionLog["action"] = "UPDATE_MERCHANT_ACCOUNT_PAY_PASSWORD"
		actionLog["actionBeforeData"], _ = json.Marshal(beforeMerchantAccount)
		actionLog["actionAfterData"], _ = json.Marshal(beforeMerchantAccount)
		actionLog["status"] = "Success"
		actionLog["accountId"] = LoginAdminId
		actionLog["ip"] = UserIP
		actionLog["ipDesc"] = "" //TODO:ip描述
		err = tx.Table("system_account_action_log").Create(actionLog).Error

		if err != nil {
			logrus.Error("修改日志添加失败 : ", err.Error())
			return err
		}
		systemCheckLog.Status = "1"
		systemCheckLog.AdminId = LoginAdminId
		systemCheckLog.CheckIp = UserIP
		systemCheckLog.CheckTime = time.Now().Format("2006-01-02 15:04:05")
		err = tx.Save(systemCheckLog).Error
		if err != nil {
			logrus.Error("systemCheckLog-修改失败 : ", err.Error())
			return err
		}

		return nil
	})

	if err != nil {
		logrus.Error(rowId, err.Error())
		redisClient.Del(c.Ctx, unicheckKey)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	go cache.MerchantAccount.RefreshOne(content.AccountId)
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "操作成功"})
	return
}

func (c MerchantController) GetAuditDisagreepassword(request *http.Request) {
	rowId, err := c.Ctx.URLParamInt64("id")
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	checkPwd := c.Ctx.URLParamTrim("checkPwd")
	auditType := c.Ctx.URLParamTrim("passwordtype")
	if len(checkPwd) > 10 || len(auditType) > 20 {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "参数错误"})
		return
	}
	unicheckKey := "passwordAuditCkeck" + LoginAdmin
	redisClient := config.NewRedis()
	unicheck, err := redisClient.Exists(c.Ctx, unicheckKey).Result()
	if unicheck > 0 {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "重复提交"})
		return
	}
	redisClient.SetEX(c.Ctx, unicheckKey, 1, 10*time.Second)

	result := redisClient.Get(c.Ctx, "checkPwd:check:count:"+strconv.FormatInt(LoginAdminId, 10))
	if result.Err() != nil && result.Err() != redis.Nil {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": result.Err().Error()})
		return
	}
	checkcount, err := result.Int64()
	if checkcount >= 5 {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "密码错误次数过多"})
		return
	}
	systemaccount := model.SystemAccount{}
	err = sqls.DB().First(&systemaccount, LoginAdminId).Error
	if err != nil {
		logrus.Error("查询系统账号失败：", err)
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "查询系统账号失败-" + err.Error()})
		return
	}
	checkPwd = mytool.GetHashPassword(checkPwd)
	if checkPwd != systemaccount.CheckPwd {
		checkcount = checkcount + 1
		redisClient.SetEX(c.Ctx, "checkPwd:check:count:"+strconv.FormatInt(LoginAdminId, 10), checkcount, 72*time.Hour)
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "审核密码错误"})
		return
	}

	systemCheckLog := model.SystemCheckLog{}
	err = sqls.DB().Where("id = ?", rowId).Where("type = ? ", auditType).First(&systemCheckLog).Error
	if err != nil {
		logrus.Error("查询数据失败：", err)
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "查询数据失败-" + err.Error()})
		return
	}

	err = sqls.DB().Transaction(func(tx *gorm.DB) error {

		UserIP := utils.GetRealIp(request)
		actionLog := make(map[string]interface{})
		actionLog["action"] = "UPDATE_MERCHANT_ACCOUNT_PASSWORD"
		actionLog["actionBeforeData"], _ = json.Marshal(systemCheckLog)
		actionLog["status"] = "Success"
		actionLog["accountId"] = LoginAdminId
		actionLog["ip"] = UserIP
		actionLog["ipDesc"] = "" //TODO:ip描述
		err = tx.Table("system_account_action_log").Create(actionLog).Error

		if err != nil {
			logrus.Error("修改日志添加失败 : ", err.Error())
			return err
		}
		systemCheckLog.Status = "2"
		systemCheckLog.AdminId = LoginAdminId
		systemCheckLog.CheckIp = UserIP
		systemCheckLog.CheckTime = time.Now().Format("2006-01-02 15:04:05")
		err = tx.Save(systemCheckLog).Error
		if err != nil {
			logrus.Error("systemCheckLog-修改失败 : ", err.Error())
			return err
		}

		return nil
	})

	if err != nil {
		logrus.Error(rowId, err.Error())
		redisClient.Del(c.Ctx, unicheckKey)
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "操作成功"})
	return
}

/*func (c *MerchantController) PostRateImportBak(r *http.Request){
	if err := r.ParseMultipartForm(1 * MB); err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	merchantNo := c.Ctx.PostValueInt64Default("merchantNo",0)
	if merchantNo <= 0{
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户号错误"})
		return
	}
	merchantNoStr := fmt.Sprintf("%d", merchantNo)
	_,res :=services.MerchantService.GetCacheByMerchantNo(merchantNoStr)
	if !res {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": "商户不存在" + merchantNoStr})
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
	reader := csv.NewReader(file)

	csvData, err := reader.ReadAll() // 读取全部数据
	if err != nil {
		fmt.Println(err)
	}
	for _, line := range csvData {
		fmt.Println(line)
	}
}*/

package gm

import (
	"encoding/json"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"luckypay/cache"
	"luckypay/config"
	"luckypay/model"
	"strings"
)

type ApiController struct {
	Ctx iris.Context
}

func (c *ApiController) GetBasedata() {
	//items := $request->getParam('requireItems');
	requireItems := c.Ctx.FormValueDefault("requireItems", "payType")
	items := strings.Split(strings.TrimSpace(requireItems), ",")
	dataMap := make(map[string]interface{})
	for _, item := range items {
		//childMap = append(dataMap)
		childs := []map[string]string{}
		if item == "merchants" {
			childs = c.GetCacheMerchantData()
		} else {
			for key, value := range config.BaseData[item] {
				child := make(map[string]string)
				child["key"] = key
				child["value"] = value
				childs = append(childs, child)
			}
		}

		dataMap[item] = childs
	}
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": dataMap})
	//fmt.Print(items)
}

/*func (c *ApiController) GetFreshCache(){
	services.MerchantService.RefreshCache()
	services.MerchantService.SetCacheMerchantData()
	services.MerchantAmount.RefreshCache()
	services.MerchantAccountService.RefreshCache()
	services.MerchantRate.RefreshCache()
	services.MerchantChannel.RefreshCache()
	services.ChannelMerchant.RefreshCache()
	services.ChannelMerchantRate.RefreshCache()
}*/

func (c *ApiController) GetCacheMerchantData() (rows []map[string]string) {
	//services.MerchantService.SetCacheMerchantData()
	key := "merchants"
	rdb := config.NewRedis()
	cacheData, err := rdb.Get(c.Ctx, key).Result()
	if err != nil {
		go cache.MerchantCache.SetCacheMerchantData()
		logrus.Error("GetCacheMerchantData error : " + err.Error())
		return
	}
	var merchantData []model.Merchant
	err = json.Unmarshal([]byte(cacheData), &merchantData)
	if err != nil {
		go cache.MerchantCache.SetCacheMerchantData()
		logrus.Error("GetCacheMerchantData error : " + err.Error())
		return
	}
	//var rows []map[string]string
	for _, value := range merchantData {
		child := make(map[string]string)
		child["key"] = value.MerchantNo
		child["value"] = value.ShortName
		rows = append(rows, child)
	}
	return rows
	//c.Ctx.StopWithJSON(200,iris.Map{"success":1,"result":rows})
	//return

}

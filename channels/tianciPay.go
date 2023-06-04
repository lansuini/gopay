package channels

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"luckypay/model"
	"luckypay/utils"
	"net/http"
)

var TianciPay = newTianciPay()

func newTianciPay() *tianciPay {
	return &tianciPay{}
}

type tianciPay struct {
}

func (ch *tianciPay) PayOrder(params model.PayParams, channelMerchant model.ChannelMerchant) (res model.RspPay, err error) {
	fmt.Println("%+v", params)
	fmt.Println("running aliPay PayOrder")
	res.ChannelOrderNo = ""
	res.PayUrl = ""
	res.Status = "Fail"
	res.FailReason = "Fail"

	channelParam := utils.AesCBCDecrypt(channelMerchant.Param)
	var decryptChannelParam = model.TianciPayMerchantParams{}
	err = json.Unmarshal([]byte(channelParam), &decryptChannelParam)
	if err != nil {
		res.FailReason = "商户参数错误"
		return
	}

	//path := "https://tiancip030905.com" + "/api/transaction"
	path := channelMerchant.DelegateDomain

	callbackUrl := GetPayCallbackUrl(params.PlatformOrderNo)

	tcPayStruct := model.TcPay{
		OutTradeNo:  params.PlatformOrderNo,
		Amount:      params.OrderAmoumt,
		CallbackUrl: callbackUrl,
	}
	body, _ := json.Marshal(tcPayStruct)
	response, _ := httpPostJson(path, body, decryptChannelParam.MerchantPrivateKey)
	result := make(map[string]any)
	err = json.Unmarshal(response, &result)
	if err != nil {
		logrus.Info(params.PlatformOrderNo, " : Umarshal failed:", err)
		logrus.Info(params.PlatformOrderNo, " : ", res)
		return
	}
	fmt.Println("m:", result)
	if result["success"] != true {
		fmt.Println("success :", result["success"])
		fmt.Println("message :", result["message"])
		fmt.Println("errors :", result["errors"])
		res.FailReason = string(response)
		logrus.Info(params.PlatformOrderNo, " : ", res)
		return
	}
	data := result["data"]
	md, _ := data.(map[string]any)
	res.ChannelOrderNo = fmt.Sprintf("%v", md["trade_no"])
	res.PayUrl = fmt.Sprintf("%v", md["uri"])
	res.Status = "Success"
	//fmt.Print(res)
	logrus.Info(params.PlatformOrderNo, " : ", string(response))
	return res, nil
}

func (ch *tianciPay) SettlementOrder(params model.SettleParams, channelMerechant model.ChannelMerchant) (res model.RspSettle, err error) {
	fmt.Println("running tianciPay SettlementOrder")
	return
}

func (ch *tianciPay) QuerySettlementOrder(order model.PlatformSettlementOrder, channelMerechant model.ChannelMerchant) (result model.RspQuerySettle, err error) {

	result.Status = "Fail"
	result.OrderStatus = "Execute"
	result.FailReason = "default"
	channelParam := utils.AesCBCDecrypt(channelMerechant.Param)
	//json字符串转结构体
	decryptChannelparam := model.LoroPayMerchantParams{}
	err = json.Unmarshal([]byte(channelParam), &decryptChannelparam)
	if err != nil {
		logrus.Info(order.PlatformOrderNo, " : Umarshal failed:", err)
		return
	}
	queryMap := make(map[string]interface{})
	queryMap["merchantNo"] = decryptChannelparam.MerchantNo
	queryMap["merchantOrderNo"] = order.PlatformOrderNo
	sign := ""

	queryMap["sign"] = sign
	body, _ := json.Marshal(queryMap)

	path := channelMerechant.DelegateDomain + "/api/cash/query"
	response, err, statusCode := httpPost(path, body)
	//rsp := make(map[string]interface{})
	rsp := LoroPayRsp{}
	err = json.Unmarshal(response, &rsp)
	if err != nil {
		logrus.Info(order.PlatformOrderNo, " : Umarshal failed:", err)
		result.FailReason = err.Error()
		return
	}
	if statusCode == 202 && rsp.Status == "404" && rsp.Message == "payout order not exist" {
		result.Status = "Success"
		result.OrderStatus = "Fail"
		result.FailReason = fmt.Sprintf("%v%v", rsp.Status, rsp.Message)
		return
	}
	//_, statusOk := rsp["status"]
	//if statusCode == 202 && statusOk && rsp["status"] == "404" && rsp["message"] == "payout order not exist"{
	//	result.Status = "Success"
	//	result.OrderStatus = "Fail"
	//	result.FailReason = fmt.Sprintf("%v%v", rsp["status"], rsp["message"])
	//	return
	//}
	if statusCode != 200 {
		result.Status = "Fail"
		result.FailReason = string(response)
		return
	}
	result.FailReason = rsp.Message
	if rsp.Data.OrderStatus == "SUCCESS" {
		result.Status = "Success"
		result.OrderStatus = "Success"
		return
	} else if rsp.Data.OrderStatus == "FAILED" {
		result.Status = "Success"
		result.OrderStatus = "Fail"
		return
	} else {
		result.Status = "Success"
		result.OrderStatus = "Execute"
		return
	}

	return
}

func (ch *tianciPay) QueryPayOrder(order model.PlatformPayOrder, channelMerchant model.ChannelMerchant) (res model.RspQuerySettle, err error) {
	//funcs := map[string]interface{}{
	//	"foo0": ch.PayOrder,
	//}
	return
}

func (ch *tianciPay) QueryBalance(channelMerchant model.ChannelMerchant) (res model.RspQueryBalance, err error) {
	//funcs := map[string]interface{}{
	//	"foo0": ch.PayOrder,
	//}
	return
}

func (ch *tianciPay) CallBack() {

}

func (ch *tianciPay) PayCallBack(ctx iris.Context) (res model.RspQuerySettle, err error) {
	return
}

func (ch *tianciPay) SettleCallBack(ctx iris.Context) (res model.RspQuerySettle, err error) {

	return
}

func (ch *tianciPay) Test() {
	//funcs := map[string]interface{}{
	//	"foo0": ch.PayOrder,
	//}
}

func httpPostJson(reqUrl string, requestBody []byte, token string) ([]byte, error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	//`这里请注意，使用 InsecureSkipVerify: true 来跳过证书验证`
	client = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	//log.Println("request_data:", string(body))
	logrus.Info("request_data:", string(requestBody))
	// 获取 request请求
	request, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("GetHttpSkip Request Error:", err)
		return nil, nil
	}
	// 加入 token
	//token := "Bearer F8Q6MKxWb4B0QnR2i6Wuge37oNgmhkdNexI9vWduFzauO3ZIHBk28xO1qUrZ"
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Add("Authorization", token)
	resp, err = client.Do(request.WithContext(context.TODO()))

	if err != nil {
		logrus.Info(string(requestBody), "StatusCode:", resp.StatusCode, "GetHttpSkip Response Error:", err)
		log.Println("GetHttpSkip Response Error:", err)
		return nil, nil
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	defer client.CloseIdleConnections()
	//fmt.Println("Response: ", string(responseBody))
	logrus.Info(string(requestBody), "StatusCode:", resp.StatusCode, "resp:", string(responseBody))

	return responseBody, nil
}

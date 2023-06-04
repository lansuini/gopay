package channels

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"

	"log"
	"luckypay/cache"
	"luckypay/model"
	"luckypay/utils"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var LoroPay = newLoroPay()

func newLoroPay() *loroPay {
	return &loroPay{}
}

type LoroPayRspData struct {
	MerchantNo      string  `json:"merchantNo"`
	PlatOrderNo     string  `json:"platOrderNo"`
	MerchantOrderNo string  `json:"merchantOrderNo"`
	OrderStatus     string  `json:"orderStatus"`
	MerchantFee     float64 `json:"merchantFee"`
}

type LoroPayRsp struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    LoroPayRspData `json:"data"`
}

type loroPay struct {
}

func (ch *loroPay) PayOrder(params model.PayParams, channelMerchant model.ChannelMerchant) (res model.RspPay, err error) {
	logrus.Info("running loroPay PayOrder")
	res.ChannelOrderNo = ""
	res.PayUrl = ""
	res.Status = "Fail"
	res.FailReason = "Fail"
	channelParam := utils.AesCBCDecrypt(channelMerchant.Param)
	//json字符串转结构体
	var decryptChannelParam = model.LoroPayMerchantParams{}
	err = json.Unmarshal([]byte(channelParam), &decryptChannelParam)
	if err != nil {
		res.FailReason = "商户参数错误"
		return
	}
	payMap := make(map[string]interface{})
	payMap["merchantNo"] = decryptChannelParam.MerchantNo
	payMap["merchantOrderNo"] = params.PlatformOrderNo
	payMap["payAmount"] = params.OrderAmoumt
	payMap["method"] = params.PayType
	payMap["name"] = params.PayType
	payMap["feeType"] = 0
	payMap["mobile"] = "9953073841"
	payMap["email"] = "9953073841@gmail.com"
	payMap["expiryPeriod"] = "1440"
	/*tcPayStruct := model.LoroPayParams{
		MerchantNo:  params.MerchantNo,
		MerchantOrderNo:      params.PlatformOrderNo,
		PayAmount:      params.OrderAmoumt,
		Method: params.PayType,
		Name: params.PayType,
		FeeType: 0,
		Mobile: "9953073841",
		Email: "9953073841@gmail.com",
		ExpiryPeriod: "1440",
	}*/
	signStr := ch.createSignStr(payMap)
	sign := ch.GetSign(decryptChannelParam.MerchantPrivateKey, signStr)
	if sign == "" {
		res.FailReason = "GetSign Fail"
		return
	}
	payMap["sign"] = sign
	body, _ := json.Marshal(payMap)

	path := channelMerchant.DelegateDomain + "/api/pay/code"
	response, err, statusCode := httpPost(path, body)
	logrus.Info(params.PlatformOrderNo, " : ", string(response))
	if err != nil {
		res.FailReason = err.Error()
		return
	}
	if statusCode != 200 {
		res.FailReason = "request Fail"
		return
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(response, &result)
	if err != nil {
		logrus.Info(params.PlatformOrderNo, " : Umarshal failed:", err)
		return
	}
	if (result["status"] != "200") || (result["message"] != "success") {
		res.FailReason = string(response)
		return
	}
	if _, ok := result["data"]; !ok {
		res.FailReason = string(response)
		return
	}
	data := result["data"]
	md, _ := data.(map[string]interface{})
	if (md["orderStatus"] != "PENDING") || (md["paymentLink"] == "") || md["paymentLink"] == nil {
		res.FailReason = string(response)
		return
	}
	res.ChannelOrderNo = fmt.Sprintf("%v", md["platOrderNo"])
	res.PayUrl = fmt.Sprintf("%v", md["paymentLink"])
	res.Status = "Success"
	return res, nil
}

func (ch *loroPay) SettlementOrder(params model.SettleParams, channelMerchant model.ChannelMerchant) (res model.RspSettle, err error) {

	logrus.Info("running loroPay SettlementOrder-", params.PlatformOrderNo)
	channelParam := utils.AesCBCDecrypt(channelMerchant.Param)
	//json字符串转结构体
	decryptChannelparam := model.LoroPayMerchantParams{}
	err = json.Unmarshal([]byte(channelParam), &decryptChannelparam)
	if err != nil {
		logrus.Info(params.PlatformOrderNo, " : Umarshal failed:", err)
		return
	}
	/*tcPayStruct := model.LoroSettle{
		MerchantNo:      channelMerechant.ChannelMerchantNo,
		MerchantOrderNo: params.PlatformOrderNo,
		PayAmount:       params.OrderAmoumt,
		Description:     "daifu",
		BankCode:        "baode",
		BankNumber:      params.BankAccountNo,
		AccountHoldName: "acntHoldName",
		Address:         params.City,
		Barangay:        "Maharlika",
		City:            "Pasig",
		ZipCode:         "1110",
		Gender:          "Male",
		FirstName:       params.BankAccountName,
		MiddleName:      params.BankAccountName,
		LastName:        params.BankAccountName,
		Mobile:          "9158838275",
	}*/
	payMap := make(map[string]interface{})
	payMap["merchantNo"] = decryptChannelparam.MerchantNo
	payMap["merchantOrderNo"] = params.PlatformOrderNo
	payMap["payAmount"] = params.OrderAmoumt
	payMap["description"] = "daifu"
	payMap["bankCode"] = params.BankCode
	payMap["bankNumber"] = params.BankAccountNo
	payMap["accountHoldName"] = params.BankAccountName
	payMap["address"] = params.City
	payMap["barangay"] = "Maharlika"
	payMap["city"] = "Pasig"
	payMap["zipCode"] = 1110
	payMap["gender"] = "Male"
	payMap["firstName"] = params.BankAccountName
	payMap["middleName"] = params.BankAccountName
	payMap["lastName"] = params.BankAccountName
	payMap["mobile"] = "9158838275"
	signStr := ch.createSignStr(payMap)
	sign := ch.GetSign(decryptChannelparam.MerchantPrivateKey, signStr)
	if sign == "" {
		err = errors.New("获取Sign失败")
		return
	}
	//tcPayStruct.Sign = sign
	payMap["sign"] = sign
	body, _ := json.Marshal(payMap)

	path := channelMerchant.DelegateDomain + "/api/cash"
	response, err, statusCode := httpPost(path, body)
	if err != nil {
		logrus.Error("请求代付错误：", err)
		return
	}
	result := make(map[string]any)
	err = json.Unmarshal(response, &result)
	if err != nil {
		logrus.Info(params.PlatformOrderNo, " : Umarshal failed:", err)
		logrus.Info(params.PlatformOrderNo, " : ", res)
		res.PushChannelTime = time.Now()
		res.Status = "Success"
		res.FailReason = string(response)
		return
	}
	_, statusOk := result["status"]
	if statusCode != 200 && statusOk && result["status"] != "200" {
		res.Status = "Fail"
		res.FailReason = fmt.Sprintf("%v%v", result["status"], result["message"])
		res.PushChannelTime = time.Now()
		return
	}
	if result["status"] != "200" {
		res.Status = "Fail"
		res.FailReason = fmt.Sprintf("%v%v", result["status"], result["message"])
		res.PushChannelTime = time.Now()
		return
	}
	data := result["data"]
	md, _ := data.(map[string]any)
	res.ChannelOrderNo = fmt.Sprintf("%v", md["platOrderNo"])
	res.PushChannelTime = time.Now()
	res.Status = "Success"
	//$output['status'] = 'Success';
	//$output['orderNo'] = $res['data']['platOrderNo'] ?? '';
	//$output['failReason'] = '';
	//$output['orderAmount'] = $res['data']['payAmount'];
	//$output['pushChannelTime'] = date('YmdHis');
	return
}

func (ch *loroPay) QuerySettlementOrder(order model.PlatformSettlementOrder, channelMerechant model.ChannelMerchant) (result model.RspQuerySettle, err error) {

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
	signStr := ch.createSignStr(queryMap)
	sign := ch.GetSign(decryptChannelparam.MerchantPrivateKey, signStr)
	if sign == "" {
		err = errors.New("获取Sign失败")
		return
	}
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
		result.ChannelOrderNo = rsp.Data.PlatOrderNo
		result.OrderStatus = "Success"
		return
	} else if rsp.Data.OrderStatus == "FAILED" {
		result.Status = "Success"
		result.ChannelOrderNo = rsp.Data.PlatOrderNo
		result.OrderStatus = "Fail"
		return
	} else {
		result.Status = "Success"
		result.OrderStatus = "Execute"
		return
	}

	return
}

func (ch *loroPay) QueryPayOrder(order model.PlatformPayOrder, channelMerchant model.ChannelMerchant) (result model.RspQuerySettle, err error) {
	result.Status = "Fail"
	result.OrderStatus = "Execute"
	result.FailReason = "default"

	channelParam := utils.AesCBCDecrypt(channelMerchant.Param)
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
	signStr := ch.createSignStr(queryMap)
	sign := ch.GetSign(decryptChannelparam.MerchantPrivateKey, signStr)
	if sign == "" {
		err = errors.New("获取Sign失败")
		return
	}
	queryMap["sign"] = sign
	body, _ := json.Marshal(queryMap)

	path := channelMerchant.DelegateDomain + "/api/pay/query"
	response, err, statusCode := httpPost(path, body)
	//rsp := make(map[string]interface{})
	rsp := LoroPayRsp{}
	err = json.Unmarshal(response, &rsp)
	if err != nil {
		logrus.Info(order.PlatformOrderNo, " : Umarshal failed:", err)
		result.FailReason = err.Error()
		return
	}

	if statusCode != 200 {
		result.Status = "Fail"
		result.FailReason = string(response)
		return
	}
	result.FailReason = rsp.Message
	if rsp.Data.OrderStatus == "SUCCESS" {
		result.Status = "Success"
		result.ChannelOrderNo = rsp.Data.PlatOrderNo
		result.OrderStatus = "Success"
		return
	} else {
		result.Status = "Success"
		result.OrderStatus = "Execute"
		return
	}
	return
}

func (ch *loroPay) QueryBalance(channelMerchant model.ChannelMerchant) (res model.RspQueryBalance, err error) {
	channelParam := utils.AesCBCDecrypt(channelMerchant.Param)
	//json字符串转结构体
	decryptChannelParam := model.LoroPayMerchantParams{}
	err = json.Unmarshal([]byte(channelParam), &decryptChannelParam)
	if err != nil {
		logrus.Info(" QueryBalance Umarshal failed:", err)
		return
	}
	querys := make(map[string]interface{})
	querys["merchantNo"] = decryptChannelParam.MerchantNo
	querys["timestamp"] = time.Now().Second()
	signStr := ch.createSignStr(querys)
	sign := ch.GetSign(decryptChannelParam.MerchantPrivateKey, signStr)
	if sign == "" {
		res.FailReason = "GetSign Fail"
		return
	}
	querys["sign"] = sign
	body, _ := json.Marshal(querys)

	path := channelMerchant.DelegateDomain + "/api/balance"
	logrus.Info("loroPay 代付查询-", querys)
	response, err, statusCode := httpPost(path, body)
	if err != nil {
		logrus.Error("响应状态：", statusCode, "-请求代付错误：", err)
		return
	}
	result := model.LoroPayQueryBalanceRes{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		logrus.Info("loroPay response Umarshal failed:", err)
		res.Status = "Fail"
		res.FailReason = "loroPay response Umarshal failed:" + err.Error()
		return
	}
	if result.Status == "200" {
		res.Status = "Success"
		res.Balance = result.Data.AvailableAmount
		res.FailReason = string(response)
	} else {
		res.Status = "Fail"
		res.Balance = 0
		res.FailReason = string(response)
	}
	return
}

func (ch *loroPay) CallBack() {
	//funcs := map[string]interface{}{
	//	"foo0": ch.PayOrder,
	//}

}

func (ch *loroPay) PayCallBack(ctx iris.Context) (res model.RspQuerySettle, err error) {
	callbackParams := model.LoroPayPayCallback{}
	var readParams map[string]interface{}
	if err = ctx.ReadJSON(&readParams); err != nil {
		logrus.Error("loropay PayCallBack ReadJSON Error:", err.Error())
		return
		//ctx.JSON(iris.Map{"success": 0, "result": callbackParams})
	}
	jsonStr, err := json.Marshal(readParams)
	if err != nil {
		logrus.Error("loropay SettleCallBack Marshal Error:", err)
		return
	}
	err = json.Unmarshal(jsonStr, &callbackParams)
	if err != nil {
		logrus.Error("loropay SettleCallBack Unmarshal Error:", err)
		return
	}
	if callbackParams.MerchantOrderNo == "" {
		err = errors.New("MerchantOrderNo is null")
		logrus.Error("loropay PayCallBack callbackParams Error:", err)
		return
	}
	if callbackParams.OrderStatus != "SUCCESS" && callbackParams.OrderStatus != "FAILED" {
		err = errors.New("OrderStatus is not clear")
		logrus.Error("loropay PayCallBack callbackParams Error:", err)
		return
	}

	order, cacheRes := cache.PayCache.GetCacheByPlatformOrderNo(callbackParams.MerchantOrderNo)
	if !cacheRes {
		err = errors.New("订单获取失败")
		logrus.Error("loropay PayCallBack Error:", err, callbackParams.MerchantOrderNo)
		return
	}

	if order.OrderStatus != "WaitPayment" {
		err = errors.New("订单已完成")
		logrus.Error("loropay PayCallBack Error:", err, callbackParams.MerchantOrderNo)
		return
	}
	channelMerechant, cacheRes := cache.ChannelMerchant.GetCacheByChannelMerchantNo(order.ChannelMerchantNo)
	if !cacheRes {
		err = errors.New("订单渠道商户信息获取失败")
		logrus.Error("loropay PayCallBack Error:", err, callbackParams.MerchantOrderNo)
		return
	}

	channelParam := utils.AesCBCDecrypt(channelMerechant.Param)
	//json字符串转结构体
	decryptChannelparam := model.LoroPayMerchantParams{}
	err = json.Unmarshal([]byte(channelParam), &decryptChannelparam)
	if err != nil {
		logrus.Info(callbackParams.MerchantOrderNo, "loropay PayCallBack : Umarshal failed:", err)
		return
	}
	//dump.Printf(decryptChannelparam)
	paramMap := map[string]interface{}{}
	err = json.Unmarshal(jsonStr, &paramMap)
	if err != nil {
		logrus.Error("loropay PayCallBack Unmarshal Error:", err)
		return
	}
	plainText := ch.createSignStr(paramMap)
	err = ch.VerifySign(plainText, decryptChannelparam.PlatformPublicKey, callbackParams.Sign)
	if err != nil {
		err = errors.New("loropay验签失败：" + err.Error())
		logrus.Error(callbackParams.MerchantOrderNo, "-loropay PayCallBack RsaVerify Error:", err)
		return
	}
	floatAmount, err := strconv.ParseFloat(callbackParams.Amount, 64)
	res.OrderAmount = floatAmount
	res.PlatformOrderNo = callbackParams.MerchantOrderNo
	if callbackParams.OrderStatus == "SUCCESS" {
		res.Status = "Success"
		res.OrderStatus = "Success"
		res.ChannelOrderNo = callbackParams.PlatOrderNo
	} else if callbackParams.OrderStatus == "FAILED" {
		res.Status = "Success"
		res.OrderStatus = "Fail"
		res.ChannelOrderNo = callbackParams.PlatOrderNo
	} else {
		res.Status = "Fail"
		res.FailReason = "回调通知缺少orderStatus或处理中"
	}
	return
}

func (ch *loroPay) SettleCallBack(ctx iris.Context) (res model.RspQuerySettle, err error) {
	callbackParams := model.LoroPaySettleCallback{}
	var readParams map[string]interface{}
	if err = ctx.ReadJSON(&readParams); err != nil {
		logrus.Error("loropay SettleCallBack ReadJSON Error:", err.Error())
		return
		//ctx.JSON(iris.Map{"success": 0, "result": callbackParams})
	}
	jsonStr, err := json.Marshal(readParams)
	if err != nil {
		logrus.Error("loropay SettleCallBack Marshal Error:", err)
		return
	}
	err = json.Unmarshal(jsonStr, &callbackParams)
	if err != nil {
		logrus.Error("loropay SettleCallBack Unmarshal Error:", err)
		return
	}
	if callbackParams.MerchantOrderNo == "" {
		err = errors.New("MerchantOrderNo is null")
		logrus.Error("loropay SettleCallBack callbackParams Error:", err)
		return
	}
	if callbackParams.OrderStatus != "SUCCESS" && callbackParams.OrderStatus != "FAILED" {
		err = errors.New("OrderStatus is not clear")
		logrus.Error("loropay SettleCallBack callbackParams Error:", err)
		return
	}

	order, cacheRes := cache.SettleCache.GetCacheByPlatformOrderNo(callbackParams.MerchantOrderNo)
	if !cacheRes {
		err = errors.New("订单获取失败")
		logrus.Error("loropay SettleCallBack Error:", err, callbackParams.MerchantOrderNo)
		return
	}

	if order.OrderStatus != "Transfered" {
		err = errors.New("订单已完成")
		logrus.Error("loropay SettleCallBack Error:", err, callbackParams.MerchantOrderNo)
		return
	}
	channelMerechant, cacheRes := cache.ChannelMerchant.GetCacheByChannelMerchantNo(order.ChannelMerchantNo)
	if !cacheRes {
		err = errors.New("订单渠道商户信息获取失败")
		logrus.Error("loropay SettleCallBack Error:", err, callbackParams.MerchantOrderNo)
		return
	}
	channelParam := utils.AesCBCDecrypt(channelMerechant.Param)
	//json字符串转结构体
	decryptChannelparam := model.LoroPayMerchantParams{}
	err = json.Unmarshal([]byte(channelParam), &decryptChannelparam)
	if err != nil {
		logrus.Info(callbackParams.MerchantOrderNo, "loropay PayCallBack : Umarshal failed:", err)
		return
	}
	paramMap := map[string]interface{}{}
	err = json.Unmarshal(jsonStr, &paramMap)
	if err != nil {
		logrus.Error("loropay SettleCallBack Unmarshal Error:", err)
		return
	}
	plainText := ch.createSignStr(paramMap)
	err = ch.VerifySign(plainText, decryptChannelparam.PlatformPublicKey, callbackParams.Sign)
	if err != nil {
		err = errors.New("loropay验签失败")
		logrus.Error(callbackParams.MerchantOrderNo, "-loropay SettleCallBack RsaVerify Error:", err)
		return
	}
	floatAmount, err := strconv.ParseFloat(callbackParams.Amount, 64)
	res.OrderAmount = floatAmount
	res.PlatformOrderNo = callbackParams.MerchantOrderNo
	if callbackParams.OrderStatus == "SUCCESS" {
		res.Status = "Success"
		res.OrderStatus = "Success"
		res.ChannelOrderNo = callbackParams.PlatOrderNo
	} else if callbackParams.OrderStatus == "FAILED" {
		res.Status = "Success"
		res.OrderStatus = "Fail"
		res.ChannelOrderNo = callbackParams.PlatOrderNo
	} else {
		res.Status = "Fail"
		res.FailReason = "回调通知缺少orderStatus或处理中"
	}
	return
}

func (ch *loroPay) createSignStr(settleParams map[string]interface{}) (signStr string) {

	//settleMap := structs.Map(settleParams)
	sortParam := utils.Params{}
	for key, value := range settleParams {
		if value == "" || key == "sign" {
			continue
		}
		strConv2 := fmt.Sprintf("%v", value)
		sortParam = append(sortParam, utils.Onestruct{key, strConv2})
		//u.Add(strings.ToLower(person.Key), person.Value)
	}
	sort.Sort(sortParam)
	var builder strings.Builder
	for _, one := range sortParam {
		if one.Value == "" {
			continue
		}
		if utils.FirstLower(one.Key) == "sign" {
			continue
		}

		builder.WriteString(one.Value)
	}

	signStr = builder.String()
	return signStr
}

func (ch *loroPay) createSignStr1(settleParams model.LoroSettle) (signStr string) {

	settleMap := structs.Map(settleParams)
	sortParam := utils.Params{}

	for key, value := range settleMap {
		if value == "" {
			continue
		}
		strConv2 := fmt.Sprintf("%v", value)
		sortParam = append(sortParam, utils.Onestruct{key, strConv2})
		//u.Add(strings.ToLower(person.Key), person.Value)
	}
	sort.Sort(sortParam)
	var builder strings.Builder
	for _, one := range sortParam {
		if one.Value == "" {
			continue
		}
		if utils.FirstLower(one.Key) == "sign" {
			continue
		}

		builder.WriteString(one.Value)
	}
	signStr = builder.String()

	return signStr
}

func (ch *loroPay) GetSign(pk, context string) string {
	pk = GetRsaPrivateKey(pk)
	block, _ := pem.Decode([]byte(pk))
	if block == nil {
		logrus.Info("私钥错误 : " + pk)
		return ""
	}
	private, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logrus.Error("PrivateKey error")
		return ""
	}
	chunks := split([]byte(context), 117)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		decrypted, err := rsa.SignPKCS1v15(rand.Reader, private, crypto.Hash(0), chunk)
		if err != nil {
			return ""
		}
		buffer.Write(decrypted)
	}
	//h := crypto.SHA256.New()
	//h.Write([]byte(context))
	//hashed := h.Sum(nil)
	//signedData, err := rsa.SignPKCS1v15(nil, private, crypto.Hash(0), []byte(context))
	//if err != nil {
	//	return ""
	//}
	signedData := []byte(buffer.String())
	//hash := sha256.New()
	//hash.Write([]byte(context))
	//sum := hash.Sum(nil)
	//// 从1.5版本规定，使用RSASSA-PKCS1-V1_5-SIGN 方案计算签名
	//signedData, err := rsa.SignPKCS1v15(rand.Reader, private, crypto.SHA256, sum)
	//
	//h := crypto.Hash.New(crypto.SHA1)
	//h.Write([]byte(context))
	//hashed := h.Sum(nil)
	//// 进行rsa加密签名
	//signedData, err := rsa.SignPKCS1v15(rand.Reader, private, crypto.SHA1, hashed)

	data := base64.StdEncoding.EncodeToString(signedData)
	//fmt.Println("GetSign: ", data)
	return data
}

func (ch *loroPay) PublicEncrypt(data string, publicKey *rsa.PublicKey) (string, error) {
	partLen := publicKey.N.BitLen()/8 - 11
	chunks := split([]byte(data), partLen)
	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		bytes, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, chunk)
		if err != nil {
			return "", err
		}
		buffer.Write(bytes)
	}
	sign := base64.RawURLEncoding.EncodeToString(buffer.Bytes())
	//fmt.Println("sign: -", sign)
	return sign, nil
}

func (ch *loroPay) RsaVerify(plainText string, pubKey string, signText string) error {
	var err error
	pubKey = GetRsaPublicKey(pubKey)
	block, _ := pem.Decode([]byte(pubKey))
	if block == nil {
		logrus.Error("公钥解析失败 : ")
		return err
	}

	pubKeyb, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logrus.Error("RsaVerify ParsePKIXPublicKey : ", err.Error())
		return err
	}
	publicKey, _ := pubKeyb.(*rsa.PublicKey)
	/*signText = strings.ReplaceAll(signText, "-", "+")
	signText = strings.ReplaceAll(signText, "_", "/")
	sign, err := base64.RawStdEncoding.DecodeString(signText)*/
	sign, err := base64.RawURLEncoding.DecodeString(signText)
	if err != nil {
		logrus.Error("RsaVerify DecodeString1: ", err.Error())
		return err
	}
	decrypted, err := utils.PublicDecrypt(publicKey, sign)
	if err != nil {
		logrus.Error("utils.PublicDecrypt: ", err.Error())
		return err
	}
	if string(decrypted) != plainText {
		err = errors.New(plainText + "-验签失败-" + string(decrypted))
		logrus.Error("utils.PublicDecrypt: ", err.Error())
		return err
	}
	//hash := crypto.SHA1.New()
	//hash.Write([]byte(plainText))
	//hashed := hash.Sum(nil)
	//err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hashed, sign)
	//if err != nil {
	//	logrus.Info("VerifyPKCS1v15验证失败 : " + err.Error())
	//	return err
	//}
	return nil
}

func (ch *loroPay) VerifySign(plainText string, pubKey string, signText string) error {
	var err error

	sign, err := base64.RawURLEncoding.DecodeString(signText)
	if err != nil {
		logrus.Error("RsaVerify DecodeString1: ", err.Error())
		return err
	}

	pubKey = GetRsaPublicKey(pubKey)
	block, _ := pem.Decode([]byte(pubKey))
	if block == nil {
		logrus.Error("公钥解析失败 : ")
		return err
	}
	pubKeyb, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logrus.Error("RsaVerify ParsePKIXPublicKey : ", err.Error())
		return err
	}
	publicKey, ok := pubKeyb.(*rsa.PublicKey)
	if !ok {
		errors.New("Parse PublicKey Fail")
		logrus.Error(err.Error())
		return err
	}
	reader := bytes.NewReader(sign)
	var writer bytes.Buffer
	ciphertextBytesChunk := make([]byte, 1024/8)
	for {
		n, _ := io.ReadFull(reader, ciphertextBytesChunk)
		if n == 0 {
			break
		}
		decryptChunk(ciphertextBytesChunk, &writer, publicKey)
	}
	// Concatenate decrypted signature chunks
	decrypted := writer.String()
	/*signText = strings.ReplaceAll(signText, "-", "+")
	signText = strings.ReplaceAll(signText, "_", "/")
	sign, err := base64.RawStdEncoding.DecodeString(signText)*/

	//decrypted, err := utils.PublicDecrypt(publicKey, sign)
	//dump.Printf(string(decrypted))
	//if err != nil {
	//	logrus.Error("utils.PublicDecrypt: ", err.Error())
	//	return err
	//}
	if decrypted != plainText {
		err = errors.New(plainText + "-验签失败-" + string(decrypted))
		logrus.Error("utils.PublicDecrypt: ", err.Error())
		return err
	}
	/*hash := sha1.New()
	hash.Write([]byte(plainText))
	hashed := hash.Sum(nil)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hashed, sign)
	if err != nil {
		logrus.Info("VerifyPKCS1v15验证失败 : " + err.Error())
		return err
	}*/
	return nil
}

func decryptChunk(ciphertextBytesChunk []byte, writer *bytes.Buffer, pubKey *rsa.PublicKey) {
	// Decrypt each signature chunk
	ciphertextInt := new(big.Int)
	ciphertextInt.SetBytes(ciphertextBytesChunk)
	decryptedPaddedInt := decrypt(new(big.Int), pubKey, ciphertextInt)
	// Remove padding
	decryptedPaddedBytes := make([]byte, pubKey.Size())
	decryptedPaddedInt.FillBytes(decryptedPaddedBytes)
	start := bytes.Index(decryptedPaddedBytes[1:], []byte{0}) + 1 // // 0001FF...FF00<data>: Find index after 2nd 0x00
	decryptedBytes := decryptedPaddedBytes[start:]
	// Write decrypted signature chunk
	writer.Write(bytes.Trim(decryptedBytes, "\u0000"))
}

func decrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	// Textbook RSA
	e := big.NewInt(int64(pub.E))
	c.Exp(m, e, pub.N)
	return c
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

func httpPost(reqUrl string, requestBody []byte) ([]byte, error, int) {
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
	logrus.Info("请求地址: ", reqUrl, " 请求参数: ", string(requestBody))
	// 获取 request请求
	request, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("GetHttpSkip Request Error:", err)
		return nil, err, 400
	}
	// 加入 token
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err = client.Do(request.WithContext(context.TODO()))

	if err != nil {
		logrus.Info(string(requestBody), "StatusCode:", resp.StatusCode, "GetHttpSkip Response Error:", err)
		log.Println("GetHttpSkip Response Error:", err)
		return nil, err, 400
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	defer client.CloseIdleConnections()
	//fmt.Println("Response: ", string(responseBody))
	logrus.Info("请求参数: ", string(requestBody), "StatusCode:", resp.StatusCode, "resp:", string(responseBody))

	return responseBody, nil, resp.StatusCode
}

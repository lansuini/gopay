package utils

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"github.com/syyongx/php2go"
	"github.com/tealeg/xlsx"
	"github.com/thinkeridea/go-extend/exnet"
	"io/ioutil"
	"log"
	"luckypay/model"
	"luckypay/pkg/config"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var localNetworks []*net.IPNet

type Onestruct struct {
	Key   string
	Value string
}

type Child struct {
	key   string
	value string
}

type Params []Onestruct

// Len()方法和Swap()方法不用变化
// 获取此 slice 的长度
func (p Params) Len() int { return len(p) }

// 交换数据
func (p Params) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p Params) Less(i, j int) bool {
	return p[i].Key < p[j].Key //根据key 排序
}

func CheckSign(signKey string, param model.ReqPayOrder) bool {
	if param.Sign == "" {
		return false
	}
	//fmt.Println(signKey)
	newParam := param
	originSign := newParam.Sign
	newParam.Sign = ""
	//fmt.Println("%+v", newParam)
	sign := GetSign(signKey, newParam)

	if originSign != sign {
		return false
	}
	return true
}

func CheckSettleSign(signKey string, param model.ReqSettlement) bool {
	if param.Sign == "" {
		return false
	}
	//fmt.Println(signKey)
	newParam := param
	originSign := newParam.Sign
	newParam.Sign = ""
	//newParam.OrderAmount = strconv.FormatFloat(param.OrderAmount, 'f', 2, 64)
	//newParam.OrderAmount = strconv.FormatFloat(param.OrderAmount, 'f', 2, 32)
	//fmt.Println("%+v", newParam)
	newMap := make(map[string]interface{})
	j, _ := json.Marshal(newParam)
	json.Unmarshal(j, &newMap)
	sign := GetSignStr(signKey, newMap)

	if originSign != sign {
		return false
	}
	return true
}

func GetSignStr(signKey string, param map[string]interface{}) string {
	sortParam := Params{}
	signKey = AesCBCDecrypt(signKey)
	//fmt.Println("signKey-", signKey)
	for key, value := range param {
		if value == "" {
			continue
		}
		strConv2 := fmt.Sprintf("%v", value)
		sortParam = append(sortParam, Onestruct{key, strConv2})
		//u.Add(strings.ToLower(person.Key), person.Value)
	}
	sort.Sort(sortParam)
	//fmt.Println(sortParam)
	var builder strings.Builder
	for _, one := range sortParam {
		if one.Value == "" {
			continue
		}
		if FirstLower(one.Key) == "sign" {
			continue
		}
		builder.WriteString(FirstLower(one.Key) + "=")
		builder.WriteString(one.Value + "&")
	}
	signStr := builder.String()
	signStr = strings.TrimRight(signStr, "&")
	signStr = signStr + signKey

	md5str := Md5(signStr)
	//fmt.Println("md5str-", md5str)
	return md5str
}

func GetSign(signKey string, param model.ReqPayOrder) string {
	sortParam := Params{}
	signKey = AesCBCDecrypt(signKey)
	fmt.Println("key:" + signKey)
	fmt.Println("%+v:", param)

	m := make(map[string]interface{})
	j, _ := json.Marshal(param)
	json.Unmarshal(j, &m)
	vaules := reflect.ValueOf(param)
	keys := reflect.TypeOf(param)

	count := vaules.NumField()
	for i := 0; i < count; i++ {
		f := vaules.Field(i)
		key := keys.Field(i)
		switch f.Kind() {
		case reflect.String:
			sortParam = append(sortParam, Onestruct{key.Name, f.String()})
		case reflect.Int:
			sortParam = append(sortParam, Onestruct{key.Name, string(f.Int())})
		case reflect.Float64:
			orderAmount := strconv.FormatFloat(f.Float(), 'f', 2, 32)
			sortParam = append(sortParam, Onestruct{key.Name, orderAmount})
		case reflect.Float32:
			orderAmount := strconv.FormatFloat(float64(f.Float()), 'f', 2, 32)
			sortParam = append(sortParam, Onestruct{key.Name, orderAmount})
		}

	}
	//for _,val := range m {
	//	append(sortParam, val)
	//}
	sort.Sort(sortParam)
	var builder strings.Builder
	for _, one := range sortParam {
		if one.Value == "" {
			continue
		}
		builder.WriteString(FirstLower(one.Key) + "=")
		builder.WriteString(one.Value + "&")
		//u.Add(strings.ToLower(person.Key), person.Value)
	}
	signStr := builder.String()
	signStr = strings.TrimRight(signStr, "&")
	signStr = signStr + signKey
	fmt.Println(signStr)
	//imStr := php2go.Implode("&", sortParam)
	//signStr := imStr + signKey
	md5str := Md5(signStr)
	fmt.Println(md5str)
	//data := []byte(signStr)
	//has := md5.Sum(data)
	//md5str := fmt.Sprintf("%x", has)
	//fmt.Println(md5str)
	return md5str
}

func AesCBCEncrypt(str string) string {
	src := []byte(str)
	envSalt := config.Instance.DataSalt
	envIV := config.Instance.DataSaltIV
	key := []byte(envSalt)
	iv := []byte(envIV)

	dst, _ := openssl.AesCBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
	//fmt.Println(base64.StdEncoding.EncodeToString(dst)) // 1jdzWuniG6UMtoa3T6uNLA==
	return base64.StdEncoding.EncodeToString(dst)

}

func AesCBCDecrypt(str string) string {
	src, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return str
	}
	envSalt := config.Instance.DataSalt
	envIV := config.Instance.DataSaltIV
	key := []byte(envSalt)
	iv := []byte(envIV)

	dst, err := openssl.AesCBCDecrypt(src, key, iv, openssl.PKCS7_PADDING)
	if err != nil {
		logrus.Error("AesCBCDecrypt :", err)
		return str
	}
	return string(dst)
}

// 生成 MD5
func Md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func FirstLower(str string) string {
	if str == "" {
		return ""
	}
	return strings.ToLower(str[:1]) + str[1:]
}

func GetRealIp(r *http.Request) string {
	ip := ClientPublicIP(r)
	if ip == "" {
		ip = ClientIP(r)
	}
	return ip
}

// ClientIP 尽最大努力实现获取客户端 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// ClientPublicIP 尽最大努力实现获取客户端公网 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func ClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" && !HasLocalIPddr(ip) {
			return ip
		}
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" && !HasLocalIPddr(ip) {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		if !HasLocalIPddr(ip) {
			return ip
		}
	}

	return ""
}

// HasLocalIPddr 检测 IP 地址字符串是否是内网地址
func HasLocalIPddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

// HasLocalIP 检测 IP 地址是否是内网地址
func HasLocalIP(ip net.IP) bool {
	localNetworks = make([]*net.IPNet, 19)
	for i, sNetwork := range []string{
		"10.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"172.17.0.0/12",
		"172.18.0.0/12",
		"172.19.0.0/12",
		"172.20.0.0/12",
		"172.21.0.0/12",
		"172.22.0.0/12",
		"172.23.0.0/12",
		"172.24.0.0/12",
		"172.25.0.0/12",
		"172.26.0.0/12",
		"172.27.0.0/12",
		"172.28.0.0/12",
		"172.29.0.0/12",
		"172.30.0.0/12",
		"172.31.0.0/12",
		"192.168.0.0/16",
	} {
		_, network, _ := net.ParseCIDR(sNetwork)
		localNetworks[i] = network
	}
	for _, network := range localNetworks {
		if network.Contains(ip) {
			return true
		}
	}

	return ip.IsLoopback()
}

// RemoteIp 返回远程客户端的 IP，如 192.168.1.1
func RemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := exnet.ClientPublicIP(req); ip != "" {
		remoteAddr = ip
	} else if ip := exnet.ClientIP(req); ip != "" {
		remoteAddr = ip
	} else if ip := req.Header.Get("X-Real-IP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("X-Forwarded-For"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

func IsIpWhite(reqIp string, ips string) bool {

	arrayIpWhite := php2go.Explode(",", ips)
	return php2go.InArray(reqIp, arrayIpWhite)
}

func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandomOrderTail(n int) string {
	var letterRunes = []rune("0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetTimeTick64() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetFormatTime(time time.Time) string {
	return time.Format("20060102")
}

func GetYMDHISTime(time time.Time) string {
	return time.Format("2006010215405")
}

//func HttpPostJson(reqUrl string, body []byte) ([]byte, error) {
//	var client *http.Client
//	var request *http.Request
//	var resp *http.Response
//	//`这里请注意，使用 InsecureSkipVerify: true 来跳过证书验证`
//	client = &http.Client{Transport: &http.Transport{
//		TLSClientConfig: &tls.Config{
//			InsecureSkipVerify: true,
//		},
//	}}
//	log.Println("request_data:", string(body))
//	// 获取 request请求
//	request, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(body))
//	if err != nil {
//		log.Println("GetHttpSkip Request Error:", err)
//		return nil, nil
//	}
//	// 加入 token
//	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
//	request.Header.Add("Authorization", "123")
//	resp, err = client.Do(request.WithContext(context.TODO()))
//	if err != nil {
//		log.Println("GetHttpSkip Response Error:", err)
//		return nil, nil
//	}
//	defer resp.Body.Close()
//	body, err = ioutil.ReadAll(resp.Body)
//	defer client.CloseIdleConnections()
//	fmt.Println("Response: ", string(body))
//	return body, nil
//}

func Format(s string, v interface{}) string {
	t, b := new(template.Template), new(strings.Builder)
	template.Must(t.Parse(s)).Execute(b, v)
	return b.String()
}

func ExportExcel(ctx iris.Context, filename string, titles []string, data [][]string) {

	file := xlsx.NewFile()
	sheet, _ := file.AddSheet("订单信息")

	row := sheet.AddRow()

	var cell *xlsx.Cell
	for _, title := range titles {
		cell = row.AddCell()
		cell.Value = title
	}

	for _, rows := range data {

		row = sheet.AddRow()
		for _, value := range rows {
			cell = row.AddCell()
			cell.Value = value
		}
	}

	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Transfer-Encoding", "binary")

	//回写到web 流媒体 形成下载
	_ = file.Write(ctx.ResponseWriter())
}

func JsonToMap(jsonStr string) (map[string]string, error) {
	m := make(map[string]string)
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		fmt.Printf("Unmarshal with error: %+v\n", err)
		return nil, err
	}

	//for k, v := range m {
	//	fmt.Printf("%v: %v\n", k, v)
	//}

	return m, nil
}

func HttpPostJson(reqUrl string, requestBody []byte) ([]byte, error, int) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	//`这里请注意，使用 InsecureSkipVerify: true 来跳过证书验证`
	client = &http.Client{
		Timeout: time.Second * 8,
		Transport: &http.Transport{
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
		logrus.Info(string(requestBody), "StatusCode:", "GetHttpSkip Response Error:", err)
		return nil, err, 400
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	client.CloseIdleConnections()
	//fmt.Println("Response: ", string(responseBody))
	logrus.Info("请求参数: ", string(requestBody), "StatusCode:", resp.StatusCode, "resp:", string(responseBody))

	return responseBody, nil, resp.StatusCode
}

func IsValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func IsInSlice(needle string, slice []string) error {
	err := errors.New(needle + "not in slice")
	for _, single := range slice {
		if needle == single {
			return nil
		}
	}
	return err
}

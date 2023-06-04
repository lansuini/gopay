package gate

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"luckypay/cache"
	"luckypay/channels"
	"luckypay/model"
	"luckypay/response"
	"luckypay/utils"
	"net/http"
	"net/http/httputil"
	"time"

	// "github.com/dchest/captcha"
	"luckypay/services"
	// "github.com/mlogclub/simple/web"
	// "luckypay/pkg/errs"
	// "luckypay/services"
)

type PayController struct {
	Ctx         iris.Context
	validate    *validator.Validate
	RedisClient *redis.Client
}

func (c *PayController) BeginRequest(ctx iris.Context) {
	rawReq, err := httputil.DumpRequest(c.Ctx.Request(), true)
	logrus.Info(string(rawReq))
	if err != nil {
		logrus.Error(err)
		return
	}
}

func (c *PayController) EndRequest(ctx iris.Context) {}

func (c *PayController) PostPay(httpWriter http.ResponseWriter, r *http.Request) {

	var p model.ReqPayOrder
	//实例化验证器
	//validate := validator.New()
	valid_err := utils.ValidateParam(c.Ctx, &p, "post")
	if valid_err != nil {
		return
	}

	merchantData, res := cache.MerchantCache.GetCacheByMerchantNo(p.MerchantNo)
	if !res {
		logrus.Error("商户不存在")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2001", "msg": "商户不存在"})
		return
	} else if merchantData.OpenPay != true {
		logrus.Error("商户未开通支付")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2002", "msg": "商户未开通支付"})
		return
	}

	paramMap := make(map[string]interface{})
	j, _ := json.Marshal(p)
	json.Unmarshal(j, &paramMap)
	sign := utils.GetSignStr(merchantData.SignKey, paramMap)
	//dump.Printf(sign)
	if sign != p.Sign {
		logrus.Error("验签失败 : ", p.MerchantOrderNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2100", "msg": "验签失败"})
		return
	}

	//if !utils.CheckSign(merchantData.SignKey, p) {
	//	logrus.Error(merchantData.MerchantNo, "-验签失败")
	//	c.Ctx.StopWithJSON(200, iris.Map{"code": "E1005", "msg": "验签失败！"})
	//}

	UserIP := utils.GetRealIp(r)
	//fmt.Println(merchantData.IpWhite)
	if !utils.IsIpWhite(UserIP, merchantData.IpWhite) {
		logrus.Error(merchantData.MerchantNo, "-交易 IP 异常-", UserIP)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2104", "msg": "交易IP异常：" + UserIP})
		return
	}

	_, res = services.PayService.GetCacheByMerchantOrderNo(p.MerchantNo, p.MerchantOrderNo)
	if res {
		logrus.Error("E2100订单已存在 : ", p.MerchantOrderNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2100", "msg": "订单已存在"})
		return
	}

	merchantRate, res := cache.MerchantRate.GetCacheByMerchantNo(p.MerchantNo)
	if !res {
		logrus.Error(p.MerchantNo, "E2026费率未设置", p)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2026", "msg": "商户未设置费率"})
		return
	}

	merchantChannel, res := cache.MerchantChannel.GetCacheByMerchantNo(p.MerchantNo)
	if !res {
		logrus.Error("商户未开通支付方式1")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2006", "msg": "商户未开通支付方式1"})
		return
	}
	payChannel, res := services.MerchantChannel.FetchConfig(p.MerchantNo, merchantChannel, p.PayType, p.OrderAmount, p.BankCode, p.CardType)
	if !res {
		logrus.Error("merchantChannel fetchConfig失败", merchantChannel)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2003", "msg": "商户未开通支付方式2"})
		return
	}
	channelRate, res := cache.ChannelMerchantRate.GetCacheByMerchantNo(payChannel.ChannelMerchantNo)
	if !res {
		logrus.Error("merchantChannel 没有设置费率", payChannel.ChannelMerchantNo, p)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2003", "msg": "商户未开通支付方式3"})
		return
	}

	merchantServiceCharge, res := services.MerchantRate.GetServiceCharge(merchantRate, p, "Pay")
	if !res {
		logrus.Error("GetServiceCharge 没有设置对应支付渠道费率", payChannel.ChannelMerchantNo, p)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2003", "msg": "没有设置对应支付渠道费率"})
		return
	}
	channelServiceCharge, res := services.ChannelMerchantRate.GetServiceCharge(channelRate, p, "Pay")
	if !res {
		logrus.Error("getChannelServiceCharge 没有设置对应上游支付渠道费率: ", payChannel.ChannelMerchantNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2003", "msg": "没有设置对应上游支付渠道费率"})
		return
	}

	//params := map[string]string{}
	//params["channel"] = payChannel.Channel
	//params["channelMerchantId"] = strconv.Itoa(payChannel.ChannelMerchantID)
	//params["channelMerchantNo"] = payChannel.ChannelMerchantNo
	//params["orderAmount"] = p.OrderAmount
	payOrderNo := services.PayService.GetPlatformOrderNo("P")
	//params["platformOrderNo"] = payOrderNo
	payParams := model.PayParams{
		//Channel:           payChannel.Channel,
		ChannelMerchantNo: payChannel.ChannelMerchantNo,
		PlatformOrderNo:   payOrderNo,
		OrderAmoumt:       p.OrderAmount,
		PayType:           p.PayType,
	}

	if _, ok := channels.Channels[payChannel.Channel]; !ok {
		response.Fail(c.Ctx, response.Error, "create failed", nil)
		return
	}
	cacheChannelMerchantData, res := cache.ChannelMerchant.GetCacheByChannelMerchantNo(payChannel.ChannelMerchantNo)
	if !res {
		logrus.Error(p.MerchantOrderNo, "-支付渠道获取失败 ：", payChannel.Channel)
		c.Ctx.JSON(iris.Map{"code": "E2301", "msg": "渠道配置错误"})
		return
	}
	Channel := channels.Channels[payChannel.Channel]
	payResult, err := channels.PayOrder(Channel, payParams, cacheChannelMerchantData)
	logrus.Info("请求结果：", payResult)

	if err != nil {
		logrus.Error(p.MerchantOrderNo, "-请求订单异常 ：", err.Error())
		c.Ctx.JSON(iris.Map{"code": "E2200", "msg": "订单请求失败"})
		return
	}
	if payResult.Status != "Success" {
		logrus.Error(p.MerchantOrderNo, "-请求订单失败 ：", payResult)
		go services.PayNotify.Push(0, payOrderNo)
		c.Ctx.JSON(iris.Map{"code": "E2201", "msg": "订单请求失败"})
		return
	}

	_, err = services.PayService.CreateOrder(p, payOrderNo, payChannel, payResult.ChannelOrderNo, merchantServiceCharge, channelServiceCharge)
	if err != nil {
		logrus.Error(payOrderNo, "-创建订单失败 : ", err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E9001", "msg": "系统忙，请稍后再试"})
		return
	}
	success := make(map[string]interface{})
	success["platformOrderNo"] = payOrderNo
	success["payUrl"] = payResult.PayUrl

	encrytkey := merchantData.SignKey
	backsign := utils.GetSignStr(encrytkey, success)
	c.Ctx.StopWithJSON(200, iris.Map{"code": "SUCCESS", "msg": "成功", "biz": success, "sign": backsign})

	return

}

func (c *PayController) PostQueryPay(httpWriter http.ResponseWriter, r *http.Request) {

	var p model.QueryPayOrder
	valid_err := utils.ValidateParam(c.Ctx, &p, "post")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		return
	}

	merchantData, res := cache.MerchantCache.GetCacheByMerchantNo(p.MerchantNo)
	//if merchantData == (model.Merchant{}) {
	if !res {
		logrus.Error("商户不存在")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2001", "msg": "商户不存在"})
		return
	} else if merchantData.OpenPay != true {
		logrus.Error("商户未开通支付")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2002", "msg": "商户未开通支付"})
		return
	}

	UserIP := utils.GetRealIp(r)

	if !utils.IsIpWhite(UserIP, merchantData.IpWhite) {
		logrus.Error("交易 IP 异常")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2104", "msg": "交易 IP 异常"})
		return
	}

	payOrder, res := cache.PayCache.GetCacheByMerchantOrderNo(p.MerchantNo, p.MerchantOrderNo)
	if !res {
		logrus.Error("E2300订单不存在 : ", p.MerchantOrderNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2300", "msg": "订单不存在"})
		return
	}
	biz := make(map[string]interface{})
	biz["merchantNo"] = payOrder.MerchantNo
	biz["merchantOrderNo"] = payOrder.MerchantOrderNo
	biz["platformOrderNo"] = payOrder.PlatformOrderNo
	biz["orderStatus"] = payOrder.OrderStatus

	sign := utils.GetSignStr(merchantData.SignKey, biz)
	c.Ctx.StopWithJSON(200, iris.Map{"code": "SUCCESS", "msg": "成功", "biz": biz, "sign": sign})
}

func (c *PayController) PostSettlement(httpWriter http.ResponseWriter, r *http.Request) {

	var p model.ReqSettlement

	valid_err := utils.Validate(c.Ctx, &p, "post")
	logrus.Info(p)
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E1001", "msg": valid_err.Error()})
		return
	}

	merchantData, res := cache.MerchantCache.GetCacheByMerchantNo(p.MerchantNo)
	if !res {
		logrus.Error("商户不存在")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2001", "msg": "商户不存在"})
		return
	} else if merchantData.OpenSettlement != true {
		logrus.Error("商户未开通代付")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E3002", "msg": "商户未开通代付"})
		return
	}

	UserIP := utils.GetRealIp(r)
	if !utils.IsIpWhite(UserIP, merchantData.IpWhite) {
		logrus.Error(merchantData.MerchantNo, "-交易 IP 异常", UserIP)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2104", "msg": "交易IP异常：" + UserIP})
		return
	}
	paramMap := make(map[string]interface{})
	j, _ := json.Marshal(p)
	json.Unmarshal(j, &paramMap)
	sign := utils.GetSignStr(merchantData.SignKey, paramMap)
	if sign != p.Sign {
		logrus.Error("代付验签失败 : ", p.MerchantOrderNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2100", "msg": "代付验签失败"})
		return
	}

	_, res = cache.SettleCache.GetCacheByMerchantOrderNo(p.MerchantNo, p.MerchantOrderNo)
	if res {
		logrus.Error("代付请求：商户订单号重复 : ", p.MerchantOrderNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2100", "msg": "订单已存在"})
		return
	}

	merchantRate, res := cache.MerchantRate.GetCacheByMerchantNo(p.MerchantNo)
	if !res {
		logrus.Error(p.MerchantNo, "E2026费率未设置", p)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2026", "msg": "商户未设置费率"})
		return
	}
	amountData := services.MerchantAmount.GetAmount(p.MerchantNo)
	merchantServiceCharge, res := services.MerchantRate.GetServiceChargeSettle(merchantRate, p, "Settlement")
	if !res {
		logrus.Error("GetServiceChargeSettle 没有设置对应上游代付渠道费率: ", merchantRate)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2003", "msg": "商户未设置代付费率"})
		return
	}
	if p.OrderAmount+merchantServiceCharge > amountData["availableBalance"] {
		logrus.Error(p.MerchantOrderNo, "代付请求：代付余额不足")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2105", "msg": "代付余额不足"})
		return
	}

	merchantChannel, res := cache.MerchantChannelSettlement.GetCacheByMerchantNo(p.MerchantNo)
	if !res {
		logrus.Error("代付请求：未配置商户代付通道")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2006", "msg": "未配置商户代付通道"})
		return
	}

	//if ($code == 'SUCCESS') {
	//	$blackUserSettlement = new BlackUserSettlement();
	//	$isblackUserExists = $blackUserSettlement->checkBlackUser($request->getParam('bankCode'),$request->getParam('bankAccountName'),$request->getParam('bankAccountNo'));
	//	if($isblackUserExists){
	//		$code = 'E2201';
	//		$logger->error("代付请求：代付黑名单用户！");
	//	}
	//}
	//fmt.Print(merchantChannel)

	payChannel, res := services.MerchantChannelSettlement.FetchConfig(p.MerchantNo, merchantChannel, "D0Settlement", p.OrderAmount, p.BankCode)
	if !res {
		logrus.Error("merchantChannel fetchConfig失败", merchantChannel)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2003", "msg": "商户未开通支付方式2"})
		return
	}
	channelRate, res := cache.ChannelMerchantRate.GetCacheByMerchantNo(payChannel.ChannelMerchantNo)
	if !res {
		logrus.Error("merchantChannelSettlement 商户费率未设置", payChannel.ChannelMerchantNo, p)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2003", "msg": "商户费率未设置"})
		return
	}

	channelServiceCharge, res := services.ChannelMerchantRate.GetServiceChargeSettle(channelRate, p, "Settlement")
	if !res {
		logrus.Error("GetServiceChargeSettle 没有设置对应上游代付渠道费率: ", payChannel.ChannelMerchantNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2003", "msg": "商户渠道费率未设置"})
		return
	}

	platformOrderNo := services.SettleService.GetPlatformOrderNo("S")
	settleParams := model.SettleParams{
		//Channel:           payChannel.Channel,
		ChannelMerchantNo: payChannel.ChannelMerchantNo,
		PlatformOrderNo:   platformOrderNo,
		OrderAmoumt:       p.OrderAmount,
		BankCode:          p.BankCode,
		BankAccountName:   p.BankAccountName,
		BankAccountNo:     p.BankAccountNo,
		City:              p.City,
	}
	if _, ok := channels.Channels[payChannel.Channel]; !ok {
		logrus.Error(p.MerchantOrderNo, "-渠道配置错误 ：", payChannel.Channel)
		c.Ctx.JSON(iris.Map{"code": "E2300", "msg": "渠道配置错误"})
		return
	}
	cacheChannelMerchantData, res := cache.ChannelMerchant.GetCacheByChannelMerchantNo(payChannel.ChannelMerchantNo)
	if !res {
		logrus.Error(p.MerchantOrderNo, "-代付渠道获取失败 ：", payChannel.Channel)
		c.Ctx.JSON(iris.Map{"code": "E2301", "msg": "渠道配置错误"})
		return
	}
	/*sdjads := "{\"description\":\"loropay\",\"merchantNo\":\"TEST081313058620\",\"platformPublicKey\":\"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCwIxdd2KkbY2VeEYTF7PU9BdeixwuZxdD86P\\/Pb\\/HpPnDXQrqUkRU8itoXfjFPeME9y4rbpNgHGkd15S\\/7fUiWyKBUnkQK5wPp40azY8xTGf5SzRQRUwGDQ+w3IBITrN0tzJ4Lc\\/rNCTCd2pGf1+a3zEOXLLe6NJaFvTDqDbM2NQIDAQAB\",\"merchantPrivateKey\":\"MIICXAIBAAKBgQCLJtfTZ\\/fH1SfnXjc24\\/axLB2wPEZXtHKdInZpOeRfLp\\/fvFbvm0cd3Say1T5xrmXHFx2qofRx8rpVmXcYrr6hw+aCG7jV7hC41uQXeNpvh+hcRTAQObhyceRQw3PEaLZp2jPs7+Xbvu8h83gG5smEHnlb2jqEdV5u\\/N5p2ZbzbQIDAQABAoGAUegbQiUAhG\\/DfTzH41dr7f25u\\/K+tQFSNYwDhwy8kAoxsNB7m64avklebgV3LBMrdXT10WpjKG9nntsmbzDspAxEbbglNOhQxKbHJmXdz9eSCJGEM7jpq1ejUc4KuUL4kPT6FZdcv+ttontgihrGjezJq3f9BjvQTPqgK3tzm0ECQQDOe0k6j\\/WV2Uxcx5ooy4+\\/6QA\\/\\/xSjlyr89rjUwdN52r+2Fy6mCkBs\\/BGxmb\\/aCqFPTPA+YfpuqL0o1MsfhSyvAkEArIXpzFQ3pBj4X0aONxPu0rWQPL6yxB+NUIRpoP6N\\/RL1VTNKHjPLQltzy1nw3Nqlu8xg+tIS8JS4cfI4SsyAowJAZdmMYpW2NydLsoxGr47RpoFBPVAOly8u5j6xJ0lAjl\\/npuNCgGaYJuojtC4540zRCvPRoYPk6wbS37wvQaoIQwJABhdsW+SVWlvvWR3ao6M2iYYTo7FwCnC6wp8KQ775MHhc5Tc8ZLibcqpb+lAgqwulUm4y9mg4dvopUQymZC24VQJBAJk+b6FVR3SI\\/YnDR0Me5FYgGC333Xus3rhohw16TCtDzEdSRrLWcS6AGGYssU7K8ihp59TzH2PGHNVDuacIUR4=\",\"merchantPublicKey\":\"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCLJtfTZ\\/fH1SfnXjc24\\/axLB2wPEZXtHKdInZpOeRfLp\\/fvFbvm0cd3Say1T5xrmXHFx2qofRx8rpVmXcYrr6hw+aCG7jV7hC41uQXeNpvh+hcRTAQObhyceRQw3PEaLZp2jPs7+Xbvu8h83gG5smEHnlb2jqEdV5u\\/N5p2ZbzbQIDAQAB\",\"ipWhite\":\"\"}"
	utils.AesCBCEncrypt(sdjads)*/
	Channel := channels.Channels[payChannel.Channel]
	//var platformSettlementOrder model.PlatformSettlementOrder
	var merchantAmount model.MerchantAmount
	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		var err error

		settleOrder, err := services.SettleService.CreateSettlementOrder(tx, p, platformOrderNo, payChannel, merchantServiceCharge, channelServiceCharge)
		if err != nil {
			logrus.Error(p.MerchantOrderNo, "-创建订单失败 : ", err.Error())
			//c.Ctx.StopWithJSON(200, iris.Map{"code": "E9001", "msg": "系统忙，请稍后再试"})
			return err
		}
		settlementFetchTask := model.SettlementFetchTask{
			Status:          "Execute",
			RetryCount:      0,
			PlatformOrderNo: platformOrderNo,
		}
		err = tx.Create(&settlementFetchTask).Error
		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, p.MerchantNo, p.MerchantOrderNo)
			logrus.Error(p.MerchantOrderNo, "-插入代付查询任务 : ", err.Error())
			return err
		}
		settlementNotifyTask := model.SettlementNotifyTask{
			Status:          "Execute",
			RetryCount:      0,
			PlatformOrderNo: platformOrderNo,
		}
		err = tx.Create(&settlementNotifyTask).Error
		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, p.MerchantNo, p.MerchantOrderNo)
			logrus.Error(p.MerchantOrderNo, "-插入代付回调任务 : ", err.Error())
			return err
		}
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", p.MerchantNo).First(&merchantAmount).Error

		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, p.MerchantNo, p.MerchantOrderNo)
			logrus.Error(p.MerchantOrderNo, "-查询商户余额失败 : ", err.Error())
			return err
		}
		balance := merchantAmount.SettlementAmount - p.OrderAmount - merchantServiceCharge

		err = tx.Model(&merchantAmount).Update("settlementAmount", balance).Error
		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, p.MerchantNo, p.MerchantOrderNo)
			logrus.Error(p.MerchantOrderNo, "-商户余额更新失败 : ", err.Error())
			log.Println(err)
			return err
		}
		now := time.Now()
		//accountDate := utils.GetFormatTime(now)
		//sqls.DB().Save(&merchantAmount)
		settleFinance := model.Finance{
			MerchantID:      merchantData.MerchantID,
			MerchantNo:      merchantData.MerchantNo,
			PlatformOrderNo: platformOrderNo,
			Amount:          p.OrderAmount,
			Balance:         merchantAmount.SettlementAmount - p.OrderAmount,
			FinanceType:     "PayOut",
			AccountDate:     now.Format("2006-01-02"),
			AccountType:     "SettlementAccount",
			SourceID:        settleOrder.OrderID,
			SourceDesc:      "代付",
			MerchantOrderNo: p.MerchantOrderNo,
			OperateSource:   "ports",
			Summary:         p.TradeSummary,
		}
		feeFinance := model.Finance{
			MerchantID:      merchantData.MerchantID,
			MerchantNo:      merchantData.MerchantNo,
			PlatformOrderNo: platformOrderNo,
			Amount:          merchantServiceCharge,
			Balance:         balance,
			FinanceType:     "PayOut",
			AccountDate:     now.Format("2006-01-02"),
			AccountType:     "SettlementAccount",
			SourceID:        settleOrder.OrderID,
			SourceDesc:      "代付手续费",
			MerchantOrderNo: p.MerchantOrderNo,
			OperateSource:   "ports",
			Summary:         p.TradeSummary,
		}
		var finances = []model.Finance{settleFinance, feeFinance}

		err = tx.Create(finances).Error
		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, p.MerchantNo, p.MerchantOrderNo)
			logrus.Error(p.MerchantOrderNo, "-插入金流日志失败 : ", err.Error())
			return err
		}
		//	now := time.Now()
		accountDate := utils.GetFormatTime(now)
		err = tx.Table("amount_pay").Where("merchantNo = ?", p.MerchantNo).Where("accountDate = ?", accountDate).Update("balance", balance).Error
		if err != nil {
			services.SettleService.DelCacheByPlatformOrderNo(platformOrderNo, p.MerchantNo, p.MerchantOrderNo)
			logrus.Error(p.MerchantOrderNo, "-更新amount_pay失败: ", err.Error())
			log.Println(err)
			return err
		}
		services.SettleService.SetCacheByPlatformOrderNo(platformOrderNo, settleOrder)
		return nil
	})

	if err != nil {
		logrus.Error("创建代付订单失败")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E9002", "msg": "系统忙，请稍后再试"})
		return
	}
	//return
	go services.SettleService.Settlement(Channel, settleParams, cacheChannelMerchantData)
	//settleResult, err := channels.SettlementOrder(Channel, settleParams, cacheChannelMerchantData)
	//
	//if err != nil {
	//	logrus.Error(p.MerchantOrderNo, "-请求订单异常 ：", err.Error())
	//	c.Ctx.JSON(iris.Map{"code": "E2302", "msg": "订单请求失败"})
	//	return
	//}
	//if settleResult.Status != "Success" {
	//	logrus.Error(p.MerchantOrderNo, "-请求代付订单失败 ：", settleResult)
	//	c.Ctx.JSON(iris.Map{"code": "E2303", "msg": "订单请求失败"})
	//	return
	//}

	success := make(map[string]interface{})
	success["platformOrderNo"] = platformOrderNo
	/*for _, value := range success {
		fmt.Println("type:", reflect.TypeOf(value))

	}*/
	encrytkey := merchantData.SignKey
	backSign := utils.GetSignStr(encrytkey, success)
	c.Ctx.StopWithJSON(200, iris.Map{"code": "SUCCESS", "msg": "成功", "biz": success, "sign": backSign})

	return

}

func (c *PayController) PostQuerySettlement(responseWriter http.ResponseWriter, request *http.Request) {

	var p model.QueryPayOrder
	valid_err := utils.ValidateParam(c.Ctx, &p, "post")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		return
	}

	merchantData, res := cache.MerchantCache.GetCacheByMerchantNo(p.MerchantNo)
	if !res {
		logrus.Error("商户不存在")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2001", "msg": "商户不存在"})
		return
	} else if merchantData.OpenSettlement != true {
		logrus.Error("商户未开通代付")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2002", "msg": "商户未开通代付"})
		return
	}

	UserIP := utils.GetRealIp(request)

	if !utils.IsIpWhite(UserIP, merchantData.IpWhite) {
		logrus.Error(merchantData.MerchantNo, "-交易 IP 异常：", UserIP)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2104", "msg": "交易IP异常: " + UserIP})
		return
	}
	paramMap := make(map[string]interface{})
	j, _ := json.Marshal(p)
	json.Unmarshal(j, &paramMap)
	sign := utils.GetSignStr(merchantData.SignKey, paramMap)
	if sign != p.Sign {
		logrus.Error("查询代付验签失败 : ", p.MerchantOrderNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2100", "msg": "查询代付验签失败"})
		return
	}

	settleOrder, res := cache.SettleCache.GetCacheByMerchantOrderNo(p.MerchantNo, p.MerchantOrderNo)
	if !res {
		logrus.Error("E2300代付订单不存在 : ", p.MerchantOrderNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2300", "msg": "代付订单不存在"})
		return
	}
	biz := make(map[string]interface{})
	biz["merchantNo"] = settleOrder.MerchantNo
	biz["merchantOrderNo"] = settleOrder.MerchantOrderNo
	biz["platformOrderNo"] = settleOrder.PlatformOrderNo
	biz["orderStatus"] = settleOrder.OrderStatus

	backSign := utils.GetSignStr(merchantData.SignKey, biz)
	c.Ctx.StopWithJSON(200, iris.Map{"code": "SUCCESS", "msg": "成功", "biz": biz, "sign": backSign})
}

func (c *PayController) PostQueryBalance(responseWriter http.ResponseWriter, request *http.Request) {

	var p model.QueryBalance
	valid_err := utils.ValidateParam(c.Ctx, &p, "post")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		return
	}

	merchantData, res := cache.MerchantCache.GetCacheByMerchantNo(p.MerchantNo)
	if !res {
		logrus.Error("商户不存在")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2001", "msg": "商户不存在"})
		return
	} else if merchantData.OpenQuery != true {
		logrus.Error("商户未开通查询功能", p.MerchantNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2002", "msg": "商户未开通查询功能"})
		return
	}

	UserIP := utils.GetRealIp(request)

	if !utils.IsIpWhite(UserIP, merchantData.IpWhite) {
		logrus.Error("交易 IP 异常")
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2104", "msg": "交易 IP 异常"})
		return
	}
	paramMap := make(map[string]interface{})
	j, _ := json.Marshal(p)
	json.Unmarshal(j, &paramMap)
	sign := utils.GetSignStr(merchantData.SignKey, paramMap)
	if sign != p.Sign {
		logrus.Error("查询余额验签失败 : ", p.MerchantNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E1005", "msg": "查询余额验签失败"})
		return
	}
	merchantAmount, res := cache.MerchantAmount.GetCacheByMerchantNo(p.MerchantNo)
	if res != true {
		logrus.Error("查询余额失败 : ", p.MerchantNo)
		c.Ctx.StopWithJSON(200, iris.Map{"code": "E2400", "msg": "查询失败"})
		return
	}
	biz := make(map[string]interface{})
	biz["merchantNo"] = p.MerchantNo
	biz["balance"] = merchantAmount.SettlementAmount

	backSign := utils.GetSignStr(merchantData.SignKey, biz)
	c.Ctx.StopWithJSON(200, iris.Map{"code": "SUCCESS", "msg": "成功", "biz": biz, "sign": backSign})
}

/*func (c *PayController) PostTest() {
	var payOrder model.ReqPayOrder
	err := c.Ctx.ReadJSON(&payOrder)
	if err != nil {
		// This check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				c.Ctx.StatusCode(iris.StatusInternalServerError)
				c.Ctx.WriteString(err.Error())
				return
			}

			c.Ctx.StatusCode(iris.StatusBadRequest)
			for _, err := range err.(validator.ValidationErrors) {
				fmt.Println()
				fmt.Println(err.Namespace())
				fmt.Println(err.Field())
				fmt.Println(err.StructNamespace())
				fmt.Println(err.StructField())
				fmt.Println(err.Tag())
				fmt.Println(err.ActualTag())
				fmt.Println(err.Kind())
				fmt.Println(err.Type())
				fmt.Println(err.Value())
				fmt.Println(err.Param())
				fmt.Println()
			}
			c.Ctx.Writef(config.Viper.GetString("D0Settlement"))
			return
		}
	}
	c.Ctx.Writef(config.Viper.GetString("qr"))
}*/

/*func (c *PayController) GetTest() {
	//payChannel    := params.FormValue(c.Ctx, "pay")
	//var passage channels.Passage
	//
	//passage = new(channels.AliPay)
	//passage.PayOrder()

	payChannel := params.FormValueDefault(c.Ctx, "pay", "loroPay")
	if _, ok := channels.Channels[payChannel]; !ok {
		response.Fail(c.Ctx, response.Error, "create failed", nil)
		return
	}
	Channel := channels.Channels[payChannel]
	params := channels.PayParams{
		//Channel:           payChannel,
		ChannelMerchantNo: "dfajosd232131",
		PlatformOrderNo:   "2381sad",
		OrderAmoumt:       200.32,
	}
	channels.PayOrder(Channel, params)

	//passage.PayOrder(params)
	//passage = new(channels.AliPaY)
	//passage.PayOrder()
}*/

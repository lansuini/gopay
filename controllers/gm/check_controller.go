package gm

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"luckypay/config"
	"luckypay/model"
	"luckypay/mytool"
	"luckypay/services"
	"luckypay/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
	//"encoding/csv"
)

type CheckController struct {
	Ctx iris.Context
}

func (c *CheckController) GetSearch() {
	queryParams := model.SearchCheckList{}
	valid_err := utils.Validate(c.Ctx, &queryParams, "get")
	if valid_err != nil {
		logrus.Info(valid_err.Error())
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": valid_err.Error()})
		return
	}
	var builder strings.Builder
	builder.WriteString("1 = 1 ")
	cnd := sqls.DB()
	if queryParams.Relevance != "" {
		builder.WriteString(" and relevance like %'" + string(queryParams.Relevance) + "'%")
		cnd = cnd.Where("relevance like %?%", queryParams.Relevance)
	}
	if queryParams.CheckStatus != "-1" {
		builder.WriteString(" and status = '" + string(queryParams.CheckStatus) + "'")
		cnd = cnd.Where("status = ?", queryParams.CheckStatus)
	}
	if queryParams.CheckType != "" {
		builder.WriteString(" and type = '" + string(queryParams.CheckType) + "'")
		cnd = cnd.Where("type = ?", queryParams.CheckType)
	}

	var logs []model.SystemCheckLog
	var count int64
	err := cnd.Limit(queryParams.Limit).Offset(*queryParams.Offset).Order("id desc").Find(&logs).Offset(-1).Count(&count).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	var result []map[string]interface{}
	err = sqls.DB().Table("system_account").Find(&result).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"success": 0, "result": err.Error()})
		return
	}
	admins := make(map[string]interface{})
	for _, value := range result {
		idstr := fmt.Sprintf("%v", value["id"])
		admins[idstr] = value["userName"]
	}

	rows := []map[string]interface{}{}
	for _, log := range logs {
		row := make(map[string]interface{})
		row["url"] = ""
		if log.Type == "支付补单" {
			row["url"] = "/payorder/makeupcheck"
		}
		if log.Type == "代付补单" {
			row["url"] = "/settlementorder/makeupcheck"
		}
		admin_id := fmt.Sprintf("%v", log.AdminId)
		if _, ok := admins[admin_id]; ok {
			row["admin_id"] = admins[admin_id]
		} else {
			row["admin_id"] = ""
		}
		CommiterId := fmt.Sprintf("%v", log.CommiterId)
		if _, ok := admins[CommiterId]; ok {
			row["commiter_id"] = admins[CommiterId]
		} else {
			row["commiter_id"] = ""
		}
		row["ip"] = log.Ip
		row["ipDesc"] = log.IpDesc
		row["check_time"] = log.CheckTime
		row["check_ip"] = log.CheckIp
		row["status"] = config.CheckStatusCode[log.Status]
		row["type"] = log.Type
		row["id"] = log.Id
		row["content"] = log.Content
		row["content"] = log.Content
		row["desc"] = log.Desc
		row["relevance"] = log.Relevance
		row["created_at"] = log.CreatedAt
		rows = append(rows, row)
	}
	//fmt.Print(admins)
	//arrays := map[string]interface{}{
	//	"支付补单" => "/payorder/makeupcheck",
	//	"代付补单" => "/settlementorder/makeupcheck",
	//}

	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "total": count, "rows": rows})
}

func (c *CheckController) GetBalanceAudit(req *http.Request) {
	rowId, err := c.Ctx.URLParamInt64("id")
	checkPwd := c.Ctx.URLParamTrim("checkPwd")
	auditType := c.Ctx.URLParamTrim("auditType")
	if len(checkPwd) > 10 || len(auditType) > 20 {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": "参数错误"})
		return
	}
	unicheckKey := "balanceAuditCkeck" + LoginAdmin
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
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": err.Error()})
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
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": err.Error()})
		return
	}
	content := make(map[string]interface{})
	err = json.Unmarshal([]byte(systemCheckLog.Content), &content)
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": err.Error()})
		return
	}
	BalanceAdjustment := model.BalanceAdjustment{}
	whereMap := map[string]interface{}{
		"platformOrderNo": content["platformOrderNo"],
		"status":          "Unaudit",
		"amount":          content["amount"],
		"merchantId":      content["merchantId"],
	}
	err = sqls.DB().Where(whereMap).First(&BalanceAdjustment).Error
	if err != nil {
		c.Ctx.StopWithJSON(200, iris.Map{"status": 0, "result": err.Error()})
		return
	}
	var merchantAmount model.MerchantAmount
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", content["merchantNo"]).First(&merchantAmount).Error
		if err != nil {
			//tx.Rollback()
			logrus.Error(content["merchantNo"], "-查询商户余额失败 : ", err.Error())
			return err
		}
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("merchantNo = ?", content["merchantNo"]).First(&merchantAmount).Error
		if err != nil {
			//tx.Rollback()
			logrus.Error(content["merchantNo"], "-查询商户余额失败 : ", err.Error())
			return err
		}
		amount := (content["amount"]).(float64)

		if content["bankrollDirection"] == "Restore" || content["bankrollDirection"] == "Unfreeze" {
			merchantAmount.SettlementAmount = merchantAmount.SettlementAmount + amount
			if content["bankrollDirection"] == "Unfreeze" {
				merchantAmount.FreezeAmount = merchantAmount.FreezeAmount - amount
			}
			accountDate := time.Now().Format("2006-01-02")
			tx.Table("amount_pay").Where("merchantId = ?", content["merchantId"]).Where("accountDate = ?", accountDate).Update("balance", merchantAmount.SettlementAmount)
			finance := model.Finance{}
			sourceDesc := config.BankrollDirection[content["bankrollDirection"].(string)] + "-账户资金"
			finance.MerchantID = int64(content["merchantId"].(float64))
			finance.MerchantNo = content["merchantNo"].(string)
			finance.PlatformOrderNo = content["platformOrderNo"].(string)
			finance.Amount = amount
			finance.Balance = merchantAmount.SettlementAmount
			finance.FinanceType = "PayIn"
			finance.AccountDate = accountDate
			finance.AccountType = "SettlementAccount"
			finance.SourceID = int64(content["sourceId"].(float64))
			finance.SourceDesc = "余额调整-" + sourceDesc
			finance.OperateSource = "admin"
			finance.Summary = content["summary"].(string)
			err = tx.Create(&finance).Error
			if err != nil {
				logrus.Error(content["merchantNo"].(string), "-插入财务记录失败 : ", err.Error())
				return err
			}
		}
		//TODO:充值判断
		//if($content['bankrollDirection'] == 'Recharge'){//充值订单}

		err = tx.Save(&merchantAmount).Error
		if err != nil {
			logrus.Error(content["merchantNo"].(string), "-更改资金失败 : ", err.Error())
			return err
		}
		actionLog := make(map[string]interface{})
		actionLog["action"] = "UPDATE_BALANCE_ADJUSTMENT"
		actionLog["actionBeforeData"] = ""
		actionLog["actionAfterData"] = "UPDATE_BALANCE_ADJUSTMENT"
		actionLog["status"] = "Success"
		actionLog["accountId"] = LoginAdminId
		UserIP := utils.GetRealIp(req)
		actionLog["ip"] = UserIP
		actionLog["ipDesc"] = "" //TODO:ip描述
		err = tx.Table("system_account_action_log").Create(actionLog).Error
		//SystemAccountActionLog::insert([
		//		'action' => 'CREATE_BALANCE_ADJUSTMENT',
		//		'actionBeforeData' => '',
		//		'actionAfterData' => $merchantAmountData->toJson(),
		//		'status' => 'Success',
		//		'accountId' => $_SESSION['accountId'],
		//		'ip' => Tools::getIp(),
		//		'ipDesc' => Tools::getIpDesc(),
		//]);

		BalanceAdjustment.Status = "Success"
		BalanceAdjustment.AuditPerson = LoginAdmin
		BalanceAdjustment.AuditTime = time.Now().Format("2006-01-02 15:04:05")
		err = tx.Save(BalanceAdjustment).Error
		if err != nil {
			logrus.Error("BalanceAdjustment-修改失败 : ", err.Error())
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
	services.MerchantAmount.RefreshCache()
	c.Ctx.StopWithJSON(200, iris.Map{"success": 1, "result": "操作成功"})
}

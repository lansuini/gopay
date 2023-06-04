package services

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"luckypay/cache"
	"luckypay/model"
	"luckypay/utils"
	"math"
	"reflect"
	"sync"
	"time"

	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var mutex sync.Mutex

var SettlementNotify = newSettlementNotify()

func newSettlementNotify() *settlementNotify {
	return &settlementNotify{}
}

type settlementNotify struct {
	RedisClient *redis.Client
}

func (s *settlementNotify) Push(taskId int64, platformOrderNo string) {
	data := ChildTask{}
	if taskId == 0 {
		task := model.SettlementNotifyTask{
			Status:          "Execute",
			PlatformOrderNo: platformOrderNo,
		}

		err := sqls.DB().Where("platformOrderNo = ?", platformOrderNo).FirstOrCreate(&task).Error
		if err != nil {
			logrus.Error("settlementNotify Push FirstOrCreate Error-", platformOrderNo, err)
			return
		}
		data.TaskId = task.ID
	} else {
		data.TaskId = taskId
	}
	data.PlatformOrderNo = platformOrderNo
	jsonStr, err := json.Marshal(data)
	if err != nil {
		logrus.Error("settlementNotify Push Marshal Error-", platformOrderNo, err)
		return
	}
	res, err := redisServer.LPush(ctx, "settlementnotify:queue", jsonStr).Result()
	if err != nil {
		logrus.Error("settlementNotify LPush Error-", platformOrderNo, err)
		return
	}
	logrus.Info("settlementnotify:queue ,lpRes:", res, ",platformOrderNo:", platformOrderNo, ",taskId:", data.TaskId)
	return
}

func (s *settlementNotify) Pop() {
	mutex.Lock()         // 添加互斥锁
	defer mutex.Unlock() // 使用结束时解锁
	data, err := redisServer.RPop(ctx, "settlementnotify:queue").Result()
	if err != nil && err != redis.Nil {
		logrus.Error("settlementnotify RPop ：", err)
		return
	}
	if len(data) == 0 {
		//logrus.Info("settlementNotify RPop Empty data-", data)
		return
	}
	cacheKey := "settlementnotify:queue:lasttime"
	redisServer.SetEX(ctx, cacheKey, time.Now(), 60*time.Second)
	if data == "" {
		logrus.Error("settlementNotify RPop 数据为空")
		return
	}
	task := ChildTask{}
	err = json.Unmarshal([]byte(data), &task)
	if err != nil {
		logrus.Error("settlementNotify Unmarshal Error-", err)
		return
	}
	if reflect.DeepEqual(task, ChildTask{}) {
		logrus.Info("settlementNotify RPop task is empty")
		return
	}

	taskData := model.SettlementNotifyTask{}
	err = sqls.DB().Where("id = ?", task.TaskId).First(&taskData).Error
	if err != nil {
		logrus.Info(task.TaskId, "-settlementNotify RPop 代付查询任务查询异常-", err.Error())
		return
	}

	if reflect.DeepEqual(taskData, model.SettlementNotifyTask{}) {
		logrus.Info("settlementnotify:queue,empty taskData:", taskData)
		return
	}

	if taskData.RetryCount >= 5 {
		logrus.Info("settlementNotify RPop 代付订单回调次数达到上限", taskData)
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": taskData.RetryCount + 1,
		}
		go s.UpdateTask(taskData, updates)
		return
	}

	orderData, res := cache.SettleCache.GetCacheByPlatformOrderNo(task.PlatformOrderNo)
	if !res {
		logrus.Info("settlementNotify RPop 获取代付订单失败", taskData)
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": taskData.RetryCount + 1,
			"failReason": "获取代付订单失败",
		}
		go s.UpdateTask(taskData, updates)

		return
	}

	if orderData.CallbackSuccess {
		logrus.Info("settlementNotify RPop 代付订单已回调成功", taskData)
		return
	}

	if orderData.OrderStatus == "Transfered" || orderData.OrderStatus == "Exception" {

		logrus.Info("settlementNotify RPop 代付订单状态未完成")
		return
	}

	if orderData.BackNoticeURL == "" {
		logrus.Error("settlementNotify RPop", orderData.PlatformOrderNo, "-代付回调失败：回调地址为空", taskData)
	}
	isValid := utils.IsValidUrl(orderData.BackNoticeURL)
	if !isValid {
		logrus.Error("settlementNotify RPop", orderData.PlatformOrderNo, "-代付回调失败：回调地址格式错误-", orderData.BackNoticeURL, taskData)
		return
		//order.BackNoticeURL = "http://cb.luckypay.mm:8082/paycallback"
	}
	go SettleService.CallbackMerchant(task.PlatformOrderNo, orderData, taskData)
	//updates := map[string]interface{}{
	//	"status":     "Success",
	//	"retryCount": taskData.RetryCount + 1,
	//	"failReason": "订单已完成",
	//}
	//go s.UpdateTask(taskData, updates)

	return
}

func (s *settlementNotify) UpdateTask(task model.SettlementNotifyTask, updates map[string]interface{}) {
	if reflect.DeepEqual(task, model.SettlementNotifyTask{}) {
		return
	}

	err := sqls.DB().Model(task).Updates(updates).Error
	if err != nil {
		logrus.Info("settlementNotify RPop 更新taskData2失败", err)
	}
}

func (s *settlementNotify) AutoPushTask() {
	sdate := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	edate := time.Now().Add(-1 * time.Minute).Format("2006-01-02 15:04:05")
	whereMap := map[string]interface{}{
		"orderStatus in":    []string{"Success", "Fail"},
		"callbackSuccess <": 1,
		"callbackLimit <":   5,
		"backNoticeUrl !=":  "",
		"created_at >=":     sdate,
		"created_at <=":     edate,
	}
	cond, vals, err := utils.WhereBuild(whereMap)
	if err != nil {
		logrus.Error("统计未回调代付订单失败：", err)
		return
	}

	var count int64
	err = sqls.DB().Table("platform_settlement_order").Where(cond, vals...).Count(&count).Error
	if err != nil {
		logrus.Error("统计未回调代付订单失败：", err)
		return
	}
	if count == 0 {
		logrus.Error("暂无符合条件的代付订单")
		return
	}
	limit := 50
	var orders []string
	if int(count) > limit {
		totalpages := int(math.Ceil(float64(count) / float64(limit))) //page总数
		for i := 1; i < totalpages; i++ {
			err = sqls.DB().Table("platform_settlement_order").Limit(limit).Offset(i).Order("orderId asc").Where(cond, vals...).Pluck("platformOrderNo", &orders).Error
			if err != nil {
				logrus.Error("查询未完成代付订单失败")
				break
			}
			go s.BatchPush(orders)
		}
		return
	} else {
		err = sqls.DB().Table("platform_settlement_order").Where(cond, vals...).Pluck("platformOrderNo", &orders).Error
		if err != nil {
			logrus.Error("查询未回调代付订单失败：", err)
		}
		go s.BatchPush(orders)
		return
	}

}

func (s *settlementNotify) AutoCacheOrder() {
	whereMap := map[string]interface{}{
		"orderStatus in":    []string{"Success", "Fail"},
		"callbackSuccess <": 1,
		"callbackLimit <":   5,
		"backNoticeUrl !=":  "",
	}
	cond, vals, err := utils.WhereBuild(whereMap)
	if err != nil {
		logrus.Error("统计未回调代付订单失败：", err)
		return
	}

	var count int64
	err = sqls.DB().Table("platform_settlement_order").Where(cond, vals...).Count(&count).Error
	if err != nil {
		logrus.Error("统计未回调代付订单失败：", err)
		return
	}
	if count == 0 {
		logrus.Error("暂无未完成订单")
		return
	}
	limit := 50
	var orders []model.PlatformSettlementOrder
	if int(count) > limit {
		totalpages := int(math.Ceil(float64(count) / float64(limit))) //page总数
		for i := 1; i < totalpages; i++ {
			err = sqls.DB().Table("platform_settlement_order").Limit(limit).Offset(i).Order("orderId asc").Where(cond, vals...).Find(&orders).Error
			if err != nil {
				logrus.Error("查询未回调代付订单失败1")
				break
			}
			for _, order := range orders {
				go cache.SettleCache.SetCacheByPlatformOrderNo(order.PlatformOrderNo, order)
			}
		}
		return
	} else {
		err = sqls.DB().Table("platform_settlement_order").Where(cond, vals...).Find(&orders).Error
		if err != nil {
			logrus.Error("查询未回调代付订单失败2：", err)
		}
		for _, order := range orders {
			go cache.SettleCache.SetCacheByPlatformOrderNo(order.PlatformOrderNo, order)
		}
		return
	}

}

func (s *settlementNotify) BatchPush(orders []string) {
	for _, platformOrderNo := range orders {
		go s.Push(0, platformOrderNo)
	}
	return
}

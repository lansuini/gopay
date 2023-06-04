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
	"time"

	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var PayNotify = newPayNotify()

func newPayNotify() *payNotify {
	return &payNotify{}
}

type payNotify struct {
	RedisClient *redis.Client
}

func (s *payNotify) Push(taskId int64, platformOrderNo string) {
	data := ChildTask{}
	if taskId == 0 {
		task := model.PayNotifyTask{
			Status:          "Execute",
			PlatformOrderNo: platformOrderNo,
		}

		err := sqls.DB().Where("platformOrderNo = ?", platformOrderNo).FirstOrCreate(&task).Error
		if err != nil {
			logrus.Error("payNotify Push FirstOrCreate Error-", platformOrderNo, err)
			return
		}
		data.TaskId = task.ID
	} else {
		data.TaskId = taskId
	}
	data.PlatformOrderNo = platformOrderNo
	jsonStr, err := json.Marshal(data)
	if err != nil {
		logrus.Error("payNotify Push Marshal Error-", platformOrderNo, err)
		return
	}

	res, err := redisServer.LPush(ctx, "paynotify:queue", jsonStr).Result()
	if err != nil {
		logrus.Error("payNotify LPush Error-", platformOrderNo, err)
		return
	}
	logrus.Info("paynotify:queue ,lpRes:", res, ",platformOrderNo:", platformOrderNo, ",taskId:", data.TaskId)
	return
}

func (s *payNotify) Pop() {
	mutex.Lock()         // 添加互斥锁
	defer mutex.Unlock() // 使用结束时解锁
	data, err := redisServer.RPop(ctx, "paynotify:queue").Result()
	if err != nil && err != redis.Nil {
		logrus.Error("payNotify RPop ：", err)
		return
	}
	if len(data) == 0 {
		return
	}
	cacheKey := "paynotify:queue:lasttime"
	redisServer.SetEX(ctx, cacheKey, time.Now(), 60*time.Second)
	if data == "" {
		logrus.Error("payNotify RPop 数据为空")
		return
	}
	task := ChildTask{}
	err = json.Unmarshal([]byte(data), &task)
	if err != nil {
		logrus.Error(data, "-payNotify Unmarshal Error-", err)
		return
	}
	if reflect.DeepEqual(task, ChildTask{}) {
		logrus.Info(data, "-payNotify RPop task is empty")
		return
	}

	taskData := model.PayNotifyTask{}
	for i := 1; i <= 5; i++ {
		err = sqls.DB().Where("id = ?", task.TaskId).First(&taskData).Error
		if err != nil {
			logrus.Info("payNotify RPop 支付查询任务查询异常-", err.Error())
			break
		}
		if !reflect.DeepEqual(taskData, model.PayNotifyTask{}) {
			logrus.Info("paynotify:queue,paynotify task count:", i, taskData)
			break
		}
		time.Sleep(500 * time.Millisecond) //0.5秒，500毫秒
	}

	if reflect.DeepEqual(taskData, model.PayNotifyTask{}) {
		logrus.Info("paynotify:queue,empty taskData:", taskData)
		return
	}

	if taskData.RetryCount >= 5 {
		logrus.Info("payNotify RPop 支付订单回调次数达到上限", taskData)
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": taskData.RetryCount + 1,
		}
		go s.UpdateTask(taskData, updates)
		return
	}

	orderData, res := cache.PayCache.GetCacheByPlatformOrderNo(task.PlatformOrderNo)
	if !res {
		logrus.Info("payNotify RPop 获取支付订单失败")
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": taskData.RetryCount + 1,
			"failReason": "获取支付订单失败",
		}
		go s.UpdateTask(taskData, updates)

		return
	}

	if orderData.CallbackSuccess {
		logrus.Info("payNotify RPop 支付订单已回调成功-", taskData)
		return
	}

	if orderData.OrderStatus != "Success" {

		logrus.Info("payNotify RPop 支付订单状态未完成-", taskData)
		return
	}

	if orderData.BackNoticeUrl == "" {
		logrus.Error("payNotify RPop", orderData.PlatformOrderNo, "-支付回调失败：回调地址为空")
	}
	isValid := utils.IsValidUrl(orderData.BackNoticeUrl)
	if !isValid {
		logrus.Error("payNotify RPop", orderData.PlatformOrderNo, "-支付回调失败：回调地址格式错误-", orderData.BackNoticeUrl)
		return
		//order.BackNoticeURL = "http://cb.luckypay.mm:8082/paycallback"
	}
	go PayService.CallbackMerchant(task.PlatformOrderNo, orderData, taskData)
	//updates := map[string]interface{}{
	//	"status":     "Success",
	//	"retryCount": taskData.RetryCount + 1,
	//	"failReason": "订单已完成",
	//}
	//go s.UpdateTask(taskData, updates)

	return
}

func (s *payNotify) UpdateTask(task model.PayNotifyTask, updates map[string]interface{}) {
	if reflect.DeepEqual(task, model.PayNotifyTask{}) {
		return
	}
	err := sqls.DB().Model(task).Updates(updates).Error
	if err != nil {
		logrus.Info("payNotify RPop 更新taskData2失败", err)
	}
}

func (s *payNotify) AutoPushTask() {

	sdate := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	whereMap := map[string]interface{}{
		"orderStatus =":     "Success",
		"callbackSuccess <": 1,
		"callbackLimit <":   5,
		"backNoticeUrl !=":  "",
		"created_at >=":     sdate,
	}
	cond, vals, err := utils.WhereBuild(whereMap)
	if err != nil {
		logrus.Error("统计未回调支付订单失败：", err)
		return
	}

	var count int64
	err = sqls.DB().Table("platform_pay_order").Where(cond, vals...).Count(&count).Error
	if err != nil {
		logrus.Error("统计未回调支付订单失败：", err)
		return
	}
	if count == 0 {
		logrus.Error("暂无未回调订单")
		return
	}
	limit := 50
	var orders []string
	if int(count) > limit {
		totalpages := int(math.Ceil(float64(count) / float64(limit))) //page总数
		for i := 1; i < totalpages; i++ {
			err = sqls.DB().Table("platform_pay_order").Limit(limit).Offset(i).Order("orderId asc").Where(cond, vals...).Pluck("platformOrderNo", &orders).Error
			if err != nil {
				logrus.Error("查询未完成支付订单失败")
				break
			}
			go s.BatchPush(orders)
		}
		return
	} else {
		err = sqls.DB().Table("platform_pay_order").Where(cond, vals...).Pluck("platformOrderNo", &orders).Error
		if err != nil {
			logrus.Error("查询未回调支付订单失败：", err)
		}
		go s.BatchPush(orders)
		return
	}

}

func (s *payNotify) AutoCacheOrder() {
	whereMap := map[string]interface{}{
		"orderStatus =":     "Success",
		"callbackSuccess <": 1,
		"callbackLimit <":   5,
		"backNoticeUrl !=":  "",
	}
	cond, vals, err := utils.WhereBuild(whereMap)
	if err != nil {
		logrus.Error("统计未回调支付订单失败：", err)
		return
	}

	var count int64
	err = sqls.DB().Table("platform_pay_order").Where(cond, vals...).Count(&count).Error
	if err != nil {
		logrus.Error("统计未回调支付订单失败：", err)
		return
	}
	if count == 0 {
		logrus.Error("暂无符合条件的支付订单")
		return
	}
	limit := 50
	var orders []model.PlatformPayOrder
	if int(count) > limit {
		totalpages := int(math.Ceil(float64(count) / float64(limit))) //page总数
		for i := 1; i < totalpages; i++ {
			err = sqls.DB().Table("platform_pay_order").Limit(limit).Offset(i).Order("orderId asc").Where(cond, vals...).Find(&orders).Error
			if err != nil {
				logrus.Error("查询未回调支付订单失败1")
				break
			}
			for _, order := range orders {
				go cache.PayCache.SetCacheByPlatformOrderNo(order.PlatformOrderNo, order)
			}
		}
		return
	} else {
		err = sqls.DB().Table("platform_pay_order").Where(cond, vals...).Find(&orders).Error
		if err != nil {
			logrus.Error("查询未回调支付订单失败2：", err)
		}
		for _, order := range orders {
			go cache.PayCache.SetCacheByPlatformOrderNo(order.PlatformOrderNo, order)
		}
		return
	}

}

func (s *payNotify) BatchPush(orders []string) {
	for _, platformOrderNo := range orders {
		go s.Push(0, platformOrderNo)
	}
	return
}

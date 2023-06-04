package services

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"luckypay/cache"
	"luckypay/channels"
	"luckypay/model"
	systemConfig "luckypay/pkg/config"
	"math"
	"reflect"
	"time"

	//"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
)

var SettlementFetch = newSettlementFetch()

func newSettlementFetch() *settlementFetch {
	return &settlementFetch{}
}

type settlementFetch struct {
	RedisClient *redis.Client
}
type ChildTask struct {
	TaskId          int64  `json:"taskId"`
	PlatformOrderNo string `json:"platformOrderNo"`
}

func (s *settlementFetch) Push(taskId int64, platformOrderNo string) {
	data := ChildTask{}
	if taskId == 0 {
		task := model.SettlementFetchTask{
			Status:          "Execute",
			PlatformOrderNo: platformOrderNo,
		}

		err := sqls.DB().Where("platformOrderNo = ?", platformOrderNo).FirstOrCreate(&task).Error
		if err != nil {
			logrus.Error("settlementFetch Push FirstOrCreate Error-", platformOrderNo, err)
			return
		}
		data.TaskId = task.ID
	} else {
		data.TaskId = taskId
	}
	data.PlatformOrderNo = platformOrderNo
	jsonStr, err := json.Marshal(data)
	if err != nil {
		logrus.Error("settlementFetch Push Marshal Error-", platformOrderNo, err)
		return
	}
	res, err := redisServer.LPush(ctx, "settlementfetch:queue", jsonStr).Result()
	if err != nil {
		logrus.Error("settlementFetch LPush Error-", platformOrderNo, err)
		return
	}
	logrus.Info("settlementfetch:queue ,lpRes:", res, ",platformOrderNo:", platformOrderNo, ",taskId:", data.TaskId)
	return
}

func (s *settlementFetch) Pop() {
	mutex.Lock()         // 添加互斥锁
	defer mutex.Unlock() // 使用结束时解锁
	data, err := redisServer.RPop(ctx, "settlementfetch:queue").Result()
	if err != nil && err != redis.Nil {
		logrus.Error("settlementFetch RPop ：", err)
		return
	}
	if len(data) == 0 {
		//logrus.Info("settlementFetch RPop 数据为空")
		return
	}
	cacheKey := "settlementfetch:queue:lasttime"
	redisServer.SetEX(ctx, cacheKey, time.Now(), 60*time.Second)
	if data == "" {
		//redisServer = nil
		logrus.Error("settlementFetch RPop 数据为空")
		return
	}
	task := ChildTask{}
	err = json.Unmarshal([]byte(data), &task)
	if err != nil {
		logrus.Error("settlementFetch Unmarshal Error-", err)
		return
	}
	if reflect.DeepEqual(task, ChildTask{}) {
		logrus.Info("settlementFetch RPop task is empty")
		return
	}

	taskData := model.SettlementFetchTask{}
	err = sqls.DB().Where("id = ?", task.TaskId).First(&taskData).Error
	if err != nil {
		logrus.Info("settlementFetch RPop 代付查询任务查询异常-", err.Error())
		return
	}

	if reflect.DeepEqual(taskData, model.SettlementFetchTask{}) {
		logrus.Info("settlementfetch:queue,empty taskData:", taskData)
		return
	}

	if taskData.RetryCount >= 5 && systemConfig.Viper.GetBool("Crontab.SETTLE_FETCH_LIMIT") {
		logrus.Info("settlementFetch RPop 订单查询已达5次", taskData)
		//updates := map[string]interface{}{
		//	"status":     "Fail",
		//	"retryCount": taskData.RetryCount + 1,
		//}
		//go s.UpdateTask(taskData, updates)
		return
	}

	orderData, res := cache.SettleCache.GetCacheByPlatformOrderNo(task.PlatformOrderNo)
	if !res {
		logrus.Info("settlementFetch RPop 获取代付订单失败")
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": taskData.RetryCount + 1,
			"failReason": "获取代付订单失败",
		}
		go s.UpdateTask(taskData, updates)

		/*err = sqls.DB().Model(taskData).Updates(updates).Error
		if err != nil {
			logrus.Info("settlementFetch RPop 更新taskData1失败", err)
		}*/
		return
	}

	if orderData.OrderStatus != "Transfered" {
		//go SettleService.CallbackMerchant(task.PlatformOrderNo, orderData, model.SettlementNotifyTask{})
		logrus.Info("settlementFetch RPop 代付订单不是处理中")
		updates := map[string]interface{}{
			"status":     "Success",
			"retryCount": taskData.RetryCount + 1,
			"failReason": "订单已完成",
		}
		go s.UpdateTask(taskData, updates)

		return
	}

	lockKey := "settlementfetch:queue:" + task.PlatformOrderNo
	lockNum, _ := redisServer.Incr(ctx, lockKey).Result()
	if lockNum > 1 {
		logrus.Info("settlementFetch RPop 任务锁定", taskData)
		go s.Push(taskData.ID, taskData.PlatformOrderNo)
		return
	}
	redisServer.Expire(ctx, lockKey, 120*time.Second)
	//查询订单状态
	if _, ok := channels.Channels[orderData.Channel]; !ok {
		logrus.Info("settlementFetch RPop 订单代付渠道不存在:", taskData)
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": taskData.RetryCount + 1,
			"failReason": "订单代付渠道不存在",
		}
		go s.UpdateTask(taskData, updates)
		return
	}
	cacheChannelMerchantData, res := ChannelMerchant.GetCacheByChannelMerchantNo(orderData.ChannelMerchantNo)
	if !res {
		logrus.Error(orderData.PlatformOrderNo, "-代付渠道数据获取失败 ：", orderData.ChannelMerchantNo)
		updates := map[string]interface{}{
			"status":     "Fail",
			"retryCount": taskData.RetryCount + 1,
			"failReason": "代付渠道数据获取失败",
		}
		go s.UpdateTask(taskData, updates)
		return
	}
	Channel := channels.Channels[orderData.Channel]
	go SettleService.QueryOrder(Channel, orderData, cacheChannelMerchantData, taskData)
	return
}

func (s *settlementFetch) UpdateTask(task model.SettlementFetchTask, updates map[string]interface{}) {
	err := sqls.DB().Model(task).Updates(updates).Error
	if err != nil {
		logrus.Info("settlementFetch RPop 更新taskData2失败", err)
	}
}

func (s *settlementFetch) AutoPushTask() {

	var count int64
	sdate := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	edate := time.Now().Add(-2 * time.Minute).Format("2006-01-02 15:04:05")
	err := sqls.DB().Table("platform_settlement_order").Where("orderStatus = ?", "Transfered").Where("created_at >= ?", sdate).Where("created_at <= ?", edate).Count(&count).Error
	if err != nil {
		logrus.Error("统计未完成代付订单失败", err)
		return
	}
	if count == 0 {
		logrus.Error("暂无未完成订单")
		return
	}
	limit := 50
	var orders []string
	if int(count) > limit {
		totalpages := int(math.Ceil(float64(count) / float64(limit))) //page总数
		for i := 1; i < totalpages; i++ {
			err = sqls.DB().Table("platform_settlement_order").Limit(limit).Offset(i).Order("orderId asc").Where("orderStatus = ?", "Transfered").Where("created_at >= ?", sdate).Where("created_at <= ?", edate).Pluck("platformOrderNo", &orders).Error
			if err != nil {
				logrus.Error("查询未完成代付订单失败")
				break
			}
			go s.BatchPush(orders)
		}
		return
	} else {
		sqls.DB().Table("platform_settlement_order").Order("orderId asc").Where("orderStatus = ?", "Transfered").Where("created_at >= ?", sdate).Where("created_at <= ?", edate).Pluck("platformOrderNo", &orders)
		if err != nil {
			logrus.Error("查询未完成代付订单失败")
			return
		}
		go s.BatchPush(orders)
		return
	}

}

func (s *settlementFetch) BatchPush(orders []string) {
	for _, platformOrderNo := range orders {
		go s.Push(0, platformOrderNo)
	}
	return
}

func (s *settlementFetch) AutoCacheOrder() {
	var count int64

	err := sqls.DB().Table("platform_settlement_order").Where("orderStatus = ?", "Transfered").Count(&count).Error
	if err != nil {
		logrus.Error("统计未完成代付订单失败", err)
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
			err = sqls.DB().Table("platform_settlement_order").Limit(limit).Offset(i).Order("orderId asc").Where("orderStatus = ?", "Transfered").Find(&orders).Error
			if err != nil {
				logrus.Error("查询未完成代付订单失败1")
				break
			}
			for _, order := range orders {
				go cache.SettleCache.SetCacheByPlatformOrderNo(order.PlatformOrderNo, order)
			}
		}
		return
	} else {
		err = sqls.DB().Table("platform_settlement_order").Order("orderId asc").Where("orderStatus = ?", "Transfered").Find(&orders).Error
		if err != nil {
			logrus.Error("查询未完成代付订单失败2")
			return
		}
		for _, order := range orders {
			go cache.SettleCache.SetCacheByPlatformOrderNo(order.PlatformOrderNo, order)
		}
		return
	}
}

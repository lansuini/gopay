package channels

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"luckypay/model"
)

type AliPay struct {
	RedisClient *redis.Client
}

func (ch *AliPay) PayOrder(params model.PayParams) {
	fmt.Println("%+v", params)
	fmt.Println("running aliPay PayOrder")
}

func (ch *AliPay) SettlementOrder() {
	fmt.Println("running aliPay PayOrder")
}

func (ch *AliPay) QueryBalance() {
	//funcs := map[string]interface{}{
	//	"foo0": ch.PayOrder,
	//}
}

func (ch *AliPay) CallBack() {
	//funcs := map[string]interface{}{
	//	"foo0": ch.PayOrder,
	//}
}

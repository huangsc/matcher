package main

import (
	"log"
	"time"

	"github.com/huangsc/matcher"
	"github.com/shopspring/decimal"
)

type ExampleHandler struct{}

func (h *ExampleHandler) OnTrade(trade *matcher.Trade) {
	log.Printf("Trade: Price=%v Quantity=%v\n",
		trade.Price, trade.Quantity)
}

func (h *ExampleHandler) OnOrderUpdate(order *matcher.Order) {
	log.Printf("Order Update: ID=%v Status=%v\n",
		order.ID, order.Status)
}

func main() {
	engine := matcher.NewMatchEngine(&ExampleHandler{})
	engine.Start()

	// 提交订单示例
	buyOrder := &matcher.Order{
		ID:       "1",
		Side:     matcher.OrderSideBuy,
		Price:    decimal.NewFromFloat(100.0),
		Quantity: decimal.NewFromFloat(1.0),
	}

	engine.SubmitOrder(buyOrder)

	time.Sleep(time.Second)
	engine.Stop()
}

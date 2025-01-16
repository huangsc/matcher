package main

import (
	"fmt"
	"log"
	"time"

	"github.com/huangsc/matcher"
	"github.com/shopspring/decimal"
)

type ExampleHandler struct{}

func (h *ExampleHandler) OnTrade(trade *matcher.Trade) {
	log.Printf("Trade: BuyOrderID=%v SellOrderID=%v Price=%v Quantity=%v\n",
		trade.BuyOrder.ID, trade.SellOrder.ID, trade.Price, trade.Quantity)
}

func (h *ExampleHandler) OnOrderUpdate(order *matcher.Order) {
	log.Printf("Order Update: ID=%v Side=%v Price=%v Quantity=%v FilledQuantity=%v Status=%v\n",
		order.ID, order.Side, order.Price, order.Quantity, order.FilledQuantity, order.Status)
}

func main() {
	engine := matcher.NewMatchEngine(&ExampleHandler{})
	engine.Start()

	// 添加一些卖单
	sellOrders := []*matcher.Order{
		{
			ID:       "sell-1",
			Side:     matcher.OrderSideSell,
			Price:    decimal.NewFromFloat(100.5),
			Quantity: decimal.NewFromFloat(2.0),
		},
		{
			ID:       "sell-2",
			Side:     matcher.OrderSideSell,
			Price:    decimal.NewFromFloat(100.8),
			Quantity: decimal.NewFromFloat(1.5),
		},
		{
			ID:       "sell-3",
			Side:     matcher.OrderSideSell,
			Price:    decimal.NewFromFloat(100.2),
			Quantity: decimal.NewFromFloat(3.0),
		},
	}

	for _, order := range sellOrders {
		engine.SubmitOrder(order)
	}

	time.Sleep(time.Millisecond * 100) // 等待卖单处理完成

	// 添加一些买单
	buyOrders := []*matcher.Order{
		{
			ID:       "buy-1",
			Side:     matcher.OrderSideBuy,
			Price:    decimal.NewFromFloat(100.3),
			Quantity: decimal.NewFromFloat(1.0),
		},
		{
			ID:       "buy-2",
			Side:     matcher.OrderSideBuy,
			Price:    decimal.NewFromFloat(100.8),
			Quantity: decimal.NewFromFloat(2.0),
		},
		{
			ID:       "buy-3",
			Side:     matcher.OrderSideBuy,
			Price:    decimal.NewFromFloat(100.1),
			Quantity: decimal.NewFromFloat(1.5),
		},
	}

	// 间隔提交买单，观察撮合过程
	for _, order := range buyOrders {
		engine.SubmitOrder(order)
		time.Sleep(time.Millisecond * 100) // 等待每个订单撮合完成
	}

	// 测试订单取消
	cancelOrder := &matcher.Order{
		ID:       "sell-4",
		Side:     matcher.OrderSideSell,
		Price:    decimal.NewFromFloat(101.0),
		Quantity: decimal.NewFromFloat(1.0),
	}
	engine.SubmitOrder(cancelOrder)
	time.Sleep(time.Millisecond * 50)

	// 打印最终订单簿状态
	fmt.Println("\nFinal OrderBook Status:")
	printOrderBook(engine)

	time.Sleep(time.Second)
	engine.Stop()
}

// 添加辅助函数打印订单簿状态
func printOrderBook(engine *matcher.MatchEngine) {
	fmt.Println("Buy Orders:")
	for _, order := range engine.GetBuyOrders() {
		fmt.Printf("  ID=%v Price=%v Quantity=%v FilledQuantity=%v Status=%v\n",
			order.ID, order.Price, order.Quantity, order.FilledQuantity, order.Status)
	}

	fmt.Println("Sell Orders:")
	for _, order := range engine.GetSellOrders() {
		fmt.Printf("  ID=%v Price=%v Quantity=%v FilledQuantity=%v Status=%v\n",
			order.ID, order.Price, order.Quantity, order.FilledQuantity, order.Status)
	}
}

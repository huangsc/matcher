package matcher

import (
	"sync/atomic"
	"time"
)

type MatchEngine struct {
	disruptor    *Disruptor
	orderBook    *OrderBook
	eventHandler EventHandler
	running      atomic.Bool
}

func NewMatchEngine(handler EventHandler) *MatchEngine {
	return &MatchEngine{
		disruptor:    NewDisruptor(defaultBufferSize),
		orderBook:    NewOrderBook(),
		eventHandler: handler,
	}
}

func (e *MatchEngine) Start() {
	if !e.running.CompareAndSwap(false, true) {
		return
	}

	go e.process()
}

func (e *MatchEngine) Stop() {
	e.running.Store(false)
}

// 处理事件循环
func (e *MatchEngine) process() {
	e.disruptor.Process(func(event *Event) {
		switch event.Type {
		case EventTypeNew:
			e.handleNewOrder(event.Order)
		case EventTypeCancel:
			e.handleCancelOrder(event.Order)
		case EventTypeMatch:
			e.handleMatch(event.Order)
		}
	})
}

// 提交新订单
func (e *MatchEngine) SubmitOrder(order *Order) bool {
	event := &Event{
		Type:      EventTypeNew,
		Order:     order,
		Timestamp: time.Now().UnixNano(),
	}
	return e.disruptor.TryPublish(event)
}

// 处理新订单
func (e *MatchEngine) handleNewOrder(order *Order) {
	trades := e.orderBook.Match(order)

	// 通知成交结果
	for _, trade := range trades {
		e.eventHandler.OnTrade(trade)
	}

	// 如果订单未完全成交，加入订单簿
	if !order.IsFullyFilled() {
		e.orderBook.Add(order)
	}

	e.eventHandler.OnOrderUpdate(order)
}

// 处理撤单
func (e *MatchEngine) handleCancelOrder(order *Order) {
	if e.orderBook.Remove(order.ID) {
		order.Status = OrderStatusCanceled
		e.eventHandler.OnOrderUpdate(order)
	}
}

// 处理撮合
func (e *MatchEngine) handleMatch(order *Order) {
	if trades := e.orderBook.Match(order); len(trades) > 0 {
		for _, trade := range trades {
			e.eventHandler.OnTrade(trade)
		}
		e.eventHandler.OnOrderUpdate(order)
	}
}

func (e *MatchEngine) GetBuyOrders() []*Order {
	orders := make([]*Order, 0)
	for _, level := range e.orderBook.buyLevels {
		orders = append(orders, level.Orders...)
	}
	return orders
}

func (e *MatchEngine) GetSellOrders() []*Order {
	orders := make([]*Order, 0)
	for _, level := range e.orderBook.sellLevels {
		orders = append(orders, level.Orders...)
	}
	return orders
}

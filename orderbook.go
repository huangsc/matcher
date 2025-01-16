package matcher

import (
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

var (
	tradePool = sync.Pool{
		New: func() interface{} {
			return &Trade{}
		},
	}
)

type OrderBook struct {
	buyLevels  []*PriceLevel
	sellLevels []*PriceLevel
}

type PriceLevel struct {
	Price  decimal.Decimal // 买单按价格降序排列，卖单按价格升序排列
	Orders []*Order        // 同一价格的订单按时间优先排序
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		buyLevels:  make([]*PriceLevel, 0, 64),
		sellLevels: make([]*PriceLevel, 0, 64),
	}
}

func (ob *OrderBook) updateLevels(side OrderSide, levels []*PriceLevel) {
	// 更新指定方向的价格档位，同时清理已完全成交的订单和空档位
	if side == OrderSideBuy {
		ob.sellLevels = levels
	} else {
		ob.buyLevels = levels
	}
}

func (ob *OrderBook) Match(order *Order) []*Trade {
	trades := make([]*Trade, 0, 64)

	var levels []*PriceLevel
	if order.Side == OrderSideBuy {
		levels = ob.sellLevels
	} else {
		levels = ob.buyLevels
	}

	// 保持未空的价格档位
	validLevels := make([]*PriceLevel, 0, len(levels))

	for _, level := range levels {
		if order.Side == OrderSideBuy && level.Price.GreaterThan(order.Price) {
			validLevels = append(validLevels, level)
			continue
		}
		if order.Side == OrderSideSell && level.Price.LessThan(order.Price) {
			validLevels = append(validLevels, level)
			continue
		}

		validOrders := make([]*Order, 0, len(level.Orders))
		for _, matchOrder := range level.Orders {
			if matchOrder.IsFullyFilled() {
				continue
			}

			matchQty := decimal.Min(order.RemainQuantity(), matchOrder.RemainQuantity())
			if matchQty.IsZero() {
				validOrders = append(validOrders, matchOrder)
				continue
			}

			trade := tradePool.Get().(*Trade)
			trade.Price = level.Price
			trade.Quantity = matchQty
			trade.BuyOrder = order
			trade.SellOrder = matchOrder
			trade.Timestamp = time.Now().UnixMilli()

			trades = append(trades, trade)

			order.FilledQuantity = order.FilledQuantity.Add(matchQty)
			matchOrder.FilledQuantity = matchOrder.FilledQuantity.Add(matchQty)

			if order.IsFullyFilled() {
				order.Status = OrderStatusFilled
				if !matchOrder.IsFullyFilled() {
					validOrders = append(validOrders, matchOrder)
				}
				level.Orders = validOrders
				if len(validOrders) > 0 {
					validLevels = append(validLevels, level)
				}
				ob.updateLevels(order.Side, validLevels)
				return trades
			}

			if !matchOrder.IsFullyFilled() {
				validOrders = append(validOrders, matchOrder)
			}
		}

		if len(validOrders) > 0 {
			level.Orders = validOrders
			validLevels = append(validLevels, level)
		}
	}

	if !order.IsFullyFilled() {
		order.Status = OrderStatusPartial
		ob.Add(order)
	}

	ob.updateLevels(order.Side, validLevels)
	return trades
}

func (ob *OrderBook) Add(order *Order) {
	var levels *[]*PriceLevel
	if order.Side == OrderSideBuy {
		levels = &ob.buyLevels
	} else {
		levels = &ob.sellLevels
	}

	// 快速路径：如果是空的或者应该放在末尾
	if len(*levels) == 0 || (order.Side == OrderSideBuy && order.Price.LessThan((*levels)[len(*levels)-1].Price)) ||
		(order.Side == OrderSideSell && order.Price.GreaterThan((*levels)[len(*levels)-1].Price)) {
		*levels = append(*levels, &PriceLevel{
			Price:  order.Price,
			Orders: []*Order{order},
		})
		return
	}

	// 快速路径：如果应该放在开头
	if (order.Side == OrderSideBuy && order.Price.GreaterThan((*levels)[0].Price)) ||
		(order.Side == OrderSideSell && order.Price.LessThan((*levels)[0].Price)) {
		*levels = append([]*PriceLevel{{
			Price:  order.Price,
			Orders: []*Order{order},
		}}, (*levels)...)
		return
	}

	// 二分查找
	idx := sort.Search(len(*levels), func(i int) bool {
		if order.Side == OrderSideBuy {
			return (*levels)[i].Price.LessThan(order.Price)
		}
		return (*levels)[i].Price.GreaterThan(order.Price)
	})

	// 如果找到相同价格
	if idx > 0 && (*levels)[idx-1].Price.Equal(order.Price) {
		(*levels)[idx-1].Orders = append((*levels)[idx-1].Orders, order)
		return
	}

	// 插入新价格档位
	newLevel := &PriceLevel{
		Price:  order.Price,
		Orders: []*Order{order},
	}

	*levels = append(*levels, nil)
	copy((*levels)[idx+1:], (*levels)[idx:])
	(*levels)[idx] = newLevel
}

func (ob *OrderBook) Remove(orderID string) bool {
	// 提取公共的移除逻辑
	removeFromLevel := func(levels *[]*PriceLevel) bool {
		for i := 0; i < len(*levels); i++ {
			level := (*levels)[i]
			for j, order := range level.Orders {
				if order.ID == orderID {
					level.Orders = append(level.Orders[:j], level.Orders[j+1:]...)
					// 如果价格档位为空，移除该档位
					if len(level.Orders) == 0 {
						*levels = append((*levels)[:i], (*levels)[i+1:]...)
					}
					return true
				}
			}
		}
		return false
	}

	// 先检查买单
	if removeFromLevel(&ob.buyLevels) {
		return true
	}
	// 再检查卖单
	return removeFromLevel(&ob.sellLevels)
}

func ReleaseTrades(trades []*Trade) {
	for _, t := range trades {
		tradePool.Put(t)
	}
}

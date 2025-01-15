package matcher

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderBook struct {
	buyOrders  *PriceLevel
	sellOrders *PriceLevel
}

type PriceLevel struct {
	Price  decimal.Decimal
	Orders []*Order
	Next   *PriceLevel
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		buyOrders:  nil,
		sellOrders: nil,
	}
}

func (ob *OrderBook) Match(order *Order) []*Trade {
	trades := make([]*Trade, 0)

	var matchPrice *PriceLevel
	if order.Side == OrderSideBuy {
		matchPrice = ob.sellOrders
	} else {
		matchPrice = ob.buyOrders
	}

	// 撮合循环
	for matchPrice != nil {
		// 检查价格是否匹配
		if !ob.isPriceMatch(order, matchPrice.Price) {
			break
		}

		// 遍历该价格档位的所有订单
		for i := 0; i < len(matchPrice.Orders); i++ {
			matchOrder := matchPrice.Orders[i]

			// 计算成交量
			matchQty := decimal.Min(
				order.RemainQuantity(),
				matchOrder.RemainQuantity(),
			)

			if matchQty.IsZero() {
				continue
			}

			// 创建成交记录
			trade := &Trade{
				Price:     matchPrice.Price,
				Quantity:  matchQty,
				BuyOrder:  order,
				SellOrder: matchOrder,
				Timestamp: time.Now().UnixMilli(),
			}
			trades = append(trades, trade)

			// 更新订单状态
			order.FilledQuantity = order.FilledQuantity.Add(matchQty)
			matchOrder.FilledQuantity = matchOrder.FilledQuantity.Add(matchQty)

			// 检查是否完全成交
			if order.IsFullyFilled() {
				order.Status = OrderStatusFilled
				break
			}
		}

		matchPrice = matchPrice.Next
	}

	// 如果订单未完全成交，加入订单簿
	if !order.IsFullyFilled() {
		order.Status = OrderStatusPartial
	}

	return trades
}

func (ob *OrderBook) isPriceMatch(order *Order, price decimal.Decimal) bool {
	if order.Side == OrderSideBuy {
		return order.Price.GreaterThanOrEqual(price)
	}
	return order.Price.LessThanOrEqual(price)
}

func (ob *OrderBook) Add(order *Order) {
	// 根据订单方向添加到对应的订单簿中
	if order.Side == OrderSideBuy {
		ob.buyOrders = ob.buyOrders.Add(order)
	} else {
		ob.sellOrders = ob.sellOrders.Add(order)
	}
}

func (ob *OrderBook) Remove(orderID string) bool {
	// 根据订单方向从对应的订单簿中移除订单
	if ob.buyOrders != nil && ob.buyOrders.Orders != nil {
		for i, order := range ob.buyOrders.Orders {
			if order.ID == orderID {
				ob.buyOrders.Orders = append(ob.buyOrders.Orders[:i], ob.buyOrders.Orders[i+1:]...)
				return true
			}
		}
	}

	if ob.sellOrders != nil && ob.sellOrders.Orders != nil {
		for i, order := range ob.sellOrders.Orders {
			if order.ID == orderID {
				ob.sellOrders.Orders = append(ob.sellOrders.Orders[:i], ob.sellOrders.Orders[i+1:]...)
				return true
			}
		}
	}

	return false
}

func (pl *PriceLevel) Add(order *Order) *PriceLevel {
	if pl == nil {
		return &PriceLevel{
			Price:  order.Price,
			Orders: []*Order{order},
		}
	}

	if pl.Price.Equal(order.Price) {
		pl.Orders = append(pl.Orders, order)
		return pl
	}

	if (order.Side == OrderSideBuy && pl.Price.LessThan(order.Price)) ||
		(order.Side == OrderSideSell && pl.Price.GreaterThan(order.Price)) {
		return &PriceLevel{
			Price:  order.Price,
			Orders: []*Order{order},
			Next:   pl,
		}
	}

	pl.Next = pl.Next.Add(order)
	return pl
}

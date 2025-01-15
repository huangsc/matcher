package matcher

import "github.com/shopspring/decimal"

type Order struct {
	ID             string
	Side           OrderSide
	Price          decimal.Decimal
	Quantity       decimal.Decimal
	FilledQuantity decimal.Decimal
	Status         OrderStatus
}

type OrderStatus string

const (
	OrderStatusNew      OrderStatus = "new"
	OrderStatusPartial  OrderStatus = "partial"
	OrderStatusFilled   OrderStatus = "filled"
	OrderStatusCanceled OrderStatus = "canceled"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

func (o *Order) RemainQuantity() decimal.Decimal {
	return o.Quantity.Sub(o.FilledQuantity)
}

func (o *Order) IsFullyFilled() bool {
	return o.RemainQuantity().IsZero()
}

type Trade struct {
	ID        string
	BuyOrder  *Order
	SellOrder *Order
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Timestamp int64
}

type Event struct {
	Type      EventType
	Order     *Order
	Timestamp int64
}

type EventType string

const (
	EventTypeNew    EventType = "new"
	EventTypeCancel EventType = "cancel"
	EventTypeMatch  EventType = "match"
)

package matcher

type EventHandler interface {
	OnTrade(trade *Trade)
	OnOrderUpdate(order *Order)
}

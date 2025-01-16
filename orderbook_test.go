package matcher

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

func BenchmarkOrderBook_Match(b *testing.B) {
	ob := NewOrderBook()

	// 预先添加一些卖单
	for i := 0; i < 1000; i++ {
		price := decimal.NewFromInt(int64(100 + i%10))
		order := &Order{
			ID:       fmt.Sprintf("sell-%d", i),
			Side:     OrderSideSell,
			Price:    price,
			Quantity: decimal.NewFromInt(1),
		}
		ob.Add(order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buyOrder := &Order{
			ID:       fmt.Sprintf("buy-%d", i),
			Side:     OrderSideBuy,
			Price:    decimal.NewFromInt(105),
			Quantity: decimal.NewFromInt(1),
		}
		trades := ob.Match(buyOrder)
		ReleaseTrades(trades)
	}
}

func BenchmarkOrderBook_Add(b *testing.B) {
	ob := NewOrderBook()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order := &Order{
			ID:       fmt.Sprintf("order-%d", i),
			Side:     OrderSideBuy,
			Price:    decimal.NewFromInt(int64(100 + i%10)),
			Quantity: decimal.NewFromInt(1),
		}
		ob.Add(order)
	}
}

func BenchmarkOrderBook_AddScenarios(b *testing.B) {
	scenarios := []struct {
		name string
		fn   func(ob *OrderBook, i int) *Order
	}{
		{
			name: "AppendToEnd",
			fn: func(ob *OrderBook, i int) *Order {
				return &Order{
					ID:       fmt.Sprintf("order-%d", i),
					Side:     OrderSideBuy,
					Price:    decimal.NewFromInt(int64(90 + i%5)),
					Quantity: decimal.NewFromInt(1),
				}
			},
		},
		{
			name: "InsertAtBeginning",
			fn: func(ob *OrderBook, i int) *Order {
				return &Order{
					ID:       fmt.Sprintf("order-%d", i),
					Side:     OrderSideBuy,
					Price:    decimal.NewFromInt(int64(110 - i%5)),
					Quantity: decimal.NewFromInt(1),
				}
			},
		},
		{
			name: "InsertInMiddle",
			fn: func(ob *OrderBook, i int) *Order {
				return &Order{
					ID:       fmt.Sprintf("order-%d", i),
					Side:     OrderSideBuy,
					Price:    decimal.NewFromInt(int64(100 + (i%3 - 1))),
					Quantity: decimal.NewFromInt(1),
				}
			},
		},
	}

	for _, s := range scenarios {
		b.Run(s.name, func(b *testing.B) {
			ob := NewOrderBook()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				order := s.fn(ob, i)
				ob.Add(order)
			}
		})
	}
}

// Package strategies contains all strategies supported by machina
package strategies

import (
	"fmt"
	"time"

	v1 "github.com/lgarciaaco/machina-api/business/strategies/api/v1"

	"github.com/lgarciaaco/machina-api/business/strategies/financial"

	"github.com/lgarciaaco/machina-api/business/broker"
)

// Candle represents an individual candle
type Candle struct {
	ID         string    `json:"id"`
	SymbolID   string    `json:"-"`
	Symbol     string    `json:"symbol"`
	Interval   string    `json:"interval"`
	OpenTime   time.Time `json:"open_time"`
	OpenPrice  float64   `json:"open_price"`
	ClosePrice float64   `json:"close_price"`
	CloseTime  time.Time `json:"close_time"`
	Low        float64   `json:"low"`
	High       float64   `json:"high"`
	Volume     float64   `json:"volume"`
}

func toFinancialCandle(cdl Candle) financial.Candle {
	fc := (*financial.Candle)(&cdl)
	return *fc
}

func toCandle(v1Cdl v1.Candle) Candle {
	fc := (*Candle)(&v1Cdl)
	return *fc
}

func toCandleSlice(v1Cdls []v1.Candle) []Candle {
	cdls := make([]Candle, len(v1Cdls))
	for i, dbCdl := range v1Cdls {
		cdls[i] = toCandle(dbCdl)
	}
	return cdls
}

// Order represent an order in a position
type Order struct {
	ID           string    `json:"order_id"`
	SymbolID     string    `json:"symbol_id"`
	PositionID   string    `json:"position_id"`
	CreationTime time.Time `json:"creation_time"`
	Price        float64   `json:"price"`
	Quantity     float64   `json:"quantity"`
	Status       string    `json:"status"`
	Type         string    `json:"type"`
	Side         string    `json:"side"`
}

func toOrder(v1Odr v1.Order) Order {
	o := (*Order)(&v1Odr)
	return *o
}

func toOrderSlice(v1Odrs []v1.Order) []Order {
	odrs := make([]Order, len(v1Odrs))
	for i, v1Odr := range v1Odrs {
		odrs[i] = toOrder(v1Odr)
	}
	return odrs
}

// Position represents a single position
type Position struct {
	ID           string    `json:"position_id"`   // Position ID
	SymbolID     string    `json:"-"`             // SymbolID this position is trading on, used to preload Symbol
	UserID       string    `json:"-"`             // UserID who created this position, used to preload User
	Side         string    `json:"side"`          // Position side: SELL / BUY
	Status       string    `json:"status"`        // Status open / closed
	CreationTime time.Time `json:"creation_time"` // CreationTime of the position
	User         string    `json:"user"`          // Name of the owner
	Symbol       string    `json:"symbol"`        // Symbol this position is trading on
	Orders       []Order   `json:"orders"`        // Orders belonging to this position
}

// close return the order to close a position.
func (p *Position) close() (Order, error) {
	if len(p.Orders) != 1 {
		return Order{}, fmt.Errorf("wrong orders.len, need 1 but got %d", len(p.Orders))
	}
	rOdr := p.Orders[0]

	rOdr.Side = financial.SideBuy
	if p.Side == financial.SideBuy {
		rOdr.Side = financial.SideSell
	}

	return rOdr, nil
}

func toPosition(v1Pos *v1.Position) *Position {
	return &Position{
		ID:           v1Pos.ID,
		SymbolID:     v1Pos.SymbolID,
		UserID:       v1Pos.UserID,
		Side:         v1Pos.Side,
		Status:       v1Pos.Status,
		CreationTime: v1Pos.CreationTime,
		User:         v1Pos.User,
		Symbol:       v1Pos.Symbol,
		Orders:       toOrderSlice(v1Pos.Orders),
	}
}

func (p Position) Profit() float64 {
	var profit float64
	if p.Status == "closed" && len(p.Orders) == 2 {
		if p.Side == broker.OrderSideBuy {
			profit = (p.Orders[1].Price * p.Orders[1].Quantity) - (p.Orders[0].Price * p.Orders[0].Quantity)
		}

		if p.Side == broker.OrderSideSell {
			profit = (p.Orders[0].Price * p.Orders[0].Quantity) - (p.Orders[1].Price * p.Orders[1].Quantity)
		}
	}

	return profit
}

// Package financial holds all the logic to operate on assets
package financial

import (
	"fmt"
)

const (
	SideSell = "SELL"
	SideBuy  = "BUY"
)

// Budget is the interface that manages the budget granted to a strategy.
// A strategy is granted an initial amount of a base coin and alt coin, with that
// the strategy starts working and in in the event of positive trades, these amounts should
// increase over time. If the strategy is plain trash and looses money, at some point the
// budget is exhausted and the strategy stops
type Budget interface {
	// Close prepares the position based on what the BaseBudget determines so it is executed
	// (closed) further down the line
	Close(o Position, c Candle) error

	// Open prepares the position based on what the BaseBudget determines so it is executed
	// (opened) further down the line
	Open(o Position, c Candle) (float64, error)
}

// Position is a simplified trading position. For the budget purpose, only status and
// side are relevant
type Position struct {
	Side   string
	Status string
}

// BaseBudget defines the data type that manages how much is expensed
// when opening or closing a position. It prepares the position
// so it can be further processed by the client
type BaseBudget struct {
	// Base is the base coin we are trading in, ETH, BTC
	Base float64

	// Alt is the alt coin we are trading in, USDT, BUST
	Alt float64

	// Lot is amount of Base that is used in every order
	Lot float64
}

// FixBudget will always work on a fix amount. Lets say we set .CalculateLot = 0.2,
// every time we Open a buy position we buy 0.2 Base Coin at Close price, and every time
// we sell, we sell 0.2 at Close price
type FixBudget struct {
	BaseBudget
}

// RatioBudget strategy operates on a percentage of the current Base coin. For instance, if
// we start trading with .CalculateLot = 0.6 and .Base = 2, starting a sell order would sell 1.2 and starting a buy
// order would buy 1.2. In the other hand, closing the order will attempt to sell or buy what was targeted in the order
type RatioBudget struct {
	BaseBudget
}

// empty: when base coin or tether are close to 0
func (b BaseBudget) empty() bool {
	return b.Base <= 0.001 || b.Alt <= 1
}

// Open will setup the position and adjust Budget and order details. It returns
// the quantity the position should use to open an order
func (b *FixBudget) Open(p Position, c Candle) (float64, error) {
	// When opening a sell order, we deduct .CalculateLot from our Base Coin Budgeter
	if p.Side == SideSell {
		b.Base -= b.Lot
		b.Alt += b.Lot * c.ClosePrice
	}

	if p.Side == SideBuy {
		b.Base += b.Lot
		b.Alt -= b.Lot * c.ClosePrice
	}

	if b.empty() {
		return 0.0, fmt.Errorf("not enough in fund to buy. In fund: B[%v] T[%v]", b.Base, b.Alt)
	}

	return b.Lot, nil
}

// Close will close an order and adjust Budgeter and order details
func (b *FixBudget) Close(p Position, c Candle) error {
	if p.Status == "closed" {
		return fmt.Errorf("can't close an order that is already closed")
	}

	if p.Side == SideSell {
		b.Base += b.Lot
		b.Alt -= b.Lot * c.ClosePrice
	}

	if p.Side == SideBuy {
		b.Base -= b.Lot
		b.Alt += b.Lot * c.ClosePrice
	}

	if b.empty() {
		return fmt.Errorf("not enough in fund to buy. In fund: B[%v] T[%v]", b.Base, b.Alt)
	}

	return nil
}

func (b *FixBudget) String() string {
	return fmt.Sprintf("[lot: %f, base: %f, alt: %f]", b.Lot, b.Base, b.Alt)
}

func (b *RatioBudget) Open(p *Position, c Candle) error {
	openTether := c.ClosePrice * (b.Lot * b.Base)
	openBase := b.Lot * b.Base

	if p.Side == SideSell {
		b.Base -= openBase
		b.Alt += openTether
	}

	if p.Side == SideBuy {
		b.Base += openBase
		b.Alt -= openTether
	}

	if b.empty() {
		return fmt.Errorf("not enough in fund to buy. In fund: B[%v] T[%v]", b.Base, b.Alt)
	}

	return nil
}

package db

import "time"

// Order defines a trading order
type Order struct {
	ID           string    `db:"order_id"`      // Order ID
	SymbolID     string    `db:"symbol_id"`     // Symbol ID, this orders trades on
	PositionID   string    `db:"position_id"`   // Order ID this order belongs to
	CreationTime time.Time `db:"creation_time"` // Order creation time
	Price        float64   `db:"price"`         // Price of the base asset
	Quantity     float64   `db:"quantity"`      // Amount of the base asset
	Status       string    `db:"status"`        // Status received from binance: FILLED, NEW, CANCELED
	Type         string    `db:"type"`          // For the moment only MARKET is supported
	Side         string    `db:"side"`          // Either SELL or BUY
}

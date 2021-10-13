package db

import "time"

// Order defines a trading order
type Order struct {
	ID           string    `db:"order_id"`
	SymbolID     string    `db:"symbol_id"`
	PositionID   string    `db:"position_id"`
	CreationTime time.Time `db:"creation_time"`
	Price        float64   `db:"price"`
	Quantity     float64   `db:"quantity"`
	Status       string    `db:"status"`
	Type         string    `db:"type"`
	Side         string    `db:"side"`
}

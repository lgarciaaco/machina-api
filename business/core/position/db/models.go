package db

import (
	"time"
)

// Position is the amount of a security, asset, or property that is owned (or sold short)
// by some individual or other entity. A trader or investor takes a position when they make a purchase through
type Position struct {
	ID           string    `db:"position_id"`   // Position ID
	SymbolID     string    `db:"symbol_id"`     // Symbol this position is trading on
	UserID       string    `db:"user_id"`       // User who created this position
	Side         string    `db:"side"`          // Position side: SELL / BUY
	Status       string    `db:"status"`        // Status open / closed
	CreationTime time.Time `db:"creation_time"` // CreationTime of the position
	User         string    `db:"user"`
	Symbol       string    `db:"symbol"`
	Orders       string    `db:"orders"`
}

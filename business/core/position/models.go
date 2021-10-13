package position

import (
	"time"

	"github.com/lgarciaaco/machina-api/business/core/position/db"
)

// Position represents a single position
type Position struct {
	ID           string    `json:"position_id"`   // Position ID
	SymbolID     string    `json:"-"`             // SymbolID this position is trading on, used to preload Symbol
	UserID       string    `json:"-"`             // UserID who created this position, used to preload User
	Side         string    `json:"side"`          // Position side: SELL / BUY
	Status       string    `json:"status"`        // Status open / closed
	CreationTime time.Time `json:"creation_time"` // CreationTime of the position
	User         string    `json:"user"`
	Symbol       string    `json:"symbol"`
}

func toPosition(dbPos db.Position) Position {
	pc := (*Position)(&dbPos)
	return *pc
}

func toPositionSlice(dbPoss []db.Position) []Position {
	poss := make([]Position, len(dbPoss))
	for i, dbPos := range dbPoss {
		poss[i] = toPosition(dbPos)
	}
	return poss
}

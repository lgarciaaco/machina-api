package position

import (
	"encoding/json"
	"fmt"
	"strings"
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
	User         string    `json:"user"`          // Name of the owner
	Symbol       string    `json:"symbol"`        // Symbol this position is trading on
	Orders       []Order   `json:"orders"`        // Orders belonging to this position
}

// Order represent an order in a position
type Order struct {
	ID           string    `json:"order_id"`
	SymbolID     string    `json:"symbol_id"`
	PositionID   string    `json:"position_id"`
	CreationTime orderTime `json:"creation_time"`
	Price        float64   `json:"price"`
	Quantity     float64   `json:"quantity"`
	Status       string    `json:"status"`
	Type         string    `json:"type"`
	Side         string    `json:"side"`
}

// orderTime is a custom time implementation to be able to Unmarshal psql
// time format
type orderTime time.Time

func toPosition(dbPos db.Position) Position {
	var ords []Order
	if dbPos.Orders != "" {
		if err := json.Unmarshal([]byte(dbPos.Orders), &ords); err != nil {
			return Position{}
		}
	}

	return Position{
		ID:           dbPos.ID,
		SymbolID:     dbPos.SymbolID,
		UserID:       dbPos.UserID,
		Side:         dbPos.Side,
		Status:       dbPos.Side,
		CreationTime: dbPos.CreationTime,
		User:         dbPos.User,
		Symbol:       dbPos.Symbol,
		Orders:       ords,
	}
}

func toPositionSlice(dbPoss []db.Position) []Position {
	poss := make([]Position, len(dbPoss))
	for i, dbPos := range dbPoss {
		poss[i] = toPosition(dbPos)
	}
	return poss
}

func (t orderTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339Nano))
	return []byte(stamp), nil
}

func (j *orderTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02T15:04:05.999999", s)
	if err != nil {
		return err
	}

	*j = orderTime(t)
	return nil
}

package order

import (
	"time"

	"github.com/lgarciaaco/machina-api/business/core/order/db"
)

// Order represents an individual order
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

func toOrder(dbOdr db.Order) Order {
	pc := (*Order)(&dbOdr)
	return *pc
}

func toOrderSlice(dbOdrs []db.Order) []Order {
	cdls := make([]Order, len(dbOdrs))
	for i, dbCdl := range dbOdrs {
		cdls[i] = toOrder(dbCdl)
	}
	return cdls
}

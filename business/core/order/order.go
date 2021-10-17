// Package order provides middleware utilities around orders.
// It wraps calls to the database and to the binance endpoints.
package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lgarciaaco/machina-api/business/core/order/binance"

	"github.com/lgarciaaco/machina-api/business/broker"

	"github.com/jmoiron/sqlx"
	"github.com/lgarciaaco/machina-api/business/core/order/db"
	"github.com/lgarciaaco/machina-api/business/sys/database"
	"github.com/lgarciaaco/machina-api/business/sys/validate"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("order not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrInvalidID             = errors.New("ID is not in its proper form")
)

// Core manages the set of API's for candle access.
type Core struct {
	dbAgent  db.Agent
	bkrAgent binance.Agent
}

// NewCore constructs a core for user api access.
func NewCore(log *zap.SugaredLogger, sqlxDB *sqlx.DB, brk broker.Broker) Core {
	return Core{
		dbAgent:  db.NewAgent(log, sqlxDB),
		bkrAgent: binance.NewAgent(log, brk),
	}
}

// Create inserts a new order into the database.
func (c Core) Create(ctx context.Context, nOdr NewOrder, now time.Time) (Order, error) {
	if err := validate.Check(nOdr); err != nil {
		return Order{}, fmt.Errorf("validating data: %w", err)
	}

	// Create order with the broker
	bkrOdr := binance.Order{
		Symbol:   nOdr.Symbol,
		Side:     nOdr.Side,
		Type:     broker.OrderTypeMarket,
		Quantity: nOdr.Quantity,
	}
	or, err := c.bkrAgent.Create(ctx, bkrOdr)
	if err != nil {
		return Order{}, fmt.Errorf("create: %w", err)
	}

	dbOdr := db.Order{
		ID:           validate.GenerateID(),
		SymbolID:     nOdr.SymbolID,
		PositionID:   nOdr.PositionID,
		CreationTime: now,
		Price:        or.Price,
		Quantity:     nOdr.Quantity,
		Status:       or.Status,
		Type:         broker.OrderTypeMarket,
		Side:         nOdr.Side,
	}

	if err := c.dbAgent.Create(ctx, dbOdr); err != nil {
		return Order{}, fmt.Errorf("create: %w", err)
	}

	return toOrder(dbOdr), nil
}

// QueryByID gets the specified order from the database.
func (c Core) QueryByID(ctx context.Context, odrID string) (Order, error) {
	if err := validate.CheckID(odrID); err != nil {
		return Order{}, ErrInvalidID
	}

	odr, err := c.dbAgent.QueryByID(ctx, odrID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Order{}, ErrNotFound
		}
		return Order{}, fmt.Errorf("query: %w", err)
	}

	return toOrder(odr), nil
}

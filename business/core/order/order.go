package order

import (
	"context"
	"errors"
	"fmt"

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
	agent db.Agent
}

// NewCore constructs a core for user api access.
func NewCore(log *zap.SugaredLogger, sqlxDB *sqlx.DB) Core {
	return Core{
		agent: db.NewAgent(log, sqlxDB),
	}
}

// Create inserts a new order into the database.
func (c Core) Create(ctx context.Context, nOdr Order) (Order, error) {
	if err := validate.Check(nOdr); err != nil {
		return Order{}, fmt.Errorf("validating data: %w", err)
	}

	dbOdr := db.Order{
		ID:           validate.GenerateID(),
		SymbolID:     nOdr.SymbolID,
		PositionID:   nOdr.PositionID,
		CreationTime: nOdr.CreationTime,
		Price:        nOdr.Price,
		Quantity:     nOdr.Quantity,
		Status:       nOdr.Status,
		Type:         nOdr.Type,
		Side:         nOdr.Side,
	}

	if err := c.agent.Create(ctx, dbOdr); err != nil {
		return Order{}, fmt.Errorf("create: %w", err)
	}

	return toOrder(dbOdr), nil
}

// QueryByID gets the specified order from the database.
func (c Core) QueryByID(ctx context.Context, odrId string) (Order, error) {
	if err := validate.CheckID(odrId); err != nil {
		return Order{}, ErrInvalidID
	}

	odr, err := c.agent.QueryByID(ctx, odrId)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Order{}, ErrNotFound
		}
		return Order{}, fmt.Errorf("query: %w", err)
	}

	return toOrder(odr), nil
}

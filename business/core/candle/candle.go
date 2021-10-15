// Package candle provides middleware utilities around candles.
// It wraps calls to the database.
package candle

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lgarciaaco/machina-api/business/core/candle/db"
	"github.com/lgarciaaco/machina-api/business/sys/database"
	"github.com/lgarciaaco/machina-api/business/sys/validate"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("candle not found")
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

// Create inserts a new candle into the database.
func (c Core) Create(ctx context.Context, nCdl Candle) (Candle, error) {
	if err := validate.Check(nCdl); err != nil {
		return Candle{}, fmt.Errorf("validating data: %w", err)
	}

	dbCdl := db.Candle{
		ID:         validate.GenerateID(),
		Symbol:     nCdl.Symbol,
		Interval:   nCdl.Interval,
		OpenTime:   nCdl.OpenTime,
		OpenPrice:  nCdl.OpenPrice,
		CloseTime:  nCdl.CloseTime,
		ClosePrice: nCdl.ClosePrice,
		High:       nCdl.High,
		Low:        nCdl.Low,
		Volume:     nCdl.Volume,
	}

	if err := c.agent.Create(ctx, dbCdl); err != nil {
		return Candle{}, fmt.Errorf("create: %w", err)
	}

	return toCandle(dbCdl), nil
}

// QueryByID gets the specified candle from the database.
func (c Core) QueryByID(ctx context.Context, userID string) (Candle, error) {
	if err := validate.CheckID(userID); err != nil {
		return Candle{}, ErrInvalidID
	}

	cdl, err := c.agent.QueryByID(ctx, userID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Candle{}, ErrNotFound
		}
		return Candle{}, fmt.Errorf("query: %w", err)
	}

	return toCandle(cdl), nil
}

// Query gets the specified candle from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Candle, error) {
	dbCdl, err := c.agent.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toCandleSlice(dbCdl), nil
}

// QueryBySymbolAndInterval gets the specified candle from the database.
func (c Core) QueryBySymbolAndInterval(ctx context.Context, pageNumber int, rowsPerPage int, cSmb string, cItv string) ([]Candle, error) {
	dbCdl, err := c.agent.QueryBySymbolAndInterval(ctx, pageNumber, rowsPerPage, cSmb, cItv)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toCandleSlice(dbCdl), nil
}

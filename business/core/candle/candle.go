// Package candle provides middleware utilities around candles.
// It wraps calls to the database.
package candle

import (
	"context"
	"errors"
	"fmt"

	"github.com/lgarciaaco/machina-api/business/core/candle/binance"

	"github.com/lgarciaaco/machina-api/business/broker"

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
	ErrInvalidCandle         = errors.New("candle is not valid")
)

// Core manages the set of API's for candle access.
type Core struct {
	dbAgent  db.Agent
	bkrAgent binance.Agent
}

// NewCore constructs a core for user api access.
func NewCore(log *zap.SugaredLogger, sqlxDB *sqlx.DB, broker broker.Broker) Core {
	return Core{
		dbAgent:  db.NewAgent(log, sqlxDB),
		bkrAgent: binance.NewAgent(log, broker),
	}
}

// Create inserts a new candle into the database.
func (c Core) Create(ctx context.Context, nCdl NewCandle) (Candle, error) {
	if err := validate.Check(nCdl); err != nil {
		return Candle{}, fmt.Errorf("validating data: %w", err)
	}

	// Fetch candle from binance api
	bkrCdls, err := c.bkrAgent.QueryBySymbolAndInterval(ctx, nCdl.Symbol, nCdl.Interval, 1)
	if err != nil {
		return Candle{}, ErrInvalidCandle
	}

	// Insert candle into the database
	dbCdl := *(*db.Candle)(&bkrCdls[0])
	dbCdl.ID = validate.GenerateID()
	dbCdl.SymbolID = nCdl.SymbolID

	if err := c.dbAgent.Create(ctx, dbCdl); err != nil {
		return Candle{}, fmt.Errorf("create candle in database: %w", err)
	}

	return toCandle(dbCdl), nil
}

// QueryByID gets the specified candle from the database.
func (c Core) QueryByID(ctx context.Context, cdlID string) (Candle, error) {
	if err := validate.CheckID(cdlID); err != nil {
		return Candle{}, ErrInvalidID
	}

	cdl, err := c.dbAgent.QueryByID(ctx, cdlID)
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
	dbCdl, err := c.dbAgent.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toCandleSlice(dbCdl), nil
}

// QueryBySymbolAndInterval gets the specified candle from the database.
func (c Core) QueryBySymbolAndInterval(ctx context.Context, pageNumber int, rowsPerPage int, sblID string, cItv string) ([]Candle, error) {
	if err := validate.CheckID(sblID); err != nil {
		return []Candle{}, ErrInvalidID
	}

	dbCdl, err := c.dbAgent.QueryBySymbolAndInterval(ctx, pageNumber, rowsPerPage, sblID, cItv)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toCandleSlice(dbCdl), nil
}

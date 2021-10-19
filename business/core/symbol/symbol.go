// Package symbol provides middleware utilities around symbol.
//
// Symbol is a function whereby you have two different currencies that can be traded between one another.
// When buying and selling a cryptocurrency, it is often swapped with local currency. For example,
// If you're looking to buy or sell Bitcoin with U.S. Dollar, the trading pair would be BTC to USD
package symbol

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lgarciaaco/machina-api/business/broker"
	"github.com/lgarciaaco/machina-api/business/core/symbol/binance"
	"github.com/lgarciaaco/machina-api/business/core/symbol/db"
	"github.com/lgarciaaco/machina-api/business/sys/database"
	"github.com/lgarciaaco/machina-api/business/sys/validate"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("symbol not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrInvalidSymbol         = errors.New("symbol is not valid")
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

// Create fetch a symbol from the binance api and
// inserts it into the database.
func (c Core) Create(ctx context.Context, nSbl NewSymbol) (Symbol, error) {
	if err := validate.Check(nSbl); err != nil {
		return Symbol{}, fmt.Errorf("validating data: %w", err)
	}

	// Fetch symbol from binance
	bkrSbl, err := c.bkrAgent.QueryBySymbol(ctx, nSbl.Symbol)
	if err != nil {
		return Symbol{}, ErrInvalidSymbol
	}

	// Insert symbol into database
	dbSbl := *(*db.Symbol)(&bkrSbl)
	dbSbl.ID = validate.GenerateID()
	if err := c.dbAgent.Create(ctx, dbSbl); err != nil {
		return Symbol{}, fmt.Errorf("create symbol in database %w", err)
	}

	return toSymbol(dbSbl), nil
}

// Query retrieves a list of existing symbols from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Symbol, error) {
	dbSbls, err := c.dbAgent.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toSymbolSlice(dbSbls), nil
}

// QueryByID gets the specified symbol from the database.
func (c Core) QueryByID(ctx context.Context, sblID string) (Symbol, error) {
	if err := validate.CheckID(sblID); err != nil {
		return Symbol{}, ErrInvalidID
	}

	cdl, err := c.dbAgent.QueryByID(ctx, sblID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Symbol{}, ErrNotFound
		}
		return Symbol{}, fmt.Errorf("query: %w", err)
	}

	return toSymbol(cdl), nil
}

// QueryBySymbol gets the specified symbol from the database.
func (c Core) QueryBySymbol(ctx context.Context, sSbl string) (Symbol, error) {
	sbl, err := c.dbAgent.QueryBySymbol(ctx, sSbl)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Symbol{}, ErrNotFound
		}
		return Symbol{}, fmt.Errorf("query: %w", err)
	}

	return toSymbol(sbl), nil
}

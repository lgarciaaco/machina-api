package symbol

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
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

// Create inserts a new symbol into the database.
func (c Core) Create(ctx context.Context, nSbl Symbol) (Symbol, error) {
	if err := validate.Check(nSbl); err != nil {
		return Symbol{}, fmt.Errorf("validating data: %w", err)
	}

	dbSbl := db.Symbol{
		ID:                         validate.GenerateID(),
		Symbol:                     nSbl.Symbol,
		Status:                     nSbl.Status,
		BaseAsset:                  nSbl.BaseAsset,
		BaseAssetPrecision:         nSbl.BaseAssetPrecision,
		QuoteAsset:                 nSbl.QuoteAsset,
		QuotePrecision:             nSbl.QuotePrecision,
		BaseCommissionPrecision:    nSbl.BaseCommissionPrecision,
		QuoteCommissionPrecision:   nSbl.BaseCommissionPrecision,
		IcebergAllowed:             nSbl.IcebergAllowed,
		OcoAllowed:                 nSbl.OcoAllowed,
		QuoteOrderQtyMarketAllowed: nSbl.QuoteOrderQtyMarketAllowed,
		IsSpotTradingAllowed:       nSbl.IsSpotTradingAllowed,
		IsMarginTradingAllowed:     nSbl.IsMarginTradingAllowed,
	}

	if err := c.agent.Create(ctx, dbSbl); err != nil {
		return Symbol{}, fmt.Errorf("create: %w", err)
	}

	return toSymbol(dbSbl), nil
}

// Query retrieves a list of existing symbols from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Symbol, error) {
	dbSbls, err := c.agent.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toSymbolSlice(dbSbls), nil
}

// QueryByID gets the specified symbol from the database.
func (c Core) QueryByID(ctx context.Context, sblId string) (Symbol, error) {
	if err := validate.CheckID(sblId); err != nil {
		return Symbol{}, ErrInvalidID
	}

	cdl, err := c.agent.QueryByID(ctx, sblId)
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
	sbl, err := c.agent.QueryBySymbol(ctx, sSbl)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Symbol{}, ErrNotFound
		}
		return Symbol{}, fmt.Errorf("query: %w", err)
	}

	return toSymbol(sbl), nil
}

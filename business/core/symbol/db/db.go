package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lgarciaaco/machina-api/business/sys/database"
	"go.uber.org/zap"
)

// Agent manages the set of API's for candle access.
type Agent struct {
	log    *zap.SugaredLogger
	sqlxDB *sqlx.DB
}

// NewAgent constructs a data for api access.
func NewAgent(log *zap.SugaredLogger, sqlxDB *sqlx.DB) Agent {
	return Agent{
		log:    log,
		sqlxDB: sqlxDB,
	}
}

// Create inserts a new symbol into the database.
func (s Agent) Create(ctx context.Context, sbl Symbol) error {
	const q = `
	INSERT INTO symbols
		(symbol_id, symbol, status, base_asset, base_asset_precision, quote_asset, quote_precision, base_commission_precision, 
		 quote_commission_precision, iceberg_allowed, oco_allowed, quote_order_qty_market_allowed, is_spot_trading_allowed, is_margin_trading_allowed)
	VALUES
		(:symbol_id, :symbol, :status, :base_asset, :base_asset_precision, :quote_asset, :quote_precision, :base_commission_precision, 
		 :quote_commission_precision, :iceberg_allowed, :oco_allowed, :quote_order_qty_market_allowed, :is_spot_trading_allowed, :is_margin_trading_allowed)`

	if err := database.NamedExecContext(ctx, s.log, s.sqlxDB, q, sbl); err != nil {
		return fmt.Errorf("inserting symbol: %w", err)
	}

	return nil
}

// Query retrieves a list of existing candles from the database.
func (s Agent) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Symbol, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		symbols
	ORDER BY
		symbol
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var sbls []Symbol
	if err := database.NamedQuerySlice(ctx, s.log, s.sqlxDB, q, data, &sbls); err != nil {
		return nil, fmt.Errorf("selecting symbol: %w", err)
	}

	return sbls, nil
}

// QueryBySymbol gets the specified symbols from the database.
func (s Agent) QueryBySymbol(ctx context.Context, sSbl string) (Symbol, error) {
	data := struct {
		Symbol string `db:"symbol"`
	}{
		Symbol: sSbl,
	}

	const q = `
	SELECT
		*
	FROM
		symbols
	WHERE 
		symbol = :symbol`

	var sbl Symbol
	if err := database.NamedQueryStruct(ctx, s.log, s.sqlxDB, q, data, &sbl); err != nil {
		return Symbol{}, fmt.Errorf("selecting symbols [%q]: %w", sSbl, err)
	}

	return sbl, nil
}

// QueryByID gets the specified symbol from the database.
func (s Agent) QueryByID(ctx context.Context, sblId string) (Symbol, error) {
	data := struct {
		SymbolID string `db:"symbol_id"`
	}{
		SymbolID: sblId,
	}

	const q = `
	SELECT
		*
	FROM
		symbols
	WHERE 
		symbol_id = :symbol_id`

	var sbl Symbol
	if err := database.NamedQueryStruct(ctx, s.log, s.sqlxDB, q, data, &sbl); err != nil {
		return Symbol{}, fmt.Errorf("selecting sblId[%q]: %w", sblId, err)
	}

	return sbl, nil
}

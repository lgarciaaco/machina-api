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

// Create inserts a new candle into the database.
func (s Agent) Create(ctx context.Context, cdl Candle) error {
	const q = `
	INSERT INTO candles
		(candle_id, symbol, interval, open_time, open_price, close_time, close_price, low, high, volume)
	VALUES
		(:candle_id, :symbol, :interval, :open_time, :open_price, :close_time, :close_price, :low, :high, :volume)`

	if err := database.NamedExecContext(ctx, s.log, s.sqlxDB, q, cdl); err != nil {
		return fmt.Errorf("inserting candle: %w", err)
	}

	return nil
}

// Query retrieves a list of existing candles from the database.
func (s Agent) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Candle, error) {
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
		candles
	ORDER BY
		close_time DESC
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var cdls []Candle
	if err := database.NamedQuerySlice(ctx, s.log, s.sqlxDB, q, data, &cdls); err != nil {
		return nil, fmt.Errorf("selecting candle: %w", err)
	}

	return cdls, nil
}

// QueryBySymbolAndInterval gets the specified candles from the database.
func (s Agent) QueryBySymbolAndInterval(ctx context.Context, pageNumber int, rowsPerPage int, smb string, itv string) ([]Candle, error) {
	data := struct {
		Symbol      string `db:"symbol"`
		Interval    string `db:"interval"`
		Offset      int    `db:"offset"`
		RowsPerPage int    `db:"rows_per_page"`
	}{
		Symbol:      smb,
		Interval:    itv,
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		candles
	WHERE 
		interval = :interval AND symbol = :symbol
	ORDER BY
		close_time
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var cdls []Candle
	if err := database.NamedQuerySlice(ctx, s.log, s.sqlxDB, q, data, &cdls); err != nil {
		return nil, fmt.Errorf("selecting candles [%q]: %w", smb, err)
	}

	return cdls, nil
}

// QueryByID gets the specified candle from the database.
func (s Agent) QueryByID(ctx context.Context, cdlID string) (Candle, error) {
	data := struct {
		CandleID string `db:"candle_id"`
	}{
		CandleID: cdlID,
	}

	const q = `
	SELECT
		*
	FROM
		candles
	WHERE 
		candle_id = :candle_id`

	var cdl Candle
	if err := database.NamedQueryStruct(ctx, s.log, s.sqlxDB, q, data, &cdl); err != nil {
		return Candle{}, fmt.Errorf("selecting cdlID[%q]: %w", cdlID, err)
	}

	return cdl, nil
}

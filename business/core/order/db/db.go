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
	log *zap.SugaredLogger
	tr  database.Transactor
	db  sqlx.ExtContext
}

// NewAgent constructs a data for api access.
func NewAgent(log *zap.SugaredLogger, db *sqlx.DB) Agent {
	return Agent{
		log: log,
		tr:  db,
		db:  db,
	}
}

// Create inserts a new order into the database.
func (s Agent) Create(ctx context.Context, odr Order) error {
	const q = `
	INSERT INTO orders
		(order_id, symbol_id, position_id, price, quantity, status, type, side, creation_time)
	VALUES
		(:order_id, :symbol_id, :position_id, :price, :quantity, :status, :type, :side, :creation_time)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, odr); err != nil {
		return fmt.Errorf("inserting order: %w", err)
	}

	return nil
}

// Query retrieves a list of existing orders from the database.
func (s Agent) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Order, error) {
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
		orders
	ORDER BY
		creation_time
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var odrs []Order
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &odrs); err != nil {
		return nil, fmt.Errorf("selecting orders: %w", err)
	}

	return odrs, nil
}

// QueryByID gets the specified order from the database.
func (s Agent) QueryByID(ctx context.Context, odrId string) (Order, error) {
	data := struct {
		OrderID string `db:"order_id"`
	}{
		OrderID: odrId,
	}

	const q = `
	SELECT
		*
	FROM
		orders
	WHERE 
		order_id = :order_id`

	var odr Order
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &odr); err != nil {
		return Order{}, fmt.Errorf("selecting odrId[%q]: %w", odrId, err)
	}

	return odr, nil
}

// QueryByPosition retrieves all order for a given position.
func (s Agent) QueryByPosition(ctx context.Context, odrId string) ([]Order, error) {
	data := struct {
		PositionID string `db:"position_id"`
	}{
		PositionID: odrId,
	}

	const q = `
	SELECT
		*
	FROM
		orders
	WHERE 
		position_id = :position_id
	ORDER BY
		creation_time`

	var ords []Order
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &ords); err != nil {
		return []Order{}, fmt.Errorf("selecting odrId[%q]: %w", odrId, err)
	}

	return ords, nil
}

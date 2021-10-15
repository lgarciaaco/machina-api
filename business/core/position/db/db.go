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

// Create inserts a new position into the database.
func (s Agent) Create(ctx context.Context, pos Position) error {
	const q = `
	INSERT INTO positions
		(position_id, symbol_id, user_id, creation_time, side, status)
	VALUES
		(:position_id, :symbol_id, :user_id, :creation_time, :side, :status)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, pos); err != nil {
		return fmt.Errorf("inserting position: %w", err)
	}

	return nil
}

// Query retrieves a list of existing positions from the database.
func (s Agent) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Position, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		p.*,
		u.name AS user,
		s.symbol AS symbol,
		json_agg(o.*) AS orders
	FROM
		positions AS p
	LEFT JOIN
		users AS u ON p.user_id = u.user_id
	LEFT JOIN
		symbols AS s ON p.symbol_id = s.symbol_id
	LEFT JOIN
		orders AS o ON p.position_id = o.position_id
	GROUP BY
		p.position_id, u.name, s.symbol
	ORDER BY
		creation_time
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var poss []Position
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &poss); err != nil {
		return nil, fmt.Errorf("selecting positions: %w", err)
	}

	return poss, nil
}

// QueryByUser gets the specified candles from the database.
func (s Agent) QueryByUser(ctx context.Context, pageNumber int, rowsPerPage int, usrId string) ([]Position, error) {
	data := struct {
		UserID      string `db:"user_id"`
		Offset      int    `db:"offset"`
		RowsPerPage int    `db:"rows_per_page"`
	}{
		UserID:      usrId,
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		p.*,
		u.name AS user,
		s.symbol AS symbol,
		json_agg(o.*) AS orders
	FROM
		positions AS p
	LEFT JOIN
		users AS u ON p.user_id = u.user_id
	LEFT JOIN
		symbols AS s ON p.symbol_id = s.symbol_id
	WHERE 
		p.user_id = :user_id
	LEFT JOIN
		orders AS o ON p.position_id = o.position_id
	GROUP BY
		p.position_id, u.name, s.symbol
	ORDER BY
		p.creation_time
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var poss []Position
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &poss); err != nil {
		return nil, fmt.Errorf("selecting positions [%q]: %w", usrId, err)
	}

	return poss, nil
}

// QueryByID gets the specified position from the database.
func (s Agent) QueryByID(ctx context.Context, posId string) (Position, error) {
	data := struct {
		PositionID string `db:"position_id"`
	}{
		PositionID: posId,
	}

	const q = `
	SELECT
		p.*,
		u.name AS user,
		s.symbol AS symbol,
		json_agg(o.*) AS orders
	FROM
		positions AS p
	LEFT JOIN
		users AS u ON p.user_id = u.user_id
	LEFT JOIN
		symbols AS s ON p.symbol_id = s.symbol_id
	LEFT JOIN
		orders AS o ON p.position_id = o.position_id
	WHERE 
		p.position_id = :position_id
	GROUP BY
		p.position_id, u.name, s.symbol`

	var pos Position
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &pos); err != nil {
		return Position{}, fmt.Errorf("selecting posId[%q]: %w", posId, err)
	}

	return pos, nil
}

// Update modifies data about a Position. It will error if the specified ID is
// invalid or does not reference an existing Position. When updating a position,
// it only makes sense to change its status from open to closed.
func (s Agent) Update(ctx context.Context, pos Position) error {
	const q = `
	UPDATE
		positions
	SET
		"status" = :status,
	WHERE
		position_id = :position_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, pos); err != nil {
		return fmt.Errorf("updating position positionID[%s]: %w", pos.ID, err)
	}

	return nil
}

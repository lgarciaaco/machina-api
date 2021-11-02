// Package db contains user related CRUD functionality.
package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lgarciaaco/machina-api/business/sys/database"
	"go.uber.org/zap"
)

// Store manages the set of API's for user access.
type Store struct {
	log          *zap.SugaredLogger
	tr           database.Transactor
	db           sqlx.ExtContext
	isWithinTran bool
}

// NewStore constructs a data for api access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) Store {
	return Store{
		log: log,
		tr:  db,
		db:  db,
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s Store) WithinTran(ctx context.Context, fn func(sqlx.ExtContext) error) error {
	if s.isWithinTran {
		return fn(s.db)
	}
	return database.WithinTran(ctx, s.log, s.tr, fn)
}

// Tran return new Store with transaction in it.
func (s Store) Tran(tx sqlx.ExtContext) Store {
	return Store{
		log:          s.log,
		tr:           s.tr,
		db:           tx,
		isWithinTran: true,
	}
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, usr User) error {
	const q = `
	INSERT INTO users
		(user_id, name, description, password_hash, roles, date_created, date_updated)
	VALUES
		(:user_id, :name, :description, :password_hash, :roles, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, usr); err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s Store) Update(ctx context.Context, usr User) error {
	const q = `
	UPDATE
		users
	SET 
		"name" = :name,
		"roles" = :roles,
		"description" = :description,
		"password_hash" = :password_hash,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, usr); err != nil {
		return fmt.Errorf("updating userID[%s]: %w", usr.ID, err)
	}

	return nil
}

// Delete removes a user from the database.
func (s Store) Delete(ctx context.Context, userID string) error {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting userID[%s]: %w", userID, err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]User, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		u.*,
		COUNT(p.position_id) AS positions_total
	FROM
		users AS u
	LEFT JOIN
		positions AS p ON p.user_id = u.user_id
	GROUP BY
		u.user_id
	ORDER BY
		u.date_created DESC
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var usrs []User
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &usrs); err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return usrs, nil
}

// QueryByID gets the specified user from the database.
func (s Store) QueryByID(ctx context.Context, userID string) (User, error) {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
	SELECT
		u.*,
		COUNT(p.position_id) AS positions_total
	FROM
		users AS u
	LEFT JOIN
		positions AS p ON p.user_id = u.user_id
	WHERE 
		u.user_id = :user_id
	GROUP BY
		u.user_id`

	var usr User
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		return User{}, fmt.Errorf("selecting userID[%q]: %w", userID, err)
	}

	return usr, nil
}

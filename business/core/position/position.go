// Package position provides middleware utilities around positions.
// It wraps calls to the database and to the binance endpoints.
package position

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lgarciaaco/machina-api/business/core/position/db"
	"github.com/lgarciaaco/machina-api/business/sys/database"
	"github.com/lgarciaaco/machina-api/business/sys/validate"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound      = errors.New("position not found")
	ErrInvalidID     = errors.New("ID is not in its proper form")
	ErrAlreadyClosed = errors.New("can't close a position that is already closed")

	CLOSED = "CLOSED"
	OPEN   = "OPEN"
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

// Create inserts a new position into the database.
func (c Core) Create(ctx context.Context, nPos NewPosition, now time.Time) (Position, error) {
	if err := validate.Check(nPos); err != nil {
		return Position{}, fmt.Errorf("validating data: %w", err)
	}

	dbPos := db.Position{
		ID:           validate.GenerateID(),
		SymbolID:     nPos.SymbolID,
		UserID:       nPos.UserID,
		Side:         nPos.Side,
		Status:       OPEN,
		CreationTime: now,
	}

	if err := c.agent.Create(ctx, dbPos); err != nil {
		return Position{}, fmt.Errorf("create: %w", err)
	}

	// Load position with user and symbol
	rPos, err := c.agent.QueryByID(ctx, dbPos.ID)
	if err != nil {
		return Position{}, ErrNotFound
	}

	return toPosition(rPos), nil
}

// Query gets the specified positions.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Position, error) {
	dbPoss, err := c.agent.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toPositionSlice(dbPoss), nil
}

// QueryByID gets the specified position from the database.
func (c Core) QueryByID(ctx context.Context, posID string) (Position, error) {
	if err := validate.CheckID(posID); err != nil {
		return Position{}, ErrInvalidID
	}

	dbPos, err := c.agent.QueryByID(ctx, posID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Position{}, ErrNotFound
		}
		return Position{}, fmt.Errorf("query: %w", err)
	}

	return toPosition(dbPos), nil
}

// QueryByUser gets the specified position from the database given a userId.
func (c Core) QueryByUser(ctx context.Context, pageNumber int, rowsPerPage int, usrID string) ([]Position, error) {
	dbPoss, err := c.agent.QueryByUser(ctx, pageNumber, rowsPerPage, usrID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toPositionSlice(dbPoss), nil
}

// Close closes a position identified by a given ID.
// Closing a position consist on figuring out the open balance and creating a position
// to set balance to 0.
func (c Core) Close(ctx context.Context, posID string) error {
	if err := validate.CheckID(posID); err != nil {
		return ErrInvalidID
	}

	dbPos, err := c.agent.QueryByID(ctx, posID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("updating product posID[%s]: %w", posID, err)
	}

	if dbPos.Status == CLOSED {
		return ErrAlreadyClosed
	}
	dbPos.Status = CLOSED

	if err := c.agent.Update(ctx, dbPos); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

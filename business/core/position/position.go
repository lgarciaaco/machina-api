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
	ErrNotFound              = errors.New("position not found")
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

// Create inserts a new position into the database.
func (c Core) Create(ctx context.Context, nPos Position) (Position, error) {
	if err := validate.Check(nPos); err != nil {
		return Position{}, fmt.Errorf("validating data: %w", err)
	}

	dbPos := db.Position{
		ID:           validate.GenerateID(),
		SymbolID:     nPos.SymbolID,
		UserID:       nPos.UserID,
		Side:         nPos.Side,
		Status:       nPos.Status,
		CreationTime: time.Now(),
	}

	if err := c.agent.Create(ctx, dbPos); err != nil {
		return Position{}, fmt.Errorf("create: %w", err)
	}

	return toPosition(dbPos), nil
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
func (c Core) QueryByID(ctx context.Context, posId string) (Position, error) {
	if err := validate.CheckID(posId); err != nil {
		return Position{}, ErrInvalidID
	}

	dbPos, err := c.agent.QueryByID(ctx, posId)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Position{}, ErrNotFound
		}
		return Position{}, fmt.Errorf("query: %w", err)
	}

	return toPosition(dbPos), nil
}

// QueryByUser gets the specified position from the database given a userId.
func (c Core) QueryByUser(ctx context.Context, pageNumber int, rowsPerPage int, usrId string) ([]Position, error) {
	dbPoss, err := c.agent.QueryByUser(ctx, pageNumber, rowsPerPage, usrId)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query: %w", err)
	}

	return toPositionSlice(dbPoss), nil
}

// Package candlegrp maintains the group of handlers for candle access.
package candlegrp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"errors"

	"github.com/lgarciaaco/machina-api/business/core/candle"
	v1Web "github.com/lgarciaaco/machina-api/business/web/v1"
	"github.com/lgarciaaco/machina-api/foundation/web"
)

// Handlers manages the set of candle endpoints.
type Handlers struct {
	Candle candle.Core
}

// Query returns a list of products with paging.
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid page format, page[%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid rows format, rows[%s]", rows), http.StatusBadRequest)
	}

	cdls, err := h.Candle.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for cdls: %w", err)
	}

	return web.Respond(ctx, w, cdls, http.StatusOK)
}

// QueryBySymbolAndInterval returns a list of candles matching with symbol
// and interval, with paging.
func (h Handlers) QueryBySymbolAndInterval(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid page format, page[%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid rows format, rows[%s]", rows), http.StatusBadRequest)
	}
	symbol := web.Param(r, "symbol")
	interval := web.Param(r, "interval")

	cdls, err := h.Candle.QueryBySymbolAndInterval(ctx, pageNumber, rowsPerPage, symbol, interval)
	if err != nil {
		return fmt.Errorf("unable to query for cdls: %w", err)
	}

	return web.Respond(ctx, w, cdls, http.StatusOK)
}

// QueryByID returns a candle by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	cdlID := web.Param(r, "id")

	usr, err := h.Candle.QueryByID(ctx, cdlID)
	if err != nil {
		switch {
		case errors.Is(err, candle.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, candle.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", cdlID, err)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

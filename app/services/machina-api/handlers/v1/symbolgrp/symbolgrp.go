// Package symbolgrp maintains the group of handlers for order access.
package symbolgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	v1Web "github.com/lgarciaaco/machina-api/business/web/v1"

	"github.com/lgarciaaco/machina-api/business/core/symbol"
	"github.com/lgarciaaco/machina-api/foundation/web"
)

// Handlers manages the set of symbol endpoints.
type Handlers struct {
	Symbol symbol.Core
}

// Create adds a new symbol to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var nSbl symbol.NewSymbol
	if err := web.Decode(r, &nSbl); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	sSbl, err := h.Symbol.Create(ctx, nSbl)
	if err != nil {
		switch {
		case errors.Is(err, symbol.ErrInvalidSymbol):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("symbol[%+v]: %w", &sSbl, err)
		}
	}

	return web.Respond(ctx, w, sSbl, http.StatusCreated)
}

// QueryByID returns a symbol by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	sblID := web.Param(r, "id")

	sbl, err := h.Symbol.QueryByID(ctx, sblID)
	if err != nil {
		switch {
		case errors.Is(err, symbol.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, symbol.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", sblID, err)
		}
	}

	return web.Respond(ctx, w, sbl, http.StatusOK)
}

// Query returns a list of symbols with paging.
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid page format [%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid rows format [%s]", rows), http.StatusBadRequest)
	}

	sbls, err := h.Symbol.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for sbls: %w", err)
	}

	return web.Respond(ctx, w, sbls, http.StatusOK)
}

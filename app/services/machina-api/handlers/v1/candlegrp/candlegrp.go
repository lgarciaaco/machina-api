// Package candlegrp maintains the group of handlers for candle access.
package candlegrp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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

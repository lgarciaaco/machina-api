// Package ordergrp maintains the group of handlers for order access.
package ordergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/lgarciaaco/machina-api/business/core/position"

	"github.com/lgarciaaco/machina-api/business/core/order"
	"github.com/lgarciaaco/machina-api/business/sys/auth"
	v1Web "github.com/lgarciaaco/machina-api/business/web/v1"
	"github.com/lgarciaaco/machina-api/foundation/web"
)

// Handlers manages the set of position endpoints.
type Handlers struct {
	Order    order.Core
	Position position.Core
}

// Create adds a new order to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	var nOdr order.NewOrder
	if err := web.Decode(r, &nOdr); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	// If it is not possible to fetch the position referenced in the document
	// we return 404
	pos, err := h.Position.QueryByID(ctx, nOdr.PositionID)
	if err != nil {
		return v1Web.NewRequestError(position.ErrNotFound, http.StatusNotFound)
	}
	nOdr.SymbolID = pos.SymbolID

	// If you are not an admin and looking to create an order for a position that doesn't belong
	// to you
	if pos.UserID != claims.Subject && !claims.Authorized(auth.RoleAdmin) {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	sOdr, err := h.Order.Create(ctx, nOdr, v.Now)
	if err != nil {
		return fmt.Errorf("orders[%+v]: %w", &sOdr, err)
	}

	return web.Respond(ctx, w, sOdr, http.StatusCreated)
}

// QueryByID returns an order by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	odrID := web.Param(r, "id")

	odr, err := h.Order.QueryByID(ctx, odrID)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, order.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", odrID, err)
		}
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	pos, err := h.Position.QueryByID(ctx, odr.PositionID)
	if err != nil {
		return fmt.Errorf("unable to fetch position[%s]", odr.PositionID)
	}
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != pos.UserID {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	return web.Respond(ctx, w, odr, http.StatusOK)
}

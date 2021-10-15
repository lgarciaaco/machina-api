package positiongrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lgarciaaco/machina-api/business/core/user"

	v1Web "github.com/lgarciaaco/machina-api/business/web/v1"

	"github.com/lgarciaaco/machina-api/business/core/position"
	"github.com/lgarciaaco/machina-api/business/sys/auth"
	"github.com/lgarciaaco/machina-api/foundation/web"
)

// Handlers manages the set of position endpoints.
type Handlers struct {
	Position position.Core
	Auth     *auth.Auth
}

// Create adds a new position to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	var nPos position.Position
	if err := web.Decode(r, &nPos); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}
	nPos.UserID = claims.Subject

	usr, err := h.Position.Create(ctx, nPos)
	if err != nil {
		return fmt.Errorf("user[%+v]: %w", &usr, err)
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

// Query returns a list of positions with paging. If an administrator is
// issuing the request, then a list with all positions is returned, otherwise
// only the positions from the logged user is returned
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

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	var poss []position.Position

	// If you are an admin you get a list with positions for all users.
	if claims.Authorized(auth.RoleAdmin) {
		poss, err = h.Position.Query(ctx, pageNumber, rowsPerPage)
		if err != nil {
			return fmt.Errorf("unable to query for products: %w", err)
		}
	} else {
		poss, err = h.Position.QueryByUser(ctx, pageNumber, rowsPerPage, claims.Subject)
		if err != nil {
			return fmt.Errorf("unable to query for products: %w", err)
		}
	}

	return web.Respond(ctx, w, poss, http.StatusOK)
}

// QueryByID returns a positions by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	userID := web.Param(r, "id")

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	usr, err := h.Position.QueryByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, user.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

// Close closes a position setting its balance to 0. Positions persist in database.
func (h Handlers) Close(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	posId := web.Param(r, "id")

	// If you are not an admin and looking to delete someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != posId {
		return v1Web.NewRequestError(auth.ErrForbidden, http.StatusForbidden)
	}

	if err := h.Position.Close(ctx, posId); err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, user.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", posId, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

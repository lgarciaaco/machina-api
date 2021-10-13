// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers/v1/candlegrp"
	"github.com/lgarciaaco/machina-api/business/core/candle"

	"github.com/jmoiron/sqlx"
	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers/v1/productgrp"
	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers/v1/usergrp"
	"github.com/lgarciaaco/machina-api/business/core/product"
	"github.com/lgarciaaco/machina-api/business/core/user"
	"github.com/lgarciaaco/machina-api/business/sys/auth"
	"github.com/lgarciaaco/machina-api/business/web/v1/mid"
	"github.com/lgarciaaco/machina-api/foundation/web"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log  *zap.SugaredLogger
	Auth *auth.Auth
	DB   *sqlx.DB
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.Auth)
	admin := mid.Authorize(auth.RoleAdmin)

	// Register user management and authentication endpoints.
	ugh := usergrp.Handlers{
		User: user.NewCore(cfg.Log, cfg.DB),
		Auth: cfg.Auth,
	}
	app.Handle(http.MethodGet, version, "/users/token", ugh.Token)
	app.Handle(http.MethodGet, version, "/users/:page/:rows", ugh.Query, authen, admin)
	app.Handle(http.MethodGet, version, "/users/:id", ugh.QueryByID, authen)
	app.Handle(http.MethodPost, version, "/users", ugh.Create, authen, admin)
	app.Handle(http.MethodPut, version, "/users/:id", ugh.Update, authen, admin)
	app.Handle(http.MethodDelete, version, "/users/:id", ugh.Delete, authen, admin)

	// Register candle endpoints
	cgh := candlegrp.Handlers{
		Candle: candle.NewCore(cfg.Log, cfg.DB),
	}
	app.Handle(http.MethodGet, version, "/candles/:page/:rows", cgh.Query)

	// Register product and sale endpoints.
	pgh := productgrp.Handlers{
		Product: product.NewCore(cfg.Log, cfg.DB),
	}
	app.Handle(http.MethodGet, version, "/products/:page/:rows", pgh.Query, authen)
	app.Handle(http.MethodGet, version, "/products/:id", pgh.QueryByID, authen)
	app.Handle(http.MethodPost, version, "/products", pgh.Create, authen)
	app.Handle(http.MethodPut, version, "/products/:id", pgh.Update, authen)
	app.Handle(http.MethodDelete, version, "/products/:id", pgh.Delete, authen)
}

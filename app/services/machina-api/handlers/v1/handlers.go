// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers/v1/symbolgrp"
	"github.com/lgarciaaco/machina-api/business/core/symbol"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers/v1/ordergrp"
	"github.com/lgarciaaco/machina-api/business/broker"
	"github.com/lgarciaaco/machina-api/business/core/order"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers/v1/positiongrp"
	"github.com/lgarciaaco/machina-api/business/core/position"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers/v1/candlegrp"
	"github.com/lgarciaaco/machina-api/business/core/candle"

	"github.com/jmoiron/sqlx"
	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers/v1/usergrp"
	"github.com/lgarciaaco/machina-api/business/core/user"
	"github.com/lgarciaaco/machina-api/business/sys/auth"
	"github.com/lgarciaaco/machina-api/business/web/v1/mid"
	"github.com/lgarciaaco/machina-api/foundation/web"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log    *zap.SugaredLogger
	Auth   *auth.Auth
	DB     *sqlx.DB
	Broker broker.Broker
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
	app.Handle(http.MethodGet, version, "/users/token", ugh.Token, mid.Cors("*"))
	app.Handle(http.MethodGet, version, "/users/:page/:rows", ugh.Query, authen, admin, mid.Cors("*"))
	app.Handle(http.MethodGet, version, "/users/:id", ugh.QueryByID, authen, mid.Cors("*"))
	app.Handle(http.MethodPost, version, "/users", ugh.Create, authen, admin, mid.Cors("*"))
	app.Handle(http.MethodPut, version, "/users/:id", ugh.Update, authen, admin, mid.Cors("*"))
	app.Handle(http.MethodDelete, version, "/users/:id", ugh.Delete, authen, admin, mid.Cors("*"))

	// Register symbol endpoints
	sbl := symbolgrp.Handlers{
		Symbol: symbol.NewCore(cfg.Log, cfg.DB, cfg.Broker),
	}
	app.Handle(http.MethodGet, version, "/symbols/:page/:rows", sbl.Query)
	app.Handle(http.MethodGet, version, "/symbols/:id", sbl.QueryByID)
	app.Handle(http.MethodPost, version, "/symbols", sbl.Create, authen)

	// Register candle endpoints
	cgh := candlegrp.Handlers{
		Candle: candle.NewCore(cfg.Log, cfg.DB, cfg.Broker),
	}
	app.Handle(http.MethodGet, version, "/candles/:page/:rows", cgh.Query, mid.Cors("*"))
	app.Handle(http.MethodGet, version, "/candles/:symbol/:interval/:page/:rows", cgh.QueryBySymbolAndInterval, mid.Cors("*"))
	app.Handle(http.MethodGet, version, "/candles/:id", cgh.QueryByID, mid.Cors("*"))

	// Register position endpoints
	pos := positiongrp.Handlers{
		Position: position.NewCore(cfg.Log, cfg.DB),
	}
	app.Handle(http.MethodGet, version, "/positions/:page/:rows", pos.Query, authen, mid.Cors("*"))
	app.Handle(http.MethodGet, version, "/positions/:id", pos.QueryByID, authen, mid.Cors("*"))
	app.Handle(http.MethodPost, version, "/positions", pos.Create, authen, mid.Cors("*"))
	app.Handle(http.MethodDelete, version, "/positions/:id", pos.Close, authen, mid.Cors("*"))

	// Register order endpoints
	odr := ordergrp.Handlers{
		Order:    order.NewCore(cfg.Log, cfg.DB, cfg.Broker),
		Position: position.NewCore(cfg.Log, cfg.DB),
	}
	app.Handle(http.MethodGet, version, "/orders/:page/:rows", odr.Query, authen, mid.Cors("*"))
	app.Handle(http.MethodGet, version, "/orders/:id", odr.QueryByID, authen, mid.Cors("*"))
	app.Handle(http.MethodPost, version, "/orders", odr.Create, authen, mid.Cors("*"))
}

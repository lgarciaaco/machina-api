package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/lgarciaaco/machina-api/foundation/test/cucumber"

	"github.com/lgarciaaco/machina-api/business/data/dbschema"
	"github.com/lgarciaaco/machina-api/business/data/dbtest"

	"github.com/ardanlabs/conf/v2"
	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers"
	"github.com/lgarciaaco/machina-api/business/sys/auth"
	"github.com/lgarciaaco/machina-api/foundation/keystore"
	"github.com/lgarciaaco/machina-api/foundation/logger"
	"go.uber.org/zap"
)

/*
Need to figure out timeouts for http service.
*/

func Test_integration(t *testing.T) {
	log, err := logger.New("MACHINA-API")
	if err != nil {
		t.Fatal("unable to initialize logger")
	}
	defer log.Sync()

	var cfg struct {
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:3000"`
		}
		Auth struct {
			KeysFolder string `conf:"default:../../../../zarf/keys/"`
			ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
		}
	}

	const prefix = "MACHINA-API"
	_, err = conf.Parse(prefix, &cfg)
	if err != nil {
		t.Fatal("unable to parse configuration")
	}

	// =========================================================================
	// Initialize authentication support

	log.Infow("startup", "status", "initializing authentication support")

	// Construct a key store based on the key files stored in
	// the specified directory.
	ks, err := keystore.NewFS(os.DirFS(cfg.Auth.KeysFolder))
	if err != nil {
		t.Fatal("reading keys: %w", err)
	}

	auth, err := auth.New(cfg.Auth.ActiveKID, ks)
	if err != nil {
		t.Fatal("constructing auth: %w", err)
	}

	// =========================================================================
	// Database Support

	// Create connectivity to the database.
	log.Infow("startup", "status", "initializing database support using docker")

	c, err := dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	log, db, teardown := dbtest.NewUnit(t, c, "integration")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	// =========================================================================
	// Start API Service

	log.Infow("startup", "status", "initializing V1 API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Construct the mux for the API calls.
	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
		Auth:     auth,
		DB:       db,
	})

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	defer func() {
		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			t.Logf("unable to properly shutdown the api")
		}
	}()

	// Finally, execute the tests
	t.Logf("starting integration test for machina api")
	cucumber.RunSuite("features", log, db, auth, t)
}

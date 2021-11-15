package main

import (
	"context"
	"errors"
	"expvar" // Calls init function.
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/lgarciaaco/machina-api/app/strategies/moving-average/handlers"

	"github.com/lgarciaaco/machina-api/business/strategies"
	v1 "github.com/lgarciaaco/machina-api/business/strategies/api/v1"

	"github.com/lgarciaaco/machina-api/business/strategies/financial"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ardanlabs/conf/v2"
	"github.com/lgarciaaco/machina-api/foundation/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

/*
Need to figure out timeouts for http service.
*/

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log, err := logger.New("MACHINA_STRATEGY", zapcore.InfoLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// =========================================================================
	// GOMAXPROCS

	// Set the correct number of threads for the service
	// based on what is available either by the machine or quotas.
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// =========================================================================
	// Configuration

	cfg := struct {
		conf.Version
		Strategy struct {
			TradingPair   string  `conf:"default:97514fb4-4ff5-4561-91d1-c8da711d8f32,trading par for this strategy"`
			Interval      string  `conf:"default:5m,candle range"`
			Base          float64 `conf:"default:0.2,base coin this strategy will trade on"`
			Alt           float64 `conf:"default:500,alt coin this strategy will trade on"`
			Lot           float64 `conf:"default:0.1,lot to open orders per position"`
			WindowFast    int     `conf:"default:20,fast moving average"`
			WindowSlow    int     `conf:"default:100,slow moving average"`
			WindowWarming int     `conf:"default:100,how many candles are required to start trading"`
		}
		Web struct {
			DebugHost string `conf:"default:0.0.0.0:4000"`
		}
		API struct {
			Username string `conf:"noprint,required"`
			Password string `conf:"noprint,required"`
			Endpoint string `conf:"default:http://machina-trader.local"`
		}
		Zipkin struct {
			ReporterURI string  `conf:"default:http://localhost:9411/api/v2/spans"`
			ServiceName string  `conf:"default:machina-api"`
			Probability float64 `conf:"default:0.05"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "STRATEGY_MA"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// =========================================================================
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	expvar.NewString("build").Set(build)

	// =========================================================================
	// Start Tracing Support

	log.Infow("startup", "status", "initializing OT/Zipkin tracing support")

	traceProvider, err := startTracing(
		cfg.Zipkin.ServiceName,
		cfg.Zipkin.ReporterURI,
		cfg.Zipkin.Probability,
	)
	if err != nil {
		return fmt.Errorf("starting tracing: %w", err)
	}
	defer traceProvider.Shutdown(context.Background())

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// =========================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This include the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugMux(build, log)

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// =========================================================================
	// Start running the strategy

	// Client required for trader and puller
	client := &v1.Client{
		Client:    retryablehttp.NewClient(),
		Username:  cfg.API.Username,
		Password:  cfg.API.Password,
		TraderAPI: cfg.API.Endpoint,
	}
	err = client.Authenticate()
	if err != nil {
		return fmt.Errorf("authenticating agains trader api %v", err)
	}

	// Puller
	pInterval, err := time.ParseDuration(cfg.Strategy.Interval)
	if err != nil {
		log.Errorw("determining duration", "err", err)
	}
	puller := strategies.FromAPI{
		Log:          log,
		PullInterval: pInterval,
		TradingPair: strategies.TradingPair{
			Interval: cfg.Strategy.Interval,
			Symbol:   cfg.Strategy.TradingPair,
			Fast:     cfg.Strategy.WindowFast,
			Slow:     cfg.Strategy.WindowSlow,
			Warming:  cfg.Strategy.WindowWarming,
		},
		Client: client,
	}

	// Trader
	trader := &strategies.ToAPI{
		Log: log,
		Budget: &financial.FixBudget{
			BaseBudget: financial.BaseBudget{
				Base: cfg.Strategy.Base,
				Alt:  cfg.Strategy.Alt,
				Lot:  cfg.Strategy.Lot,
			},
		},
		Client: client,
	}

	// Print total profit to date
	defer log.Infof("main : total profit to date: %f", trader.Profit())

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Make a channel to shutdown the strategy
	done := make(chan bool)

	// Run the strategy
	go func() {
		s := financial.Strategy{
			Log: log,
			Rule: financial.NewMovingAverageRule(
				cfg.Strategy.WindowFast, cfg.Strategy.WindowSlow, cfg.Strategy.WindowWarming, &financial.TimeSeries{}),
		}

		serverErrors <- s.Run(done, puller, trader)
	}()
	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)
		close(done)

		// Give outstanding requests a deadline for completion.
		// ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		// defer cancel()
	}

	return nil
}

// =============================================================================

// startTracing configure open telemetry to be used with zipkin.
func startTracing(serviceName string, reporterURI string, probability float64) (*trace.TracerProvider, error) {

	// WARNING: The current settings are using defaults which may not be
	// compatible with your project. Please review the documentation for
	// opentelemetry.

	exporter, err := zipkin.New(
		reporterURI,
		// zipkin.WithLogger(zap.NewStdLog(log)),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(probability)),
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultBatchTimeout),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				attribute.String("exporter", "zipkin"),
			),
		),
	)

	// I can only get this working properly using the singleton :(
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

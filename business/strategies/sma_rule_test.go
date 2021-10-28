package strategies

import (
	"math"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/lgarciaaco/machina-api/foundation/logger"

	"github.com/lgarciaaco/machina-api/business/strategies/financial"

	"github.com/stretchr/testify/assert"
)

// Test by running the strategy, looking at the results and visually making sure
// they make sense.
func TestMovingAverageStrategy_Run(t *testing.T) {
	log, err := logger.New("TEST")
	if err != nil {
		t.Fatal("unable to create logger")
	}
	puller := FromFile{
		Log:  log,
		File: "./../../zarf/binance/ETHUSDT_4h_200.json",
	}
	trader := &ToStdout{Log: log, Lot: 0.5}

	// Create strategy and run it
	s := financial.Strategy{
		Rule: financial.NewMovingAverageRule(20, 100, 100, &financial.TimeSeries{Candles: make([]financial.Candle, 0)}),
	}

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Stop puller and trader by notifying this channel
	done := make(chan bool)
	serverErrors <- s.Run(done, puller, trader)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		assert.Nil(t, err)
		assert.Equal(t, 11.05, math.Floor(trader.Profit()*100)/100)

	case sig := <-shutdown:
		done <- true
		log.Infof("main : %v : Start shutdown", sig)
	}
}

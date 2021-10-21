// Package sync synchronizes candles from between database and binance API
// For each symbol, it pulls 100 candles per interval and the last candle afterwards
package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/lgarciaaco/machina-api/business/core/candle"

	"github.com/lgarciaaco/machina-api/business/core/symbol"

	"go.uber.org/zap"
)

// Synchronizer exposes functionalities to synchronize a resource. An example is
// synchronizing candles between binance api and the system
type Synchronizer interface {
	Run(ctx context.Context)
}

var (
	intervals       = []time.Duration{5 * time.Minute, time.Hour, 2 * time.Hour, 4 * time.Hour}
	intervalsString = map[time.Duration]string{
		5 * time.Minute: "5m", 10 * time.Minute: "10m", 15 * time.Minute: "15m",
		time.Hour: "1h", 2 * time.Hour: "2h", 4 * time.Hour: "4h",
	}
)

// CandleSynchronizer synchronizes candles between binance api and the system
type CandleSynchronizer struct {
	Log        *zap.SugaredLogger
	Symbol     symbol.Core
	Candle     candle.Core
	SyncPeriod time.Duration
}

// Run pulls candles from binance api and inserts them into the system
func (b *CandleSynchronizer) Run(ctx context.Context) {
	// Start synchronizing for all symbols
	go func() {
		ticker := time.NewTicker(b.SyncPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				b.Log.Infof("gracefully shutting down synchronizer")
				return
			case t := <-ticker.C:
				b.Log.Infof("sync at %s", t.String())
				func() {
					// In case this thread blocks, we want to release it before the next iteration
					// kicks in
					ctx, cancel := context.WithTimeout(ctx, b.SyncPeriod-b.SyncPeriod/10)
					defer cancel()

					if err := b.sync(ctx); err != nil {
						b.Log.Errorf("sync %s", err)
					}
				}()
			}
		}
	}()
}

// sync fetches symbols from database and checks weather it is time
// to pull a new candle for a given interval. If no candles exist
// for a symbol/interval pair, it seeds(fetches 100 candles) the pair
func (b CandleSynchronizer) sync(ctx context.Context) error {
	sbls, err := b.Symbol.Query(ctx, 1, 10)
	if err != nil {
		return fmt.Errorf("query symbols %w", err)
	}

	for _, s := range sbls {
		for _, i := range intervals {
			//
			dbCdl, err := b.Candle.QueryBySymbolAndInterval(ctx, 1, 1, s.ID, intervalsString[i])
			if err != nil {
				b.Log.Errorf("getting candles for symbol %s, interval %s", s.Symbol, intervalsString[i])
				continue
			}

			// If we dont get any candles from db, it means that
			// we never seed candles for the symbol / interval
			if len(dbCdl) == 0 {
				nCdl := candle.NewCandle{
					SymbolID: s.ID,
					Symbol:   s.Symbol,
					Interval: intervalsString[i],
				}
				if err := b.Candle.Seed(ctx, nCdl, 100); err != nil {
					b.Log.Errorf("seeding candles for symbol %s, interval %s", s.Symbol, intervalsString[i])
				}
				continue
			}

			// We check whether it is time to add a new candle by fetching the last
			// candle for the symbol
			if dbCdl[0].CloseTime.Add(i).Before(time.Now()) {
				_, err := b.Candle.Create(ctx, candle.NewCandle{
					SymbolID: s.ID,
					Symbol:   s.Symbol,
					Interval: intervalsString[i],
				})
				if err != nil {
					b.Log.Errorf("creating candles for symbol %s, interval %s", s.Symbol, intervalsString[i])
				}
			}
		}
	}

	return nil
}

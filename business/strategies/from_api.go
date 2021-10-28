package strategies

import (
	"time"

	"github.com/lgarciaaco/machina-api/business/strategies/financial"

	v1 "github.com/lgarciaaco/machina-api/business/strategies/api/v1"

	"go.uber.org/zap"
)

type FromAPI struct {
	Log          *zap.SugaredLogger
	PullInterval time.Duration
	TradingPair  TradingPair
	Client       *v1.Client
}

func (f FromAPI) Pull(done <-chan bool, candles chan<- financial.Candle) error {
	for _, c := range f.seed() {
		candles <- toFinancialCandle(c)
	}

	// We now regularly pull the last candle. if a candle is new, we pass it to the
	// strategy. The strategy takes over, evaluates and if a condition is met, it creates
	// a position via the positions channel
	ticker := time.NewTicker(f.PullInterval)
	for {
		select {
		case <-ticker.C:
			if c, err := f.Client.RetrieveCandle(f.TradingPair.Symbol, f.TradingPair.Interval, 1, 1); err != nil {
				f.Log.Errorf("puller: error pulling candle from api, retrying ...")
			} else {
				if len(c) != 0 {
					candles <- toFinancialCandle(toCandle(c[0]))
				}
			}
		case <-done:
			f.Log.Infof("puller : gracefully shutting down the puller")
			ticker.Stop()
			return nil
		}
	}
}

// seed fills in the data required for the strategy to work, namely
// as many candles as Strategy.Candle.Warning states
// It also validates the data for consistency
func (f FromAPI) seed() []Candle {
	tsc, err := f.Client.RetrieveCandle(f.TradingPair.Symbol, f.TradingPair.Interval, 1, f.TradingPair.Warming)
	if err != nil {
		f.Log.Errorf("puller : seed : error pulling candle from api")
		return []Candle{}
	} else {
		return toCandleSlice(tsc)
	}
}

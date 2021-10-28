package financial

type MovingAverageRuleCondition = func(fast MovingAverageIndicator, slow MovingAverageIndicator) (OrderType, ActionType)

// MovingAverageRule will tell when to start or to close an order. orders can be either buy or close
type MovingAverageRule struct {
	TS            *TimeSeries
	fast          MovingAverageIndicator
	slow          MovingAverageIndicator
	condition     MovingAverageRuleCondition
	warmingPeriod int
}

func NewMovingAverageRule(wfast int, wslow int, warming int, ts *TimeSeries) (r MovingAverageRule) {
	r = MovingAverageRule{
		TS:            ts,
		warmingPeriod: warming,
		fast: &Ema{
			MovingAverage: MovingAverage{
				TS:     ts,
				Window: wfast,
			}},
		slow: &Ema{
			MovingAverage: MovingAverage{
				TS:     ts,
				Window: wslow,
			}},
		condition: func(fast MovingAverageIndicator, slow MovingAverageIndicator) (OrderType, ActionType) {
			/*
			 * Im a downtrend, we open a SELL position if price crossover down the fast media
			 */
			if slow.Downtrend() {
				if fast.CrossOver() == Down {
					return Sell, Open
				}

				if fast.CrossOver() == Up {
					return Sell, Close
				}
			}

			// In an uptrend, we want to buy when the price cross over the fast media
			if slow.Uptrend() {
				if fast.CrossOver() == Up {
					return Buy, Open
				}

				if fast.CrossOver() == Down {
					return Buy, Close
				}
			}

			return 0, 0
		},
	}

	return
}

// Assert the sma condition.
func (mar MovingAverageRule) Assert(candle Candle) (OrderType, ActionType) {
	// If there is no previous candles, we just add the candle this time and skip assertions
	if len(mar.TS.Candles) == 0 || (len(mar.TS.Candles) != 0 && mar.TS.LastCandle().OpenTime != candle.OpenTime) {
		mar.TS.AddCandle(candle)
		mar.slow.Calculate()
		mar.fast.Calculate()

		// We can only assert a Rule if we already have calculated all EMAs,
		if mar.fast.Value() == 0 || mar.slow.Value() == 0 {
			return 0, 0
		}

		// Having a warning up period might help to smooth the moving averages
		if mar.slow.Position() < mar.warmingPeriod {
			return 0, 0
		}

		return mar.condition(mar.fast, mar.slow)
	}

	return 0, 0
}

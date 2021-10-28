package financial

// MovingAverage is the trend indicator.
// It takes average price figures and, as a result, smooth price action from fluctuations.
type MovingAverage struct {
	TS       *TimeSeries
	Window   int
	current  int
	previous float64
	value    float64
}

// Uptrend we are in an uptrend if values are increasing
func (ma *MovingAverage) Uptrend() bool {
	return ma.previous < ma.value
}

// Downtrend if we aren't in a uptrend, we must be in a downtrend
func (ma *MovingAverage) Downtrend() bool {
	return !ma.Uptrend()
}

// CrossOver the price can cross down a moving average only when it is red candle, same way, an
// up crossover can only occur with a green candle
func (ma *MovingAverage) CrossOver() (r Cross) {
	candle := ma.TS.LastCandle()
	r = None

	// For a green candle
	if candle.ClosePrice > candle.OpenPrice {
		if (candle.OpenPrice < ma.value) && (ma.value < candle.ClosePrice) {
			r = Up
		}
	}

	if candle.ClosePrice < candle.OpenPrice {
		if (candle.OpenPrice > ma.value) && (ma.value > candle.ClosePrice) {
			r = Down
		}
	}

	return
}

func (ma *MovingAverage) Value() float64 {
	return ma.value
}

func (ma *MovingAverage) Previous() float64 {
	return ma.previous
}

func (ma *MovingAverage) Position() int {
	return ma.current
}

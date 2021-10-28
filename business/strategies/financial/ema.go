package financial

// Ema (exponential moving average (EMA)) is a type of moving average (MA) that places a greater weight and significance
// on the most recent data points. The exponential moving average is also referred to as the exponentially weighted
// moving average. An exponentially weighted moving average reacts more significantly to recent price changes than a
// simple moving average (SMA), which applies an equal weight to all observations in the period.
type Ema struct {
	MovingAverage
}

// Calculate the formula for an EMA incorporates the previous period's EMA value, which in turn incorporates the value for the EMA value
// before that, and so on. Each previous EMA value accounts for a small portion of the current value. Therefore, the current EMA
// value will change depending on how much past data you use in your EMA calculation. Ideally, for a 100% accurate EMA,
// you should use every data point the stock has ever had in calculating the EMA, starting your calculations from the first day
// the stock existed. This is not always practical, but the more data points you use, the more accurate your EMA will be
func (ei *Ema) Calculate() float64 {
	ei.current++
	ei.previous = ei.value

	// If we dont have enough data for the current Window there is nothing we can do
	if ei.Window > len(ei.TS.Candles) {
		ei.value = 0.0
		return ei.value
	}

	// The first time we have enough data, EMA is equal to SMA
	if ei.Window == len(ei.TS.Candles) {
		ei.value = sma(ei.TS.Candles, ei.Window)
		return ei.value
	}

	// Majority of cases should land here where there is a previous EMA calculated
	// Multiplier: (2 / (Time periods + 1) )
	multiplier := 2.0 / float64(ei.Window+1.0)

	// close price
	close := ei.TS.LastCandle().ClosePrice

	// EMA: {close - EMA(previous day)} x multiplier + EMA(previous day).
	ema := (close-ei.previous)*multiplier + ei.previous

	ei.value = ema
	return ei.value
}

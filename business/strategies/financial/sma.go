package financial

// Sma (simple moving average (SMA)) calculates the average of a selected range of prices,
//usually closing prices, by the number of periods in that range.
type Sma struct {
	MovingAverage
}

// Calculate SMA. Formula for SMA https://www.investopedia.com/terms/s/sma.asp
// There are 3 cases:
// - When we start in a timeline with only one candle, we can safely take the value of the close as SMA
// - When the Window is bigger than the amount of Candles, for instance there are 5 Candles but the Window
//   is 20. For this case the best effort is to resize the Window to len(Candles) and calculate SMA. For the next iteration
//   if the scenario repeats (6 Candles, windows 20) the same methodology is applied.
// - Lastly, calculate an ordinary SMA
func (si *Sma) Calculate() float64 {
	si.current++
	si.previous = si.value

	// The first time there is no value, no calculation is needed so taking the values as it is should be fine
	if si.value == 0 {
		// Make sure that there is only one time series, if this is not the case, there should be already
		// some SMA calculated
		if len(si.TS.Candles) == 1 {
			si.value = si.TS.LastCandle().ClosePrice
		}

		return si.value
	}

	// If we dont have enough Candles to satisfy the windows, we calculate with what we have
	// TODO: Find a way better than iterating through the whole array
	if si.Window > len(si.TS.Candles)-1 {
		si.value = sma(si.TS.Candles, si.Window)
		return si.value
	}

	// We have calculated the previous SMA and we have enough Candles to calculate future SMAs
	drop := si.TS.Candles[len(si.TS.Candles)-(si.Window+1)].ClosePrice
	si.value = si.value + ((si.TS.LastCandle().ClosePrice - drop) / float64(si.Window))

	return si.value
}

// Calculate sma by reverse iterating the whole array
func sma(candles []Candle, window int) (r float64) {
	sum := 0.0
	if len(candles) < window {
		window = len(candles)
	}

	for i := len(candles) - 1; i >= len(candles)-window; i-- {
		sum += candles[i].ClosePrice
	}

	r = sum / float64(window)
	return
}

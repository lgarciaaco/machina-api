package financial

type TimeSeries struct {
	Candles []Candle
}

func (ts *TimeSeries) AddCandle(c Candle) {
	if ts.Candles == nil {
		ts.Candles = make([]Candle, 0)
	}

	if len(ts.Candles) == 0 {
		ts.Candles = append(ts.Candles, c)
		return
	}

	if len(ts.Candles) != 0 {
		if ts.LastCandle().OpenTime == c.OpenTime {
			return
		}
	}

	ts.Candles = append(ts.Candles, c)
}

func (ts *TimeSeries) LastCandle() Candle {
	if len(ts.Candles) > 0 {
		return ts.Candles[len(ts.Candles)-1]
	}

	return Candle{}
}

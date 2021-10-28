package financial

import (
	"testing"
	"time"
)

func TestMovingAverage_CrossOver(t *testing.T) {
	opentime := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name   string
		value  float64
		candle Candle
		wantR  Cross
	}{
		{
			name:   "Crossover up in a green candle",
			value:  10.0,
			candle: Candle{OpenPrice: 5.0, ClosePrice: 15.0, OpenTime: opentime},
			wantR:  Up,
		},
		{
			name:   "No crossover in a green candle",
			value:  1.0,
			candle: Candle{OpenPrice: 5.0, ClosePrice: 15.0, OpenTime: opentime.Add(time.Hour)},
			wantR:  None,
		},
		{
			name:   "No crossover in a green candle",
			value:  18.0,
			candle: Candle{OpenPrice: 5.0, ClosePrice: 15.0, OpenTime: opentime.Add(2 * time.Hour)},
			wantR:  None,
		},
		{
			name:   "Crossover down in a red candle",
			value:  10.0,
			candle: Candle{OpenPrice: 15.0, ClosePrice: 5.0, OpenTime: opentime.Add(3 * time.Hour)},
			wantR:  Down,
		},
		{
			name:   "No crossover in a red candle",
			value:  1.0,
			candle: Candle{OpenPrice: 15.0, ClosePrice: 5.0, OpenTime: opentime.Add(4 * time.Hour)},
			wantR:  None,
		},
	}
	ts := &TimeSeries{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := &MovingAverage{
				value: tt.value,
				TS:    ts,
			}

			ts.AddCandle(tt.candle)
			if gotR := ma.CrossOver(); gotR != tt.wantR {
				t.Errorf("CrossOver() = %v, want %v", gotR, tt.wantR)
			}
		})
	}
}

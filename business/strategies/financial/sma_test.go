package financial

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
This test runs a scenario where we start
with a close price of 1 and adding successive Candles with
closes as following:

1(initial), 3, 5, 3, 8, 1, 13

The SMA should be for a Window of 5 should be
1, 2, 3, 4, 4, 4, 6
*/
func TestSma_Calculate(t *testing.T) {
	si := &Sma{
		MovingAverage{
			TS:      &TimeSeries{},
			Window:  5,
			current: 0,
			value:   0,
		},
	}

	tests := []struct {
		value float64
		sma   float64
	}{
		{1, 1}, {3, 2}, {5, 3},
		{3, 3}, {8, 4}, {1, 4},
		{13, 6},
	}

	// Run calculating SMA
	opentime := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.Local)
	for i, tt := range tests {
		t.Run(fmt.Sprintf("with a close value or %v EMA now should be %v", tt.value, tt.sma), func(t *testing.T) {
			si.TS.AddCandle(Candle{ClosePrice: tt.value, OpenTime: opentime})
			si.Calculate()
			opentime = opentime.Add(time.Hour)
			assert.Equal(t, tt.sma, si.value)
			assert.Equal(t, i, si.current-1)
		})
	}
}

package financial

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmaIndicator_Calculate(t *testing.T) {
	ei := &Ema{
		MovingAverage{
			TS:      &TimeSeries{},
			Window:  10,
			current: 0,
			value:   0,
		},
	}

	// Because we set a Window of 10, the first 9 times we should get an EMA of 0.0
	opentime := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.Local)
	for i, v := range []float64{22.27, 22.19, 22.08, 22.17, 22.18, 22.13, 22.23, 22.43, 22.24} {
		ei.TS.AddCandle(Candle{ClosePrice: v, OpenTime: opentime})
		ei.Calculate()
		opentime = opentime.Add(time.Hour)

		assert.Equal(t, 0.0, ei.value)
		assert.Equal(t, i+1, ei.current)
	}

	// Now with enough data
	tests := []struct {
		close float64
		ema   float64
	}{
		{22.29, 22.22}, {22.15, 22.21}, {22.39, 22.24},
		{22.38, 22.27}, {22.61, 22.33}, {23.36, 22.52},
		{24.05, 22.80}, {23.75, 22.97}, {23.83, 23.13},
		{23.95, 23.28}, {23.63, 23.34}, {23.82, 23.43},
		{23.87, 23.51}, {23.65, 23.53}, {23.19, 23.47},
		{23.10, 23.40}, {23.33, 23.39}, {22.68, 23.26},
		{23.10, 23.23}, {22.40, 23.08}, {22.17, 22.92},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("with a close value or %v EMA now should be %v", tt.close, tt.ema), func(t *testing.T) {
			ei.TS.AddCandle(Candle{ClosePrice: tt.close, OpenTime: opentime})
			opentime = opentime.Add(time.Hour)
			calcEma := ei.Calculate()
			assert.Equal(t, tt.ema, math.Round(calcEma*100)/100)
		})
	}
}

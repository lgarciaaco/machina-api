package financial

import (
	"time"
)

// Candle Candlestick charts are used by traders to determine possible price movement based on past patterns.
type Candle struct {
	ID         string    `json:"id"`
	SymbolID   string    `json:"-"`
	Symbol     string    `json:"symbol"`
	Interval   string    `json:"interval"`
	OpenTime   time.Time `json:"open_time"`
	OpenPrice  float64   `json:"open_price"`
	ClosePrice float64   `json:"close_price"`
	CloseTime  time.Time `json:"close_time"`
	Low        float64   `json:"low"`
	High       float64   `json:"high"`
	Volume     float64   `json:"volume"`
}

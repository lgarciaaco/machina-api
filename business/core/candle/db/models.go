package db

import "time"

// Candle is a type of price chart used in technical analysis that displays the high,
// low, open, and closing prices of a security for a specific period
type Candle struct {
	ID         string    `db:"candle_id"`
	Symbol     string    `db:"symbol"`
	Interval   string    `db:"interval"`
	OpenTime   time.Time `db:"open_time"`
	OpenPrice  float64   `db:"open_price"`
	ClosePrice float64   `db:"close_price"`
	CloseTime  time.Time `db:"close_time"`
	Low        float64   `db:"low"`
	High       float64   `db:"high"`
	Volume     float64   `db:"volume"`
}

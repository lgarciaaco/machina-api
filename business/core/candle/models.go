package candle

import (
	"time"

	"github.com/lgarciaaco/machina-api/business/core/candle/db"
)

// Candle represents an individual candle
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

type NewCandle struct {
	SymbolID string `json:"symbol_id" validate:"required"`
	Symbol   string `json:"symbol"`
	Interval string `json:"interval" validate:"required"`
}

func toCandle(dbCdl db.Candle) Candle {
	pc := (*Candle)(&dbCdl)
	pc.CloseTime = time.Date(
		pc.CloseTime.Year(), pc.CloseTime.Month(), pc.CloseTime.Day(),
		pc.CloseTime.Hour(), pc.CloseTime.Minute(), pc.CloseTime.Second(),
		pc.CloseTime.Nanosecond(), time.Local)

	pc.OpenTime = time.Date(
		pc.OpenTime.Year(), pc.OpenTime.Month(), pc.OpenTime.Day(),
		pc.OpenTime.Hour(), pc.OpenTime.Minute(), pc.OpenTime.Second(),
		pc.OpenTime.Nanosecond(), time.Local)

	return *pc
}

func toCandleSlice(dbCdls []db.Candle) []Candle {
	cdls := make([]Candle, len(dbCdls))
	for i, dbCdl := range dbCdls {
		cdls[i] = toCandle(dbCdl)
	}
	return cdls
}

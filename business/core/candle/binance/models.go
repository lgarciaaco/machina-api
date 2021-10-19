package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/lgarciaaco/machina-api/business/broker"
)

// Candle charts display the high, low, open, and closing prices of a
// security for a specific period. Independently of the answer from Binance api,
// we want to return structured json
type Candle struct {
	ID         string    `db:"candle_id"` // Not used but set for easy data transformation
	SymbolID   string    `db:"symbol_id"` // Not used but set for easy data transformation
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

// toCandle marshals the body of a response from binance klines api into a Candle struct
// it returns the array of candles
//
// https://github.com/binance-exchange/binance-official-api-docs/blob/master/rest-api.md#klinecandlestick-data
func toCandle(rd io.Reader, symbol string, interval string) (cs []Candle, err error) {
	/*
		We expect a response like:
		[
		  [
			1499040000000,      // open time
			"0.01634790",       // open
			"0.80000000",       // High
			"0.01575800",       // Low
			"0.01577100",       // close
			"148976.11427815",  // Volume
			1499644799999,      // close time
			"2434.19055334",    // Quote asset volume
			308,                // Number of trades
			"1756.87402397",    // Taker buy base asset volume
			"28.46694368",      // Taker buy quote asset volume
			"17928899.62484339" // Ignore.
		  ]
		]
	*/
	bi := make([][]interface{}, 0)
	if err := json.NewDecoder(rd).Decode(&bi); err != nil {
		return nil, fmt.Errorf("unable to unmarshal binance response, err %w", err)
	}

	cs = make([]Candle, 0, len(bi))
	for _, res := range bi {
		// The response from binance should have exactly 12 fields, having anything different than 12
		// is an indication that api might have changed
		if len(res) != 12 {
			return nil, fmt.Errorf("binance response wrong formated")
		}

		// Set open and  and close time
		candle := Candle{
			Symbol:    symbol,
			Interval:  interval,
			OpenTime:  broker.ToTime(res[0].(float64)),
			CloseTime: broker.ToTime(res[6].(float64)),
		}

		// Set open price
		candle.OpenPrice, err = strconv.ParseFloat(res[1].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse candle openPrice out of binance response")
		}

		// Set high
		candle.High, err = strconv.ParseFloat(res[2].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse candle high out of binance response")
		}

		// Set low
		candle.Low, err = strconv.ParseFloat(res[3].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse candle low out of binance response")
		}

		// Set close price
		candle.ClosePrice, err = strconv.ParseFloat(res[4].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse candle closePrice out of binance response")
		}

		// Set volume
		candle.Volume, err = strconv.ParseFloat(res[5].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse candle volume out of binance response")
		}

		cs = append(cs, candle)
	}

	return cs, nil
}

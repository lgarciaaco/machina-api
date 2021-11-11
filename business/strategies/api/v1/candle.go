package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Candle Candlestick charts are used by traders to determine possible price movement based on past patterns.
type Candle struct {
	ID         string    `json:"id"`
	SymbolID   string    `json:"symbol_id"`
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

func (c *Client) RetrieveCandle(par string, interval string, pageNumber int, rowsPerPage int) (candle []Candle, err error) {
	resp, err := http.Get(fmt.Sprintf("%s%s/%s/%s/%d/%d", c.TraderAPI, "/v1/candles", par, interval, pageNumber, rowsPerPage))
	if err != nil {
		return candle, err
	}
	defer resp.Body.Close()

	// we care only about status codes in 2xx range, anything else we can't process
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return candle, fmt.Errorf("status code [%d] out of range, expecting 200 <= status code <= 299", resp.StatusCode)
	}

	// finally, return the body
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return candle, fmt.Errorf("unable to read from response")
	} else {
		if err = json.Unmarshal(b, &candle); err != nil {
			return candle, fmt.Errorf("unable to unmarshal reposns %s into json", b)
		}
	}

	return
}

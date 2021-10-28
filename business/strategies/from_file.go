package strategies

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"

	"go.uber.org/zap"

	"github.com/lgarciaaco/machina-api/business/strategies/financial"

	"github.com/lgarciaaco/machina-api/business/broker"
)

// FromFile is a puller that loads candles from a file. It blocks the thread and
// waits for exit signal, making sure the puller gets all the candles
type FromFile struct {
	Log  *zap.SugaredLogger
	File string
}

func (f FromFile) Pull(done <-chan bool, candles chan<- financial.Candle) error {
	data, err := LoadTimeSeriesFromFile(f.File)
	if err != nil {
		log.Fatal("cant open file")
		return err
	}

	for _, c := range data {
		c.Symbol = "BNBUSDT"
		c.Interval = "1h"
		c.SymbolID = "97514fb4-4ff5-4561-91d1-c8da711d8f32"
		candles <- toFinancialCandle(c)
	}

	<-done
	f.Log.Infof("puller : gracefully shutting down strategy")
	return nil
}

// LoadTimeSeriesFromFile candles from file
func LoadTimeSeriesFromFile(path string) (r []Candle, err error) {
	// read file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	candles := make([][]interface{}, 0)
	if err = json.Unmarshal(data, &candles); err != nil {
		return
	}

	for _, v := range candles {
		cs := Candle{
			OpenTime:  broker.ToTime(v[0].(float64)),
			CloseTime: broker.ToTime(v[6].(float64)),
		}

		cs.OpenPrice, err = strconv.ParseFloat(v[1].(string), 64)
		if err != nil {
			return nil, err
		}

		cs.High, err = strconv.ParseFloat(v[2].(string), 64)
		if err != nil {
			return nil, err
		}

		cs.Low, err = strconv.ParseFloat(v[3].(string), 64)
		if err != nil {
			return nil, err
		}

		cs.ClosePrice, err = strconv.ParseFloat(v[4].(string), 64)
		if err != nil {
			return nil, err
		}

		cs.Volume, err = strconv.ParseFloat(v[5].(string), 64)
		if err != nil {
			return nil, err
		}

		r = append(r, cs)
	}

	return
}

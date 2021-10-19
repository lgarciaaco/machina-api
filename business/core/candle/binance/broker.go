// Package binance manages candles using the binance api
package binance

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lgarciaaco/machina-api/business/broker"
	"go.uber.org/zap"
)

// Agent manages the set of API's for candle access.
type Agent struct {
	log    *zap.SugaredLogger
	broker broker.Broker
}

// NewAgent constructs a data for api access.
func NewAgent(log *zap.SugaredLogger, brk broker.Broker) Agent {
	return Agent{
		log:    log,
		broker: brk,
	}
}

// QueryBySymbolAndInterval fetch candles from binance api by symbol and interval
func (a Agent) QueryBySymbolAndInterval(cxt context.Context, sbl, ival string, limit int) (or []Candle, err error) {
	bncResp, err := a.broker.Request(cxt, http.MethodGet, "klines",
		"symbol", sbl,
		"interval", ival,
		"limit", strconv.Itoa(limit))
	if err != nil {
		return nil, fmt.Errorf("fetching exchange info %w", err)
	}

	cdls, err := toCandle(bncResp, sbl, ival)
	if err != nil {
		return nil, fmt.Errorf("marshaling candles %w", err)
	}

	return cdls, nil
}

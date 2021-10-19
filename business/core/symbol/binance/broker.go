// Package binance manages symbol using the binance api
package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lgarciaaco/machina-api/business/broker"
	"go.uber.org/zap"
)

// Agent manages the set of API's for symbol access.
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

// QueryBySymbol fetch a symbol from binance api
func (a Agent) QueryBySymbol(cxt context.Context, sbl string) (or Symbol, err error) {
	bncResp, err := a.broker.Request(cxt, http.MethodGet, "exchangeInfo",
		"symbol", sbl)
	if err != nil {
		return Symbol{}, fmt.Errorf("fetching exchange info %w", err)
	}

	type exchangeInfo struct {
		Symbols []Symbol `json:"symbols"`
	}
	var ei exchangeInfo
	if err := json.NewDecoder(bncResp).Decode(&ei); err != nil {
		return Symbol{}, fmt.Errorf("decoding exchange info %w", err)
	}

	return ei.Symbols[0], nil
}

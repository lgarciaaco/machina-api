// Package binance manages orders using the binance api
package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lgarciaaco/machina-api/business/broker"
	"go.uber.org/zap"
)

// Agent manages the set of API's for order access.
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

// Create dispatch a POST broker call attempting to create a MARKET order. It returns the
// broker response
func (a Agent) Create(cxt context.Context, nOdr Order) (or OrderResponse, err error) {
	bncResp, err := a.broker.Request(cxt, http.MethodPost,
		"symbol", nOdr.Symbol,
		"side", nOdr.Side,
		"type", nOdr.Type,
		"quantity", strconv.FormatFloat(nOdr.Quantity, 'f', -1, 64))
	if err != nil {
		return OrderResponse{}, fmt.Errorf("creating order %w", err)
	}

	var odrResp OrderResponse
	if err := json.NewDecoder(bncResp).Decode(&odrResp); err != nil {
		return OrderResponse{}, fmt.Errorf("decoding order response %w", err)
	}

	if odrResp.Status != "FILLED" {
		return OrderResponse{}, fmt.Errorf("received unsupported status %s", odrResp.Status)
	}

	// We calculate the average price for this order iterating
	// through the different fills
	price := 0.0
	for _, f := range odrResp.Fills {
		price += f.Price
	}
	odrResp.Price = price / float64(len(odrResp.Fills))

	return odrResp, nil
}

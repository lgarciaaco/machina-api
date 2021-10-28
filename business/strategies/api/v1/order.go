package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Order represent an order in a position
type Order struct {
	ID           string    `json:"order_id"`
	SymbolID     string    `json:"symbol_id"`
	PositionID   string    `json:"position_id"`
	CreationTime time.Time `json:"creation_time"`
	Price        float64   `json:"price"`
	Quantity     float64   `json:"quantity"`
	Status       string    `json:"status"`
	Type         string    `json:"type"`
	Side         string    `json:"side"`
}

// NewOrder contains information needed to create a new Order.
type NewOrder struct {
	PositionID string  `json:"position_id"`
	Quantity   float64 `json:"quantity"`
	Side       string  `json:"side"`
}

func (c *Client) CreateOrder(no NewOrder) (o *Order, err error) {
	body, err := json.Marshal(no)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.TraderAPI, "/v1/orders"), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Set authentication headers
	req.Header = map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", c.token)},
	}
	resp, err := c.Do(&retryablehttp.Request{Request: req})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// we care only about status codes in 2xx range, anything else we can't process
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return nil, ErrUnexpectedCode
	}

	// unmarshal response
	pos := &Order{}
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		if err = json.Unmarshal(b, pos); err != nil {
			return nil, err
		}
	}

	return pos, nil
}

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

// Position represents a single position
type Position struct {
	ID           string    `json:"position_id"`   // Position ID
	SymbolID     string    `json:"-"`             // SymbolID this position is trading on, used to preload Symbol
	UserID       string    `json:"-"`             // UserID who created this position, used to preload User
	Side         string    `json:"side"`          // Position side: SELL / BUY
	Status       string    `json:"status"`        // Status open / closed
	CreationTime time.Time `json:"creation_time"` // CreationTime of the position
	User         string    `json:"user"`          // Name of the owner
	Symbol       string    `json:"symbol"`        // Symbol this position is trading on
	Orders       []Order   `json:"orders"`        // Orders belonging to this position
}

// NewPosition contains information needed to create a new position
type NewPosition struct {
	SymbolID string `json:"symbol_id"`
	Side     string `json:"side"`
}

// ListPositions return all positions belonging to this username
func (c *Client) ListPositions() (p []Position, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", c.TraderAPI, "/v1/positions/1/10"), nil)
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
	var pos []Position
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		if err = json.Unmarshal(b, &pos); err != nil {
			return nil, err
		}
	}

	// This is the case where there are no positions
	if len(pos) == 0 {
		return nil, nil
	}

	return pos, nil
}

// RetrievePosition return the last position belonging to this username
func (c *Client) RetrievePosition() (p *Position, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", c.TraderAPI, "/v1/positions/1/1"), nil)
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
	var pos []Position
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		if err = json.Unmarshal(b, &pos); err != nil {
			return nil, err
		}
	}

	// This is the case where there are no positions
	if len(pos) == 0 {
		return nil, nil
	}

	return &pos[len(pos)-1], nil
}

// CreatePosition do a POST to the trader api and Creates a position
func (c *Client) CreatePosition(np NewPosition, quantity float64) (p *Position, err error) {
	body, err := json.Marshal(np)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.TraderAPI, "/v1/positions"), bytes.NewBuffer(body))
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
	pos := &Position{}
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		if err = json.Unmarshal(b, pos); err != nil {
			return nil, err
		}
	}

	nOdr := NewOrder{
		PositionID: pos.ID,
		Quantity:   quantity,
		Side:       pos.Side,
	}
	odr, err := c.CreateOrder(nOdr)
	if err != nil {
		return pos, fmt.Errorf("unable to create order %w", err)
	}

	pos.Orders = append(pos.Orders, *odr)
	return pos, nil
}

// ClosePosition closes a position using the trader api
func (c *Client) ClosePosition(id string) (p *Position, err error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s/%s", c.TraderAPI, "/v1/positions", id), nil)
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

	// Fetch the position so we can return it
	return c.RetrievePosition()
}

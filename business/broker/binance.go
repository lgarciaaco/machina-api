// Package broker manages calls to binance api V3. Documentation about all endpoints
// and how should they be formed can be found at https://github.com/binance/binance-spot-api-docs
package broker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/lgarciaaco/machina-api/business/broker/encode"
	"github.com/pkg/errors"
)

type Broker interface {
	Request(ctx context.Context, method string, keysAndValues ...string) (rd io.Reader, err error)
}

const (
	OrderTypeMarket    = "MARKET"
	MaxIdleConnections = 10
	IdleConnTimeout    = 30 * time.Second
	TestNet            = "https://testnet.binance.vision/api/v3"
)

// Binance manages calls to binance api v3
type Binance struct {
	Endpoint string        // Endpoint: ping, time, orders ...
	APIKey   string        // APIKey is required for calls that need authentication
	Signer   encode.Signer // Signer is used to encode calls to binance that include sensitive data, like APIKey
}

// Request convert a bunch of key-value pairs into an url query, it takes the api endpoint
// and builds the binance api request. It returns the body of the response
func (as Binance) Request(ctx context.Context, method string, keysAndValues ...string) (rd io.Reader, err error) {
	// form the api request url
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/%s?", TestNet, as.Endpoint), nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create request")
	}

	q := req.URL.Query()
	for i := 0; i < len(keysAndValues); {
		// make sure this isn't a mismatched key
		if i == len(keysAndValues)-1 {
			return nil, fmt.Errorf("odd number of arguments passed as key-value pairs")
		}

		// process a key-value pair,
		key, val := keysAndValues[i], keysAndValues[i+1]
		q.Add(key, val)
		i += 2
	}
	q.Add("timestamp", strconv.FormatInt(unixMillis(time.Now()), 10))

	// If there is an Api key defined we include it in the header
	if as.APIKey != "" {
		req.Header.Add("X-MBX-APIKEY", as.APIKey)
	}

	// Add signature parameter if signature is defined
	if as.Signer != nil {
		signature, err := as.Signer.Sign([]byte(q.Encode()))
		if err != nil {
			return nil, fmt.Errorf("unable to sign")
		}
		q.Add("signature", signature)
	}

	req.URL.RawQuery = q.Encode()

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       MaxIdleConnections,
			IdleConnTimeout:    IdleConnTimeout,
			DisableCompression: true,
		}}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to issue request, err %w", err)
	}
	defer resp.Body.Close()

	// we care only about status codes in 2xx range, anything else we can't process
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return nil, fmt.Errorf("status code [%d] out of range, expecting 200 <= status code <= 299", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("creating reader %w", err)
	}
	r := bytes.NewReader(b)

	// finally, return the reader for the body
	return r, nil
}

func unixMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

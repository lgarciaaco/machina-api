package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/lgarciaaco/machina-api/business/broker/encode"
)

// TestBinance manages calls to binance test api v3 if the endpoint
// has an test api. Currently, only order endpoint has a test api
type TestBinance struct {
	APIKey string        // APIKey is required for calls that need authentication
	Signer encode.Signer // Signer is used to encode calls to binance that include sensitive data, like APIKey
}

// Request convert a bunch of key-value pairs into an url query, it takes the api endpoint
// and builds the binance api request. It returns the body of the response
func (as TestBinance) Request(ctx context.Context, method, endpoint string, keysAndValues ...string) (rd io.Reader, err error) {
	base := APIV3
	if endpoint == "order" {
		base = TestNet
	}

	// form the api request url
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/%s", base, endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request %w", err)
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

	// Order is the only authenticated endpoint so we need to pass the keys
	if endpoint == "order" {
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

	// finally, return the reader for the body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("creating reader %w", err)
	}
	r := bytes.NewReader(b)

	// finally, return the reader for the body
	return r, nil
}

// Time fetches the api time
func (as TestBinance) Time(ctx context.Context) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", TestNet, "time"), nil)
	if err != nil {
		return 0, fmt.Errorf("unable to create request %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       MaxIdleConnections,
			IdleConnTimeout:    IdleConnTimeout,
			DisableCompression: true,
		}}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to issue request, err %w", err)
	}
	defer resp.Body.Close()

	// we care only about status codes in 2xx range, anything else we can't process
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return 0, fmt.Errorf("status code [%d] out of range, expecting 200 <= status code <= 299", resp.StatusCode)
	}

	type time struct {
		Time int64 `json:"serverTime"`
	}
	var t time

	// finally, return the body
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return 0, fmt.Errorf("unable to read from response")
	} else {
		if err = json.Unmarshal(b, &t); err != nil {
			return 0, fmt.Errorf("unable to unmarshal reposns %s into json", b)
		}
	}

	return t.Time, nil
}

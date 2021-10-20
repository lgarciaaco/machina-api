package broker

import (
	"bytes"
	"context"
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
	if endpoint == "order" {
		endpoint = endpoint + "/test"
	}

	// form the api request url
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/%s", APIV3, endpoint), nil)
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

	// If the endpoint is order, there is no response sent from test api, we need
	// to fake a response
	if endpoint == "order/test" {
		return as.fakeResponse(), nil
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

// Because test api doesnt return a response ...
func (as TestBinance) fakeResponse() io.Reader {
	const resp = `
{
  "symbol": "BTCUSDT",
  "orderId": 28,
  "orderListId": -1,
  "clientOrderId": "6gCrw2kRUAF9CvJDGP16IP",
  "transactTime": 1507725176595,
  "price": "0.00000000",
  "origQty": "10.00000000",
  "executedQty": "10.00000000",
  "cummulativeQuoteQty": "10.00000000",
  "status": "FILLED",
  "timeInForce": "GTC",
  "type": "MARKET",
  "side": "SELL",
  "fills": [
    {
      "price": "4000.00000000",
      "qty": "1.00000000",
      "commission": "4.00000000",
      "commissionAsset": "USDT"
    },
    {
      "price": "3999.00000000",
      "qty": "5.00000000",
      "commission": "19.99500000",
      "commissionAsset": "USDT"
    },
    {
      "price": "3998.00000000",
      "qty": "2.00000000",
      "commission": "7.99600000",
      "commissionAsset": "USDT"
    },
    {
      "price": "3997.00000000",
      "qty": "1.00000000",
      "commission": "3.99700000",
      "commissionAsset": "USDT"
    },
    {
      "price": "3995.00000000",
      "qty": "1.00000000",
      "commission": "3.99500000",
      "commissionAsset": "USDT"
    }
  ]
}
`
	return bytes.NewReader([]byte(resp))
}

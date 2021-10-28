// Package v1 do calls to machina api v1
package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

var (
	// ErrUnexpectedCode occurs when a strategy tries opens or close a position, but the strategy's budget is out of fund.
	ErrUnexpectedCode = errors.New("unexpected code in response")
)

type Client struct {
	*retryablehttp.Client
	Username  string
	Password  string
	TraderAPI string

	PullInterval time.Duration

	token string
}

// Authenticate requests a token from the trader api
func (c *Client) Authenticate() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.TraderAPI, "/v1/users/token"), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.Do(&retryablehttp.Request{Request: req})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// we care only about status codes in 2xx range, anything else we can't process
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return ErrUnexpectedCode
	}

	// set the token
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("unable to read from response")
	} else {
		var token = struct {
			Token string `json:"token"`
		}{}

		if err = json.Unmarshal(b, &token); err != nil {
			return fmt.Errorf("unable to unmarshal reposns %s into json", b)
		}

		c.token = token.Token
	}
	return nil
}

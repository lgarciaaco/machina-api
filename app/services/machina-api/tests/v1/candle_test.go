package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lgarciaaco/machina-api/business/core/candle"

	"github.com/lgarciaaco/machina-api/business/data/dbschema"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers"
	"github.com/lgarciaaco/machina-api/business/data/dbtest"
)

// CandleTests holds methods for each candle subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type CandleTests struct {
	app http.Handler
}

// TestCandles runs a series of tests to exercise Candle behavior from the
// API level. The subtests all share the same database and application for
// speed and convenience. Given the fact that candle api only list candles, we
// use a populated database
func TestCandles(t *testing.T) {
	t.Parallel()

	test := dbtest.NewIntegration(t, c, "inttestcdls")
	t.Cleanup(test.Teardown)

	_, db, teardown := dbtest.NewUnit(t, c, "integration")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	shutdown := make(chan os.Signal, 1)
	tests := CandleTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
	}

	t.Run("getCandle200", tests.getCandle200)
	t.Run("getCandle404", tests.getCandle404)
	t.Run("getCandles200", tests.getCandles200)
	t.Run("getCandles200", tests.getCandlesWithWrongSymbol200)
	t.Run("getCandles200", tests.getCandlesWithWrongInterval200)
}

// getCandle404 validates a candle request for a candle that does not exist with the endpoint.
func (pt *CandleTests) getCandle404(t *testing.T) {
	id := "a224a8d6-3f9e-4b11-9900-e81a25d80702"

	r := httptest.NewRequest(http.MethodGet, "/v1/candles/"+id, nil)
	w := httptest.NewRecorder()

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a candle with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new candle %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", dbtest.Success, testID)

			got := w.Body.String()
			exp := "not found"
			if !strings.Contains(got, exp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// getCandle200 validates a candle request for an existing id.
func (pt *CandleTests) getCandle200(t *testing.T) {
	id := "039eee35-7463-4dbd-ae91-0428f3b89c42"
	r := httptest.NewRequest(http.MethodGet, "/v1/candles/"+id, nil)
	w := httptest.NewRecorder()

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a candle that exists.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the existing candle %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)

			var got candle.Candle
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			exp := got
			exp.ID = id
			exp.Symbol = "ETHUSDT"
			exp.Volume = 13456.00
			exp.ClosePrice = 110.50
			exp.OpenPrice = 100.50

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)

		}
	}
}

// getCandles200 validates a request for a list of candles with correct data.
// The list should be ordered and new candles should come first
func (pt *CandleTests) getCandles200(t *testing.T) {
	symbol := "ETHUSDT"
	interval := "4h"
	pageNumber, rowsPerPage := 1, 1

	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/candles/%s/%s/%d/%d", symbol, interval, pageNumber, rowsPerPage), nil)
	w := httptest.NewRecorder()

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting candles by symbol and interval.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using symbol %s and interval %s.", testID, symbol, interval)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)
		}

		var got []candle.Candle
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
		}

		if len(got) != 1 {
			t.Fatalf("\t%s\tTest : %d\tShould get exactly one candle.", dbtest.Failed, testID)
		}
		t.Logf("\t%s\tTest : %d\tShould get exactly one candle.", dbtest.Success, testID)

		// Define what we wanted to receive. We will just trust the generated
		// fields like Dates so we copy p.
		exp := got[0]
		exp.ID = "cd0f4919-2fe7-4808-8ba2-a1ea652cd591"
		exp.Symbol = "ETHUSDT"
		exp.Volume = 33456.00
		exp.ClosePrice = 310.50
		exp.OpenPrice = 200.50

		if diff := cmp.Diff(got[0], exp); diff != "" {
			t.Fatalf("\t%s\tTest : %d\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
		}
		t.Logf("\t%s\tTest : %d\tShould get the expected result.", dbtest.Success, testID)
	}
}

// getCandlesWithWrongInterval200 validates a request for a list of candles with wrong symbol.
// When wrong symbol is provided the app should return 200 and empty array.
func (pt *CandleTests) getCandlesWithWrongSymbol200(t *testing.T) {
	symbol := "some_symbol"
	interval := "4h"
	pageNumber, rowsPerPage := 1, 1

	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/candles/%s/%s/%d/%d", symbol, interval, pageNumber, rowsPerPage), nil)
	w := httptest.NewRecorder()

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to deal with wrong symbol.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using symbol %s and interval %s.", testID, symbol, interval)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)
		}

		var got []candle.Candle
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
		}

		if len(got) != 0 {
			t.Fatalf("\t%s\tTest : %d\tShouldn't get candles.", dbtest.Failed, testID)
		}
		t.Logf("\t%s\tTest : %d\tShouldn't get candles.", dbtest.Success, testID)
	}
}

// getCandlesWithWrongInterval200 validates a request for a list of candles with wrong interval.
// When wrong interval is provided the app should return 200 and empty array.
func (pt *CandleTests) getCandlesWithWrongInterval200(t *testing.T) {
	symbol := "ETHUSDT"
	interval := "13h"
	pageNumber, rowsPerPage := 1, 1

	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/candles/%s/%s/%d/%d", symbol, interval, pageNumber, rowsPerPage), nil)
	w := httptest.NewRecorder()

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to deal with wrong interval.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using symbol %s and interval %s.", testID, symbol, interval)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)
		}

		var got []candle.Candle
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
		}

		if len(got) != 0 {
			t.Fatalf("\t%s\tTest : %d\tShouldn't get candles.", dbtest.Failed, testID)
		}
		t.Logf("\t%s\tTest : %d\tShouldn't get candles.", dbtest.Success, testID)
	}
}

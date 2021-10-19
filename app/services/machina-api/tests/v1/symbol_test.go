package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lgarciaaco/machina-api/business/core/symbol"
	"github.com/lgarciaaco/machina-api/business/sys/validate"
	v1Web "github.com/lgarciaaco/machina-api/business/web/v1"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers"
	"github.com/lgarciaaco/machina-api/business/broker"
	"github.com/lgarciaaco/machina-api/business/data/dbtest"
)

// SymbolTests holds methods for each order subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type SymbolTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// TestSymbols is the entry point for testing order management functions.
func TestSymbols(t *testing.T) {
	t.Parallel()

	test := dbtest.NewIntegration(t, c, "inttestsymbols")
	t.Cleanup(test.Teardown)

	broker := broker.Binance{}

	shutdown := make(chan os.Signal, 1)
	tests := SymbolTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
			Broker:   broker,
		}),
		userToken:  test.Token("45b5fbd3-755f-4379-8f07-a58d4a30fa2f", "gophers"),
		adminToken: test.Token("5cf37266-3473-4006-984f-9325122678b7", "gophers"),
	}

	t.Run("postSymbol400", tests.postSymbol400)
	t.Run("postSymbol401", tests.postSymbol401)
	t.Run("getSymbol400", tests.getSymbol400)
	t.Run("crudSymbol", tests.crudSymbol)
}

// crudSymbol performs a complete test of CRUD against the api.
func (ot *SymbolTests) crudSymbol(t *testing.T) {
	sbl := ot.postSymbol201(t)
	ot.getSymbol200(t, sbl.ID)
}

// postSymbol400 validates a symbol can't be created with the endpoint
// unless a valid symbol document is submitted.
func (ot *SymbolTests) postSymbol400(t *testing.T) {
	body, err := json.Marshal(&symbol.NewSymbol{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/symbols", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new symbol can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete symbol value.", testID)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", dbtest.Success, testID)

			var got v1Web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type : %v", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type.", dbtest.Success, testID)

			fields := validate.FieldErrors{
				{Field: "symbol", Error: "symbol is a required field"},
			}
			exp := v1Web.ErrorResponse{
				Error:  "data validation error",
				Fields: fields.Fields(),
			}

			// We can't rely on the order of the field errors so they have to be
			// sorted. Tell the cmp package how to sort them.
			sorter := cmpopts.SortSlices(func(a, b validate.FieldError) bool {
				return a.Field < b.Field
			})

			if diff := cmp.Diff(got, exp, sorter); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// postSymbol401 validates an symbol can't be created unless the calling user is
// authenticated.
func (ot *SymbolTests) postSymbol401(t *testing.T) {
	body, err := json.Marshal(&symbol.NewSymbol{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/symbols", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting the Authorization header.
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new symbol can't be created unless the calling user is authenticated.")
	{
		testID := 0

		t.Logf("\tTest %d:\tWhen creating a new symbol without being authenticated.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", dbtest.Success, testID)
		}
	}
}

// postSymbol201 validates a symbol can be created with the endpoint.
func (ot *SymbolTests) postSymbol201(t *testing.T) symbol.Symbol {
	nSbl := symbol.NewSymbol{
		Symbol: "BNBUSDT",
	}

	body, err := json.Marshal(&nSbl)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/symbols", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	// This needs to be returned for other dbtest.
	var got symbol.Symbol

	t.Log("Given the need to create a new symbol with the positions endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared symbol value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 201 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 201 for the response.", dbtest.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy u.
			exp := got
			exp.Symbol = "BNBUSDT"
			exp.QuotePrecision = 8
			exp.BaseCommissionPrecision = 8
			exp.OcoAllowed = true

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}

	return got
}

// getPosition200 validates a symbol request for an existing positionID.
func (ot *SymbolTests) getSymbol200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/symbols/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a symbol that exists.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new symbol %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)

			var got symbol.Symbol
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			exp := got
			exp.ID = id
			exp.QuotePrecision = 8
			exp.BaseCommissionPrecision = 8
			exp.OcoAllowed = true

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// getSymbol400 validates a request for a malformed symbol_id.
func (ot *SymbolTests) getSymbol400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest(http.MethodGet, "/v1/symbols/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting an symbol with a malformed symbol_id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new symbol %s.", testID, id)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", dbtest.Success, testID)

			got := w.Body.String()
			exp := `{"error":"ID is not in its proper form"}`
			if got != exp {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

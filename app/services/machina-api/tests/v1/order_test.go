package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/lgarciaaco/machina-api/business/core/order"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lgarciaaco/machina-api/business/sys/validate"
	v1Web "github.com/lgarciaaco/machina-api/business/web/v1"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers"
	"github.com/lgarciaaco/machina-api/business/data/dbtest"
)

// OrderTests holds methods for each order subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type OrderTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// TestOrders is the entry point for testing order management functions.
func TestOrders(t *testing.T) {
	t.Parallel()

	test := dbtest.NewIntegration(t, c, "inttestorders")
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := OrderTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
		userToken:  test.Token("45b5fbd3-755f-4379-8f07-a58d4a30fa2f", "gophers"),
		adminToken: test.Token("5cf37266-3473-4006-984f-9325122678b7", "gophers"),
	}

	t.Run("postOrder400", tests.postOrder400)
	t.Run("postOrder401", tests.postOrder401)
	t.Run("postOrder403", tests.postOrder403)
	t.Run("postOrder404", tests.postOrder404)
	t.Run("getOrder400", tests.getOrder400)
	t.Run("getOrder403", tests.getOrder403)
	t.Run("getOrder404", tests.getOrder404)
	t.Run("crudOrder", tests.crudOrder)
}

// postOrder400 validates an order can't be created with the endpoint
// unless a valid position document is submitted. We provide a valid PositionID
// because this is validated post document validation. Tests around PositionID are
// executed further down
func (ot *OrderTests) postOrder400(t *testing.T) {
	body, err := json.Marshal(&order.NewOrder{PositionID: "75fabb5c-6c22-40c6-9236-0f8017a8e12d"})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new order can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete order value.", testID)
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
				{Field: "quantity", Error: "quantity is a required field"},
				{Field: "side", Error: "side is a required field"},
				{Field: "price", Error: "price is a required field"},
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

// postOrder401 validates an order can't be created unless the calling user is
// authenticated.
func (ot *OrderTests) postOrder401(t *testing.T) {
	body, err := json.Marshal(&order.NewOrder{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting the Authorization header.
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new order can't be created unless the calling user is authenticated.")
	{
		testID := 0

		t.Logf("\tTest %d:\tWhen creating a new order without being authenticated.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", dbtest.Success, testID)
		}
	}
}

// postOrder403 validates an order can't be created for a position that doesn't
// belong to the authenticated user
func (ot *OrderTests) postOrder403(t *testing.T) {
	body, err := json.Marshal(&order.NewOrder{PositionID: "75fabb5c-6c22-40c6-9236-0f8017a8e12d"})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.Header.Set("Authorization", "Bearer "+ot.userToken)

	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new order can't be created for a position that doesn't belong to the authenticated user.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen creating a new order for a position that doesn't belong to the authenticated user.", testID)
		{
			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 403 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 403 for the response.", dbtest.Success, testID)

			recv := w.Body.String()
			resp := `{"error":"attempted action is not allowed"}`
			if resp != recv {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// postOrder404 validates a valid position must exist to add the new order to.
// If no valid position matching the request is found, then 404 is returned.
func (ot *OrderTests) postOrder404(t *testing.T) {
	body, err := json.Marshal(&order.NewOrder{PositionID: "some_position_id"})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.Header.Set("Authorization", "Bearer "+ot.userToken)

	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a valid position must exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen creating a new order for a position that doesn't exist.", testID)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", dbtest.Success, testID)

			recv := w.Body.String()
			resp := `{"error":"position not found"}`
			if resp != recv {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// postOrder201 validates a user can be created with the endpoint.
func (ot *OrderTests) postOrder201(t *testing.T) order.Order {
	nOdr := order.NewOrder{
		PositionID: "75fabb5c-6c22-40c6-9236-0f8017a8e12d",
		Quantity:   2,
		Price:      1450,
		Side:       "SELL",
	}

	body, err := json.Marshal(&nOdr)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/orders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	// This needs to be returned for other dbtest.
	var got order.Order

	t.Log("Given the need to create a new order with the positions endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared order value.", testID)
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
			exp.Status = "opening"
			exp.Side = "SELL"
			exp.SymbolID = "5f25aa33-e294-4353-92b4-246e3bacdfc7"
			exp.Type = "MARKET"
			exp.Quantity = 2
			exp.Price = 1450

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}

	return got
}

// getOrder404 validates a request for an order that does not exist with the endpoint.
func (ot *OrderTests) getOrder404(t *testing.T) {
	id := "d0e7f962-7b40-4725-9e6f-34665fcd8794"

	r := httptest.NewRequest(http.MethodGet, "/v1/orders/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting an order with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new order %s.", testID, id)
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

// crudPosition performs a complete test of CRUD against the api.
func (ot *OrderTests) crudOrder(t *testing.T) {
	odr := ot.postOrder201(t)
	ot.getOrder200(t, odr.ID)
}

// getOrder400 validates a request for a malformed order_id.
func (ot *OrderTests) getOrder400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest(http.MethodGet, "/v1/orders/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting an order with a malformed order_id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new order %s.", testID, id)
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

// getOrder403 validates a request fetching an order for a position that does not belong to the authenticated user
func (ot *OrderTests) getOrder403(t *testing.T) {
	id := "8a89e4ec-4b51-44ac-be9f-f15910d93682"

	r := httptest.NewRequest(http.MethodGet, "/v1/orders/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.userToken)
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting an order from other user.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new order %s.", testID, id)
		{
			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 403 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 403 for the response.", dbtest.Success, testID)

			got := w.Body.String()
			exp := `{"error":"attempted action is not allowed"}`
			if got != exp {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// getPosition200 validates a position request for an existing positionID.
func (ot *OrderTests) getOrder200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/orders/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ot.adminToken)
	ot.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting an order that exists.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new order %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)

			var got order.Order
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			exp := got
			exp.ID = id
			exp.Type = "MARKET"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

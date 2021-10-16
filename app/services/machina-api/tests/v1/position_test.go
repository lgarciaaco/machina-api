package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/lgarciaaco/machina-api/business/core/position"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lgarciaaco/machina-api/business/sys/validate"
	v1Web "github.com/lgarciaaco/machina-api/business/web/v1"

	"github.com/lgarciaaco/machina-api/app/services/machina-api/handlers"
	"github.com/lgarciaaco/machina-api/business/data/dbtest"
)

// PositionTests holds methods for each position subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type PositionTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// TestPositions is the entry point for testing position management functions.
func TestPositions(t *testing.T) {
	t.Parallel()

	test := dbtest.NewIntegration(t, c, "inttestpositions")
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := PositionTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
		userToken:  test.Token("45b5fbd3-755f-4379-8f07-a58d4a30fa2f", "gophers"),
		adminToken: test.Token("5cf37266-3473-4006-984f-9325122678b7", "gophers"),
	}

	t.Run("postPosition400", tests.postPosition400)
	t.Run("postPosition401", tests.postPosition401)
	t.Run("getPosition400", tests.getPosition400)
	t.Run("getPosition403", tests.getPosition403)
	t.Run("getPosition404", tests.getPosition404)
	t.Run("closePosition404", tests.closePosition404)
	t.Run("closePosition403", tests.closePosition403)
	t.Run("crudPosition", tests.crudPosition)
}

// postPosition400 validates a position can't be created with the endpoint
// unless a valid position document is submitted.
func (pt *PositionTests) postPosition400(t *testing.T) {
	body, err := json.Marshal(&position.NewPosition{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/positions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.adminToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new position can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete position value.", testID)
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
				{Field: "side", Error: "side is a required field"},
				{Field: "symbol_id", Error: "symbol_id is a required field"},
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

// postPosition401 validates a position can't be created unless the calling user is
// authenticated.
func (pt *PositionTests) postPosition401(t *testing.T) {
	body, err := json.Marshal(&position.NewPosition{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/positions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting the Authorization header.
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new position can't be created unless the calling user is authenticated.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen creating a new position without being authenticated.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", dbtest.Success, testID)
		}
	}
}

// getPosition400 validates a position request for a malformed position_id.
func (pt *PositionTests) getPosition400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest(http.MethodGet, "/v1/positions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.adminToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a position with a malformed position_id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new position %s.", testID, id)
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

// getPosition403 validates a regular user can't fetch positions from other users.
func (pt *PositionTests) getPosition403(t *testing.T) {
	t.Log("Given the need to validate regular users can't fetch other user's positions.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching other user's position.", testID)
		{
			const otherUsrPos = "75fabb5c-6c22-40c6-9236-0f8017a8e12d"
			r := httptest.NewRequest(http.MethodGet, "/v1/positions/"+otherUsrPos, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)
			pt.app.ServeHTTP(w, r)

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

		testID = 1
		t.Logf("\tTest %d:\tWhen fetching user's own position.", testID)
		{
			const usrOwnPos = "989efd27-3da5-43ba-abf5-89dabcf4d298"
			r := httptest.NewRequest(http.MethodGet, "/v1/positions/"+usrOwnPos, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)
			pt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)
		}
	}
}

// getPosition404 validates a position request for a position that does not exist with the endpoint.
func (pt *PositionTests) getPosition404(t *testing.T) {
	id := "c50a5d66-3c4d-453f-af3f-bc960ed1a503"

	r := httptest.NewRequest(http.MethodGet, "/v1/positions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.adminToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a position with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new position %s.", testID, id)
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

// closePosition404 validates closing a position that does not exist.
func (pt *PositionTests) closePosition404(t *testing.T) {
	id := "3097c45e-780a-421b-9eae-43c2fda2bf14"

	r := httptest.NewRequest(http.MethodDelete, "/v1/positions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.adminToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a positions that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new position %s.", testID, id)
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

// closePosition403 validates a position can only be closed by its owner
// or an admin.
func (pt *PositionTests) closePosition403(t *testing.T) {
	t.Log("Given the need to validate regular users can't close other user's positions.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen closing other user's position.", testID)
		{
			const otherUsrPos = "75fabb5c-6c22-40c6-9236-0f8017a8e12d"
			r := httptest.NewRequest(http.MethodDelete, "/v1/positions/"+otherUsrPos, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)
			pt.app.ServeHTTP(w, r)

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

		testID = 1
		t.Logf("\tTest %d:\tWhen closing user's own position.", testID)
		{
			const usrOwnPos = "989efd27-3da5-43ba-abf5-89dabcf4d298"
			r := httptest.NewRequest(http.MethodDelete, "/v1/positions/"+usrOwnPos, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)
			pt.app.ServeHTTP(w, r)

			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", dbtest.Success, testID)
		}
	}
}

// crudPosition performs a complete test of CRUD against the api.
func (pt *PositionTests) crudPosition(t *testing.T) {
	nu := pt.postPosition201(t)
	defer pt.closePosition204(t, nu.ID)

	pt.getPosition200(t, nu.ID)
}

// postPosition201 validates a position can be created with the endpoint.
func (pt *PositionTests) postPosition201(t *testing.T) position.Position {
	nPos := position.NewPosition{
		SymbolID: "125240c0-7f7f-4d0f-b30d-939fd93cf027",
		Side:     "SELL",
	}

	body, err := json.Marshal(&nPos)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/positions", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	// This needs to be returned for other dbtest.
	var got position.Position

	t.Log("Given the need to create a new position with the positions endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared position value.", testID)
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
			exp.Status = position.OPEN

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}

	return got
}

// getPosition200 validates a position request for an existing positionID.
func (pt *PositionTests) getPosition200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/positions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a position that exists.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new position %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", dbtest.Success, testID)

			var got position.Position
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			exp := got
			exp.ID = id
			exp.User = "User Gopher"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", dbtest.Success, testID)
		}
	}
}

// closePosition204 validates closing a position that does exist.
func (pt *PositionTests) closePosition204(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodDelete, "/v1/positions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to close a position with the positions endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the modified position value.", testID)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", dbtest.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/positions/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)
			pt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve : %v", dbtest.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve.", dbtest.Success, testID)

			var rp position.Position
			if err := json.NewDecoder(w.Body).Decode(&rp); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", dbtest.Failed, testID, err)
			}

			if rp.Status != position.CLOSED {
				t.Fatalf("\t%s\tTest %d:\tShould see an updated status : got %q want %q", dbtest.Failed, testID, rp.Status, position.CLOSED)
			}
			t.Logf("\t%s\tTest %d:\tShould see an updated status.", dbtest.Success, testID)

			t.Logf("\t%s\tTest %d:\tShould not affect other fields.", dbtest.Success, testID)
		}
	}
}

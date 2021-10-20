package position

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lgarciaaco/machina-api/foundation/docker"

	"github.com/google/go-cmp/cmp"

	"github.com/lgarciaaco/machina-api/business/data/dbschema"
	"github.com/lgarciaaco/machina-api/business/data/dbtest"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func TestPosition(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testpos")
	t.Cleanup(teardown)

	core := NewCore(log, db)

	t.Log("Given the need to work with Positions records.")
	{
		testID := 0

		t.Logf("\tTest %d:\tWhen handling a single Order.", testID)
		{
			ctx := context.Background()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			nPos := NewPosition{
				SymbolID: "125240c0-7f7f-4d0f-b30d-939fd93cf027", // SymbolID is seeded in db
				UserID:   "45b5fbd3-755f-4379-8f07-a58d4a30fa2f", // UserID is seeded in db
				Side:     "BUY",
			}
			pos, err := core.Create(ctx, nPos, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create position : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create position.", dbtest.Success, testID)

			clsPos, err := core.QueryByID(ctx, pos.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve position by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve position by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(pos.ID, clsPos.ID); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same position. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same position.", dbtest.Success, testID)

			err = core.Close(ctx, pos.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to close position : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create position.", dbtest.Success, testID)

			clsPos, err = core.QueryByID(ctx, pos.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve position by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve position by ID.", dbtest.Success, testID)

			if clsPos.Status != CLOSED {
				t.Fatalf("\t%s\tTest %d:\tShould get CLOSED status for position but got %s.", dbtest.Failed, testID, clsPos.Status)
			}
			t.Logf("\t%s\tTest %d:\tShould get CLOSED status for position.", dbtest.Success, testID)
		}
	}
}

func TestPagingPosition(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testpospg")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	core := NewCore(log, db)

	t.Log("Given the need to page through Positions records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 positions.", testID)
		{
			ctx := context.Background()

			pos1, err := core.Query(ctx, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve positions for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve positions for page 1.", dbtest.Success, testID)

			if len(pos1) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single position : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single position.", dbtest.Success, testID)

			pos2, err := core.QueryByUser(ctx, 2, 1, "45b5fbd3-755f-4379-8f07-a58d4a30fa2f")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve positions for page 2 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve positions for page 2.", dbtest.Success, testID)

			if len(pos2) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single position : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single position.", dbtest.Success, testID)

			if pos1[0].ID == pos2[0].ID {
				t.Logf("\t\tTest %d:\tPosition1: %v", testID, pos1[0].ID)
				t.Logf("\t\tTest %d:\tPosition1: %v", testID, pos2[0].ID)
				t.Fatalf("\t%s\tTest %d:\tShould have different positions : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different positions.", dbtest.Success, testID)
		}
	}
}

func TestPositionOrders(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testposodr")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	ID := "891c178b-3dbf-4f99-a8f0-99a86cb578b7"
	core := NewCore(log, db)

	t.Log("Given the need to retrieve Positions records with Orders.")
	{
		testID := 0
		seedPos, err := core.QueryByID(ctx, ID)
		if err != nil {
			t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve position by ID: %s.", dbtest.Failed, testID, err)
		}
		t.Logf("\t%s\tTest %d:\tShould be able to retrieve position by ID.", dbtest.Success, testID)

		odrs := seedPos.Orders
		if len(odrs) != 2 {
			t.Fatalf("\t%s\tTest %d:\tShould get two orders but got %d.", dbtest.Failed, testID, len(odrs))
		}
		t.Logf("\t%s\tTest %d:\tShould get two orders and got %d.", dbtest.Success, testID, len(odrs))

		ts, _ := time.Parse("2006-01-02T15:04:05.999999", "2019-04-01T00:00:01.000001")
		odr1 := Order{
			ID:           "ef984be8-da66-4d52-b659-591b95d92591",
			SymbolID:     "125240c0-7f7f-4d0f-b30d-939fd93cf027",
			PositionID:   "891c178b-3dbf-4f99-a8f0-99a86cb578b7",
			CreationTime: orderTime(ts),
			Price:        1500,
			Quantity:     2,
			Status:       "FILLED",
			Type:         "MARKET",
			Side:         "SELL",
		}

		if diff := cmp.Diff(odr1, odrs[0]); diff != "" {
			t.Fatalf("\t%s\tTest %d:\tShould get back the same order. Diff:\n%s", dbtest.Failed, testID, diff)
		}
		t.Logf("\t%s\tTest %d:\tShould get back the same order.", dbtest.Success, testID)
	}
}

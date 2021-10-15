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
	log, db, teardown := dbtest.NewUnit(t, c, "testposition")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	core := NewCore(log, db)

	t.Log("Given the need to work with Positions records.")
	{
		testID := 0
		now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
		t.Logf("\tTest %d:\tWhen handling a single Position.", testID)
		{
			ctx := context.Background()
			nPos := Position{
				SymbolID:     "35aee552-a5bf-42a1-9d40-b6a9d4a5f342", // SymbolID is seeded in db
				UserID:       "45b5fbd3-755f-4379-8f07-a58d4a30fa2f", // UserID is seeded in db
				Side:         "BUY",
				Status:       "open",
				CreationTime: now,
			}
			pos, err := core.Create(ctx, nPos)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create position : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create position.", dbtest.Success, testID)

			saved, err := core.QueryByID(ctx, pos.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve position by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve position by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(pos.ID, saved.ID); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same position. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same position.", dbtest.Success, testID)
		}
	}
}

func TestPagingPosition(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testposition")
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

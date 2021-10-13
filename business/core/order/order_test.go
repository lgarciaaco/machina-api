package order

import (
	"context"
	"testing"
	"time"

	"github.com/lgarciaaco/machina-api/business/data/dbschema"

	"github.com/google/go-cmp/cmp"

	"github.com/lgarciaaco/machina-api/business/data/dbtest"
)

var dbc = dbtest.DBContainer{
	Image: "postgres:14-alpine",
	Port:  "5432",
	Args:  []string{"-e", "POSTGRES_PASSWORD=postgres"},
}

func TestOrder(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, dbc)
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	core := NewCore(log, db)

	t.Log("Given the need to work with Order records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Order.", testID)
		{
			ctx := context.Background()

			now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
			nOdr := Order{
				SymbolID:     "125240c0-7f7f-4d0f-b30d-939fd93cf027",
				PositionID:   "75fabb5c-6c22-40c6-9236-0f8017a8e12d",
				CreationTime: now,
				Price:        0,
				Quantity:     0,
				Status:       "FILLED",
				Type:         "MARKET",
				Side:         "SELL",
			}
			odr, err := core.Create(ctx, nOdr)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create order : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create order.", dbtest.Success, testID)

			saved, err := core.QueryByID(ctx, odr.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve order by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve order by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(odr, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same order. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same order.", dbtest.Success, testID)
		}
	}
}

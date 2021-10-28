package symbol

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lgarciaaco/machina-api/business/broker"
	"github.com/lgarciaaco/machina-api/business/data/dbschema"
	"github.com/lgarciaaco/machina-api/business/data/dbtest"
	"github.com/lgarciaaco/machina-api/foundation/docker"
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

func TestSymbol(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testsbl")
	t.Cleanup(teardown)

	core := NewCore(log, db, broker.TestBinance{})

	t.Log("Given the need to work with SymbolID records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single SymbolID.", testID)
		{
			ctx := context.Background()
			symbol := "BNBBTC"
			sbl, err := core.Create(ctx, NewSymbol{Symbol: symbol})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create symbol : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create symbol.", dbtest.Success, testID)

			saved, err := core.QueryByID(ctx, sbl.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve symbol by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve symbol by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(sbl, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same symbol. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same symbol.", dbtest.Success, testID)

			saved, err = core.QueryBySymbol(ctx, symbol)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve symbol by symbol: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve symbol by symbol.", dbtest.Success, testID)

			if diff := cmp.Diff(sbl, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same symbol. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same symbol.", dbtest.Success, testID)
		}
	}
}

func TestPagingSymbol(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testsblpg")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	candle := NewCore(log, db, broker.TestBinance{})

	t.Log("Given the need to page through Symbols records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 symbols.", testID)
		{
			ctx := context.Background()

			sbl1, err := candle.Query(ctx, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve symbols for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve symbol for page 1.", dbtest.Success, testID)

			if len(sbl1) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single symbol : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single symbol.", dbtest.Success, testID)

			sbl2, err := candle.Query(ctx, 2, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve symbols for page 2 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve symbols for page 2.", dbtest.Success, testID)

			if len(sbl2) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single symbol : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single symbol.", dbtest.Success, testID)

			if sbl1[0].ID == sbl2[0].ID {
				t.Logf("\t\tTest %d:\tSymbol1: %v", testID, sbl1[0].ID)
				t.Logf("\t\tTest %d:\tSymbol2: %v", testID, sbl2[0].ID)
				t.Fatalf("\t%s\tTest %d:\tShould have different symbols : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different symbols.", dbtest.Success, testID)
		}
	}
}

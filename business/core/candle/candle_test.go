package candle

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lgarciaaco/machina-api/business/broker"

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

func TestCandle(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testcdl")
	t.Cleanup(teardown)

	core := NewCore(log, db, broker.TestBinance{})

	t.Log("Given the need to work with Candle records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Candle.", testID)
		{
			ctx := context.Background()
			nCdl := NewCandle{
				SymbolID: "125240c0-7f7f-4d0f-b30d-939fd93cf027",
				Symbol:   "ETHUSDT",
				Interval: "4h",
			}
			cdl, err := core.Create(ctx, nCdl)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create candle : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create candle.", dbtest.Success, testID)

			saved, err := core.QueryByID(ctx, cdl.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve candle by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve candle by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(cdl, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same candle. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same candle.", dbtest.Success, testID)
		}
	}
}

func TestPagingCandle(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testcdlpg")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	candle := NewCore(log, db, broker.Binance{})

	t.Log("Given the need to page through Candle records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 candles.", testID)
		{
			ctx := context.Background()

			// Query
			cdl1, err := candle.Query(ctx, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve candles for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve candles for page 1.", dbtest.Success, testID)

			if len(cdl1) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single candle : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single candle.", dbtest.Success, testID)

			cdl2, err := candle.Query(ctx, 2, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve candles for page 2 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve candles for page 2.", dbtest.Success, testID)

			if len(cdl2) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single candle : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single candle.", dbtest.Success, testID)

			if cdl1[0].CloseTime == cdl2[0].CloseTime {
				t.Logf("\t\tTest %d:\tCandle1: %v", testID, cdl1[0].CloseTime)
				t.Logf("\t\tTest %d:\tCandle2: %v", testID, cdl2[0].CloseTime)
				t.Fatalf("\t%s\tTest %d:\tShould have different candles : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different candles.", dbtest.Success, testID)

			// Query by symbol and interval
			cdl3, err := candle.QueryBySymbolAndInterval(ctx, 1, 1, "125240c0-7f7f-4d0f-b30d-939fd93cf027", "4h")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve candles by symbol and interval for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve candles by symbol and interval for page 1.", dbtest.Success, testID)

			if len(cdl1) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single candle by symbol and interval: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single candle by symbol and interval.", dbtest.Success, testID)

			cdl4, err := candle.QueryBySymbolAndInterval(ctx, 2, 1, "125240c0-7f7f-4d0f-b30d-939fd93cf027", "4h")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve candles by symbol and interval for page 2 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve candles by symbol and interval for page 2.", dbtest.Success, testID)

			if len(cdl2) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single candle by symbol and interval: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single candle by symbol and interval.", dbtest.Success, testID)

			if cdl3[0].CloseTime == cdl4[0].CloseTime {
				t.Logf("\t\tTest %d:\tCandle1: %v", testID, cdl3[0].CloseTime)
				t.Logf("\t\tTest %d:\tCandle2: %v", testID, cdl4[0].CloseTime)
				t.Fatalf("\t%s\tTest %d:\tShould have different candles : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different candles.", dbtest.Success, testID)
		}
	}
}

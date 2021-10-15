package symbol

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

func TestCandle(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testsbl")
	t.Cleanup(teardown)

	core := NewCore(log, db)

	t.Log("Given the need to work with Symbol records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Symbol.", testID)
		{
			ctx := context.Background()
			nSbl := Symbol{
				Symbol:                     "BTCETH",
				Status:                     "",
				BaseAsset:                  "BTC",
				BaseAssetPrecision:         1,
				QuoteAsset:                 "ETH",
				QuotePrecision:             1,
				BaseCommissionPrecision:    1,
				QuoteCommissionPrecision:   1,
				IcebergAllowed:             true,
				OcoAllowed:                 false,
				QuoteOrderQtyMarketAllowed: true,
				IsSpotTradingAllowed:       false,
				IsMarginTradingAllowed:     false,
			}
			cdl, err := core.Create(ctx, nSbl)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create symbol : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create symbol.", dbtest.Success, testID)

			saved, err := core.QueryByID(ctx, cdl.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve symbol by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve symbol by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(cdl, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same symbol. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same symbol.", dbtest.Success, testID)

			saved, err = core.QueryBySymbol(ctx, "BTCETH")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve symbol by symbol: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve symbol by symbol.", dbtest.Success, testID)

			if diff := cmp.Diff(cdl, saved); diff != "" {
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

	candle := NewCore(log, db)

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

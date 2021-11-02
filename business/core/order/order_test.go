package order

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lgarciaaco/machina-api/business/broker"
	"github.com/lgarciaaco/machina-api/business/broker/encode"

	"github.com/lgarciaaco/machina-api/foundation/docker"

	"github.com/lgarciaaco/machina-api/business/data/dbschema"

	"github.com/google/go-cmp/cmp"

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

func TestOrder(t *testing.T) {
	key, present := os.LookupEnv("MACHINA_BROKER_BINANCE_KEY")
	if !present {
		t.Skipf("skipping order tests, environment variable MACHINA_BROKER_BINANCE_KEY is required")
	}

	secret, present := os.LookupEnv("MACHINA_BROKER_BINANCE_SECRET")
	if !present {
		t.Skipf("skipping order tests, environment variable MACHINA_BROKER_BINANCE_KEY is required")
	}

	log, db, teardown := dbtest.NewUnit(t, c, "testodr")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	core := NewCore(log, db, broker.TestBinance{
		APIKey: key,
		Signer: &encode.Hmac{Key: []byte(secret)},
	})

	t.Log("Given the need to work with Order records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Order.", testID)
		{
			ctx := context.Background()

			now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
			nOdr := NewOrder{
				SymbolID:   "97514fb4-4ff5-4561-91d1-c8da711d8f32",
				Symbol:     "BNBUSDT",
				PositionID: "75fabb5c-6c22-40c6-9236-0f8017a8e12d",
				Quantity:   0.1,
				Side:       "SELL",
			}
			odr, err := core.Create(ctx, nOdr, now)
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

func TestPagingOrders(t *testing.T) {
	key, present := os.LookupEnv("MACHINA_BROKER_BINANCE_KEY")
	if !present {
		t.Skipf("skipping order tests, environment variable MACHINA_BROKER_BINANCE_KEY is required")
	}

	secret, present := os.LookupEnv("MACHINA_BROKER_BINANCE_SECRET")
	if !present {
		t.Skipf("skipping order tests, environment variable MACHINA_BROKER_BINANCE_KEY is required")
	}

	log, db, teardown := dbtest.NewUnit(t, c, "testodrs")
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	core := NewCore(log, db, broker.TestBinance{
		APIKey: key,
		Signer: &encode.Hmac{Key: []byte(secret)},
	})

	t.Log("Given the need to page through Positions records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 positions.", testID)
		{
			ctx := context.Background()

			odr1, err := core.Query(ctx, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve orders for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve orders for page 1.", dbtest.Success, testID)

			if len(odr1) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single orders : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single order.", dbtest.Success, testID)

			odr2, err := core.QueryByUser(ctx, 2, 1, "45b5fbd3-755f-4379-8f07-a58d4a30fa2f")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve orders for page 2 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve orders for page 2.", dbtest.Success, testID)

			if len(odr2) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single order : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single order.", dbtest.Success, testID)

			if odr1[0].ID == odr2[0].ID {
				t.Logf("\t\tTest %d:\tPosition1: %v", testID, odr1[0].ID)
				t.Logf("\t\tTest %d:\tPosition1: %v", testID, odr2[0].ID)
				t.Fatalf("\t%s\tTest %d:\tShould have different positions : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different positions.", dbtest.Success, testID)
		}
	}
}

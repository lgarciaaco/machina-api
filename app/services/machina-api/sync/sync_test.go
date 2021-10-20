package sync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lgarciaaco/machina-api/business/core/candle"
	"github.com/lgarciaaco/machina-api/business/core/symbol"

	"github.com/lgarciaaco/machina-api/business/broker"
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

func TestSync(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testsync")
	t.Cleanup(teardown)

	candleSync := CandleSynchronizer{
		Log:        log,
		Symbol:     symbol.NewCore(log, db, broker.TestBinance{}),
		Candle:     candle.NewCore(log, db, broker.TestBinance{}),
		SyncPeriod: time.Second,
	}

	candleCore := candle.NewCore(log, db, broker.TestBinance{})

	t.Log("Given the need to sync with Candle.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen sync is ran, it should seed for symbols.", testID)
		{
			ctx := context.Background()

			cdls, err := candleCore.QueryBySymbolAndInterval(ctx, 1, 1, "5f25aa33-e294-4353-92b4-246e3bacdfc7", "4h")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query candles for symbol BTCUSDT and interval 4h: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query candles for symbol BTCUSDT and interval 4h.", dbtest.Success, testID)

			if len(cdls) != 0 {
				t.Fatalf("\t%s\tTest %d:\tShould get 0 candles", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get 0 candles.", dbtest.Success, testID)

			err = candleSync.sync(ctx)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to sync: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to sync.", dbtest.Success, testID)

			cdls, err = candleCore.QueryBySymbolAndInterval(ctx, 1, 100, "5f25aa33-e294-4353-92b4-246e3bacdfc7", "1h")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query candles for symbol BTCUSDT and interval 1h: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query candles for symbol BTCUSDT and interval 1h.", dbtest.Success, testID)

			if len(cdls) != 100 {
				t.Fatalf("\t%s\tTest %d:\tShould get 100 candles for symbol BTCUSDT and interval 1h.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get 100 candles for symbol BTCUSDT and interval 1h", dbtest.Success, testID)

			cdls, err = candleCore.QueryBySymbolAndInterval(ctx, 1, 100, "5f25aa33-e294-4353-92b4-246e3bacdfc7", "2h")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query candles for symbol BTCUSDT and interval 2h: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query candles for symbol BTCUSDT and interval 2h.", dbtest.Success, testID)

			if len(cdls) != 100 {
				t.Fatalf("\t%s\tTest %d:\tShould get 100 candles for symbol BTCUSDT and interval 2h.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get 100 candles for symbol BTCUSDT and interval 2h", dbtest.Success, testID)

			cdls, err = candleCore.QueryBySymbolAndInterval(ctx, 1, 100, "5f25aa33-e294-4353-92b4-246e3bacdfc7", "4h")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query candles for symbol BTCUSDT and interval 4h: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query candles for symbol BTCUSDT and interval 4h.", dbtest.Success, testID)

			if len(cdls) != 100 {
				t.Fatalf("\t%s\tTest %d:\tShould get 100 candles for symbol BTCUSDT and interval 4h.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get 100 candles for symbol BTCUSDT and interval 4h", dbtest.Success, testID)

			// ETHUSDT has already 3 candles in 4h intervals, after sync it should have 4
			cdls, err = candleCore.QueryBySymbolAndInterval(ctx, 1, 100, "125240c0-7f7f-4d0f-b30d-939fd93cf027", "4h")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query candles for symbol ETHUSDT and interval 4h: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query candles for symbol ETHUSDT and interval 4h.", dbtest.Success, testID)

			if len(cdls) != 4 {
				t.Fatalf("\t%s\tTest %d:\tShould get 100 candles for symbol ETHUSDT and interval 4h but got %d", dbtest.Failed, testID, len(cdls))
			}
			t.Logf("\t%s\tTest %d:\tShould get 100 candles for symbol ETHUSDT and interval 4h", dbtest.Success, testID)
		}
	}
}

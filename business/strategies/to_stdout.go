package strategies

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/lgarciaaco/machina-api/business/strategies/financial"

	"github.com/lgarciaaco/machina-api/business/broker"
)

// ToStdout is a trader that writes operations to log
type ToStdout struct {
	Log       *zap.SugaredLogger
	Lot       float64
	positions []*Position
}

// Trade will wait 3 seconds for candles and exit. It prints the profit
func (t *ToStdout) Trade(done <-chan bool, candles <-chan financial.Candle, rule financial.Rule) error {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			return nil

		case <-done:
			return nil

		case c := <-candles:
			ot, at := rule.Assert(c)

			if at == financial.Open {
				if err := t.open(Position{Side: ot.String()}, c); err != nil {
					return err
				}
			}

			if at == financial.Close {
				if err := t.close(c); err != nil {
					return err
				}
			}
		}
	}
}

func (t *ToStdout) open(p Position, c financial.Candle) error {
	if len(t.positions) != 0 {
		pos := t.positions[len(t.positions)-1]
		if pos.Status != "closed" {
			// we can't leave a position open, therefore we have to force close it
			t.Log.Infof("force closing position")
			t.close(c)
		}
	}

	t.positions = append(t.positions, &Position{
		Orders: []Order{
			{
				CreationTime: c.CloseTime,
				SymbolID:     c.SymbolID,
				Price:        c.ClosePrice,
				Quantity:     t.Lot,
				Type:         broker.OrderTypeMarket,
				Side:         broker.OrderSideBuy,
			},
		},
		Symbol: c.Symbol,
		Side:   p.Side,
		Status: "open",
	})
	return nil
}

func (t *ToStdout) close(c financial.Candle) error {
	if len(t.positions) != 0 {
		pos := t.positions[len(t.positions)-1]
		if len(pos.Orders) != 1 {
			return fmt.Errorf("can't close position with len(orders) != 1")
		}

		pos.Orders = append(pos.Orders, Order{
			CreationTime: c.OpenTime,
			SymbolID:     c.SymbolID,
			Price:        c.ClosePrice,
			Quantity:     t.Lot,
			Type:         broker.OrderTypeMarket,
			Side:         broker.OrderSideSell,
		})
		pos.Status = "closed"
	}

	return nil
}

func (t *ToStdout) Profit() float64 {
	var total float64
	for i, p := range t.positions {
		if p.Status == "closed" {
			t.Log.Infof("Position [OPEN %s, CLOSE %s] : [type: %s, open: %f, close %f], Profit for operation %d: %f",
				p.Orders[0].CreationTime.Format("Mon Jan 2 15:04"),
				p.Orders[1].CreationTime.Format("Mon Jan 2 15:04"),
				p.Side,
				p.Orders[0].Price*p.Orders[0].Quantity,
				p.Orders[1].Price*p.Orders[1].Quantity,
				i, p.Profit())

			total += p.Profit()
		}
	}

	t.Log.Infof("Total profit: %f", total)
	return total
}

package financial

import (
	"go.uber.org/zap"
)

type OrderType int
type ActionType int

const (
	Buy OrderType = iota + 1
	Sell
)

const (
	Open ActionType = iota + 1
	Close
)

// Puller knows how to pull candles
type Puller interface {
	// Pull will lock the thread and wait for a signal to done to exit.
	// In the meantime, it will pull candles and write to candles. In case of
	// error, it returns
	Pull(done <-chan bool, candles chan<- Candle) error
}

// Rule has all the logic on when to start or close a position.
type Rule interface {
	Assert(candle Candle) (OrderType, ActionType)
}

// Trader holds all the information to make a trade, and how to do it.
type Trader interface {
	// Trade blocks the thread and reads candles from chan candles, it relies on  Rule.Assert(candle)
	// to start / close a position.
	Trade(done <-chan bool, candles <-chan Candle, rule Rule) error

	// Profit returns the trader's profit
	Profit() float64
}

// Runner is the interface to be implemented for each strategy. Every strategy run on the same bases:
// A strategy run passing a puller who knows how to pull candles and a trader who knows hot to trade.
// A strategy must have a Rule needed by the trader to make decisions
type Runner interface {
	// Run is the only method a strategy runner implements. In Run, spawns two one thread for
	// the puller in which the puller pulls candles, and one thread for the trader in
	// which the trader executes trades as candles arrive.
	//
	// If puller or trader return an error or naturally finish, the Run ends
	Run(done <-chan bool, puller Puller, trader Trader) error
}

// Strategy implements the Runner interface. Strategy
type Strategy struct {
	Log  *zap.SugaredLogger
	Rule Rule
}

func (mas *Strategy) Run(done <-chan bool, puller Puller, trader Trader) error {
	candles := make(chan Candle)

	// Make a channel to listen for errors coming from the puller or trader. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	pullerDone := make(chan bool)
	go func() {
		serverErrors <- puller.Pull(pullerDone, candles)
	}()

	traderDone := make(chan bool)
	go func() {
		serverErrors <- trader.Trade(traderDone, candles, mas.Rule)
	}()

	// Block thread and wait either for errors or for a shutdown signal
	select {
	case err := <-serverErrors:
		return err

	case <-done:
		close(pullerDone)
		close(traderDone)
		return nil
	}
}

func (a ActionType) String() string {
	actionTypeToString := map[ActionType]string{
		0:     "none",
		Open:  "OPEN",
		Close: "CLOSE",
	}

	return actionTypeToString[a]
}

func (o OrderType) String() string {
	orderTypeToString := map[OrderType]string{
		0:    "none",
		Buy:  "BUY",
		Sell: "SELL",
	}

	return orderTypeToString[o]
}

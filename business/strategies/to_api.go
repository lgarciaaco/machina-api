package strategies

import (
	"errors"

	v1 "github.com/lgarciaaco/machina-api/business/strategies/api/v1"

	"github.com/lgarciaaco/machina-api/business/strategies/financial"

	"go.uber.org/zap"
)

var (
	ErrPositionMissing   = errors.New("position not found")
	ErrNotSufficientFund = errors.New("no enough fund to execute position")
	ErrPositionOpened    = errors.New("an open position already exists")
)

type TradingPair struct {
	Symbol   string
	Interval string
	Fast     int
	Slow     int
	Warming  int
}

type ToAPI struct {
	Log    *zap.SugaredLogger
	Budget financial.Budget
	Client *v1.Client

	currentPosition *Position
}

func (s *ToAPI) Trade(done <-chan bool, candles <-chan financial.Candle, rule financial.Rule) error {
	s.Log.Infof("trade : starting to trade with budget %s", s.Budget)
	for {
		select {
		case <-done:
			s.Log.Infof("trader : gracefully shutting down trader")
			return nil

		case c := <-candles:
			ot, at := rule.Assert(c)

			if at == financial.Open {
				pos, err := s.open(Position{
					Side:     ot.String(),
					SymbolID: c.SymbolID,
				}, c)
				if err != nil {
					s.Log.Errorf("trader : error : unable to open position [%s, %f]", ot.String(), c.ClosePrice)
					return err
				}

				s.Log.Infof("trader : opened : %s [%s, %f, %f] | budget: %s", pos.ID, pos.Side, c.ClosePrice, pos.Orders[0].Quantity, s.Budget)
			}

			if at == financial.Close {
				pos, err := s.close(c)
				if err != nil {
					s.Log.Errorf("trader : error : unable to close position [%s, %f]", ot.String(), c.ClosePrice)
					return err
				}

				if pos != nil {
					s.Log.Infof("trader : closed : %s [%s, %f, %f] | budget: %s", pos.ID, pos.Side, c.ClosePrice, pos.Orders[1].Quantity, s.Budget)
				}
			}
		}
	}
}

func (s *ToAPI) close(c financial.Candle) (p *Position, err error) {
	// A position needs to exist in order to be closed
	if s.currentPosition == nil {
		pos, err := s.retrieveCurrentPosition()
		if err != nil {
			return nil, err
		}

		if pos == nil {
			s.Log.Infof("trader : close : unable to find an opened position, skipping this iteration")
			return nil, nil
		}

		s.currentPosition = pos
	}

	// If for any reason current position is already closed
	if s.currentPosition.Status == "CLOSED" {
		s.Log.Infof("trader : close : unable to close a closed position, skipping this iteration, pos[%s]", s.currentPosition.ID)
		return nil, nil
	}

	// Run the position by the budget and this will validate and set the lot
	// is validation passes
	if err := s.Budget.Close(financial.Position{
		Side:   s.currentPosition.Side,
		Status: s.currentPosition.Status,
	}, c); err != nil {
		return s.currentPosition, ErrNotSufficientFund
	}

	cOdr, err := s.currentPosition.close()
	if err != nil {
		return nil, err
	}

	nOdr := v1.NewOrder{
		PositionID: s.currentPosition.ID,
		Quantity:   cOdr.Quantity,
		Side:       cOdr.Side,
	}

	_, err = s.Client.CreateOrder(nOdr)
	if err != nil {
		return nil, err
	}

	// call the trader api and close the position
	v1Pos, err := s.Client.ClosePosition(s.currentPosition.ID)
	if err != nil {
		return nil, err
	}

	s.currentPosition = nil
	return toPosition(v1Pos), nil
}

// open creates a position with the pair this strategy is trading on, lot
// and side
func (s *ToAPI) open(p Position, c financial.Candle) (pos *Position, err error) {
	// We can only open a position if there is no position already open
	if s.currentPosition != nil {
		s.Log.Infof("trader : open : force closing previous position [%s, %s] | budget: %s",
			s.currentPosition.ID, s.currentPosition.Side,
			s.Budget)
		s.close(c)
	}

	// Build the payload based on what we got from the strategy
	payload := v1.NewPosition{
		Side:     p.Side,
		SymbolID: p.SymbolID,
	}

	// Run the new order by the budget and this will validate and set the lot
	// is validation passes
	q, err := s.Budget.Open(financial.Position{Side: p.Side, Status: p.Status}, c)
	if err != nil {
		return pos, ErrNotSufficientFund
	}

	v1Pos, err := s.Client.CreatePosition(payload, q)
	if err != nil {
		return nil, err
	}

	s.currentPosition = toPosition(v1Pos)
	return s.currentPosition, nil
}

func (s *ToAPI) Profit() float64 {
	v1Poss, err := s.Client.ListPositions()
	if err != nil {
		s.Log.Infof("trader : unable to fetch v1Poss from api")
		return 0.0
	}

	var profit float64
	for _, p := range v1Poss {
		pos := toPosition(&p)
		profit += pos.Profit()
	}
	return profit
}

// In case of a restart or new start, we need to get from api the last opened
// position we were working on.
func (s *ToAPI) retrieveCurrentPosition() (pos *Position, err error) {
	v1Pos, err := s.Client.RetrievePosition()
	if err != nil {
		return nil, err
	}

	if v1Pos != nil {
		s.currentPosition = toPosition(v1Pos)
	}

	return s.currentPosition, nil
}

package financial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixBudget_Close(t *testing.T) {
	type fields struct {
		Budget BaseBudget
	}

	type args struct {
		p Position
		c Candle
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantBase float64
		wantAlt  float64
	}{
		{
			name:     "Closing a SELL position should increase Base and decrease Alt",
			fields:   fields{Budget: BaseBudget{Base: 10, Alt: 500, Lot: 2}},
			args:     args{p: Position{Side: "SELL"}, c: Candle{ClosePrice: 50}},
			wantAlt:  400,
			wantBase: 12,
			wantErr:  false,
		},
		{
			name:     "Closing a SELL position with not enough funds should error",
			fields:   fields{Budget: BaseBudget{Base: 10, Alt: 500, Lot: 2}},
			args:     args{p: Position{Side: "SELL"}, c: Candle{ClosePrice: 350}},
			wantAlt:  -200,
			wantBase: 12,
			wantErr:  true,
		},
		{
			name:     "Closing a BUY position should decrease Base and increase Alt",
			fields:   fields{Budget: BaseBudget{Base: 10, Alt: 500, Lot: 2}},
			args:     args{p: Position{Side: "BUY"}, c: Candle{ClosePrice: 50}},
			wantAlt:  600,
			wantBase: 8,
			wantErr:  false,
		},
		{
			name:     "Closing a BUY position with not enough funds should error",
			fields:   fields{Budget: BaseBudget{Base: 1, Alt: 500, Lot: 2}},
			args:     args{p: Position{Side: "BUY"}, c: Candle{ClosePrice: 50}},
			wantAlt:  600,
			wantBase: -1,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &FixBudget{
				BaseBudget: tt.fields.Budget,
			}

			err := b.Close(tt.args.p, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("close() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.wantBase, b.Base)
			assert.Equal(t, tt.wantAlt, b.Alt)
		})
	}
}

func TestFixBudget_Open(t *testing.T) {
	type fields struct {
		Budget BaseBudget
	}

	type args struct {
		p Position
		c Candle
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantBase float64
		wantAlt  float64
	}{
		{
			name:     "Opening a SELL position should decrease Base and increase Alt",
			fields:   fields{Budget: BaseBudget{Base: 10, Alt: 500, Lot: 2}},
			args:     args{p: Position{Side: "SELL"}, c: Candle{ClosePrice: 50}},
			wantAlt:  600,
			wantBase: 8,
			wantErr:  false,
		},
		{
			name:     "Opening a SELL position with not enough funds should error",
			fields:   fields{Budget: BaseBudget{Base: 10, Alt: 500, Lot: 11}},
			args:     args{p: Position{Side: "SELL"}, c: Candle{ClosePrice: 50}},
			wantAlt:  1050,
			wantBase: -1,
			wantErr:  true,
		},
		{
			name:     "Opening a BUY position should decrease Base and increase Alt",
			fields:   fields{Budget: BaseBudget{Base: 10, Alt: 500, Lot: 2}},
			args:     args{p: Position{Side: "BUY"}, c: Candle{ClosePrice: 50}},
			wantAlt:  400,
			wantBase: 12,
			wantErr:  false,
		},
		{
			name:     "Opening a BUY position with not enough funds should error",
			fields:   fields{Budget: BaseBudget{Base: 1, Alt: 100, Lot: 2}},
			args:     args{p: Position{Side: "BUY"}, c: Candle{ClosePrice: 150}},
			wantAlt:  -200,
			wantBase: 3,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &FixBudget{
				BaseBudget: tt.fields.Budget,
			}

			_, err := b.Open(tt.args.p, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("close() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.wantBase, b.Base)
			assert.Equal(t, tt.wantAlt, b.Alt)
		})
	}
}

package financial

type Cross int

const (
	None Cross = iota
	Up
	Down
)

// Indicator In technical analysis in finance, a technical indicator is a mathematical calculation based on historic price,
// volume, or Open interest information that aims to forecast financial market direction.
type Indicator interface {
	// Calculate the current value of the indicator given the set of Candles
	Calculate() float64
	// Value  of the indicator
	Value() float64
	// Previous value of the indicator
	Previous() float64
	// Position Current position in the time series
	Position() int
}

// MovingAverageIndicator moving averages are the most common indicator in technical analysis. The moving average itself may also be the most
// important indicator, as it serves as the foundation of countless others, such as the Moving Average Convergence Divergence (MACD).
type MovingAverageIndicator interface {
	Indicator
	// Uptrend Indicates whether this indicator is in uptrend
	Uptrend() bool
	// Downtrend Indicates whether this indicator is in uptrend
	Downtrend() bool
	// CrossOver Indicates whether price and indicator value intersect
	CrossOver() (r Cross)
}

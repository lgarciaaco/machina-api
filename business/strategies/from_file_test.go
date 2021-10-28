package strategies

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_load(t *testing.T) {
	for _, i := range []string{"5m", "15m", "30m", "1h", "4h"} {
		data, err := LoadTimeSeriesFromFile("./../../zarf/binance/" + i + ".json")
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	}
}

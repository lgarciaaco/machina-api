package binance

type Symbol struct {
	ID                         string `json:"symbol_id"`
	Symbol                     string `json:"symbol"`
	Status                     string `json:"status"`
	BaseAsset                  string `json:"baseAsset"`
	BaseAssetPrecision         int    `json:"baseAssetPrecision"`
	QuoteAsset                 string `json:"quoteAsset"`
	QuotePrecision             int    `json:"quotePrecision"`
	BaseCommissionPrecision    int    `json:"baseCommissionPrecision"`
	QuoteCommissionPrecision   int    `json:"quoteCommissionPrecision"`
	IcebergAllowed             bool   `json:"icebergAllowed"`
	OcoAllowed                 bool   `json:"ocoAllowed"`
	QuoteOrderQtyMarketAllowed bool   `json:"quoteOrderQtyMarketAllowed"`
	IsSpotTradingAllowed       bool   `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed     bool   `json:"isMarginTradingAllowed"`
}

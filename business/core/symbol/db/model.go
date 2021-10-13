package db

// Symbol is a function whereby you have two different currencies that can be traded between one another.
// When buying and selling a cryptocurrency, it is often swapped with local currency. For example,
// If you're looking to buy or sell Bitcoin with U.S. Dollar, the trading pair would be BTC to USD
type Symbol struct {
	ID                         string `db:"symbol_id"`
	Symbol                     string `db:"symbol"`
	Status                     string `db:"status"`
	BaseAsset                  string `db:"base_asset"`
	BaseAssetPrecision         int    `db:"base_asset_precision"`
	QuoteAsset                 string `db:"quote_asset"`
	QuotePrecision             int    `db:"quote_precision"`
	BaseCommissionPrecision    int    `db:"base_commission_precision"`
	QuoteCommissionPrecision   int    `db:"quote_commission_precision"`
	IcebergAllowed             bool   `db:"iceberg_allowed"`
	OcoAllowed                 bool   `db:"oco_allowed"`
	QuoteOrderQtyMarketAllowed bool   `db:"quote_order_qty_market_allowed"`
	IsSpotTradingAllowed       bool   `db:"is_spot_trading_allowed"`
	IsMarginTradingAllowed     bool   `db:"is_margin_trading_allowed"`
}

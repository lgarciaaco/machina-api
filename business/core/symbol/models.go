package symbol

import (
	"github.com/lgarciaaco/machina-api/business/core/symbol/db"
)

type Symbol struct {
	ID                         string `json:"symbol_id"`
	Symbol                     string `json:"symbol"`
	Status                     string `json:"status"`
	BaseAsset                  string `json:"base_asset"`
	BaseAssetPrecision         int    `json:"base_asset_precision"`
	QuoteAsset                 string `json:"quote_asset"`
	QuotePrecision             int    `json:"quote_precision"`
	BaseCommissionPrecision    int    `json:"base_commission_precision"`
	QuoteCommissionPrecision   int    `json:"quote_commission_precision"`
	IcebergAllowed             bool   `json:"iceberg_allowed"`
	OcoAllowed                 bool   `json:"oco_allowed"`
	QuoteOrderQtyMarketAllowed bool   `json:"quote_order_qty_market_allowed"`
	IsSpotTradingAllowed       bool   `json:"is_spot_trading_allowed"`
	IsMarginTradingAllowed     bool   `json:"is_margin_trading_allowed"`
}

type NewSymbol struct {
	Symbol string `json:"symbol" validate:"required"`
}

func toSymbol(dbSbl db.Symbol) Symbol {
	pc := (*Symbol)(&dbSbl)
	return *pc
}

func toSymbolSlice(dbSbls []db.Symbol) []Symbol {
	sbls := make([]Symbol, len(dbSbls))
	for i, dbCdl := range dbSbls {
		sbls[i] = toSymbol(dbCdl)
	}
	return sbls
}

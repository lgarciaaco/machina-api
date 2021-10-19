package binance

// Order defines a trading order
type Order struct {
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Type     string  `json:"type"`
	Quantity float64 `json:"quantity"`
}

// OrderResponse defines the response from broker api when an order is created
type OrderResponse struct {
	Symbol              string  `json:"symbol"`
	OrderID             int     `json:"orderId"`
	OrderListID         int     `json:"orderListId"`
	ClientOrderID       string  `json:"clientOrderId"`
	TransactTime        int64   `json:"transactTime"`
	Price               float64 `json:"price,string"`
	OrigQty             float64 `json:"origQty,string"`
	ExecutedQty         float64 `json:"executedQty,string"`
	CummulativeQuoteQty string  `json:"cummulativeQuoteQty"`
	Status              string  `json:"status"`
	TimeInForce         string  `json:"timeInForce"`
	Type                string  `json:"type"`
	Side                string  `json:"side"`
	Fills               []Fill  `json:"fills"`
}

type Fill struct {
	Price           float64 `json:"price,string"`
	Qty             float64 `json:"qty,string"`
	Commission      float64 `json:"commission,string"`
	CommissionAsset string  `json:"commissionAsset"`
}

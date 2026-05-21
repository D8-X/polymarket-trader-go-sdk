package models

type PriceRequest struct {
	TokenID string `json:"token_id"`
	Side    string `json:"side"`
}

type SpreadRequest struct {
	TokenID string `json:"token_id"`
}

type LastTradePrice struct {
	Price   string `json:"price"`
	Side    string `json:"side"`
	TokenID string `json:"token_id,omitempty"`
}

type PriceHistoryEntry struct {
	Timestamp int64   `json:"t"`
	Price     float64 `json:"p"`
}

type PricesHistoryParams struct {
	Market   string
	Interval string
	StartTS  int64
	EndTS    int64
	Fidelity int
}

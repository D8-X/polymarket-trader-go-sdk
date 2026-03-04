package polytrade

import "encoding/json"

type Fill struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
	Cost  float64 `json:"cost"`
}

type FillRoute struct {
	TokenID      string  `json:"token_id"`
	Outcome      string  `json:"outcome"`
	BookSide     string  `json:"book_side"`
	AvgPrice     float64 `json:"avg_price"`
	BestPrice    float64 `json:"best_price"`
	WorstPrice   float64 `json:"worst_price"`
	FilledSize   float64 `json:"filled_size"`
	TotalCost    float64 `json:"total_cost"`
	BookSize     float64 `json:"book_size"`
	BookDepth    int     `json:"book_depth"`
	BookConsumed float64 `json:"book_consumed"`
	Fillable     bool    `json:"fillable"`
	Fills        []Fill  `json:"fills"`
}

type SportsQuote struct {
	ConditionID     string          `json:"contract_id"`
	Team           string          `json:"team"`
	Side           string          `json:"side"`
	Size           float64         `json:"requested_size"`
	Position       float64         `json:"position"`
	EffectivePrice float64         `json:"effective_price"`
	MidPrice       float64         `json:"mid_price"`
	MarketWidth    float64         `json:"market_width"`
	Slippage       float64         `json:"slippage"`
	Fillable       bool            `json:"fillable"`
	Route          string          `json:"route"`
	Direct         *FillRoute `json:"direct"`
	Opposite       *FillRoute `json:"opposite"`
	Confidence     int             `json:"confidence"`
	MarketClosed   bool            `json:"market_closed"`
}

type L2Credentials struct {
	Address    string `json:"address,omitempty"`
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

type L2Headers struct {
	Address    string
	APIKey     string
	Passphrase string
	Signature  string
	Timestamp  string
}

type OrderFields struct {
	Salt          int64  `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBps    string `json:"feeRateBps"`
	Side          string `json:"side"`
	SignatureType int    `json:"signatureType"`
	Signature     string `json:"signature"`
	sideNumeric   int
}

type SignedOrder struct {
	Order     OrderFields `json:"order"`
	Owner     string      `json:"owner"`
	OrderType string      `json:"orderType"`
}

type PlaceOrderResponse struct {
	OrderID  string `json:"orderID"`
	Success  bool   `json:"success"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

type OrderStatus struct {
	ID           string      `json:"id"`
	Status       string      `json:"status"`
	SizeMatched  json.Number `json:"size_matched"`
	OriginalSize json.Number `json:"original_size"`
	Price        json.Number `json:"price"`
	Side         string      `json:"side"`
	TokenID      string      `json:"asset_id"`
}

type BalanceEntry struct {
	AssetID string  `json:"asset_id"`
	Balance float64 `json:"balance"`
}

type PositionEntry struct {
	Asset       string  `json:"asset"`
	ConditionID string  `json:"conditionId"`
	Size        float64 `json:"size"`
	AvgPrice    float64 `json:"avgPrice"`
	CurPrice    float64 `json:"curPrice"`
	Outcome     string  `json:"outcome"`
	Title       string  `json:"title"`
}

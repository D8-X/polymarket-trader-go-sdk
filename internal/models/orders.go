package models

type OrderFields struct {
	Salt          int64  `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Timestamp     string `json:"timestamp"`
	Metadata      string `json:"metadata"`
	Builder       string `json:"builder"`
	Side          string `json:"side"`
	SignatureType int    `json:"signatureType"`
	Signature     string `json:"signature"`
	SideNumeric   int    `json:"-"`
}

type SignedOrder struct {
	Order     OrderFields `json:"order"`
	Owner     string      `json:"owner"`
	OrderType string      `json:"orderType"`
	PostOnly  bool        `json:"postOnly,omitempty"`
	DeferExec bool        `json:"deferExec,omitempty"`
}

type PlaceOrderResponse struct {
	OrderID           string   `json:"orderID"`
	Success           bool     `json:"success"`
	ErrorMsg          string   `json:"errorMsg,omitempty"`
	TransactionHashes []string `json:"transactionsHashes,omitempty"`
	TradeIDs          []string `json:"tradeIDs,omitempty"`
	Status            string   `json:"status,omitempty"`
	TakingAmount      string   `json:"takingAmount,omitempty"`
	MakingAmount      string   `json:"makingAmount,omitempty"`
}

type OrderStatus struct {
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	SizeMatched      string   `json:"size_matched"`
	OriginalSize     string   `json:"original_size"`
	Price            string   `json:"price"`
	Side             string   `json:"side"`
	TokenID          string   `json:"asset_id"`
	Market           string   `json:"market,omitempty"`
	Outcome          string   `json:"outcome,omitempty"`
	OrderType        string   `json:"order_type,omitempty"`
	MakerAddress     string   `json:"maker_address,omitempty"`
	Owner            string   `json:"owner,omitempty"`
	Expiration       string   `json:"expiration,omitempty"`
	AssociatedTrades []string `json:"associate_trades,omitempty"`
	CreatedAt        int64    `json:"created_at,omitempty"`
}

type CancelResponse struct {
	Canceled    []string          `json:"canceled"`
	NotCanceled map[string]string `json:"not_canceled"`
}

type OrderScoringResult struct {
	Scoring bool `json:"scoring"`
}

type Trade struct {
	ID              string       `json:"id"`
	TakerOrderID    string       `json:"taker_order_id"`
	Market          string       `json:"market"`
	AssetID         string       `json:"asset_id"`
	Side            string       `json:"side"`
	Size            string       `json:"size"`
	FeeRateBps      string       `json:"fee_rate_bps"`
	Price           string       `json:"price"`
	Status          string       `json:"status"`
	MatchTime       string       `json:"match_time"`
	MatchTimeNano   string       `json:"match_time_nano"`
	LastUpdate      string       `json:"last_update"`
	Outcome         string       `json:"outcome"`
	BucketIndex     int          `json:"bucket_index"`
	Owner           string       `json:"owner"`
	MakerAddress    string       `json:"maker_address"`
	TransactionHash string       `json:"transaction_hash"`
	TraderSide      string       `json:"trader_side"`
	ErrMsg          *string      `json:"err_msg"`
	MakerOrders     []MakerOrder `json:"maker_orders"`
}

type MakerOrder struct {
	OrderID       string `json:"order_id"`
	Owner         string `json:"owner"`
	MakerAddress  string `json:"maker_address"`
	MatchedAmount string `json:"matched_amount"`
	Price         string `json:"price"`
	FeeRateBps    string `json:"fee_rate_bps"`
	AssetID       string `json:"asset_id"`
	Outcome       string `json:"outcome"`
	Side          string `json:"side"`
}

type ClosePositionOpts struct {
	OrderType string
	TickSize  string
	PostOnly  bool
	DeferExec bool
}

package polytrade

import (
	"encoding/json"
	"time"
)

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

// OrderFields is the CLOB order wire body. Expiration and Side (string form)
// are sent to the API but are not part of the EIP-712 signed struct.
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
	sideNumeric   int
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

type PollOpts struct {
	Interval time.Duration
	Timeout  time.Duration
}

type PollResult struct {
	OrderID     string
	Status      *OrderStatus
	PlaceStatus string
	Err         error
}

const (
	RelayerStateNew       = "STATE_NEW"
	RelayerStateExecuted  = "STATE_EXECUTED"
	RelayerStateMined     = "STATE_MINED"
	RelayerStateConfirmed = "STATE_CONFIRMED"
	RelayerStateInvalid   = "STATE_INVALID"
	RelayerStateFailed    = "STATE_FAILED"
)

type nonceResponse struct {
	Nonce string `json:"nonce"`
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

type CancelResponse struct {
	Canceled    []string          `json:"canceled"`
	NotCanceled map[string]string `json:"not_canceled"`
}

type PaginatedResponse[T any] struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor"`
	Count      int    `json:"count"`
	Data       []T    `json:"data"`
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

type OrderBook struct {
	Market         string           `json:"market"`
	AssetID        string           `json:"asset_id"`
	Timestamp      string           `json:"timestamp"`
	Hash           string           `json:"hash"`
	Bids           []OrderBookLevel `json:"bids"`
	Asks           []OrderBookLevel `json:"asks"`
	MinOrderSize   string           `json:"min_order_size"`
	TickSize       string           `json:"tick_size"`
	NegRisk        bool             `json:"neg_risk"`
	LastTradePrice string           `json:"last_trade_price"`
}

type OrderBookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

type BalanceAllowanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances"`
}

// ClobMarketInfo is the per-market metadata returned by the CLOB.
// Effective match-time fee rate = FeeDetails.Rate * 10^-FeeDetails.Exponent;
// applied only to takers when FeeDetails.TakerOnly is true.
type ClobMarketInfo struct {
	MinTickSize  json.Number           `json:"mts"`
	MinOrderSize json.Number           `json:"mos"`
	FeeDetails   ClobMarketFeeDetails  `json:"fd"`
	Tokens       []ClobMarketInfoToken `json:"t"`
	RFQEnabled   bool                  `json:"rfqe"`
}

type ClobMarketFeeDetails struct {
	Rate      float64 `json:"r"`
	Exponent  int     `json:"e"`
	TakerOnly bool    `json:"to"`
}

type ClobMarketInfoToken struct {
	TokenID string `json:"t"`
	Outcome string `json:"o"`
}

type PriceLevel struct {
	Price float64
	Size  float64
}

type SweepLevel struct {
	Price    float64
	Size     float64
	Slippage float64
}

type SweepEstimate struct {
	Levels     []SweepLevel
	Side       string
	BestPrice  float64
	WorstPrice float64
	AvgPrice   float64
	TotalSize  float64
}

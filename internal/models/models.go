package models

import (
	"encoding/json"
	"time"
)

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

type MarketRewardsRate struct {
	AssetAddress     string  `json:"asset_address"`
	RewardsDailyRate float64 `json:"rewards_daily_rate"`
}

type MarketRewards struct {
	Rates     []MarketRewardsRate `json:"rates"`
	MinSize   float64             `json:"min_size"`
	MaxSpread float64             `json:"max_spread"`
}

type MarketToken struct {
	TokenID string  `json:"token_id"`
	Outcome string  `json:"outcome"`
	Price   float64 `json:"price"`
	Winner  bool    `json:"winner"`
}

type MarketInfo struct {
	EnableOrderBook         bool          `json:"enable_order_book"`
	Active                  bool          `json:"active"`
	Closed                  bool          `json:"closed"`
	Archived                bool          `json:"archived"`
	AcceptingOrders         bool          `json:"accepting_orders"`
	AcceptingOrderTimestamp string        `json:"accepting_order_timestamp"`
	MinimumOrderSize        float64       `json:"minimum_order_size"`
	MinimumTickSize         float64       `json:"minimum_tick_size"`
	ConditionID             string        `json:"condition_id"`
	QuestionID              string        `json:"question_id"`
	Question                string        `json:"question"`
	Description             string        `json:"description"`
	MarketSlug              string        `json:"market_slug"`
	EndDateISO              string        `json:"end_date_iso"`
	GameStartTime           string        `json:"game_start_time"`
	SecondsDelay            int           `json:"seconds_delay"`
	FPMM                    string        `json:"fpmm"`
	MakerBaseFee            float64       `json:"maker_base_fee"`
	TakerBaseFee            float64       `json:"taker_base_fee"`
	NotificationsEnabled    bool          `json:"notifications_enabled"`
	NegRisk                 bool          `json:"neg_risk"`
	NegRiskMarketID         string        `json:"neg_risk_market_id"`
	NegRiskRequestID        string        `json:"neg_risk_request_id"`
	Icon                    string        `json:"icon"`
	Image                   string        `json:"image"`
	Rewards                 MarketRewards `json:"rewards"`
	Is5050Outcome           bool          `json:"is_50_50_outcome"`
	Tokens                  []MarketToken `json:"tokens"`
	Tags                    []string      `json:"tags,omitempty"`
}

type SimplifiedMarketInfo struct {
	ConditionID string        `json:"condition_id"`
	Rewards     MarketRewards `json:"rewards"`
	Tokens      []MarketToken `json:"tokens"`
}

type MarketByTokenInfo struct {
	ConditionID      string `json:"condition_id"`
	PrimaryTokenID   string `json:"primary_token_id"`
	SecondaryTokenID string `json:"secondary_token_id"`
}

type MarketLiveActivity struct {
	ConditionID string         `json:"condition_id"`
	ID          int64          `json:"id"`
	Question    string         `json:"question"`
	MarketSlug  string         `json:"market_slug"`
	EventSlug   string         `json:"event_slug"`
	SeriesSlug  string         `json:"series_slug"`
	Icon        string         `json:"icon"`
	Image       string         `json:"image"`
	Tags        []string       `json:"tags"`
	Raw         map[string]any `json:"-"`
}

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
	StartTs  int64
	EndTs    int64
	Fidelity int
}

type MarketRewardConfig struct {
	AssetAddress string  `json:"asset_address"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	RatePerDay   float64 `json:"rate_per_day"`
	TotalRewards float64 `json:"total_rewards"`
	ID           int     `json:"id"`
}

type CurrentRewardMarket struct {
	ConditionID      string               `json:"condition_id"`
	RewardsConfig    []MarketRewardConfig `json:"rewards_config"`
	RewardsMaxSpread float64              `json:"rewards_max_spread"`
	RewardsMinSize   float64              `json:"rewards_min_size"`
	NativeDailyRate  float64              `json:"native_daily_rate"`
	TotalDailyRate   float64              `json:"total_daily_rate"`
}

type OrderScoringResult struct {
	Scoring bool `json:"scoring"`
}

type NonceResponse struct {
	Nonce string `json:"nonce"`
}

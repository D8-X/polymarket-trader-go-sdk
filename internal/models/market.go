package models

import (
	"encoding/json"
	"strconv"
)

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

func (b *OrderBook) BestBid() (price, size float64, ok bool) {
	return bestLevel(b.Bids, true)
}

func (b *OrderBook) BestAsk() (price, size float64, ok bool) {
	return bestLevel(b.Asks, false)
}

func (b *OrderBook) Mid() (float64, bool) {
	bp, _, bok := b.BestBid()
	ap, _, aok := b.BestAsk()
	if !bok || !aok {
		return 0, false
	}
	return (bp + ap) / 2, true
}

func bestLevel(levels []OrderBookLevel, highest bool) (float64, float64, bool) {
	var bestP, bestS float64
	first := true
	for _, lvl := range levels {
		p, err := strconv.ParseFloat(lvl.Price, 64)
		if err != nil {
			continue
		}
		s, _ := strconv.ParseFloat(lvl.Size, 64)
		if first || (highest && p > bestP) || (!highest && p < bestP) {
			bestP, bestS = p, s
			first = false
		}
	}
	return bestP, bestS, !first
}

type OrderBookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
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

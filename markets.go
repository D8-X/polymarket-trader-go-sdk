package polytrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (c *CLOBClient) GetMarkets(ctx context.Context, nextCursor string) (*PaginatedResponse[MarketInfo], error) {
	return getPaginatedJSON[MarketInfo](ctx, c, "/markets", nextCursor)
}

func (c *CLOBClient) GetMarket(ctx context.Context, conditionID string) (*MarketInfo, error) {
	return getJSON[MarketInfo](ctx, c, "/markets/"+conditionID)
}

func (c *CLOBClient) GetSamplingMarkets(ctx context.Context, nextCursor string) (*PaginatedResponse[MarketInfo], error) {
	return getPaginatedJSON[MarketInfo](ctx, c, "/sampling-markets", nextCursor)
}

func (c *CLOBClient) GetSimplifiedMarkets(ctx context.Context, nextCursor string) (*PaginatedResponse[SimplifiedMarketInfo], error) {
	return getPaginatedJSON[SimplifiedMarketInfo](ctx, c, "/simplified-markets", nextCursor)
}

func (c *CLOBClient) GetSamplingSimplifiedMarkets(ctx context.Context, nextCursor string) (*PaginatedResponse[SimplifiedMarketInfo], error) {
	return getPaginatedJSON[SimplifiedMarketInfo](ctx, c, "/sampling-simplified-markets", nextCursor)
}

func (c *CLOBClient) GetMarketByToken(ctx context.Context, tokenID string) (*MarketByTokenInfo, error) {
	return getJSON[MarketByTokenInfo](ctx, c, "/markets-by-token/"+tokenID)
}

func (c *CLOBClient) GetMarketLiveActivity(ctx context.Context, conditionID string) (*MarketLiveActivity, error) {
	return getJSON[MarketLiveActivity](ctx, c, "/markets/live-activity/"+conditionID)
}

func getJSON[T any](ctx context.Context, c *CLOBClient, path string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("GET %s: build request: %w", path, err)
	}
	body, err := c.doRequest(req, "GET "+path)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("GET %s: unmarshal: %w", path, err)
	}
	return &out, nil
}

func getPaginatedJSON[T any](ctx context.Context, c *CLOBClient, path, nextCursor string) (*PaginatedResponse[T], error) {
	fullPath := path + "?next_cursor=" + nextCursor
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("GET %s: build request: %w", path, err)
	}
	body, err := c.doRequest(req, "GET "+path)
	if err != nil {
		return nil, err
	}
	var out PaginatedResponse[T]
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("GET %s: unmarshal: %w", path, err)
	}
	return &out, nil
}

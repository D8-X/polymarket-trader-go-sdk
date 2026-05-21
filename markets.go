package polytrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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

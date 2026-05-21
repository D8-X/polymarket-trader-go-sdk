package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func (c *Client) GetMarkets(ctx context.Context, nextCursor string) (*models.PaginatedResponse[models.MarketInfo], error) {
	return getPaginatedJSON[models.MarketInfo](ctx, c, "/markets", nextCursor)
}

func (c *Client) GetMarket(ctx context.Context, conditionID string) (*models.MarketInfo, error) {
	return getJSON[models.MarketInfo](ctx, c, "/markets/"+conditionID)
}

func (c *Client) GetSamplingMarkets(ctx context.Context, nextCursor string) (*models.PaginatedResponse[models.MarketInfo], error) {
	return getPaginatedJSON[models.MarketInfo](ctx, c, "/sampling-markets", nextCursor)
}

func (c *Client) GetSimplifiedMarkets(ctx context.Context, nextCursor string) (*models.PaginatedResponse[models.SimplifiedMarketInfo], error) {
	return getPaginatedJSON[models.SimplifiedMarketInfo](ctx, c, "/simplified-markets", nextCursor)
}

func (c *Client) GetSamplingSimplifiedMarkets(ctx context.Context, nextCursor string) (*models.PaginatedResponse[models.SimplifiedMarketInfo], error) {
	return getPaginatedJSON[models.SimplifiedMarketInfo](ctx, c, "/sampling-simplified-markets", nextCursor)
}

func (c *Client) GetMarketByToken(ctx context.Context, tokenID string) (*models.MarketByTokenInfo, error) {
	return getJSON[models.MarketByTokenInfo](ctx, c, "/markets-by-token/"+tokenID)
}

func (c *Client) GetMarketLiveActivity(ctx context.Context, conditionID string) (*models.MarketLiveActivity, error) {
	return getJSON[models.MarketLiveActivity](ctx, c, "/markets/live-activity/"+conditionID)
}

func getJSON[T any](ctx context.Context, c *Client, path string) (*T, error) {
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

func getPaginatedJSON[T any](ctx context.Context, c *Client, path, nextCursor string) (*models.PaginatedResponse[T], error) {
	fullPath := path + "?next_cursor=" + nextCursor
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("GET %s: build request: %w", path, err)
	}
	body, err := c.doRequest(req, "GET "+path)
	if err != nil {
		return nil, err
	}
	var out models.PaginatedResponse[T]
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("GET %s: unmarshal: %w", path, err)
	}
	return &out, nil
}

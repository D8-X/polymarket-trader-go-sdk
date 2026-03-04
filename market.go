package polytrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (c *CLOBClient) GetServerTime(ctx context.Context) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/time", nil)
	if err != nil {
		return 0, fmt.Errorf("get server time: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /time")
	if err != nil {
		return 0, fmt.Errorf("get server time: %w", err)
	}

	ts, err := strconv.ParseInt(strings.TrimSpace(string(respBody)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("get server time: parse response: %w", err)
	}

	return ts, nil
}

func (c *CLOBClient) GetOrderBook(ctx context.Context, tokenID string) (*OrderBook, error) {
	path := "/book?token_id=" + tokenID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("get order book: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /book")
	if err != nil {
		return nil, fmt.Errorf("get order book: %w", err)
	}

	var book OrderBook
	if err := json.Unmarshal(respBody, &book); err != nil {
		return nil, fmt.Errorf("get order book: unmarshal response: %w", err)
	}

	return &book, nil
}

func (c *CLOBClient) GetPrice(ctx context.Context, tokenID, side string) (string, error) {
	path := fmt.Sprintf("/price?token_id=%s&side=%s", tokenID, side)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return "", fmt.Errorf("get price: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /price")
	if err != nil {
		return "", fmt.Errorf("get price: %w", err)
	}

	var result struct {
		Price json.Number `json:"price"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("get price: unmarshal response: %w", err)
	}

	return result.Price.String(), nil
}

func (c *CLOBClient) GetMidpoint(ctx context.Context, tokenID string) (string, error) {
	path := "/midpoint?token_id=" + tokenID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return "", fmt.Errorf("get midpoint: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /midpoint")
	if err != nil {
		return "", fmt.Errorf("get midpoint: %w", err)
	}

	var result struct {
		MidPrice string `json:"mid_price"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("get midpoint: unmarshal response: %w", err)
	}

	return result.MidPrice, nil
}

func (c *CLOBClient) GetSpread(ctx context.Context, tokenID string) (string, error) {
	path := "/spread?token_id=" + tokenID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return "", fmt.Errorf("get spread: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /spread")
	if err != nil {
		return "", fmt.Errorf("get spread: %w", err)
	}

	var result struct {
		Spread string `json:"spread"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("get spread: unmarshal response: %w", err)
	}

	return result.Spread, nil
}

func (c *CLOBClient) GetTickSize(ctx context.Context, tokenID string) (string, error) {
	path := "/tick-size?token_id=" + tokenID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return "", fmt.Errorf("get tick size: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /tick-size")
	if err != nil {
		return "", fmt.Errorf("get tick size: %w", err)
	}

	var result struct {
		TickSize json.Number `json:"minimum_tick_size"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("get tick size: unmarshal response: %w", err)
	}

	return result.TickSize.String(), nil
}

func (c *CLOBClient) GetFeeRate(ctx context.Context, tokenID string) (int, error) {
	path := "/fee-rate"
	if tokenID != "" {
		path += "?token_id=" + tokenID
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, fmt.Errorf("get fee rate: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /fee-rate")
	if err != nil {
		return 0, fmt.Errorf("get fee rate: %w", err)
	}

	var result struct {
		BaseFee int `json:"base_fee"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("get fee rate: unmarshal response: %w", err)
	}

	return result.BaseFee, nil
}

func (c *CLOBClient) GetNegRisk(ctx context.Context, tokenID string) (bool, error) {
	path := "/neg-risk"
	if tokenID != "" {
		path += "?token_id=" + tokenID
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return false, fmt.Errorf("get neg risk: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /neg-risk")
	if err != nil {
		return false, fmt.Errorf("get neg risk: %w", err)
	}

	var result struct {
		NegRisk bool `json:"neg_risk"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, fmt.Errorf("get neg risk: unmarshal response: %w", err)
	}

	return result.NegRisk, nil
}

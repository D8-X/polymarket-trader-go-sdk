package polytrade

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

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

func (c *CLOBClient) GetPrices(ctx context.Context, params []PriceRequest) (map[string]map[string]string, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("get prices: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/prices", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("get prices: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	respBody, err := c.doRequest(req, "POST /prices")
	if err != nil {
		return nil, err
	}
	var out map[string]map[string]string
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("get prices: unmarshal: %w", err)
	}
	return out, nil
}

func (c *CLOBClient) GetSpreads(ctx context.Context, params []SpreadRequest) (map[string]string, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("get spreads: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/spreads", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("get spreads: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	respBody, err := c.doRequest(req, "POST /spreads")
	if err != nil {
		return nil, err
	}
	var out map[string]string
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("get spreads: unmarshal: %w", err)
	}
	return out, nil
}

func (c *CLOBClient) GetLastTradePrice(ctx context.Context, tokenID string) (*LastTradePrice, error) {
	path := "/last-trade-price?token_id=" + tokenID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("get last trade price: build request: %w", err)
	}
	respBody, err := c.doRequest(req, "GET /last-trade-price")
	if err != nil {
		return nil, err
	}
	var out LastTradePrice
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("get last trade price: unmarshal: %w", err)
	}
	return &out, nil
}

func (c *CLOBClient) GetLastTradePrices(ctx context.Context, params []SpreadRequest) ([]LastTradePrice, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("get last trades prices: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/last-trades-prices", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("get last trades prices: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	respBody, err := c.doRequest(req, "POST /last-trades-prices")
	if err != nil {
		return nil, err
	}
	var out []LastTradePrice
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("get last trades prices: unmarshal: %w", err)
	}
	return out, nil
}

func (c *CLOBClient) GetPricesHistory(ctx context.Context, p PricesHistoryParams) ([]PriceHistoryEntry, error) {
	if p.Interval == "" && (p.StartTs == 0 || p.EndTs == 0) {
		return nil, fmt.Errorf("get prices history: requires either Interval or both StartTs and EndTs")
	}
	q := "?market=" + p.Market
	if p.Interval != "" {
		q += "&interval=" + p.Interval
	}
	if p.StartTs != 0 {
		q += "&startTs=" + strconv.FormatInt(p.StartTs, 10)
	}
	if p.EndTs != 0 {
		q += "&endTs=" + strconv.FormatInt(p.EndTs, 10)
	}
	if p.Fidelity != 0 {
		q += "&fidelity=" + strconv.Itoa(p.Fidelity)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/prices-history"+q, nil)
	if err != nil {
		return nil, fmt.Errorf("get prices history: build request: %w", err)
	}
	respBody, err := c.doRequest(req, "GET /prices-history")
	if err != nil {
		return nil, err
	}
	var out struct {
		History []PriceHistoryEntry `json:"history"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("get prices history: unmarshal: %w", err)
	}
	return out.History, nil
}

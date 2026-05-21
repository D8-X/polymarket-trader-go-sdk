package clob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func (c *Client) GetPrices(ctx context.Context, params []models.PriceRequest) (map[string]map[string]string, error) {
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

func (c *Client) GetSpreads(ctx context.Context, params []models.SpreadRequest) (map[string]string, error) {
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

func (c *Client) GetLastTradePrice(ctx context.Context, tokenID string) (*models.LastTradePrice, error) {
	path := "/last-trade-price?token_id=" + tokenID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("get last trade price: build request: %w", err)
	}
	respBody, err := c.doRequest(req, "GET /last-trade-price")
	if err != nil {
		return nil, err
	}
	var out models.LastTradePrice
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("get last trade price: unmarshal: %w", err)
	}
	return &out, nil
}

func (c *Client) GetLastTradePrices(ctx context.Context, params []models.SpreadRequest) ([]models.LastTradePrice, error) {
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
	var out []models.LastTradePrice
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("get last trades prices: unmarshal: %w", err)
	}
	return out, nil
}

func (c *Client) GetPricesHistory(ctx context.Context, p models.PricesHistoryParams) ([]models.PriceHistoryEntry, error) {
	if p.Interval == "" && (p.StartTS == 0 || p.EndTS == 0) {
		return nil, fmt.Errorf("get prices history: requires either Interval or both StartTS and EndTS")
	}
	q := "?market=" + p.Market
	if p.Interval != "" {
		q += "&interval=" + p.Interval
	}
	if p.StartTS != 0 {
		q += "&startTs=" + strconv.FormatInt(p.StartTS, 10)
	}
	if p.EndTS != 0 {
		q += "&endTs=" + strconv.FormatInt(p.EndTS, 10)
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
		History []models.PriceHistoryEntry `json:"history"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("get prices history: unmarshal: %w", err)
	}
	return out.History, nil
}

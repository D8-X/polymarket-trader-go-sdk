package clob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/auth"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/types"
)

func (c *Client) IsOrderScoring(ctx context.Context, orderID string, creds *types.L2Credentials) (*models.OrderScoringResult, error) {
	path := "/order-scoring"
	fullPath := path + "?order_id=" + orderID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("is order scoring: build request: %w", err)
	}
	headers, err := auth.SignRequest(creds, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("is order scoring: %w", err)
	}
	auth.ApplyHeaders(req, headers)
	respBody, err := c.doRequest(req, "GET /order-scoring")
	if err != nil {
		return nil, fmt.Errorf("is order scoring: %w", err)
	}
	var result models.OrderScoringResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("is order scoring: unmarshal response: %w", err)
	}
	return &result, nil
}

// AreOrdersScoring returns scoring status for multiple orders, keyed by order
// ID. Requires L2 auth.
func (c *Client) AreOrdersScoring(ctx context.Context, orderIDs []string, creds *types.L2Credentials) (map[string]bool, error) {
	body, err := json.Marshal(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("are orders scoring: marshal: %w", err)
	}
	path := "/orders-scoring"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("are orders scoring: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	headers, err := auth.SignRequest(creds, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("are orders scoring: %w", err)
	}
	auth.ApplyHeaders(req, headers)
	respBody, err := c.doRequest(req, "POST /orders-scoring")
	if err != nil {
		return nil, fmt.Errorf("are orders scoring: %w", err)
	}
	var result map[string]bool
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("are orders scoring: unmarshal response: %w", err)
	}
	return result, nil
}

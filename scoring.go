package polytrade

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// OrderScoringResult reports whether an order qualifies for liquidity rewards.
type OrderScoringResult struct {
	Scoring bool `json:"scoring"`
}

// IsOrderScoring returns whether the given order qualifies for liquidity
// rewards. Requires L2 auth.
func (c *CLOBClient) IsOrderScoring(ctx context.Context, orderID string, creds *L2Credentials) (*OrderScoringResult, error) {
	path := "/order-scoring"
	fullPath := path + "?order_id=" + orderID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("is order scoring: build request: %w", err)
	}
	headers, err := SignL2Request(creds, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("is order scoring: %w", err)
	}
	ApplyL2Headers(req, headers)
	respBody, err := c.doRequest(req, "GET /order-scoring")
	if err != nil {
		return nil, fmt.Errorf("is order scoring: %w", err)
	}
	var result OrderScoringResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("is order scoring: unmarshal response: %w", err)
	}
	return &result, nil
}

// AreOrdersScoring returns scoring status for multiple orders, keyed by order
// ID. Requires L2 auth.
func (c *CLOBClient) AreOrdersScoring(ctx context.Context, orderIDs []string, creds *L2Credentials) (map[string]bool, error) {
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
	headers, err := SignL2Request(creds, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("are orders scoring: %w", err)
	}
	ApplyL2Headers(req, headers)
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

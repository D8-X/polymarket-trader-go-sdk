package polytrade

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CLOBClient struct {
	baseURL string
	client  *http.Client
}

func NewCLOBClient() *CLOBClient {
	return &CLOBClient{
		baseURL: ClobBaseURL,
		client:  &http.Client{Timeout: CLOBTimeout},
	}
}

func (c *CLOBClient) PlaceOrder(ctx context.Context, signedOrder *SignedOrder, creds *L2Credentials) (*PlaceOrderResponse, error) {
	body, err := json.Marshal(signedOrder)
	if err != nil {
		return nil, fmt.Errorf("place order: marshal: %w", err)
	}

	path := "/order"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("place order: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	headers, err := SignL2Request(creds, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}
	ApplyL2Headers(req, headers)

	respBody, err := c.doRequest(req, "POST /order")
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	var result PlaceOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("place order: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *CLOBClient) PlaceOrders(ctx context.Context, signedOrders []*SignedOrder, creds *L2Credentials) ([]PlaceOrderResponse, error) {
	body, err := json.Marshal(signedOrders)
	if err != nil {
		return nil, fmt.Errorf("place orders: marshal: %w", err)
	}

	path := "/orders"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("place orders: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	headers, err := SignL2Request(creds, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("place orders: %w", err)
	}
	ApplyL2Headers(req, headers)

	respBody, err := c.doRequest(req, "POST /orders")
	if err != nil {
		return nil, fmt.Errorf("place orders: %w", err)
	}

	var results []PlaceOrderResponse
	if err := json.Unmarshal(respBody, &results); err != nil {
		return nil, fmt.Errorf("place orders: unmarshal response: %w", err)
	}

	return results, nil
}

func (c *CLOBClient) GetOrder(ctx context.Context, orderID string, creds *L2Credentials) (*OrderStatus, error) {
	path := "/order/" + orderID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("get order: build request: %w", err)
	}

	headers, err := SignL2Request(creds, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	ApplyL2Headers(req, headers)

	respBody, err := c.doRequest(req, "GET /order")
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	var status OrderStatus
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("get order: unmarshal response: %w", err)
	}

	return &status, nil
}

func (c *CLOBClient) CancelOrder(ctx context.Context, orderID string, creds *L2Credentials) (*CancelResponse, error) {
	body, err := json.Marshal(map[string]string{"orderID": orderID})
	if err != nil {
		return nil, fmt.Errorf("cancel order: marshal: %w", err)
	}

	path := "/order"
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cancel order: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	headers, err := SignL2Request(creds, http.MethodDelete, path, body)
	if err != nil {
		return nil, fmt.Errorf("cancel order: %w", err)
	}
	ApplyL2Headers(req, headers)

	respBody, err := c.doRequest(req, "DELETE /order")
	if err != nil {
		return nil, fmt.Errorf("cancel order: %w", err)
	}

	var result CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cancel order: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *CLOBClient) CancelOrders(ctx context.Context, orderIDs []string, creds *L2Credentials) (*CancelResponse, error) {
	body, err := json.Marshal(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("cancel orders: marshal: %w", err)
	}

	path := "/orders"
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cancel orders: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	headers, err := SignL2Request(creds, http.MethodDelete, path, body)
	if err != nil {
		return nil, fmt.Errorf("cancel orders: %w", err)
	}
	ApplyL2Headers(req, headers)

	respBody, err := c.doRequest(req, "DELETE /orders")
	if err != nil {
		return nil, fmt.Errorf("cancel orders: %w", err)
	}

	var result CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cancel orders: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *CLOBClient) CancelAll(ctx context.Context, creds *L2Credentials) (*CancelResponse, error) {
	path := "/cancel-all"
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("cancel all: build request: %w", err)
	}

	headers, err := SignL2Request(creds, http.MethodDelete, path, nil)
	if err != nil {
		return nil, fmt.Errorf("cancel all: %w", err)
	}
	ApplyL2Headers(req, headers)

	respBody, err := c.doRequest(req, "DELETE /cancel-all")
	if err != nil {
		return nil, fmt.Errorf("cancel all: %w", err)
	}

	var result CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cancel all: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *CLOBClient) CancelMarketOrders(ctx context.Context, market, assetID string, creds *L2Credentials) (*CancelResponse, error) {
	body, err := json.Marshal(map[string]string{"market": market, "asset_id": assetID})
	if err != nil {
		return nil, fmt.Errorf("cancel market orders: marshal: %w", err)
	}

	path := "/cancel-market-orders"
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cancel market orders: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	headers, err := SignL2Request(creds, http.MethodDelete, path, body)
	if err != nil {
		return nil, fmt.Errorf("cancel market orders: %w", err)
	}
	ApplyL2Headers(req, headers)

	respBody, err := c.doRequest(req, "DELETE /cancel-market-orders")
	if err != nil {
		return nil, fmt.Errorf("cancel market orders: %w", err)
	}

	var result CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cancel market orders: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *CLOBClient) GetOpenOrders(ctx context.Context, market, assetID string, creds *L2Credentials) ([]OrderStatus, error) {
	var all []OrderStatus
	cursor := ""

	for {
		path := "/orders"
		query := "?"
		if market != "" {
			query += "market=" + market + "&"
		}
		if assetID != "" {
			query += "asset_id=" + assetID + "&"
		}
		if cursor != "" {
			query += "next_cursor=" + cursor + "&"
		}
		fullPath := path + query[:len(query)-1]

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
		if err != nil {
			return nil, fmt.Errorf("get open orders: build request: %w", err)
		}

		headers, err := SignL2Request(creds, http.MethodGet, fullPath, nil)
		if err != nil {
			return nil, fmt.Errorf("get open orders: %w", err)
		}
		ApplyL2Headers(req, headers)

		respBody, err := c.doRequest(req, "GET /orders")
		if err != nil {
			return nil, fmt.Errorf("get open orders: %w", err)
		}

		var page PaginatedResponse[OrderStatus]
		if err := json.Unmarshal(respBody, &page); err != nil {
			return nil, fmt.Errorf("get open orders: unmarshal response: %w", err)
		}

		all = append(all, page.Data...)

		if page.NextCursor == "" || page.NextCursor == "LTE=" || len(page.Data) == 0 {
			break
		}
		cursor = page.NextCursor
	}

	return all, nil
}

func (c *CLOBClient) GetTrades(ctx context.Context, makerAddress, market, assetID string, creds *L2Credentials) ([]Trade, error) {
	var all []Trade
	cursor := ""

	for {
		path := "/trades"
		query := "?maker_address=" + makerAddress + "&"
		if market != "" {
			query += "market=" + market + "&"
		}
		if assetID != "" {
			query += "asset_id=" + assetID + "&"
		}
		if cursor != "" {
			query += "next_cursor=" + cursor + "&"
		}
		fullPath := path + query[:len(query)-1]

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
		if err != nil {
			return nil, fmt.Errorf("get trades: build request: %w", err)
		}

		headers, err := SignL2Request(creds, http.MethodGet, fullPath, nil)
		if err != nil {
			return nil, fmt.Errorf("get trades: %w", err)
		}
		ApplyL2Headers(req, headers)

		respBody, err := c.doRequest(req, "GET /trades")
		if err != nil {
			return nil, fmt.Errorf("get trades: %w", err)
		}

		var page PaginatedResponse[Trade]
		if err := json.Unmarshal(respBody, &page); err != nil {
			return nil, fmt.Errorf("get trades: unmarshal response: %w", err)
		}

		all = append(all, page.Data...)

		if page.NextCursor == "" || page.NextCursor == "LTE=" || len(page.Data) == 0 {
			break
		}
		cursor = page.NextCursor
	}

	return all, nil
}

func (c *CLOBClient) GetBalances(ctx context.Context, creds *L2Credentials) ([]BalanceEntry, error) {
	positions, err := c.GetPositions(ctx, creds.Address)
	if err != nil {
		return nil, fmt.Errorf("get balances: %w", err)
	}
	var balances []BalanceEntry
	for _, p := range positions {
		if p.Size > 0 {
			balances = append(balances, BalanceEntry{
				AssetID: p.Asset,
				Balance: p.Size,
			})
		}
	}
	return balances, nil
}

func (c *CLOBClient) GetBalanceAllowance(ctx context.Context, assetType string, tokenID string, sigType int, creds *L2Credentials) (*BalanceAllowanceResponse, error) {
	path := fmt.Sprintf("/balance-allowance?asset_type=%s&signature_type=%d", assetType, sigType)
	if tokenID != "" {
		path += "&token_id=" + tokenID
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("get balance allowance: build request: %w", err)
	}

	headers, err := SignL2Request(creds, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get balance allowance: %w", err)
	}
	ApplyL2Headers(req, headers)

	respBody, err := c.doRequest(req, "GET /balance-allowance")
	if err != nil {
		return nil, fmt.Errorf("get balance allowance: %w", err)
	}

	var result BalanceAllowanceResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("get balance allowance: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *CLOBClient) UpdateBalanceAllowance(ctx context.Context, assetType string, tokenID string, sigType int, creds *L2Credentials) error {
	path := fmt.Sprintf("/balance-allowance?asset_type=%s&signature_type=%d", assetType, sigType)
	if tokenID != "" {
		path += "&token_id=" + tokenID
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("update balance allowance: build request: %w", err)
	}

	headers, err := SignL2Request(creds, http.MethodPut, path, nil)
	if err != nil {
		return fmt.Errorf("update balance allowance: %w", err)
	}
	ApplyL2Headers(req, headers)

	if _, err := c.doRequest(req, "PUT /balance-allowance"); err != nil {
		return fmt.Errorf("update balance allowance: %w", err)
	}

	return nil
}

func (c *CLOBClient) GetPositions(ctx context.Context, walletAddress string) ([]PositionEntry, error) {
	fullURL := fmt.Sprintf("%s/positions?user=%s&sizeThreshold=0", DataAPIBaseURL, walletAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get positions: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /positions")
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}

	var positions []PositionEntry
	if err := json.Unmarshal(respBody, &positions); err != nil {
		return nil, fmt.Errorf("get positions: unmarshal response: %w", err)
	}

	return positions, nil
}

func (c *CLOBClient) doRequest(req *http.Request, endpoint string) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http %s: read body: %w", endpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   endpoint,
			Body:       string(body),
		}
	}

	return body, nil
}

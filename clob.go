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
	Verbose bool
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
		return nil, fmt.Errorf("marshal order: %w", err)
	}

	if c.Verbose {
		fmt.Printf("\n--- POST /order REQUEST ---\n%s\n", string(body))
	}

	path := "/order"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	headers := SignL2Request(creds, http.MethodPost, path, body)
	ApplyL2Headers(req, headers)

	if c.Verbose {
		fmt.Printf("--- HEADERS ---\n")
		for k, v := range req.Header {
			if len(k) > 4 && k[:4] == "Poly" {
				fmt.Printf("  %s: %s\n", k, v[0])
			}
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("clob request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.Verbose {
		fmt.Printf("--- RESPONSE %d ---\n%s\n---\n", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clob returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result PlaceOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *CLOBClient) GetBalances(ctx context.Context, creds *L2Credentials) ([]BalanceEntry, error) {
	positions, err := c.GetPositions(ctx, creds.Address)
	if err != nil {
		return nil, err
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

func (c *CLOBClient) GetPositions(ctx context.Context, walletAddress string) ([]PositionEntry, error) {
	fullURL := fmt.Sprintf("%s/positions?user=%s&sizeThreshold=0", DataAPIBaseURL, walletAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("data-api request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("data-api returned %d: %s", resp.StatusCode, string(respBody))
	}

	var positions []PositionEntry
	if err := json.Unmarshal(respBody, &positions); err != nil {
		return nil, fmt.Errorf("unmarshal positions: %w", err)
	}

	return positions, nil
}

func (c *CLOBClient) GetOrder(ctx context.Context, orderID string, creds *L2Credentials) (*OrderStatus, error) {
	path := "/data/order/" + orderID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	headers := SignL2Request(creds, http.MethodGet, path, nil)
	ApplyL2Headers(req, headers)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("clob request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.Verbose {
		fmt.Printf("[poll] GET %s -> %d: %s\n", path, resp.StatusCode, string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clob returned %d: %s", resp.StatusCode, string(respBody))
	}

	var status OrderStatus
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("unmarshal order status: %w", err)
	}

	return &status, nil
}

func (c *CLOBClient) CancelOrder(ctx context.Context, orderID string, creds *L2Credentials) error {
	path := "/order/" + orderID
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	headers := SignL2Request(creds, http.MethodDelete, path, nil)
	ApplyL2Headers(req, headers)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("clob request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clob returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

package clob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/auth"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/types"
)

type Client struct {
	baseURL        string
	dataAPIBaseURL string
	client         *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL:        consts.ClobBaseURL,
		dataAPIBaseURL: consts.DataAPIBaseURL,
		client:         &http.Client{Timeout: consts.CLOBTimeout},
	}
}

func (c *Client) SetBaseURL(url string)        { c.baseURL = url }
func (c *Client) SetDataAPIBaseURL(url string) { c.dataAPIBaseURL = url }

func (c *Client) PlaceOrder(ctx context.Context, signedOrder *models.SignedOrder, creds *types.L2Credentials) (*models.PlaceOrderResponse, error) {
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

	headers, err := auth.SignRequest(creds, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "POST /order")
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	var result models.PlaceOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("place order: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) PlaceOrders(ctx context.Context, signedOrders []*models.SignedOrder, creds *types.L2Credentials) ([]models.PlaceOrderResponse, error) {
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

	headers, err := auth.SignRequest(creds, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("place orders: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "POST /orders")
	if err != nil {
		return nil, fmt.Errorf("place orders: %w", err)
	}

	var results []models.PlaceOrderResponse
	if err := json.Unmarshal(respBody, &results); err != nil {
		return nil, fmt.Errorf("place orders: unmarshal response: %w", err)
	}

	return results, nil
}

func (c *Client) GetOrder(ctx context.Context, orderID string, creds *types.L2Credentials) (*models.OrderStatus, error) {
	path := "/data/order/" + orderID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("get order: build request: %w", err)
	}

	headers, err := auth.SignRequest(creds, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "GET /order")
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	var status models.OrderStatus
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("get order: unmarshal response: %w", err)
	}

	return &status, nil
}

func (c *Client) CancelOrder(ctx context.Context, orderID string, creds *types.L2Credentials) (*models.CancelResponse, error) {
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

	headers, err := auth.SignRequest(creds, http.MethodDelete, path, body)
	if err != nil {
		return nil, fmt.Errorf("cancel order: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "DELETE /order")
	if err != nil {
		return nil, fmt.Errorf("cancel order: %w", err)
	}

	var result models.CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cancel order: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) CancelOrders(ctx context.Context, orderIDs []string, creds *types.L2Credentials) (*models.CancelResponse, error) {
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

	headers, err := auth.SignRequest(creds, http.MethodDelete, path, body)
	if err != nil {
		return nil, fmt.Errorf("cancel orders: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "DELETE /orders")
	if err != nil {
		return nil, fmt.Errorf("cancel orders: %w", err)
	}

	var result models.CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cancel orders: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) CancelAll(ctx context.Context, creds *types.L2Credentials) (*models.CancelResponse, error) {
	path := "/cancel-all"
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("cancel all: build request: %w", err)
	}

	headers, err := auth.SignRequest(creds, http.MethodDelete, path, nil)
	if err != nil {
		return nil, fmt.Errorf("cancel all: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "DELETE /cancel-all")
	if err != nil {
		return nil, fmt.Errorf("cancel all: %w", err)
	}

	var result models.CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cancel all: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) CancelMarketOrders(ctx context.Context, market, assetID string, creds *types.L2Credentials) (*models.CancelResponse, error) {
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

	headers, err := auth.SignRequest(creds, http.MethodDelete, path, body)
	if err != nil {
		return nil, fmt.Errorf("cancel market orders: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "DELETE /cancel-market-orders")
	if err != nil {
		return nil, fmt.Errorf("cancel market orders: %w", err)
	}

	var result models.CancelResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cancel market orders: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetOpenOrders(ctx context.Context, market, assetID string, creds *types.L2Credentials) ([]models.OrderStatus, error) {
	var all []models.OrderStatus
	cursor := ""

	for {
		path := "/data/orders"
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

		headers, err := auth.SignRequest(creds, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("get open orders: %w", err)
		}
		auth.ApplyHeaders(req, headers)

		respBody, err := c.doRequest(req, "GET /data/orders")
		if err != nil {
			return nil, fmt.Errorf("get open orders: %w", err)
		}

		var page models.PaginatedResponse[models.OrderStatus]
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

// GetPreMigrationOrders returns the caller's V1 orders that were still open at
// the time of the V2 cutover. Use this if you held positions before the
// migration and want to inspect or reconcile them.
func (c *Client) GetPreMigrationOrders(ctx context.Context, creds *types.L2Credentials) ([]models.OrderStatus, error) {
	var all []models.OrderStatus
	cursor := ""

	for {
		path := "/data/pre-migration-orders"
		fullPath := path
		if cursor != "" {
			fullPath = path + "?next_cursor=" + cursor
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
		if err != nil {
			return nil, fmt.Errorf("get pre-migration orders: build request: %w", err)
		}

		headers, err := auth.SignRequest(creds, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("get pre-migration orders: %w", err)
		}
		auth.ApplyHeaders(req, headers)

		respBody, err := c.doRequest(req, "GET /data/pre-migration-orders")
		if err != nil {
			return nil, fmt.Errorf("get pre-migration orders: %w", err)
		}

		var page models.PaginatedResponse[models.OrderStatus]
		if err := json.Unmarshal(respBody, &page); err != nil {
			return nil, fmt.Errorf("get pre-migration orders: unmarshal response: %w", err)
		}

		all = append(all, page.Data...)

		if page.NextCursor == "" || page.NextCursor == "LTE=" || len(page.Data) == 0 {
			break
		}
		cursor = page.NextCursor
	}

	return all, nil
}

func (c *Client) GetTrades(ctx context.Context, makerAddress, market, assetID string, creds *types.L2Credentials) ([]models.Trade, error) {
	var all []models.Trade
	cursor := ""

	for {
		path := "/data/trades"
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

		headers, err := auth.SignRequest(creds, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("get trades: %w", err)
		}
		auth.ApplyHeaders(req, headers)

		respBody, err := c.doRequest(req, "GET /data/trades")
		if err != nil {
			return nil, fmt.Errorf("get trades: %w", err)
		}

		var page models.PaginatedResponse[models.Trade]
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

func (c *Client) GetBalances(ctx context.Context, creds *types.L2Credentials) ([]models.BalanceEntry, error) {
	positions, err := c.GetPositions(ctx, creds.Address)
	if err != nil {
		return nil, fmt.Errorf("get balances: %w", err)
	}
	balances := make([]models.BalanceEntry, 0, len(positions))
	for _, p := range positions {
		balances = append(balances, models.BalanceEntry{
			AssetID: p.Asset,
			Balance: p.Size,
		})
	}
	return balances, nil
}

func (c *Client) GetBalanceAllowance(ctx context.Context, assetType string, tokenID string, creds *types.L2Credentials) (*models.BalanceAllowanceResponse, error) {
	signPath := "/balance-allowance"
	fullPath := fmt.Sprintf("/balance-allowance?asset_type=%s&signature_type=%d", assetType, consts.SignatureTypePoly1271)
	if tokenID != "" {
		fullPath += "&token_id=" + tokenID
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("get balance allowance: build request: %w", err)
	}

	headers, err := auth.SignRequest(creds, http.MethodGet, signPath, nil)
	if err != nil {
		return nil, fmt.Errorf("get balance allowance: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "GET /balance-allowance")
	if err != nil {
		return nil, fmt.Errorf("get balance allowance: %w", err)
	}

	var result models.BalanceAllowanceResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("get balance allowance: unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdateBalanceAllowance(ctx context.Context, assetType string, tokenID string, creds *types.L2Credentials) error {
	signPath := "/balance-allowance/update"
	fullPath := fmt.Sprintf("/balance-allowance/update?asset_type=%s&signature_type=%d", assetType, consts.SignatureTypePoly1271)
	if tokenID != "" {
		fullPath += "&token_id=" + tokenID
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
	if err != nil {
		return fmt.Errorf("update balance allowance: build request: %w", err)
	}

	headers, err := auth.SignRequest(creds, http.MethodGet, signPath, nil)
	if err != nil {
		return fmt.Errorf("update balance allowance: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	if _, err := c.doRequest(req, "PUT /balance-allowance"); err != nil {
		return fmt.Errorf("update balance allowance: %w", err)
	}

	return nil
}

func (c *Client) GetPositions(ctx context.Context, walletAddress string) ([]models.PositionEntry, error) {
	fullURL := fmt.Sprintf("%s/positions?user=%s&sizeThreshold=0", c.dataAPIBaseURL, walletAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get positions: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /positions")
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}

	var positions []models.PositionEntry
	if err := json.Unmarshal(respBody, &positions); err != nil {
		return nil, fmt.Errorf("get positions: unmarshal response: %w", err)
	}

	return positions, nil
}

func (c *Client) doRequest(req *http.Request, endpoint string) ([]byte, error) {
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
		return nil, &types.APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   endpoint,
			Body:       string(body),
		}
	}

	return body, nil
}

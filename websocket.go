package polytrade

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ws"
)

type WSEvent = ws.Event

type MarketSubscription = ws.Subscription
type UserSubscription = ws.Subscription

func (c *Client) SubscribeMarket(ctx context.Context, assetIDs []string) (*MarketSubscription, error) {
	body, err := json.Marshal(map[string]any{
		"assets_ids":             assetIDs,
		"type":                   "market",
		"custom_feature_enabled": true,
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe market: marshal: %w", err)
	}
	return ws.Open(ctx, ws.MarketURL, body)
}

func (c *Client) SubscribeUser(ctx context.Context, conditionIDs []string) (*UserSubscription, error) {
	c.mu.RLock()
	creds := c.l2Creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	body, err := json.Marshal(map[string]any{
		"auth": map[string]string{
			"apiKey":     creds.APIKey,
			"secret":     creds.Secret,
			"passphrase": creds.Passphrase,
		},
		"markets": conditionIDs,
		"type":    "user",
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe user: marshal: %w", err)
	}
	return ws.Open(ctx, ws.UserURL, body)
}

func (c *Client) SubscribeMarketReconnecting(ctx context.Context, assetIDs []string) (*MarketSubscription, error) {
	body, err := json.Marshal(map[string]any{
		"assets_ids":             assetIDs,
		"type":                   "market",
		"custom_feature_enabled": true,
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe market reconnecting: marshal: %w", err)
	}
	return ws.OpenReconnecting(ctx, ws.MarketURL, body)
}

func (c *Client) SubscribeUserReconnecting(ctx context.Context, conditionIDs []string) (*UserSubscription, error) {
	c.mu.RLock()
	creds := c.l2Creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	body, err := json.Marshal(map[string]any{
		"auth": map[string]string{
			"apiKey":     creds.APIKey,
			"secret":     creds.Secret,
			"passphrase": creds.Passphrase,
		},
		"markets": conditionIDs,
		"type":    "user",
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe user reconnecting: marshal: %w", err)
	}
	return ws.OpenReconnecting(ctx, ws.UserURL, body)
}

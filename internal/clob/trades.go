package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/auth"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

// GetTradeByID fetches a single trade by its trade ID. It returns nil without
// an error when the trade is not visible yet.
func (c *Client) GetTradeByID(ctx context.Context, tradeID string, creds *models.L2Credentials) (*models.Trade, error) {
	if tradeID == "" {
		return nil, fmt.Errorf("get trade by id: empty trade id")
	}

	path := "/data/trades"
	fullPath := path + "?id=" + url.QueryEscape(tradeID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("get trade by id: build request: %w", err)
	}

	headers, err := auth.SignRequest(creds, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get trade by id: %w", err)
	}
	auth.ApplyHeaders(req, headers)

	respBody, err := c.doRequest(req, "GET /data/trades")
	if err != nil {
		return nil, fmt.Errorf("get trade by id: %w", err)
	}

	var page models.PaginatedResponse[models.Trade]
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("get trade by id: unmarshal response: %w", err)
	}

	for i := range page.Data {
		if page.Data[i].ID == tradeID {
			return &page.Data[i], nil
		}
	}
	return nil, nil
}

func tradeResolved(t *models.Trade) bool {
	return t != nil && (t.TransactionHash != "" || tradeFailed(t))
}

func tradeFailed(t *models.Trade) bool {
	return t != nil && strings.EqualFold(t.Status, consts.TradeStatusFailed)
}

// ResolveTradeHashes polls /data/trades for each trade ID until every trade
// carries a transaction hash or reports FAILED, or the timeout expires. It
// never returns an error for an unresolved trade, those IDs come back in
// Pending so the caller can decide what to do.
func (c *Client) ResolveTradeHashes(ctx context.Context, tradeIDs []string, creds *models.L2Credentials, opts *models.ResolveOpts) (*models.TradeResolution, error) {
	res := &models.TradeResolution{}
	if len(tradeIDs) == 0 {
		return res, nil
	}

	interval := consts.DefaultTradePollInterval
	timeout := consts.DefaultTradeResolveTimeout
	if opts != nil {
		if opts.Interval > 0 {
			interval = opts.Interval
		}
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resolved := make(map[string]*models.Trade, len(tradeIDs))
	var mu sync.Mutex

	for {
		var pending []string
		for _, id := range tradeIDs {
			if _, ok := resolved[id]; !ok {
				pending = append(pending, id)
			}
		}
		if len(pending) == 0 {
			break
		}

		var wg sync.WaitGroup
		for _, id := range pending {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				trade, err := c.GetTradeByID(ctx, id, creds)
				if err != nil || !tradeResolved(trade) {
					return
				}
				mu.Lock()
				resolved[id] = trade
				mu.Unlock()
			}(id)
		}
		wg.Wait()

		done := true
		for _, id := range tradeIDs {
			if _, ok := resolved[id]; !ok {
				done = false
				break
			}
		}
		if done {
			break
		}

		select {
		case <-ctx.Done():
			return finishResolution(res, tradeIDs, resolved), nil
		case <-time.After(interval):
		}
	}

	return finishResolution(res, tradeIDs, resolved), nil
}

func finishResolution(res *models.TradeResolution, tradeIDs []string, resolved map[string]*models.Trade) *models.TradeResolution {
	seenHash := make(map[string]bool, len(tradeIDs))
	for _, id := range tradeIDs {
		trade, ok := resolved[id]
		if !ok {
			res.Pending = append(res.Pending, id)
			continue
		}
		if tradeFailed(trade) {
			res.Failed = append(res.Failed, id)
			continue
		}
		res.Trades = append(res.Trades, *trade)
		if !seenHash[trade.TransactionHash] {
			seenHash[trade.TransactionHash] = true
			res.Hashes = append(res.Hashes, trade.TransactionHash)
		}
	}
	return res
}

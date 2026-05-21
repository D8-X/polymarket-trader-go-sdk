package polytrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *CLOBClient) GetCurrentRewards(ctx context.Context) ([]CurrentRewardMarket, error) {
	var all []CurrentRewardMarket
	cursor := ""
	for {
		path := "/rewards/markets/current"
		fullPath := path
		if cursor != "" {
			fullPath = path + "?next_cursor=" + cursor
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
		if err != nil {
			return nil, fmt.Errorf("get current rewards: build request: %w", err)
		}
		respBody, err := c.doRequest(req, "GET /rewards/markets/current")
		if err != nil {
			return nil, fmt.Errorf("get current rewards: %w", err)
		}
		var page PaginatedResponse[CurrentRewardMarket]
		if err := json.Unmarshal(respBody, &page); err != nil {
			return nil, fmt.Errorf("get current rewards: unmarshal response: %w", err)
		}
		all = append(all, page.Data...)
		if page.NextCursor == "" || page.NextCursor == "LTE=" || len(page.Data) == 0 {
			break
		}
		cursor = page.NextCursor
	}
	return all, nil
}

// GetEarningsForUserForDay returns the user's reward earnings for a given UTC
// date (YYYY-MM-DD). The shape of each entry varies and is returned as a raw
// map so callers can decode the fields they need. Requires L2 auth.
func (c *CLOBClient) GetEarningsForUserForDay(ctx context.Context, date string, sigType int, creds *L2Credentials) ([]map[string]any, error) {
	var all []map[string]any
	cursor := ""
	for {
		path := "/rewards/user"
		fullPath := fmt.Sprintf("%s?date=%s&signature_type=%d", path, date, sigType)
		if cursor != "" {
			fullPath += "&next_cursor=" + cursor
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
		if err != nil {
			return nil, fmt.Errorf("get earnings: build request: %w", err)
		}
		headers, err := SignL2Request(creds, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("get earnings: %w", err)
		}
		ApplyL2Headers(req, headers)
		respBody, err := c.doRequest(req, "GET /rewards/user")
		if err != nil {
			return nil, fmt.Errorf("get earnings: %w", err)
		}
		var page PaginatedResponse[map[string]any]
		if err := json.Unmarshal(respBody, &page); err != nil {
			return nil, fmt.Errorf("get earnings: unmarshal response: %w", err)
		}
		all = append(all, page.Data...)
		if page.NextCursor == "" || page.NextCursor == "LTE=" || len(page.Data) == 0 {
			break
		}
		cursor = page.NextCursor
	}
	return all, nil
}

// GetRewardPercentages returns the user's current reward percentage allocations
// across markets. The shape varies and is returned as a raw map. Requires L2 auth.
func (c *CLOBClient) GetRewardPercentages(ctx context.Context, sigType int, creds *L2Credentials) (map[string]any, error) {
	path := "/rewards/user/percentages"
	fullPath := fmt.Sprintf("%s?signature_type=%d", path, sigType)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("get reward percentages: build request: %w", err)
	}
	headers, err := SignL2Request(creds, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get reward percentages: %w", err)
	}
	ApplyL2Headers(req, headers)
	respBody, err := c.doRequest(req, "GET /rewards/user/percentages")
	if err != nil {
		return nil, fmt.Errorf("get reward percentages: %w", err)
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("get reward percentages: unmarshal response: %w", err)
	}
	return result, nil
}

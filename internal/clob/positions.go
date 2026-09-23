package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

const MaxPositionsLimit = 1000

// GetPositions returns every position matching opts, following the cursor to the end.
func (c *Client) GetPositions(ctx context.Context, walletAddress string, opts ...models.PositionsOpts) ([]models.PositionEntry, error) {
	var opt models.PositionsOpts
	switch len(opts) {
	case 0:
	case 1:
		opt = opts[0]
	default:
		return nil, fmt.Errorf("get positions: at most one PositionsOpts, got %d", len(opts))
	}
	if opt.Limit == 0 {
		opt.Limit = MaxPositionsLimit
	}

	var all []models.PositionEntry
	for {
		page, err := c.GetPositionsPage(ctx, walletAddress, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Positions...)
		if !page.Pagination.HasMore {
			return all, nil
		}
		if next := page.Pagination.NextCursor; next == "" || next == opt.Cursor {
			return nil, fmt.Errorf("get positions: has_more without a new cursor after %d positions", len(all))
		}
		opt.Cursor = page.Pagination.NextCursor
	}
}

// GetPositionsPage returns one page of /v2/positions. Pass the page's NextCursor back in
// opts.Cursor, with the other opts unchanged, until HasMore is false.
func (c *Client) GetPositionsPage(ctx context.Context, walletAddress string, opts models.PositionsOpts) (*models.PositionsPage, error) {
	if opts.Limit < 0 || opts.Limit > MaxPositionsLimit {
		return nil, fmt.Errorf("get positions: limit %d outside 0..%d", opts.Limit, MaxPositionsLimit)
	}

	// This might change in the future. Their doc says that only cursor is sufficient.
	// but user also needs to be sent each time.
	query := url.Values{}
	query.Set("user", walletAddress)
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		query.Set("cursor", opts.Cursor)
	}
	if opts.Status != "" {
		query.Set("status", string(opts.Status))
	}
	if len(opts.ConditionIDs) > 0 {
		query.Set("condition", strings.Join(opts.ConditionIDs, ","))
	}
	if opts.Title != "" {
		query.Set("title", opts.Title)
	}
	if opts.MinSize > 0 {
		query.Set("filter_type", "TOKENS")
		query.Set("filter_amount", strconv.FormatFloat(opts.MinSize, 'f', -1, 64))
	}
	if opts.IncludeArchived {
		query.Set("include_archived", "true")
	}
	if opts.SortBy != "" {
		query.Set("sort_by", string(opts.SortBy))
	}
	if opts.Ascending {
		query.Set("sort_direction", "ASC")
	}

	fullURL := c.dataAPIBaseURL + "/v2/positions?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get positions: build request: %w", err)
	}

	respBody, err := c.doRequest(req, "GET /v2/positions")
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}

	var raw struct {
		Data       []positionV2 `json:"data"`
		Pagination struct {
			NextCursor string `json:"next_cursor"`
			HasMore    bool   `json:"has_more"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("get positions: unmarshal response: %w", err)
	}
	page := &models.PositionsPage{
		Positions: make([]models.PositionEntry, len(raw.Data)),
		Pagination: models.PositionsPagination{
			NextCursor: raw.Pagination.NextCursor,
			HasMore:    raw.Pagination.HasMore,
		},
	}
	for i, r := range raw.Data {
		page.Positions[i] = r.entry()
	}
	return page, nil
}

type positionV2 struct {
	TokenID         string                `json:"token_id"`
	ConditionID     string                `json:"condition_id"`
	CurrentSize     float64               `json:"current_size"`
	AvgPrice        float64               `json:"avg_price"`
	CurrentPrice    float64               `json:"current_price"`
	CurrentValue    float64               `json:"current_value"`
	Outcome         string                `json:"outcome"`
	OutcomeIndex    int                   `json:"outcome_index"`
	OppositeTokenID string                `json:"opposite_token_id"`
	Title           string                `json:"title"`
	Slug            string                `json:"slug"`
	EventID         string                `json:"event_id"`
	EndDate         string                `json:"end_date"`
	EntryCostUsdc   float64               `json:"entry_cost_usdc"`
	TotalCostUsdc   float64               `json:"total_cost_usdc"`
	EntryFeesUsdc   float64               `json:"entry_fees_usdc"`
	RealizedPnl     float64               `json:"realized_pnl"`
	UnrealizedPnl   float64               `json:"unrealized_pnl"`
	Status          models.PositionStatus `json:"status"`
	Redeemable      bool                  `json:"redeemable"`
	Mergeable       bool                  `json:"mergeable"`
	NegativeRisk    bool                  `json:"negative_risk"`
	Archived        bool                  `json:"archived"`
}

func (r positionV2) entry() models.PositionEntry {
	return models.PositionEntry{
		Asset:             r.TokenID,
		ConditionID:       r.ConditionID,
		Size:              r.CurrentSize,
		AvgPrice:          r.AvgPrice,
		CurPrice:          r.CurrentPrice,
		Outcome:           r.Outcome,
		Title:             r.Title,
		InitialValue:      r.EntryCostUsdc,
		GrossInitialValue: r.TotalCostUsdc,
		EntryFeesUsdc:     r.EntryFeesUsdc,
		CurrentValue:      r.CurrentValue,
		RealizedPnl:       r.RealizedPnl,
		UnrealizedPnl:     r.UnrealizedPnl,
		OutcomeIndex:      r.OutcomeIndex,
		OppositeAsset:     r.OppositeTokenID,
		Slug:              r.Slug,
		EventID:           r.EventID,
		EndDate:           r.EndDate,
		Status:            r.Status,
		Redeemable:        r.Redeemable,
		Mergeable:         r.Mergeable,
		NegativeRisk:      r.NegativeRisk,
		Archived:          r.Archived,
	}
}

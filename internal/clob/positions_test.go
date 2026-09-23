package clob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

const positionsBody = `{"data":[{
  "token_id": "148537938949645074763140955138487924759672149592656400404380",
  "condition_id": "0xd2a70e3a310382159d16586c0138e8d8fcb9bee96ac19dee4b3881b2f",
  "current_size": 10,
  "avg_price": 0.48,
  "current_price": 0,
  "current_value": 0,
  "outcome": "Toronto Blue Jays",
  "outcome_index": 0,
  "opposite_token_id": "72899352413449242797717683017315332451624415009198038203913",
  "title": "Toronto Blue Jays vs. Baltimore Orioles",
  "entry_cost_usdc": 4.8,
  "total_cost_usdc": 4.87488,
  "entry_fees_usdc": 0.07488,
  "realized_pnl": 0,
  "unrealized_pnl": -4.8,
  "status": "REDEEMABLE",
  "redeemable": true,
  "mergeable": false
},{
  "token_id": "72899352413449242797717683017315332451624415009198038203913",
  "condition_id": "0x9345d5142a67f5541264c96515496affee02580f1c572680759eac9fd",
  "current_size": 8,
  "avg_price": 0.49,
  "current_price": 0.51,
  "current_value": 4.08,
  "outcome": "New York Mets",
  "title": "New York Mets vs. Seattle Mariners",
  "entry_cost_usdc": 3.92,
  "total_cost_usdc": 3.97997,
  "entry_fees_usdc": 0.05997,
  "status": "OPEN",
  "redeemable": false,
  "mergeable": true
}],"pagination":{"limit":2,"offset":0,"has_more":true,"next_cursor":"abc"}}`

func positionsServer(t *testing.T, seen *string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/positions" {
			t.Errorf("path = %s, want /v2/positions", r.URL.Path)
		}
		*seen = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(positionsBody))
	}))
	t.Cleanup(srv.Close)
	c := NewClient()
	c.SetDataAPIBaseURL(srv.URL)
	return c
}

func TestGetPositionsDecodesPage(t *testing.T) {
	var query string
	c := positionsServer(t, &query)

	page, err := c.GetPositions(context.Background(), "0xabc", models.PositionsOpts{Limit: 2})
	if err != nil {
		t.Fatalf("get positions: %v", err)
	}
	if len(page.Positions) != 2 || !page.Pagination.HasMore || page.Pagination.NextCursor != "abc" {
		t.Fatalf("page = %+v", page)
	}
	p := page.Positions[0]
	if p.Asset == "" || p.Size != 10 || p.InitialValue != 4.8 || p.GrossInitialValue != 4.87488 || p.EntryFeesUsdc != 0.07488 {
		t.Fatalf("first row = %+v", p)
	}
	if p.Status != models.PositionStatusRedeemable || !p.Redeemable || p.Mergeable {
		t.Fatalf("first row state = %s/%v/%v", p.Status, p.Redeemable, p.Mergeable)
	}
	live := page.Positions[1]
	if live.Status != models.PositionStatusOpen || live.CurPrice != 0.51 || live.CurrentValue != 4.08 || !live.Mergeable {
		t.Fatalf("second row = %+v", live)
	}
}

func TestGetPositionsQuery(t *testing.T) {
	cases := []struct {
		name string
		opts models.PositionsOpts
		want string
	}{
		{"server default limit", models.PositionsOpts{}, "user=0xabc"},
		{"first page", models.PositionsOpts{Limit: 100}, "limit=100&user=0xabc"},
		{"next page keeps the anchor", models.PositionsOpts{Limit: 100, Cursor: "c1"}, "cursor=c1&limit=100&user=0xabc"},
		{"status", models.PositionsOpts{Limit: 100, Status: models.PositionStatusRedeemableLost}, "limit=100&status=REDEEMABLE_LOST&user=0xabc"},
		{"conditions", models.PositionsOpts{Limit: 100, ConditionIDs: []string{"0x1", "0x2"}}, "condition=0x1%2C0x2&limit=100&user=0xabc"},
		{"title", models.PositionsOpts{Limit: 100, Title: "Mets"}, "limit=100&title=Mets&user=0xabc"},
		{"min size", models.PositionsOpts{Limit: 100, MinSize: 1.5}, "filter_amount=1.5&filter_type=TOKENS&limit=100&user=0xabc"},
		{"negative min size", models.PositionsOpts{Limit: 100, MinSize: -1}, "limit=100&user=0xabc"},
		{"include archived", models.PositionsOpts{Limit: 100, IncludeArchived: true}, "include_archived=true&limit=100&user=0xabc"},
		{"sort", models.PositionsOpts{Limit: 100, SortBy: models.PositionsSortTimestamp, Ascending: true}, "limit=100&sort_by=TIMESTAMP&sort_direction=ASC&user=0xabc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var query string
			c := positionsServer(t, &query)
			if _, err := c.GetPositions(context.Background(), "0xabc", tc.opts); err != nil {
				t.Fatalf("get positions: %v", err)
			}
			if query != tc.want {
				t.Fatalf("query = %q, want %q", query, tc.want)
			}
		})
	}
}

func TestGetPositionsRejectsBadLimit(t *testing.T) {
	for _, limit := range []int{-1, MaxPositionsLimit + 1} {
		var query string
		c := positionsServer(t, &query)
		if _, err := c.GetPositions(context.Background(), "0xabc", models.PositionsOpts{Limit: limit}); err == nil {
			t.Errorf("limit %d: expected error", limit)
		}
		if query != "" {
			t.Errorf("limit %d: request sent despite bad limit", limit)
		}
	}
}

func TestGetBalancesFollowsCursor(t *testing.T) {
	var cursors []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("user") != "0xabc" {
			t.Errorf("page %d lost the user anchor", len(cursors))
		}
		cursors = append(cursors, q.Get("cursor"))
		n := len(cursors)
		next := ""
		if n < 3 {
			next = fmt.Sprint("c", n)
		}
		body, _ := json.Marshal(map[string]any{
			"data":       []map[string]any{{"token_id": fmt.Sprint("tok-", n), "current_size": n}},
			"pagination": map[string]any{"has_more": next != "", "next_cursor": next},
		})
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	c := NewClient()
	c.SetDataAPIBaseURL(srv.URL)

	got, err := c.GetBalances(context.Background(), &models.L2Credentials{Address: "0xabc"})
	if err != nil {
		t.Fatalf("get balances: %v", err)
	}
	if len(got) != 3 || got[2].AssetID != "tok-3" || got[2].Balance != 3 {
		t.Fatalf("balances = %+v", got)
	}
	if fmt.Sprint(cursors) != "[ c1 c2]" {
		t.Fatalf("cursors = %q, want first page bare then c1, c2", cursors)
	}
}

func TestGetBalancesStopsOnStuckCursor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"token_id":"t","current_size":1}],"pagination":{"has_more":true,"next_cursor":"same"}}`))
	}))
	t.Cleanup(srv.Close)
	c := NewClient()
	c.SetDataAPIBaseURL(srv.URL)

	if _, err := c.GetBalances(context.Background(), &models.L2Credentials{Address: "0xabc"}); err == nil {
		t.Fatal("expected an error instead of looping on a repeated cursor")
	}
}

func TestGetPositionsSurfacesAPIError(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusServiceUnavailable} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":"nope","code":"x","retryable":false}`))
		}))
		c := NewClient()
		c.SetDataAPIBaseURL(srv.URL)
		_, err := c.GetPositions(context.Background(), "0xabc", models.PositionsOpts{Limit: 10})
		srv.Close()

		var apiErr *models.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != status {
			t.Fatalf("status %d: got %v, want an APIError with that status", status, err)
		}
	}
}

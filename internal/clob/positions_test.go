package clob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

const positionsBody = `[{
  "asset": "148537938949645074763140955138487924759672149592656400404380",
  "conditionId": "0xd2a70e3a310382159d16586c0138e8d8fcb9bee96ac19dee4b3881b2f",
  "size": 10,
  "avgPrice": 0.48,
  "curPrice": 0,
  "outcome": "Toronto Blue Jays",
  "title": "Toronto Blue Jays vs. Baltimore Orioles",
  "initialValue": 4.8,
  "grossInitialValue": 4.87488,
  "entryFeesUsdc": 0.07488,
  "redeemable": true,
  "mergeable": false
},{
  "asset": "72899352413449242797717683017315332451624415009198038203913",
  "conditionId": "0x9345d5142a67f5541264c96515496affee02580f1c572680759eac9fd",
  "size": 8,
  "avgPrice": 0.49,
  "curPrice": 0,
  "outcome": "New York Mets",
  "title": "New York Mets vs. Seattle Mariners",
  "initialValue": 3.9199,
  "grossInitialValue": 3.97997,
  "entryFeesUsdc": 0.05997
}]`

func positionsServer(t *testing.T, seen *string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(positionsBody))
	}))
	t.Cleanup(srv.Close)
	c := NewClient()
	c.SetDataAPIBaseURL(srv.URL)
	return c
}

func TestGetPositionsDecodesFeeFields(t *testing.T) {
	var query string
	c := positionsServer(t, &query)

	got, err := c.GetPositions(context.Background(), "0xabc", models.PositionsOpts{Limit: 100})
	if err != nil {
		t.Fatalf("get positions: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d positions, want 2", len(got))
	}
	p := got[0]
	if p.InitialValue != 4.8 || p.GrossInitialValue != 4.87488 || p.EntryFeesUsdc != 0.07488 {
		t.Fatalf("fee fields = %v/%v/%v, want 4.8/4.87488/0.07488",
			p.InitialValue, p.GrossInitialValue, p.EntryFeesUsdc)
	}
	if !p.Redeemable || p.Mergeable || got[1].Redeemable {
		t.Fatalf("redeemable/mergeable not decoded: %+v", got)
	}
	for _, e := range got {
		diff := e.GrossInitialValue - (e.InitialValue + e.EntryFeesUsdc)
		if diff > 1e-3 || diff < -1e-3 {
			t.Errorf("%s: gross is %v from initial + fees, beyond rounding", e.Title, diff)
		}
	}
	if exact := got[1].GrossInitialValue - (got[1].InitialValue + got[1].EntryFeesUsdc); exact == 0 {
		t.Error("fixture no longer covers the rounded case, the tolerance is now untested")
	}
}

func TestGetPositionsQuery(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		name string
		opts models.PositionsOpts
		want string
	}{
		{"first page", models.PositionsOpts{Limit: 100}, "limit=100&offset=0&sizeThreshold=0&user=0xabc"},
		{"include archived", models.PositionsOpts{Limit: 100, IncludeArchived: true}, "includeArchived=true&limit=100&offset=0&sizeThreshold=0&user=0xabc"},
		{"custom threshold", models.PositionsOpts{Limit: 100, SizeThreshold: 1.5}, "limit=100&offset=0&sizeThreshold=1.5&user=0xabc"},
		{"negative threshold", models.PositionsOpts{Limit: 100, SizeThreshold: -1}, "limit=100&offset=0&sizeThreshold=0&user=0xabc"},
		{"open only", models.PositionsOpts{Limit: 100, Redeemable: &no}, "limit=100&offset=0&redeemable=false&sizeThreshold=0&user=0xabc"},
		{"mergeable", models.PositionsOpts{Limit: 100, Mergeable: &yes}, "limit=100&mergeable=true&offset=0&sizeThreshold=0&user=0xabc"},
		{"paged archived", models.PositionsOpts{Limit: 50, Offset: 100, IncludeArchived: true},
			"includeArchived=true&limit=50&offset=100&sizeThreshold=0&user=0xabc"},
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

func TestGetPositionsRejectsBadPaging(t *testing.T) {
	cases := []models.PositionsOpts{
		{},
		{Limit: MaxPositionsLimit + 1},
		{Limit: 100, Offset: -1},
		{Limit: 100, Offset: MaxPositionsOffset + 1},
	}
	for _, opts := range cases {
		var query string
		c := positionsServer(t, &query)
		if _, err := c.GetPositions(context.Background(), "0xabc", opts); err == nil {
			t.Errorf("%+v: expected error", opts)
		}
		if query != "" {
			t.Errorf("%+v: request sent despite bad paging", opts)
		}
	}
}

func TestGetBalancesPagesThroughPositions(t *testing.T) {
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsets = append(offsets, r.URL.Query().Get("offset"))
		n := MaxPositionsLimit
		if len(offsets) == 2 {
			n = 3
		}
		page := make([]models.PositionEntry, n)
		_ = json.NewEncoder(w).Encode(page)
	}))
	t.Cleanup(srv.Close)
	c := NewClient()
	c.SetDataAPIBaseURL(srv.URL)

	got, err := c.GetBalances(context.Background(), &models.L2Credentials{Address: "0xabc"})
	if err != nil {
		t.Fatalf("get balances: %v", err)
	}
	if len(got) != MaxPositionsLimit+3 {
		t.Fatalf("got %d balances, want %d", len(got), MaxPositionsLimit+3)
	}
	if len(offsets) != 2 || offsets[0] != "0" || offsets[1] != "500" {
		t.Fatalf("offsets = %v, want [0 500]", offsets)
	}
}

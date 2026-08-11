package clob

import (
	"context"
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
  "entryFeesUsdc": 0.07488
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

	got, err := c.GetPositions(context.Background(), "0xabc")
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

func TestGetPositionsDefaultQuery(t *testing.T) {
	var query string
	c := positionsServer(t, &query)

	if _, err := c.GetPositions(context.Background(), "0xabc"); err != nil {
		t.Fatalf("get positions: %v", err)
	}
	if query != "sizeThreshold=0&user=0xabc" {
		t.Fatalf("query = %q, want the paging defaults with no includeArchived", query)
	}
}

func TestGetPositionsWithOptsQuery(t *testing.T) {
	threshold := 1.5
	limit, offset := 50, 100
	cases := []struct {
		name string
		opts *models.PositionsOpts
		want string
	}{
		{"nil opts matches default", nil, "sizeThreshold=0&user=0xabc"},
		{"include archived", &models.PositionsOpts{IncludeArchived: true}, "includeArchived=true&sizeThreshold=0&user=0xabc"},
		{"custom threshold", &models.PositionsOpts{SizeThreshold: &threshold}, "sizeThreshold=1.5&user=0xabc"},
		{"limit", &models.PositionsOpts{Limit: &limit}, "limit=50&sizeThreshold=0&user=0xabc"},
		{"offset", &models.PositionsOpts{Offset: &offset}, "offset=100&sizeThreshold=0&user=0xabc"},
		{"paged archived", &models.PositionsOpts{IncludeArchived: true, Limit: &limit, Offset: &offset},
			"includeArchived=true&limit=50&offset=100&sizeThreshold=0&user=0xabc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var query string
			c := positionsServer(t, &query)
			if _, err := c.GetPositionsWithOpts(context.Background(), "0xabc", tc.opts); err != nil {
				t.Fatalf("get positions: %v", err)
			}
			if query != tc.want {
				t.Fatalf("query = %q, want %q", query, tc.want)
			}
		})
	}
}

package clob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func newMockServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(handler))
	c := NewClient()
	c.SetBaseURL(srv.URL)
	return c, srv.Close
}

func TestGetMarketsRoundTrip(t *testing.T) {
	c, cleanup := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Errorf("path: got %s want /markets", r.URL.Path)
		}
		if r.URL.Query().Get("next_cursor") != "abc" {
			t.Errorf("cursor: got %s want abc", r.URL.Query().Get("next_cursor"))
		}
		_ = json.NewEncoder(w).Encode(models.PaginatedResponse[models.MarketInfo]{
			Data:       []models.MarketInfo{{ConditionID: "0xc0", Question: "Q", Active: true}},
			NextCursor: "next",
			Count:      1,
			Limit:      1000,
		})
	})
	defer cleanup()

	page, err := c.GetMarkets(context.Background(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 || page.Data[0].ConditionID != "0xc0" {
		t.Errorf("unexpected page: %+v", page)
	}
	if page.NextCursor != "next" {
		t.Errorf("next cursor: got %s want next", page.NextCursor)
	}
}

func TestGetMarketRoundTrip(t *testing.T) {
	c, cleanup := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/0xabc" {
			t.Errorf("path: got %s want /markets/0xabc", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(models.MarketInfo{ConditionID: "0xabc", Question: "Q"})
	})
	defer cleanup()

	mkt, err := c.GetMarket(context.Background(), "0xabc")
	if err != nil {
		t.Fatal(err)
	}
	if mkt.ConditionID != "0xabc" {
		t.Errorf("got %s want 0xabc", mkt.ConditionID)
	}
}

func TestGetMarketByTokenRoundTrip(t *testing.T) {
	c, cleanup := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets-by-token/12345" {
			t.Errorf("path: got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(models.MarketByTokenInfo{
			ConditionID:      "0xc1",
			PrimaryTokenID:   "12345",
			SecondaryTokenID: "67890",
		})
	})
	defer cleanup()

	info, err := c.GetMarketByToken(context.Background(), "12345")
	if err != nil {
		t.Fatal(err)
	}
	if info.ConditionID != "0xc1" || info.PrimaryTokenID != "12345" {
		t.Errorf("got %+v", info)
	}
}

func TestGetSamplingSimplifiedMarketsRoundTrip(t *testing.T) {
	c, cleanup := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sampling-simplified-markets" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(models.PaginatedResponse[models.SimplifiedMarketInfo]{
			Data: []models.SimplifiedMarketInfo{{ConditionID: "0xdead"}},
		})
	})
	defer cleanup()

	page, err := c.GetSamplingSimplifiedMarkets(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 || page.Data[0].ConditionID != "0xdead" {
		t.Errorf("got %+v", page)
	}
}

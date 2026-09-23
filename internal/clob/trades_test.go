package clob

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func testCreds() *models.L2Credentials {
	return &models.L2Credentials{Address: "0xabc", APIKey: "key", Secret: "c2VjcmV0", Passphrase: "pass"}
}

func tradesServer(t *testing.T, handler func(id string) []models.Trade) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			t.Errorf("expected id query param, got %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		data := handler(id)
		_ = json.NewEncoder(w).Encode(models.PaginatedResponse[models.Trade]{Count: len(data), Data: data})
	}))
	t.Cleanup(srv.Close)
	c := NewClient()
	c.SetBaseURL(srv.URL)
	return c
}

func TestResolveTradeHashesPollsUntilHash(t *testing.T) {
	var calls int32
	c := tradesServer(t, func(id string) []models.Trade {
		if atomic.AddInt32(&calls, 1) < 3 {
			return []models.Trade{{ID: id, Status: "MINED"}}
		}
		return []models.Trade{{ID: id, Status: "CONFIRMED", TransactionHash: "0xhash"}}
	})

	res, err := c.ResolveTradeHashes(context.Background(), []string{"t1"}, testCreds(),
		&models.ResolveOpts{Interval: time.Millisecond, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(res.Hashes) != 1 || res.Hashes[0] != "0xhash" {
		t.Fatalf("hashes = %v, want [0xhash]", res.Hashes)
	}
	if len(res.Pending) != 0 || len(res.Failed) != 0 {
		t.Fatalf("pending=%v failed=%v, want both empty", res.Pending, res.Failed)
	}
}

func TestResolveTradeHashesFailedTrade(t *testing.T) {
	c := tradesServer(t, func(id string) []models.Trade {
		return []models.Trade{{ID: id, Status: consts.TradeStatusFailed}}
	})

	res, err := c.ResolveTradeHashes(context.Background(), []string{"t1"}, testCreds(),
		&models.ResolveOpts{Interval: time.Millisecond, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(res.Failed) != 1 || res.Failed[0] != "t1" {
		t.Fatalf("failed = %v, want [t1]", res.Failed)
	}
	if len(res.Hashes) != 0 {
		t.Fatalf("hashes = %v, want empty for a failed trade", res.Hashes)
	}
}

func TestResolveTradeHashesTimesOutIntoPending(t *testing.T) {
	c := tradesServer(t, func(id string) []models.Trade {
		return []models.Trade{{ID: id, Status: "MINING"}}
	})

	res, err := c.ResolveTradeHashes(context.Background(), []string{"t1", "t2"}, testCreds(),
		&models.ResolveOpts{Interval: time.Millisecond, Timeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("resolve should not error on timeout: %v", err)
	}
	if len(res.Pending) != 2 {
		t.Fatalf("pending = %v, want both ids", res.Pending)
	}
}

func TestResolveTradeHashesDedupesSharedHash(t *testing.T) {
	c := tradesServer(t, func(id string) []models.Trade {
		return []models.Trade{{ID: id, Status: "CONFIRMED", TransactionHash: "0xsame"}}
	})

	res, err := c.ResolveTradeHashes(context.Background(), []string{"t1", "t2"}, testCreds(), nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(res.Hashes) != 1 {
		t.Fatalf("hashes = %v, want one deduped hash", res.Hashes)
	}
	if len(res.Trades) != 2 {
		t.Fatalf("trades = %d, want 2", len(res.Trades))
	}
}

func TestResolveTradeHashesNoTradeIDs(t *testing.T) {
	c := NewClient()
	res, err := c.ResolveTradeHashes(context.Background(), nil, testCreds(), nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(res.Hashes) != 0 || len(res.Pending) != 0 {
		t.Fatalf("empty input should resolve to an empty result, got %+v", res)
	}
}

func TestGetTradeByIDMissingTrade(t *testing.T) {
	c := tradesServer(t, func(id string) []models.Trade { return nil })
	trade, err := c.GetTradeByID(context.Background(), "t1", testCreds())
	if err != nil {
		t.Fatalf("get trade: %v", err)
	}
	if trade != nil {
		t.Fatalf("trade = %+v, want nil when not yet visible", trade)
	}
}

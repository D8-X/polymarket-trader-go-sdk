//go:build smoke

package smoke

import (
	"context"
	"testing"

	polytrade "github.com/D8-X/polymarket-trader-go-sdk"
)

const testConditionID = "0x1fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be"

func TestMarketDataSmoke(t *testing.T) {
	ctx := context.Background()
	clob := polytrade.NewCLOBClient()

	t.Log("GetServerTime")
	ts, err := clob.GetServerTime(ctx)
	if err != nil {
		t.Fatalf("server time: %v", err)
	}
	if ts <= 0 {
		t.Fatalf("non-positive server time: %d", ts)
	}
	t.Logf("  server time: %d", ts)

	t.Log("GetClobMarketInfo")
	info, err := clob.GetClobMarketInfo(ctx, testConditionID)
	if err != nil {
		t.Fatalf("clob market info: %v", err)
	}
	if len(info.Tokens) == 0 {
		t.Fatal("empty tokens")
	}
	if info.MinTickSize.String() == "" {
		t.Fatal("empty mts")
	}
	if info.MinOrderSize.String() == "" {
		t.Fatal("empty mos")
	}
	t.Logf("  tick=%s mos=%s tokens=%d", info.MinTickSize, info.MinOrderSize, len(info.Tokens))

	tokenID := info.Tokens[0].TokenID
	if tokenID == "" {
		t.Fatal("empty token id")
	}

	t.Log("GetOrderBook")
	book, err := clob.GetOrderBook(ctx, tokenID)
	if err != nil {
		t.Fatalf("orderbook: %v", err)
	}
	if book.AssetID != tokenID {
		t.Fatalf("asset id mismatch: got %s want %s", book.AssetID, tokenID)
	}
	t.Logf("  bids=%d asks=%d", len(book.Bids), len(book.Asks))

	t.Log("GetMidpoint")
	mid, err := clob.GetMidpoint(ctx, tokenID)
	if err != nil {
		t.Fatalf("midpoint: %v", err)
	}
	if mid == "" {
		t.Fatal("empty midpoint")
	}
	t.Logf("  midpoint: %s", mid)
}

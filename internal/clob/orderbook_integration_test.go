//go:build integration

package clob

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

// This integration test hits the CLOB. It picks a random active market and asserts
// the OrderBook returned by the SDK is sorted
func TestOrderBookFromCLOBIsSorted(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c := NewClient()

	page, err := c.GetSamplingSimplifiedMarkets(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) == 0 {
		t.Skip("no sampling markets returned")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	picks := rng.Perm(len(page.Data))
	for _, idx := range picks {
		m := page.Data[idx]
		if len(m.Tokens) == 0 {
			continue
		}
		tokenID := m.Tokens[0].TokenID
		book, err := c.GetOrderBook(ctx, tokenID)
		if err != nil {
			continue
		}
		if len(book.Bids) == 0 || len(book.Asks) == 0 {
			continue
		}
		if !book.IsSorted() {
			t.Errorf("CLOB returned a non-sorted book for token %s; BestBid/BestAsk assumption broken", tokenID)
		}
		return
	}
	t.Skip("no sampled market had a two-sided book. try again")
}

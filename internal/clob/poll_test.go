package clob

import (
	"context"
	"testing"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func TestIsTerminalStatus(t *testing.T) {
	cases := map[string]bool{
		consts.OrderStatusMatched:   true,
		consts.OrderStatusCanceled:  true,
		consts.OrderStatusUnmatched: true,
		consts.OrderStatusLive:      false,
		consts.OrderStatusDelayed:   false,
		"":                          false,
	}
	for status, want := range cases {
		if got := isTerminalStatus(status); got != want {
			t.Errorf("isTerminalStatus(%q)=%v want %v", status, got, want)
		}
	}
}

func TestAwaitManyReturnsOnUnmatchedOrEmptyID(t *testing.T) {
	c := &Client{}
	responses := []models.PlaceOrderResponse{
		{Success: true, Status: consts.OrderStatusUnmatched, OrderID: "0xabc"},
		{Success: true, Status: consts.OrderStatusDelayed, OrderID: ""},
		{Success: true, Status: "", OrderID: "0xdef", ErrorMsg: "no orders found to match with FAK order."},
	}
	done := make(chan []models.PollResult, 1)
	go func() {
		done <- c.awaitMany(context.Background(), responses, nil, &models.PollOpts{Timeout: 30 * time.Second})
	}()
	select {
	case results := <-done:
		if len(results) != len(responses) {
			t.Fatalf("got %d results, want %d", len(results), len(responses))
		}
		for i, r := range results {
			if r.Err != nil {
				t.Errorf("result %d unexpected err: %v", i, r.Err)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("awaitMany hung instead of returning immediately for unmatched/empty-id orders")
	}
}

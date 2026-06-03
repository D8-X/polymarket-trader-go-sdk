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

func TestResolvedAtPlacement(t *testing.T) {
	cases := []struct {
		name string
		r    models.PlaceOrderResponse
		want bool
	}{
		{"killed no fill", models.PlaceOrderResponse{Success: true, OrderID: "0x1", ErrorMsg: "no orders found"}, true},
		{"empty id", models.PlaceOrderResponse{Success: true, OrderID: ""}, true},
		{"killed but partially filled", models.PlaceOrderResponse{Success: true, OrderID: "0x1", ErrorMsg: "partially filled", TakingAmount: "5"}, false},
		{"matched", models.PlaceOrderResponse{Success: true, OrderID: "0x1", Status: consts.OrderStatusMatched, TakingAmount: "10"}, false},
		{"delayed", models.PlaceOrderResponse{Success: true, OrderID: "0x1", Status: consts.OrderStatusDelayed}, false},
	}
	for _, tc := range cases {
		if got := resolvedAtPlacement(tc.r); got != tc.want {
			t.Errorf("%s: resolvedAtPlacement=%v want %v", tc.name, got, tc.want)
		}
	}
}

func TestAwaitManyUniformStatus(t *testing.T) {
	c := &Client{}
	responses := []models.PlaceOrderResponse{
		{Success: true, Status: "", OrderID: "0xdef", ErrorMsg: "no orders found to match with FAK order."},
		{Success: true, OrderID: ""},
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
			if r.Status == nil {
				t.Fatalf("result %d Status is nil; want uniform status", i)
			}
			if r.Status.Status != consts.OrderStatusUnmatched {
				t.Errorf("result %d status=%q want unmatched", i, r.Status.Status)
			}
			if r.Status.SizeMatched != "0" {
				t.Errorf("result %d size_matched=%q want 0", i, r.Status.SizeMatched)
			}
		}
		if results[0].ErrorMsg == "" {
			t.Errorf("kill reason not surfaced for result 0")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("awaitMany hung instead of returning immediately for resolved-at-placement orders")
	}
}

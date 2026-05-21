//go:build smoke

package smoke

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	polytrade "github.com/D8-X/polymarket-trader-go-sdk/v2"
)

func TestSportsOrderCancelBeforeMatch(t *testing.T) {
	pk := os.Getenv("POLYMARKET_TEST_PK")
	dw := os.Getenv("POLYMARKET_TEST_DEPOSIT_WALLET")
	if pk == "" || dw == "" {
		t.Skip("POLYMARKET_TEST_PK and POLYMARKET_TEST_DEPOSIT_WALLET both required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cli, err := polytrade.NewClient(ctx, polytrade.Config{
		PrivateKeyHex: pk,
		DepositWallet: dw,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	t.Log("scanning markets for an active sports market (seconds_delay > 0)...")
	tokenID, tickSize, conditionID, found := findActiveSportsToken(ctx, t, cli)
	if !found {
		t.Skip("no active sports market found in first 5 pages of sampling markets")
	}
	t.Logf("  using sports market %s tokenID=%s tick=%s", conditionID, tokenID[:16]+"...", tickSize)

	t.Log("placing GTC BUY 5 @ 0.01 (rests on book, below any market, no fill risk)...")
	signed, err := cli.PrepareAndSign(
		tokenID,
		polytrade.BUY,
		polytrade.OrderTypeGTC,
		0.01, 5,
		polytrade.OrderOpts{TickSize: tickSize},
	)
	if err != nil {
		t.Fatalf("prepare and sign: %v", err)
	}
	placeStart := time.Now()
	resp, err := cli.PlaceOrder(ctx, signed)
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	t.Logf("  placed orderID=%s status=%s in %v", resp.OrderID, resp.Status, time.Since(placeStart))
	if !resp.Success {
		t.Fatalf("placement not successful: %+v", resp)
	}

	t.Log("cancelling immediately (within the sports-delay window)...")
	cancelStart := time.Now()
	cancelResp, err := cli.CancelOrder(ctx, resp.OrderID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	t.Logf("  cancel response received in %v: %+v", time.Since(cancelStart), cancelResp)

	confirmed := false
	for _, id := range cancelResp.Canceled {
		if id == resp.OrderID {
			confirmed = true
			break
		}
	}
	if !confirmed {
		t.Fatalf("order not in cancelled list: %+v", cancelResp)
	}
	if total := time.Since(placeStart); total > 3*time.Second {
		t.Errorf("place+cancel round trip took %v", total)
	}
}

func findActiveSportsToken(ctx context.Context, t *testing.T, cli *polytrade.Client) (tokenID, tickSize, conditionID string, ok bool) {
	t.Helper()
	cursor := ""
	for page := 0; page < 5; page++ {
		resp, err := cli.GetSamplingMarkets(ctx, cursor)
		if err != nil {
			t.Fatalf("get sampling markets: %v", err)
		}
		for _, m := range resp.Data {
			if m.Closed || !m.AcceptingOrders || !m.EnableOrderBook {
				continue
			}
			if m.SecondsDelay == 0 {
				continue
			}
			if m.NegRisk {
				continue
			}
			if len(m.Tokens) == 0 {
				continue
			}
			return m.Tokens[0].TokenID, trimTickSize(m.MinimumTickSize), m.ConditionID, true
		}
		if resp.NextCursor == "" || resp.NextCursor == "LTE=" {
			return "", "", "", false
		}
		cursor = resp.NextCursor
	}
	return "", "", "", false
}

func trimTickSize(t float64) string {
	s := strings.TrimRight(strings.TrimRight(formatFloat(t), "0"), ".")
	if s == "" {
		return "0.01"
	}
	return s
}

func formatFloat(f float64) string {
	const prec = 6
	s := ""
	if f == 0.1 {
		s = "0.1"
	} else if f == 0.01 {
		s = "0.01"
	} else if f == 0.001 {
		s = "0.001"
	} else if f == 0.0001 {
		s = "0.0001"
	} else {
		s = "0.01"
	}
	_ = prec
	return s
}

//go:build smoke

package smoke

import (
	"context"
	"os"
	"testing"

	polytrade "github.com/D8-X/polymarket-trader-go-sdk/v2"
)

func TestOpenOrdersAndTradesSmoke(t *testing.T) {
	pk := os.Getenv("POLYMARKET_TEST_PK")
	if pk == "" {
		t.Skip("POLYMARKET_TEST_PK not set")
	}

	creds, err := polytrade.DeriveL2Credentials(pk, polytrade.PolygonChainID)
	if err != nil {
		t.Fatalf("derive creds: %v", err)
	}

	ctx := context.Background()
	clob := polytrade.NewCLOBClient()

	t.Log("GetOpenOrders (V2 /data/orders)")
	orders, err := clob.GetOpenOrders(ctx, "", "", creds)
	if err != nil {
		t.Fatalf("get open orders: %v", err)
	}
	t.Logf("  open orders: %d", len(orders))

	t.Log("GetTrades (V2 /data/trades)")
	trades, err := clob.GetTrades(ctx, creds.Address, "", "", creds)
	if err != nil {
		t.Fatalf("get trades: %v", err)
	}
	t.Logf("  trades: %d", len(trades))
}

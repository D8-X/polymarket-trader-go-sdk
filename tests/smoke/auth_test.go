//go:build smoke

package smoke

import (
	"context"
	"os"
	"testing"

	polytrade "github.com/D8-X/polymarket-trader-go-sdk"
)

func TestAuthSmoke(t *testing.T) {
	pk := os.Getenv("POLYMARKET_TEST_PK")
	if pk == "" {
		t.Skip("POLYMARKET_TEST_PK not set")
	}

	creds, err := polytrade.DeriveL2Credentials(pk, polytrade.PolygonChainID)
	if err != nil {
		t.Fatalf("derive creds: %v", err)
	}
	if creds.APIKey == "" {
		t.Fatal("empty api key")
	}

	clob := polytrade.NewCLOBClient()
	if _, err := clob.GetBalanceAllowance(context.Background(), "COLLATERAL", "", polytrade.SignatureTypeEOA, creds); err != nil {
		t.Fatalf("authed call: %v", err)
	}
}

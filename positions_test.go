package polytrade

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGetPositionsRequiresDepositWallet(t *testing.T) {
	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.GetPositions(context.Background())
	if !errors.Is(err, errNoDepositWallet) {
		t.Errorf("expected errNoDepositWallet, got %v", err)
	}
}

func TestClientGetPositionsCallsDepositWalletAddress(t *testing.T) {
	var queried string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queried = r.URL.Query().Get("user")
		_ = json.NewEncoder(w).Encode([]PositionEntry{
			{Asset: "tok-1", ConditionID: "0x01", Size: 5, AvgPrice: 0.20, CurPrice: 0.22, Outcome: "Yes", Title: "Test market"},
		})
	}))
	defer srv.Close()

	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet})
	if err != nil {
		t.Fatal(err)
	}
	cli.clob.SetDataAPIBaseURL(srv.URL)

	positions, err := cli.GetPositions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(queried, testDepositWallet) {
		t.Errorf("queried user: got %s want %s", queried, testDepositWallet)
	}
	if len(positions) != 1 {
		t.Fatalf("positions: got %d want 1", len(positions))
	}
	if positions[0].Size != 5 {
		t.Errorf("size: got %v want 5", positions[0].Size)
	}
}

func TestClientGetPositionsOfQueriesArbitraryAddress(t *testing.T) {
	var queried string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queried = r.URL.Query().Get("user")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey})
	cli.clob.SetDataAPIBaseURL(srv.URL)

	other := "0x000000000000000000000000000000000000beef"
	_, err := cli.GetPositionsOf(context.Background(), other)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(queried, other) {
		t.Errorf("queried user: got %s want %s", queried, other)
	}
}

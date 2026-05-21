package polytrade

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReplaceOrderRejectsNilOrder(t *testing.T) {
	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	_, _, err := cli.ReplaceOrder(context.Background(), "0x1", nil)
	if err == nil || !strings.Contains(err.Error(), "nil new order") {
		t.Errorf("got %v", err)
	}
}

func TestReplaceOrderCancelsThenPlaces(t *testing.T) {
	var sequence []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/order":
			sequence = append(sequence, "cancel")
			_ = json.NewEncoder(w).Encode(CancelResponse{Canceled: []string{"old-1"}})
		case r.Method == http.MethodPost && r.URL.Path == "/order":
			sequence = append(sequence, "place")
			_ = json.NewEncoder(w).Encode(PlaceOrderResponse{OrderID: "new-1", Success: true, Status: "live"})
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	cli, _ := NewClient(Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		Creds:         &L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	cli.clob.baseURL = srv.URL

	signed, err := cli.PrepareAndSign("100", BUY, OrderTypeGTC, 0.5, 5, OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatal(err)
	}
	cancelResp, placeResp, err := cli.ReplaceOrder(context.Background(), "old-1", signed)
	if err != nil {
		t.Fatal(err)
	}
	if len(sequence) != 2 || sequence[0] != "cancel" || sequence[1] != "place" {
		t.Errorf("call sequence: got %v want [cancel place]", sequence)
	}
	if cancelResp == nil || len(cancelResp.Canceled) != 1 || cancelResp.Canceled[0] != "old-1" {
		t.Errorf("cancel response: %+v", cancelResp)
	}
	if placeResp == nil || placeResp.OrderID != "new-1" {
		t.Errorf("place response: %+v", placeResp)
	}
}

func TestReplaceOrderStopsOnCancelFailure(t *testing.T) {
	var sequence []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/order":
			sequence = append(sequence, "cancel")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"order not found"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/order":
			sequence = append(sequence, "place")
			_ = json.NewEncoder(w).Encode(PlaceOrderResponse{OrderID: "should-not-happen"})
		}
	}))
	defer srv.Close()

	cli, _ := NewClient(Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		Creds:         &L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	cli.clob.baseURL = srv.URL

	signed, err := cli.PrepareAndSign("100", BUY, OrderTypeGTC, 0.5, 5, OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatal(err)
	}
	_, placeResp, err := cli.ReplaceOrder(context.Background(), "old-1", signed)
	if err == nil {
		t.Fatal("expected error on cancel failure")
	}
	if placeResp != nil {
		t.Errorf("place should not have run, got %+v", placeResp)
	}
	if len(sequence) != 1 || sequence[0] != "cancel" {
		t.Errorf("call sequence: got %v want [cancel]", sequence)
	}
}

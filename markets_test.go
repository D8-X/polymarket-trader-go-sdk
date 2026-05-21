package polytrade

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientWrappersHitClob(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/markets/0xabc":
			_ = json.NewEncoder(w).Encode(MarketInfo{ConditionID: "0xabc"})
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey})
	cli.clob.SetBaseURL(srv.URL)
	mkt, err := cli.GetMarket(context.Background(), "0xabc")
	if err != nil {
		t.Fatal(err)
	}
	if mkt.ConditionID != "0xabc" {
		t.Errorf("got %s", mkt.ConditionID)
	}
}

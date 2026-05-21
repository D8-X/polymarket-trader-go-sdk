package polytrade

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetPricesPostsRequestList(t *testing.T) {
	var seenBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/prices" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		seenBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"tok1":{"BUY":"0.55","SELL":"0.60"}}`))
	}))
	defer srv.Close()
	c := NewCLOBClient()
	c.SetBaseURL(srv.URL)

	got, err := c.GetPrices(context.Background(), []PriceRequest{{TokenID: "tok1", Side: "BUY"}, {TokenID: "tok1", Side: "SELL"}})
	if err != nil {
		t.Fatal(err)
	}
	if got["tok1"]["BUY"] != "0.55" || got["tok1"]["SELL"] != "0.60" {
		t.Errorf("got %+v", got)
	}
	var sent []PriceRequest
	if err := json.Unmarshal(seenBody, &sent); err != nil {
		t.Fatal(err)
	}
	if len(sent) != 2 || sent[0].Side != "BUY" {
		t.Errorf("sent body wrong: %+v", sent)
	}
}

func TestGetSpreadsRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tok1":"0.05"}`))
	}))
	defer srv.Close()
	c := NewCLOBClient()
	c.SetBaseURL(srv.URL)

	got, err := c.GetSpreads(context.Background(), []SpreadRequest{{TokenID: "tok1"}})
	if err != nil {
		t.Fatal(err)
	}
	if got["tok1"] != "0.05" {
		t.Errorf("got %+v", got)
	}
}

func TestGetLastTradePriceRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "token_id=tok1") {
			t.Errorf("query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"price":"0.81","side":"BUY"}`))
	}))
	defer srv.Close()
	c := NewCLOBClient()
	c.SetBaseURL(srv.URL)

	got, err := c.GetLastTradePrice(context.Background(), "tok1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Price != "0.81" || got.Side != "BUY" {
		t.Errorf("got %+v", got)
	}
}

func TestGetLastTradePricesRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"price":"0.50","side":"BUY","token_id":"tok1"}]`))
	}))
	defer srv.Close()
	c := NewCLOBClient()
	c.SetBaseURL(srv.URL)

	got, err := c.GetLastTradePrices(context.Background(), []SpreadRequest{{TokenID: "tok1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Price != "0.50" {
		t.Errorf("got %+v", got)
	}
}

func TestGetPricesHistoryRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("market") != "tok1" || q.Get("interval") != "1d" {
			t.Errorf("bad query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"history":[{"t":1700000000,"p":0.55},{"t":1700003600,"p":0.6}]}`))
	}))
	defer srv.Close()
	c := NewCLOBClient()
	c.SetBaseURL(srv.URL)

	got, err := c.GetPricesHistory(context.Background(), PricesHistoryParams{Market: "tok1", Interval: "1d"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Timestamp != 1700000000 || got[1].Price != 0.6 {
		t.Errorf("got %+v", got)
	}
}

func TestGetPricesHistoryRequiresIntervalOrRange(t *testing.T) {
	c := NewCLOBClient()
	_, err := c.GetPricesHistory(context.Background(), PricesHistoryParams{Market: "tok1"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

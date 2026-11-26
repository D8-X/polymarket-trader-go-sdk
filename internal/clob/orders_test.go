package clob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func TestGetOrderNullIsNotFound(t *testing.T) {
	c, done := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("null\n"))
	})
	defer done()
	_, err := c.GetOrder(context.Background(), "0xabc", testCreds())
	if !errors.Is(err, models.ErrOrderNotFound) {
		t.Fatalf("err = %v, want ErrOrderNotFound", err)
	}
}

func TestGetOpenOrdersAddsMarketForAssetFilter(t *testing.T) {
	var gotQuery string
	c, done := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/markets-by-token/42":
			_ = json.NewEncoder(w).Encode(models.MarketByTokenInfo{ConditionID: "0xcond"})
		case "/data/orders":
			gotQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(models.PaginatedResponse[models.OrderStatus]{NextCursor: "LTE="})
		default:
			http.NotFound(w, r)
		}
	})
	defer done()
	if _, err := c.GetOpenOrders(context.Background(), "", "42", testCreds()); err != nil {
		t.Fatal(err)
	}
	if gotQuery != "market=0xcond&asset_id=42" {
		t.Errorf("query = %q, want market=0xcond&asset_id=42", gotQuery)
	}
}

func TestCancelOrdersSplitsLargeBatches(t *testing.T) {
	var sizes []int
	c, done := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var ids []string
		_ = json.NewDecoder(r.Body).Decode(&ids)
		sizes = append(sizes, len(ids))
		_ = json.NewEncoder(w).Encode(models.CancelResponse{Canceled: ids[:1], NotCanceled: map[string]string{ids[1]: "gone"}})
	})
	defer done()
	ids := make([]string, consts.MaxCancelOrderIDs+5)
	for i := range ids {
		ids[i] = fmt.Sprint(i)
	}
	res, err := c.CancelOrders(context.Background(), ids, testCreds())
	if err != nil {
		t.Fatal(err)
	}
	if len(sizes) != 2 || sizes[0] != consts.MaxCancelOrderIDs || sizes[1] != 5 {
		t.Errorf("batch sizes = %v", sizes)
	}
	if len(res.Canceled) != 2 || len(res.NotCanceled) != 2 {
		t.Errorf("merged = %+v", res)
	}
}

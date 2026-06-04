package clob

import (
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func TestFakRemainingSize(t *testing.T) {
	cases := []struct {
		name      string
		orderType string
		side      string
		maker     string
		taker     string
		status    string
		making    string
		taking    string
		want      string
	}{
		{"buy fak full fill", consts.OrderTypeFAK, consts.BUY, "1040000", "13000000", consts.OrderStatusMatched, "1.04", "13", "0"},
		{"buy fak partial fill", consts.OrderTypeFAK, consts.BUY, "1040000", "13000000", consts.OrderStatusMatched, "0.8", "10", "3"},
		{"buy fak no fill", consts.OrderTypeFAK, consts.BUY, "1040000", "13000000", "", "", "", "13"},
		{"buy fak fractional", consts.OrderTypeFAK, consts.BUY, "999600", "14280000", consts.OrderStatusMatched, "0.7", "10", "4.28"},
		{"sell fak partial", consts.OrderTypeFAK, consts.SELL, "13000000", "910000", consts.OrderStatusMatched, "10", "0.7", "3"},
		{"fak delayed -> empty", consts.OrderTypeFAK, consts.BUY, "1040000", "13000000", consts.OrderStatusDelayed, "", "", ""},
		{"gtc -> empty", consts.OrderTypeGTC, consts.BUY, "1040000", "13000000", consts.OrderStatusMatched, "1.04", "13", ""},
		{"fok -> empty", consts.OrderTypeFOK, consts.BUY, "1040000", "13000000", consts.OrderStatusMatched, "1.04", "13", ""},
	}
	for _, tc := range cases {
		signed := &models.SignedOrder{
			OrderType: tc.orderType,
			Order:     models.OrderFields{Side: tc.side, MakerAmount: tc.maker, TakerAmount: tc.taker},
		}
		resp := &models.PlaceOrderResponse{Status: tc.status, MakingAmount: tc.making, TakingAmount: tc.taking}
		if got := fakRemainingSize(signed, resp); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

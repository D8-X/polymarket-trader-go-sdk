package order

import (
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
)

const (
	testPrivateKey    = "0x2222222222222222222222222222222222222222222222222222222222222222"
	testDepositWallet = "0x000000000000000000000000000000000000d077"
	testCTFExchange   = "0xE111180000d2663C0091e4f400237545B87B996B"
)

func TestParseDecimal(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		maxDecimals int
		wantErr     bool
	}{
		{"two-decimals ok", "0.55", 2, false},
		{"two-decimals over", "0.555", 2, true},
		{"one-decimal ok", "0.5", 1, false},
		{"three-decimals ok", "0.555", 3, false},
		{"integer ok", "5", 2, false},
		{"trailing zeros ok", "0.50", 2, false},
		{"empty", "", 2, true},
		{"negative", "-0.5", 2, true},
		{"zero", "0", 2, true},
		{"leading dot", ".5", 2, true},
		{"trailing dot", "10.", 2, true},
		{"multiple dots", "1.2.3", 4, true},
		{"letters", "1a", 2, true},
		{"exponent", "1e3", 6, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseDecimal(tc.in, "v", tc.maxDecimals)
			if (err != nil) != tc.wantErr {
				t.Fatalf("got err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestGetRoundConfig(t *testing.T) {
	cases := []struct {
		tick string
		want roundConfig
	}{
		{"0.1", roundConfig{price: 1, size: 2, amount: 3}},
		{"0.01", roundConfig{price: 2, size: 2, amount: 4}},
		{"0.001", roundConfig{price: 3, size: 2, amount: 5}},
		{"0.0001", roundConfig{price: 4, size: 2, amount: 6}},
		{"", roundConfig{price: 2, size: 2, amount: 4}},
		{"unknown", roundConfig{price: 2, size: 2, amount: 4}},
	}
	for _, tc := range cases {
		t.Run("tick="+tc.tick, func(t *testing.T) {
			got := getRoundConfig(tc.tick)
			if got != tc.want {
				t.Errorf("got %+v want %+v", got, tc.want)
			}
		})
	}
}

func TestPrepareAndSignAmounts(t *testing.T) {
	ob := NewBuilder(testDepositWallet, testCTFExchange, testPrivateKey)

	cases := []struct {
		name      string
		side      string
		orderType string
		tick      string
		price     string
		size      string
		wantMaker string
		wantTaker string
	}{
		{"buy GTC tick 0.01", consts.BUY, consts.OrderTypeGTC, "0.01", "0.55", "10", "5500000", "10000000"},
		{"sell GTC tick 0.01", consts.SELL, consts.OrderTypeGTC, "0.01", "0.55", "10", "10000000", "5500000"},
		{"buy GTC tick 0.001", consts.BUY, consts.OrderTypeGTC, "0.001", "0.555", "10", "5550000", "10000000"},
		{"buy GTC tick 0.0001", consts.BUY, consts.OrderTypeGTC, "0.0001", "0.5555", "10", "5555000", "10000000"},
		{"buy GTC tick 0.1", consts.BUY, consts.OrderTypeGTC, "0.1", "0.5", "5", "2500000", "5000000"},
		{"buy GTC fractional keeps 4dp maker", consts.BUY, consts.OrderTypeGTC, "0.01", "0.07", "14.28", "999600", "14280000"},
		{"buy FAK fractional rounds maker up to 2dp", consts.BUY, consts.OrderTypeFAK, "0.01", "0.07", "14.28", "1000000", "14280000"},
		{"buy FOK fractional rounds maker up to 2dp", consts.BUY, consts.OrderTypeFOK, "0.01", "0.07", "14.28", "1000000", "14280000"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			signed, err := ob.PrepareAndSign("100", tc.side, tc.orderType, tc.price, tc.size, "k", false, Opts{TickSize: tc.tick})
			if err != nil {
				t.Fatalf("prepare: %v", err)
			}
			if signed.Order.MakerAmount != tc.wantMaker {
				t.Errorf("maker: got %s want %s", signed.Order.MakerAmount, tc.wantMaker)
			}
			if signed.Order.TakerAmount != tc.wantTaker {
				t.Errorf("taker: got %s want %s", signed.Order.TakerAmount, tc.wantTaker)
			}
		})
	}
}

func TestPrepareAndSignRejectsSubTickPrice(t *testing.T) {
	ob := NewBuilder(testDepositWallet, testCTFExchange, testPrivateKey)
	_, err := ob.PrepareAndSign("100", consts.BUY, consts.OrderTypeGTC, "0.555", "10", "k", false, Opts{TickSize: "0.01"})
	if err == nil {
		t.Fatal("expected error for sub-tick price")
	}
}

func TestBuilderSetsSignerToFunder(t *testing.T) {
	ob := NewBuilder(testDepositWallet, testCTFExchange, testPrivateKey)
	if ob.SignerAddress() != testDepositWallet {
		t.Errorf("signerAddress: got %s want %s", ob.SignerAddress(), testDepositWallet)
	}
}

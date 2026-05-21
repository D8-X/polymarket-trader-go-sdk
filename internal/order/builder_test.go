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

func TestCheckPrecision(t *testing.T) {
	cases := []struct {
		name     string
		value    float64
		decimals int
		wantErr  bool
	}{
		{"two-decimals ok", 0.55, 2, false},
		{"two-decimals over", 0.555, 2, true},
		{"one-decimal ok", 0.5, 1, false},
		{"three-decimals ok", 0.555, 3, false},
		{"four-decimals over", 0.55555, 4, true},
		{"integer at low precision", 5.0, 2, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkPrecision(tc.value, tc.decimals, "v")
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
		tick      string
		price     float64
		size      float64
		wantMaker string
		wantTaker string
	}{
		{"buy tick 0.01", consts.BUY, "0.01", 0.55, 10, "5500000", "10000000"},
		{"sell tick 0.01", consts.SELL, "0.01", 0.55, 10, "10000000", "5500000"},
		{"buy tick 0.001", consts.BUY, "0.001", 0.555, 10, "5550000", "10000000"},
		{"buy tick 0.0001", consts.BUY, "0.0001", 0.5555, 10, "5555000", "10000000"},
		{"buy tick 0.1", consts.BUY, "0.1", 0.5, 5, "2500000", "5000000"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			signed, err := ob.PrepareAndSign("100", tc.side, consts.OrderTypeGTC, tc.price, tc.size, "k", Opts{TickSize: tc.tick})
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
	_, err := ob.PrepareAndSign("100", consts.BUY, consts.OrderTypeGTC, 0.555, 10, "k", Opts{TickSize: "0.01"})
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

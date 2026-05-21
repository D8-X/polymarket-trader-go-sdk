package polytrade

import (
	"context"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum"
)

type mockCaller struct {
	lastTo   string
	lastData []byte
	result   []byte
	err      error
}

func (m *mockCaller) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if msg.To != nil {
		m.lastTo = msg.To.Hex()
	}
	m.lastData = msg.Data
	return m.result, m.err
}

func TestGetOutcomeTokenBalanceDecodesUint256(t *testing.T) {
	want := big.NewInt(123_456_789)
	mc := &mockCaller{result: ethutil.PadTo32(want.Bytes())}
	got, err := GetOutcomeTokenBalance(context.Background(), mc, testDepositWallet, "42")
	if err != nil {
		t.Fatal(err)
	}
	if got.Cmp(want) != 0 {
		t.Errorf("balance: got %s want %s", got, want)
	}
}

func TestGetOutcomeTokenBalanceCalldata(t *testing.T) {
	mc := &mockCaller{result: ethutil.PadTo32(big.NewInt(0).Bytes())}
	_, err := GetOutcomeTokenBalance(context.Background(), mc, testDepositWallet, "777")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(mc.lastTo, consts.ConditionalTokens) {
		t.Errorf("called to: got %s want %s", mc.lastTo, consts.ConditionalTokens)
	}
	gotHex := hex.EncodeToString(mc.lastData)
	const wantPrefix = "00fdd58e"
	if !strings.HasPrefix(gotHex, wantPrefix) {
		t.Errorf("selector: got %s... want %s...", gotHex[:8], wantPrefix)
	}
	if len(mc.lastData) != 4+32+32 {
		t.Errorf("calldata length: got %d want 68", len(mc.lastData))
	}
}

func TestGetOutcomeTokenBalanceRejectsBadTokenID(t *testing.T) {
	mc := &mockCaller{}
	_, err := GetOutcomeTokenBalance(context.Background(), mc, testDepositWallet, "not-a-number")
	if err == nil {
		t.Fatal("expected error for invalid tokenID")
	}
}

func TestRawBalanceToSize(t *testing.T) {
	cases := []struct {
		balance int64
		want    float64
	}{
		{0, 0},
		{1_000_000, 1.0},
		{5_000_000, 5.0},
		{500_000, 0.5},
		{12_345_678, 12.345678},
	}
	for _, c := range cases {
		got := rawBalanceToSize(big.NewInt(c.balance))
		if got != c.want {
			t.Errorf("balance %d: got %g want %g", c.balance, got, c.want)
		}
	}
	if rawBalanceToSize(nil) != 0 {
		t.Errorf("nil balance: expected 0")
	}
}

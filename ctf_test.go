package polytrade

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
	"github.com/ethereum/go-ethereum/common"
)

func TestEncodeSplitPositionCalldataGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(onchain.EncodeSplitPositionCalldata(
		PUSDAddress,
		"0x1fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be",
		[]int64{1, 2},
		big.NewInt(5_000_000),
	))
	const want = "0x72ce4275000000000000000000000000c011a7e12a19f7b1f670d46f03b03f3342e82dfb00000000000000000000000000000000000000000000000000000000000000001fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be00000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000004c4b40000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"
	if got != want {
		t.Errorf("splitPosition calldata mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestEncodeMergePositionsCalldataGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(onchain.EncodeMergePositionsCalldata(
		PUSDAddress,
		"0x1fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be",
		[]int64{1, 2},
		big.NewInt(1_000_000),
	))
	const want = "0x9e7212ad000000000000000000000000c011a7e12a19f7b1f670d46f03b03f3342e82dfb00000000000000000000000000000000000000000000000000000000000000001fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be00000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000f4240000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"
	if got != want {
		t.Errorf("mergePositions calldata mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestEncodeRedeemPositionsCalldataGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(onchain.EncodeRedeemPositionsCalldata(
		PUSDAddress,
		"0x1fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be",
		[]int64{1, 2},
	))
	const want = "0x01b7037c000000000000000000000000c011a7e12a19f7b1f670d46f03b03f3342e82dfb00000000000000000000000000000000000000000000000000000000000000001fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be0000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"
	if got != want {
		t.Errorf("redeemPositions calldata mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestParseBytes32HexHandlesShortAndLong(t *testing.T) {
	short := onchain.ParseBytes32Hex("0x1234")
	if len(short) != 32 {
		t.Errorf("short len: got %d want 32", len(short))
	}
	if short[31] != 0x34 || short[30] != 0x12 || short[0] != 0 {
		t.Errorf("short padding wrong: %x", short)
	}

	full := "1fad72fae204143ff1c3035e99e7c0f65ea8d5cd9bd1070987bd1a3316f772be"
	fb := onchain.ParseBytes32Hex("0x" + full)
	if "0x"+common.Bytes2Hex(fb) != "0x"+full {
		t.Errorf("full mismatch: got %s want 0x%s", common.Bytes2Hex(fb), full)
	}
}

func TestClientSplitPositionRequiresRelayerCreds(t *testing.T) {
	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet})
	_, err := cli.SplitPosition(context.Background(), "0xc0", big.NewInt(1))
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

func TestClientSplitPositionRejectsBadAmount(t *testing.T) {
	cli, _ := NewClient(Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		RelayerCreds:  &RelayerCredentials{APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	_, err := cli.SplitPosition(context.Background(), "0xc0", big.NewInt(0))
	if err == nil || !strings.Contains(err.Error(), "amount must be positive") {
		t.Errorf("got %v", err)
	}
}

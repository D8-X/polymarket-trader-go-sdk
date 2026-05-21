package onchain

import (
	"math/big"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/ethereum/go-ethereum/common"
)

const (
	testBatchPrivateKey    = "0x2222222222222222222222222222222222222222222222222222222222222222"
	testBatchDepositWallet = "0x000000000000000000000000000000000000d077"
)

func TestDepositWalletDomainSeparatorGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(DepositWalletDomainSeparator(testBatchDepositWallet))
	const want = "0x2e4c72f139823b32bfd3968eae8092eb4e2847319ca7f968938c408fe785281e"
	if got != want {
		t.Errorf("domain sep mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestCallStructHashGolden(t *testing.T) {
	c := models.WalletCall{
		Target: testUSDC,
		Value:  big.NewInt(0),
		Data:   []byte{0xde, 0xad, 0xbe, 0xef},
	}
	got := "0x" + common.Bytes2Hex(CallStructHash(c))
	const want = "0x0ef7aee9dac94c4364bc1795529b515b209e1e3ca4b175b71a56ff0cfc9835b3"
	if got != want {
		t.Errorf("call struct hash mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestSignBatchGolden(t *testing.T) {
	calls := []models.WalletCall{
		{Target: testUSDC, Value: big.NewInt(0), Data: []byte{0xaa}},
		{Target: testCTFExchange, Value: big.NewInt(0), Data: []byte{0xbb}},
	}
	got, err := SignBatch(testBatchPrivateKey, testBatchDepositWallet, 7, 1_750_000_000, calls)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	const want = "0x50772c33a79d98b353d4b80e5ced14adcd9a61334c2187cd84e2aff4269c257216907483123f85baa4af46482ca3e00f7192b752e814311fbf1ec804285ce3721c"
	if got != want {
		t.Errorf("batch sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

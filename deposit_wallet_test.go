package polytrade

import (
	"math/big"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const testDepositWallet = "0x000000000000000000000000000000000000d077"

func TestDepositWalletDomainSeparatorGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(depositWalletDomainSeparator(testDepositWallet))
	const want = "0x2e4c72f139823b32bfd3968eae8092eb4e2847319ca7f968938c408fe785281e"
	if got != want {
		t.Errorf("domain sep mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestCallStructHashGolden(t *testing.T) {
	c := WalletCall{
		Target: USDCAddress,
		Value:  big.NewInt(0),
		Data:   []byte{0xde, 0xad, 0xbe, 0xef},
	}
	got := "0x" + common.Bytes2Hex(callStructHash(c))
	const want = "0x0ef7aee9dac94c4364bc1795529b515b209e1e3ca4b175b71a56ff0cfc9835b3"
	if got != want {
		t.Errorf("call struct hash mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestSignBatchGolden(t *testing.T) {
	calls := []WalletCall{
		{Target: USDCAddress, Value: big.NewInt(0), Data: []byte{0xaa}},
		{Target: CTFExchange, Value: big.NewInt(0), Data: []byte{0xbb}},
	}
	got, err := signBatch(testPrivateKey, testDepositWallet, 7, 1_750_000_000, calls)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	const want = "0x50772c33a79d98b353d4b80e5ced14adcd9a61334c2187cd84e2aff4269c257216907483123f85baa4af46482ca3e00f7192b752e814311fbf1ec804285ce3721c"
	if got != want {
		t.Errorf("batch sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestWrapPoly1271OrderSignatureGolden(t *testing.T) {
	ob := NewOrderBuilder(testDepositWallet, CTFExchange, testPrivateKey, SignatureTypePoly1271)
	order := OrderFields{
		Salt:          12345,
		Maker:         testDepositWallet,
		Signer:        testDepositWallet,
		TokenID:       "100000000000000000000000000000000000000000000000000000000000000000000000000000",
		MakerAmount:   "1000000",
		TakerAmount:   "5000000",
		Expiration:    "0",
		Timestamp:     "1716000000000",
		Metadata:      ZeroBytes32,
		Builder:       ZeroBytes32,
		Side:          BUY,
		SignatureType: SignatureTypePoly1271,
		sideNumeric:   SideBuy,
	}
	got, err := ob.signOrder(order)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	const want = "0xc1c199d3822d7f465b1f213793ebb7bbc3a77810ceefa89b58ecfeb284e708953c5ec946a41c6fa2a6fac373c78b167dd6834fae6f3e37581111ffb06200f1d21b3264e159346253e26a64e00b69032db0e7d32f94628de3e6eecb50304d7af3d26814859c22020105275eba8b46528be2edd6078dea415bec6d90f2076b0c8ac64f726465722875696e743235362073616c742c61646472657373206d616b65722c61646472657373207369676e65722c75696e7432353620746f6b656e49642c75696e74323536206d616b6572416d6f756e742c75696e743235362074616b6572416d6f756e742c75696e743820736964652c75696e7438207369676e6174757265547970652c75696e743235362074696d657374616d702c62797465733332206d657461646174612c62797465733332206275696c6465722900ba"
	if got != want {
		t.Errorf("wrapped sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
	if len(got) < 600 || len(got) > 700 {
		t.Errorf("wrapped sig length %d hex chars, expected ~636", len(got))
	}
}

func TestPoly1271BuilderSetsSignerToFunder(t *testing.T) {
	ob := NewOrderBuilder(testDepositWallet, CTFExchange, testPrivateKey, SignatureTypePoly1271)
	if ob.signerAddress != testDepositWallet {
		t.Errorf("signerAddress: got %s want %s", ob.signerAddress, testDepositWallet)
	}
}

func TestNonPoly1271BuilderSetsSignerToEOA(t *testing.T) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(testPrivateKey))
	if err != nil {
		t.Fatal(err)
	}
	wantEOA := crypto.PubkeyToAddress(pk.PublicKey).Hex()
	ob := NewOrderBuilder("0xabc", CTFExchange, testPrivateKey, SignatureTypeGnosisSafe)
	if ob.signerAddress != wantEOA {
		t.Errorf("signerAddress: got %s want %s", ob.signerAddress, wantEOA)
	}
}

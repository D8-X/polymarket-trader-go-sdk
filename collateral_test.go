package polytrade

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestEncodeOnrampWrapCalldataGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(encodeOnrampWrapCalldata(USDCAddress, testDepositWallet, big.NewInt(1_250_000)))
	const want = "0x623556380000000000000000000000002791bca1f2de4661ed88a30c99a7a9449aa84174000000000000000000000000000000000000000000000000000000000000d07700000000000000000000000000000000000000000000000000000000001312d0"
	if got != want {
		t.Errorf("wrap calldata mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestEncodeOfframpUnwrapCalldataGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(encodeOfframpUnwrapCalldata(USDCAddress, testDepositWallet, big.NewInt(500_000)))
	const want = "0x8cc7104f0000000000000000000000002791bca1f2de4661ed88a30c99a7a9449aa84174000000000000000000000000000000000000000000000000000000000000d077000000000000000000000000000000000000000000000000000000000007a120"
	if got != want {
		t.Errorf("unwrap calldata mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestEncodeSetApprovalForAllCalldataTrueGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(encodeSetApprovalForAllCalldata(CTFExchange, true))
	const want = "0xa22cb465000000000000000000000000e111180000d2663c0091e4f400237545b87b996b0000000000000000000000000000000000000000000000000000000000000001"
	if got != want {
		t.Errorf("setApprovalForAll(true) calldata mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestEncodeSetApprovalForAllCalldataFalseGolden(t *testing.T) {
	got := "0x" + common.Bytes2Hex(encodeSetApprovalForAllCalldata(NegRiskCTFExchange, false))
	const want = "0xa22cb465000000000000000000000000e2222d279d744050d28e00520010520000310f590000000000000000000000000000000000000000000000000000000000000000"
	if got != want {
		t.Errorf("setApprovalForAll(false) calldata mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestEncodeApproveCalldataAmountGolden(t *testing.T) {
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	got := "0x" + common.Bytes2Hex(encodeApproveCalldataAmount(CTFExchange, maxU))
	const want = "0x095ea7b3000000000000000000000000e111180000d2663c0091e4f400237545b87b996bffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	if got != want {
		t.Errorf("approve calldata mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

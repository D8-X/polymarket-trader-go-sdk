package polytrade

import (
	"math/big"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
)

func TestEncodeWrapCalldataGolden(t *testing.T) {
	selector := ethutil.Keccak256([]byte("wrap(address,address,uint256)"))[:4]
	got := encodeWrapCalldata(selector, USDCAddress, "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", big.NewInt(1_000_000))
	const want = "0x623556380000000000000000000000002791bca1f2de4661ed88a30c99a7a9449aa84174000000000000000000000000d8da6bf26964af9d7eed9e03e53415d37aa9604500000000000000000000000000000000000000000000000000000000000f4240"
	if got != want {
		t.Errorf("encoded mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

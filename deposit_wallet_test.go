package polytrade

import (
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	testPrivateKey    = "0x2222222222222222222222222222222222222222222222222222222222222222"
	testDepositWallet = "0x000000000000000000000000000000000000d077"
)

func testEOA(t *testing.T) string {
	t.Helper()
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(testPrivateKey))
	if err != nil {
		t.Fatalf("parse key: %v", err)
	}
	return crypto.PubkeyToAddress(pk.PublicKey).Hex()
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

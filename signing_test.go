package polytrade

import (
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	testPrivateKey = "0x2222222222222222222222222222222222222222222222222222222222222222"
	testChainID    = 137
)

func testEOA(t *testing.T) string {
	t.Helper()
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(testPrivateKey))
	if err != nil {
		t.Fatalf("parse key: %v", err)
	}
	return crypto.PubkeyToAddress(pk.PublicKey).Hex()
}

func TestSignClobAuthGolden(t *testing.T) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(testPrivateKey))
	if err != nil {
		t.Fatalf("parse key: %v", err)
	}
	address := crypto.PubkeyToAddress(pk.PublicKey).Hex()

	const (
		timestamp = "1716000000"
		nonce     = int64(0)
	)

	domainSep := hashClobAuthDomain(testChainID)
	structHash := hashClobAuthStruct(address, timestamp, nonce)
	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)

	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}
	got := "0x" + common.Bytes2Hex(sig)

	const want = "0x17c74c7a204cd41c4784756cbd9c0e77e33ca9e14acfb562f520560ebbc4a74443226f0bfc99510e0738219ef67f83e79f298160297260424ed39df06b7e6fcc1b"
	if got != want {
		t.Errorf("sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}


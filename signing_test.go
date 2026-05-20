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

func TestSignOrderGolden(t *testing.T) {
	eoa := testEOA(t)
	safe := DeriveSafeAddress(eoa)

	ob := NewOrderBuilder(safe, eoa, CTFExchange, testPrivateKey, SignatureTypeGnosisSafe)
	order := OrderFields{
		Salt:          12345,
		Maker:         safe,
		Signer:        eoa,
		TokenID:       "100000000000000000000000000000000000000000000000000000000000000000000000000000",
		MakerAmount:   "1000000",
		TakerAmount:   "1818181",
		Expiration:    "0",
		Timestamp:     "1716000000000",
		Metadata:      ZeroBytes32,
		Builder:       ZeroBytes32,
		Side:          BUY,
		SignatureType: SignatureTypeGnosisSafe,
		sideNumeric:   SideBuy,
	}

	got, err := ob.signOrder(order)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	const want = "0xb9136e02806f156ad83b35df2b53af1d7c3d18f916706278ff92956f6009e4727a975e4f429a3cfbc97737a2c36453e8db028516a08f41ef804a5cd1dcfcf7ec1b"
	if got != want {
		t.Errorf("sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
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

func TestSignSafeTxGolden(t *testing.T) {
	eoa := testEOA(t)
	safe := DeriveSafeAddress(eoa)
	tx := SafeTransaction{
		To:        USDCAddress,
		Value:     "0",
		Data:      "0xa9059cbb" + leftPadHex(eoa[2:], 64) + leftPadHex("f4240", 64),
		Operation: OperationCall,
	}
	got, err := signSafeTx(testPrivateKey, safe, tx, "7")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	const want = "0x084a1308a2bcbe05ee489ca24c88280934bc457242531bad24ddf2b08ec685a220ce32691a59f7794370ee2a9c1ef3636e8836395467d89142e47dca0b0178641f"
	if got != want {
		t.Errorf("sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestSignSafeCreateGolden(t *testing.T) {
	got, err := signSafeCreate(testPrivateKey)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	const want = "0xe3331f503f0639e8cb08bb21eb434c26aa5c47b9b8106ed5d28c630d6cb656296ad9befa65c3c5a06d4a6184fe18c2380aea07a0a485b9099c342b38381e0fb71b"
	if got != want {
		t.Errorf("sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func leftPadHex(s string, width int) string {
	for len(s) < width {
		s = "0" + s
	}
	return s
}

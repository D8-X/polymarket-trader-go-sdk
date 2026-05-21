package sign

import (
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

const (
	testPrivateKey    = "0x2222222222222222222222222222222222222222222222222222222222222222"
	testDepositWallet = "0x000000000000000000000000000000000000d077"
	testCTFExchange   = "0xE111180000d2663C0091e4f400237545B87B996B"
)

func TestWrapPoly1271OrderSignatureGolden(t *testing.T) {
	order := models.OrderFields{
		Salt:          12345,
		Maker:         testDepositWallet,
		Signer:        testDepositWallet,
		TokenID:       "100000000000000000000000000000000000000000000000000000000000000000000000000000",
		MakerAmount:   "1000000",
		TakerAmount:   "5000000",
		Expiration:    "0",
		Timestamp:     "1716000000000",
		Metadata:      consts.ZeroBytes32,
		Builder:       consts.ZeroBytes32,
		Side:          "BUY",
		SignatureType: consts.SignatureTypePoly1271,
		SideNumeric:   consts.SideBuy,
	}
	got, err := Order(testPrivateKey, testCTFExchange, order)
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

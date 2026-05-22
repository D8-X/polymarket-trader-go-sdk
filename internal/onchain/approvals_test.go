package onchain

import (
	"context"
	"math/big"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type stubCaller struct {
	responses map[string][]byte
}

func (s stubCaller) CallContract(ctx context.Context, msg ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	key := msg.To.Hex() + ":" + string(msg.Data[:4])
	if v, ok := s.responses[key]; ok {
		return v, nil
	}
	return make([]byte, 32), nil
}

func makeApprovedStub(allowanceVal *big.Int, isApprovedTrue bool) stubCaller {
	pad := func(v *big.Int) []byte {
		b := make([]byte, 32)
		v.FillBytes(b)
		return b
	}
	flag := make([]byte, 32)
	if isApprovedTrue {
		flag[31] = 1
	}
	pusd := common.HexToAddress(consts.PUSDAddress)
	ct := common.HexToAddress(consts.ConditionalTokens)
	allowanceSel := string([]byte{0xdd, 0x62, 0xed, 0x3e})
	isApprovedSel := string([]byte{0xe9, 0x85, 0xe9, 0xc5})
	return stubCaller{responses: map[string][]byte{
		pusd.Hex() + ":" + allowanceSel: pad(allowanceVal),
		ct.Hex() + ":" + isApprovedSel:  flag,
	}}
}

func TestIsFullyApproved_AllSet(t *testing.T) {
	big2_200 := new(big.Int).Lsh(big.NewInt(1), 200)
	caller := makeApprovedStub(big2_200, true)
	ok, err := IsFullyApproved(context.Background(), caller, common.HexToAddress("0xf7d4a87a96c1d4b2bb05b2750ed6a6f9c7eb5e62"))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected true when all approvals are set")
	}
}

func TestIsFullyApproved_AllowanceTooLow(t *testing.T) {
	caller := makeApprovedStub(big.NewInt(1000), true)
	ok, err := IsFullyApproved(context.Background(), caller, common.HexToAddress("0xf7d4a87a96c1d4b2bb05b2750ed6a6f9c7eb5e62"))
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected false when allowance is below threshold")
	}
}

func TestIsFullyApproved_IsApprovedForAllFalse(t *testing.T) {
	big2_200 := new(big.Int).Lsh(big.NewInt(1), 200)
	caller := makeApprovedStub(big2_200, false)
	ok, err := IsFullyApproved(context.Background(), caller, common.HexToAddress("0xf7d4a87a96c1d4b2bb05b2750ed6a6f9c7eb5e62"))
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected false when ConditionalTokens isApprovedForAll is false")
	}
}

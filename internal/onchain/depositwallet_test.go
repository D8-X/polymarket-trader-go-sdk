package onchain

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestDeriveDepositWallet_KnownPairs(t *testing.T) {
	cases := []struct {
		eoa, want string
	}{
		{"0xdC110F9F3B1E0fE71A635ED4C5728f39fE61625B", "0xf7d4a87a96c1d4b2bb05b2750ed6a6f9c7eb5e62"},
		{"0x98c362C91aF97ce0dc70ad44235d7C5745cC139b", "0x31b4e1f76d880e52C399d0d58596e4E42ef90B10"},
		{"0x11c71b3be69224a6f42AC1c000478Dd8Bb611551", "0xd10A0aa0Df742C70802afaA586B9F2A0225584Ff"},
		{"0x11c17DDE0EaA6BE3851261226648e5e585CeeB3F", "0x25cA56960A98Af051A80Adc9887f20a2854563d8"},
		{"0xA6f4f0a2eeC81630601F33F2df39830EdA7AF519", "0x1503C91F0af317a08a4300677463cA395DcEb99a"},
	}
	for _, c := range cases {
		got := DeriveDepositWallet(common.HexToAddress(c.eoa))
		want := common.HexToAddress(c.want)
		if got != want {
			t.Errorf("EOA %s: got %s want %s", c.eoa, got.Hex(), want.Hex())
		}
	}
}

func TestDeriveDepositWallet_Deterministic(t *testing.T) {
	eoa := common.HexToAddress("0xdC110F9F3B1E0fE71A635ED4C5728f39fE61625B")
	a := DeriveDepositWallet(eoa)
	b := DeriveDepositWallet(eoa)
	if a != b {
		t.Errorf("non-deterministic: %s != %s", a.Hex(), b.Hex())
	}
}

type stubCodeReader struct {
	code []byte
	err  error
}

func (s stubCodeReader) CodeAt(ctx context.Context, addr common.Address, blk *big.Int) ([]byte, error) {
	return s.code, s.err
}

func TestLookupDepositWallet_DeployedTrue(t *testing.T) {
	eoa := common.HexToAddress("0xdC110F9F3B1E0fE71A635ED4C5728f39fE61625B")
	addr, deployed, err := LookupDepositWallet(context.Background(), stubCodeReader{code: []byte{0x36, 0x3d}}, eoa)
	if err != nil {
		t.Fatal(err)
	}
	if addr.Hex() != "0xF7d4a87A96C1D4B2BB05b2750ed6A6f9c7eb5E62" {
		t.Errorf("addr: got %s", addr.Hex())
	}
	if !deployed {
		t.Error("expected deployed=true when CodeAt returns bytecode")
	}
}

func TestLookupDepositWallet_DeployedFalse(t *testing.T) {
	eoa := common.HexToAddress("0xdC110F9F3B1E0fE71A635ED4C5728f39fE61625B")
	_, deployed, err := LookupDepositWallet(context.Background(), stubCodeReader{code: nil}, eoa)
	if err != nil {
		t.Fatal(err)
	}
	if deployed {
		t.Error("expected deployed=false when CodeAt returns empty")
	}
}

func TestLookupDepositWallet_NilEth(t *testing.T) {
	eoa := common.HexToAddress("0xdC110F9F3B1E0fE71A635ED4C5728f39fE61625B")
	addr, deployed, err := LookupDepositWallet(context.Background(), nil, eoa)
	if err != nil {
		t.Fatal(err)
	}
	if deployed {
		t.Error("expected deployed=false when eth is nil")
	}
	if (addr == common.Address{}) {
		t.Error("address should still be derived even with nil eth")
	}
}

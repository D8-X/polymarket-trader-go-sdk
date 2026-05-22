package onchain

import (
	"context"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
)

var (
	erc1967Prefix  = []byte{0x61, 0x00, 0x3D, 0x3D, 0x81, 0x60, 0x23, 0x3D, 0x39, 0x73}
	erc1967Push    = []byte{0x60, 0x09}
	erc1967Const1  = mustHex("cc3735a920a3ca505d382bbc545af43d6000803e6038573d6000fd5b3d6000f3")
	erc1967Const2  = mustHex("5155f3363d3d373d3d363d7f360894a13ba1a3210667c828492db98dca3e2076")
)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		panic(err)
	}
	return b
}

func padAddr(addr common.Address) []byte {
	out := make([]byte, 32)
	copy(out[12:], addr.Bytes())
	return out
}

func DeriveDepositWallet(owner common.Address) common.Address {
	factory := common.HexToAddress(consts.DepositWalletFactory)
	impl := common.HexToAddress(consts.DepositWalletImplementation)

	args := append(padAddr(factory), padAddr(owner)...)

	if len(args) > 0xC2 {
		panic("DeriveDepositWallet: args size would overflow the ERC-1967 prefix byte; constructor structure changed")
	}
	prefix := append([]byte{}, erc1967Prefix...)
	prefix[2] += byte(len(args))

	initCode := append([]byte{}, prefix...)
	initCode = append(initCode, impl.Bytes()...)
	initCode = append(initCode, erc1967Push...)
	initCode = append(initCode, erc1967Const2...)
	initCode = append(initCode, erc1967Const1...)
	initCode = append(initCode, args...)

	bytecodeHash := ethutil.Keccak256(initCode)
	salt := ethutil.Keccak256(args)

	data := []byte{0xff}
	data = append(data, factory.Bytes()...)
	data = append(data, salt...)
	data = append(data, bytecodeHash...)
	return common.BytesToAddress(ethutil.Keccak256(data)[12:])
}

type CodeReader interface {
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
}

func LookupDepositWallet(ctx context.Context, eth CodeReader, owner common.Address) (addr common.Address, deployed bool, err error) {
	addr = DeriveDepositWallet(owner)
	if eth == nil {
		return addr, false, nil
	}
	code, err := eth.CodeAt(ctx, addr, nil)
	if err != nil {
		return addr, false, err
	}
	return addr, len(code) > 0, nil
}

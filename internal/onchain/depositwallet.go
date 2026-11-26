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
	erc1967Prefix = []byte{0x61, 0x00, 0x3D, 0x3D, 0x81, 0x60, 0x23, 0x3D, 0x39, 0x73}
	erc1967Push   = []byte{0x60, 0x09}
	erc1967Const1 = mustHex("cc3735a920a3ca505d382bbc545af43d6000803e6038573d6000fd5b3d6000f3")
	erc1967Const2 = mustHex("5155f3363d3d373d3d363d7f360894a13ba1a3210667c828492db98dca3e2076")

	beaconPrefix = []byte{0x61, 0x00, 0x52, 0x3D, 0x81, 0x60, 0x23, 0x3D, 0x39, 0x73}
	beaconConst1 = mustHex("b3582b35133d50545afa5036515af43d6000803e604d573d6000fd5b3d6000f3")
	beaconConst2 = mustHex("1b60e01b36527fa3f0ad74e5423aebfd80d3ef4346578335a9a72aeaee59ff6c")
	beaconConst3 = mustHex("60195155f3363d3d373d3d363d602036600436635c60da")
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

func walletArgs(owner common.Address) []byte {
	return append(padAddr(common.HexToAddress(consts.DepositWalletFactory)), padAddr(owner)...)
}

func sizedPrefix(prefix, args []byte) []byte {
	if len(args) > 0xff-int(prefix[2]) {
		panic("deposit wallet: args size would overflow the ERC-1967 prefix byte; constructor structure changed")
	}
	out := append([]byte{}, prefix...)
	out[2] += byte(len(args))
	return out
}

func create2(args, initCode []byte) common.Address {
	data := []byte{0xff}
	data = append(data, common.HexToAddress(consts.DepositWalletFactory).Bytes()...)
	data = append(data, ethutil.Keccak256(args)...)
	data = append(data, ethutil.Keccak256(initCode)...)
	return common.BytesToAddress(ethutil.Keccak256(data)[12:])
}

// DeriveDepositWallet returns the beacon proxy address the factory deploys for owner.
func DeriveDepositWallet(owner common.Address) common.Address {
	args := walletArgs(owner)
	initCode := sizedPrefix(beaconPrefix, args)
	initCode = append(initCode, common.HexToAddress(consts.DepositWalletBeacon).Bytes()...)
	initCode = append(initCode, beaconConst3...)
	initCode = append(initCode, beaconConst2...)
	initCode = append(initCode, beaconConst1...)
	initCode = append(initCode, args...)
	return create2(args, initCode)
}

// DeriveLegacyDepositWallet returns the UUPS proxy address used by wallets the
// factory deployed before it switched to beacon proxies.
func DeriveLegacyDepositWallet(owner common.Address) common.Address {
	args := walletArgs(owner)
	initCode := sizedPrefix(erc1967Prefix, args)
	initCode = append(initCode, common.HexToAddress(consts.DepositWalletImplementation).Bytes()...)
	initCode = append(initCode, erc1967Push...)
	initCode = append(initCode, erc1967Const2...)
	initCode = append(initCode, erc1967Const1...)
	initCode = append(initCode, args...)
	return create2(args, initCode)
}

type CodeReader interface {
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
}

// LookupDepositWallet prefers a deployed legacy wallet and otherwise reports the beacon address.
func LookupDepositWallet(ctx context.Context, eth CodeReader, owner common.Address) (addr common.Address, deployed bool, err error) {
	addr = DeriveDepositWallet(owner)
	if eth == nil {
		return addr, false, nil
	}
	legacy := DeriveLegacyDepositWallet(owner)
	code, err := eth.CodeAt(ctx, legacy, nil)
	if err != nil {
		return addr, false, err
	}
	if len(code) > 0 {
		return legacy, true, nil
	}
	code, err = eth.CodeAt(ctx, addr, nil)
	if err != nil {
		return addr, false, err
	}
	return addr, len(code) > 0, nil
}

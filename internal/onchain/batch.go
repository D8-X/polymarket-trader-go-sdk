package onchain

import (
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func DepositWalletDomainSeparator(walletAddress string) []byte {
	domainTypeHash := ethutil.Keccak256([]byte(consts.EIP712DepositWalletDomainType))
	return ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(domainTypeHash),
		ethutil.PadTo32(ethutil.Keccak256([]byte(consts.EIP712DepositWalletName))),
		ethutil.PadTo32(ethutil.Keccak256([]byte(consts.EIP712DepositWalletVersion))),
		ethutil.PadTo32(big.NewInt(consts.PolygonChainID).Bytes()),
		ethutil.PadTo32(common.HexToAddress(walletAddress).Bytes()),
	))
}

func CallStructHash(c types.WalletCall) []byte {
	value := c.Value
	if value == nil {
		value = new(big.Int)
	}
	return ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(ethutil.Keccak256([]byte(consts.EIP712CallType))),
		ethutil.PadTo32(common.HexToAddress(c.Target).Bytes()),
		ethutil.PadTo32(value.Bytes()),
		ethutil.Keccak256(c.Data),
	))
}

func SignBatch(privateKeyHex, walletAddress string, nonce, deadline int64, calls []types.WalletCall) (string, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("sign batch: invalid private key: %w", err)
	}
	domainSep := DepositWalletDomainSeparator(walletAddress)
	var callHashes []byte
	for _, c := range calls {
		callHashes = append(callHashes, CallStructHash(c)...)
	}
	callsArrayHash := ethutil.Keccak256(callHashes)
	structHash := ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(ethutil.Keccak256([]byte(consts.EIP712BatchType))),
		ethutil.PadTo32(common.HexToAddress(walletAddress).Bytes()),
		ethutil.PadTo32(big.NewInt(nonce).Bytes()),
		ethutil.PadTo32(big.NewInt(deadline).Bytes()),
		callsArrayHash,
	))
	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)
	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		return "", fmt.Errorf("sign batch: sign EIP-712 digest: %w", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}
	return "0x" + common.Bytes2Hex(sig), nil
}

package sign

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func WrapPoly1271Signature(pk *ecdsa.PrivateKey, depositWallet string, appDomainSep, contentsHash []byte) (string, error) {
	soladyTypeHash := ethutil.Keccak256([]byte(consts.EIP712SoladyTypedDataSignType))
	dwNameHash := ethutil.Keccak256([]byte(consts.EIP712DepositWalletName))
	dwVersionHash := ethutil.Keccak256([]byte(consts.EIP712DepositWalletVersion))

	tsHash := ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(soladyTypeHash),
		contentsHash,
		ethutil.PadTo32(dwNameHash),
		ethutil.PadTo32(dwVersionHash),
		ethutil.PadTo32(big.NewInt(consts.PolygonChainID).Bytes()),
		ethutil.PadTo32(common.HexToAddress(depositWallet).Bytes()),
		make([]byte, 32),
	))

	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, appDomainSep, tsHash)
	innerSig, err := crypto.Sign(digest, pk)
	if err != nil {
		return "", fmt.Errorf("sign order: sign poly1271 digest: %w", err)
	}
	if innerSig[64] < 27 {
		innerSig[64] += 27
	}

	typeBytes := []byte(consts.EIP712OrderType)
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(len(typeBytes)))

	wrapped := ethutil.Concat(innerSig, appDomainSep, contentsHash, typeBytes, lenBytes)
	return "0x" + common.Bytes2Hex(wrapped), nil
}

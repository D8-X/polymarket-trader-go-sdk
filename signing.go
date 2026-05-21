package polytrade

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

func (ob *OrderBuilder) signOrder(order OrderFields) (string, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(ob.privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("sign order: invalid private key: %w", err)
	}

	domainTypeHash := ethutil.Keccak256([]byte(consts.EIP712DomainType))
	nameHash := ethutil.Keccak256([]byte(consts.EIP712OrderDomainName))
	versionHash := ethutil.Keccak256([]byte(consts.EIP712OrderVersion))
	chainID := ethutil.PadTo32(big.NewInt(PolygonChainID).Bytes())

	ctfAddr := new(big.Int)
	if len(ob.ctfExchangeAddress) > 2 {
		ctfAddr.SetString(ob.ctfExchangeAddress[2:], 16)
	}

	domainSep := ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(domainTypeHash),
		ethutil.PadTo32(nameHash),
		ethutil.PadTo32(versionHash),
		chainID,
		ethutil.PadTo32(ctfAddr.Bytes()),
	))

	orderTypeHash := ethutil.Keccak256([]byte(consts.EIP712OrderType))

	structHash := ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(orderTypeHash),
		ethutil.PadTo32(big.NewInt(order.Salt).Bytes()),
		ethutil.PadTo32(ethutil.ParseAddress(order.Maker)),
		ethutil.PadTo32(ethutil.ParseAddress(order.Signer)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.TokenID)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.MakerAmount)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.TakerAmount)),
		ethutil.PadTo32(big.NewInt(int64(order.sideNumeric)).Bytes()),
		ethutil.PadTo32(big.NewInt(int64(order.SignatureType)).Bytes()),
		ethutil.PadTo32(ethutil.ParseBigInt(order.Timestamp)),
		ethutil.ParseBytes32(order.Metadata),
		ethutil.ParseBytes32(order.Builder),
	))

	if order.SignatureType == SignatureTypePoly1271 {
		return wrapPoly1271Signature(pk, order.Signer, domainSep, structHash)
	}

	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)

	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		return "", fmt.Errorf("sign order: sign EIP-712 digest: %w", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}

	return "0x" + common.Bytes2Hex(sig), nil
}

func wrapPoly1271Signature(pk *ecdsa.PrivateKey, depositWallet string, appDomainSep, contentsHash []byte) (string, error) {
	soladyTypeHash := ethutil.Keccak256([]byte(consts.EIP712SoladyTypedDataSignType))
	dwNameHash := ethutil.Keccak256([]byte(consts.EIP712DepositWalletName))
	dwVersionHash := ethutil.Keccak256([]byte(consts.EIP712DepositWalletVersion))

	tsHash := ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(soladyTypeHash),
		contentsHash,
		ethutil.PadTo32(dwNameHash),
		ethutil.PadTo32(dwVersionHash),
		ethutil.PadTo32(big.NewInt(PolygonChainID).Bytes()),
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

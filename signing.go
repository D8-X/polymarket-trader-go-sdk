package polytrade

import (
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (ob *OrderBuilder) signOrder(order OrderFields) (string, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(ob.privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("sign order: invalid private key: %w", err)
	}

	domainTypeHash := ethutil.Keccak256([]byte(eip712DomainType))
	nameHash := ethutil.Keccak256([]byte(eip712OrderDomainName))
	versionHash := ethutil.Keccak256([]byte(eip712Version))
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

	orderTypeHash := ethutil.Keccak256([]byte(eip712OrderType))

	structHash := ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(orderTypeHash),
		ethutil.PadTo32(big.NewInt(order.Salt).Bytes()),
		ethutil.PadTo32(ethutil.ParseAddress(order.Maker)),
		ethutil.PadTo32(ethutil.ParseAddress(order.Signer)),
		ethutil.PadTo32(ethutil.ParseAddress(order.Taker)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.TokenID)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.MakerAmount)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.TakerAmount)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.Expiration)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.Nonce)),
		ethutil.PadTo32(ethutil.ParseBigInt(order.FeeRateBps)),
		ethutil.PadTo32(big.NewInt(int64(order.sideNumeric)).Bytes()),
		ethutil.PadTo32(big.NewInt(int64(order.SignatureType)).Bytes()),
	))

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

func hashSafeFactoryDomain() []byte {
	typeHash := ethutil.Keccak256([]byte(eip712SafeFactoryDomainType))
	nameHash := ethutil.Keccak256([]byte(SafeFactoryName))
	chainIDBytes := ethutil.PadTo32(big.NewInt(PolygonChainID).Bytes())

	factoryAddr := new(big.Int)
	factoryAddr.SetString(SafeFactory[2:], 16)

	return ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(typeHash),
		ethutil.PadTo32(nameHash),
		chainIDBytes,
		ethutil.PadTo32(factoryAddr.Bytes()),
	))
}

func hashCreateProxyStruct(paymentToken, payment, paymentReceiver string) []byte {
	typeHash := ethutil.Keccak256([]byte(eip712CreateProxyType))

	tokenAddr := new(big.Int)
	if len(paymentToken) > 2 {
		tokenAddr.SetString(paymentToken[2:], 16)
	}

	paymentBig := new(big.Int)
	paymentBig.SetString(payment, 10)

	receiverAddr := new(big.Int)
	if len(paymentReceiver) > 2 {
		receiverAddr.SetString(paymentReceiver[2:], 16)
	}

	return ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(typeHash),
		ethutil.PadTo32(tokenAddr.Bytes()),
		ethutil.PadTo32(paymentBig.Bytes()),
		ethutil.PadTo32(receiverAddr.Bytes()),
	))
}

func signSafeCreate(privateKeyHex string) (string, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("sign safe create: invalid private key: %w", err)
	}

	domainSep := hashSafeFactoryDomain()
	structHash := hashCreateProxyStruct(ZeroAddress, "0", ZeroAddress)
	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)

	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		return "", fmt.Errorf("sign safe create: sign EIP-712 digest: %w", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}

	return "0x" + common.Bytes2Hex(sig), nil
}

func hashSafeTxDomain(safeAddress string) []byte {
	typeHash := ethutil.Keccak256([]byte(eip712SafeTxDomainType))
	chainIDBytes := ethutil.PadTo32(big.NewInt(PolygonChainID).Bytes())

	safeAddr := new(big.Int)
	if len(safeAddress) > 2 {
		safeAddr.SetString(safeAddress[2:], 16)
	}

	return ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(typeHash),
		chainIDBytes,
		ethutil.PadTo32(safeAddr.Bytes()),
	))
}

func hashSafeTxStruct(to, value, data string, operation int, nonce string) []byte {
	typeHash := ethutil.Keccak256([]byte(eip712SafeTxType))

	toAddr := new(big.Int)
	if len(to) > 2 {
		toAddr.SetString(to[2:], 16)
	}

	valueBig := new(big.Int)
	valueBig.SetString(value, 10)

	dataBytes := common.FromHex(data)
	dataHash := ethutil.Keccak256(dataBytes)

	nonceBig := new(big.Int)
	nonceBig.SetString(nonce, 10)

	zeroAddr := new(big.Int) // gasToken and refundReceiver = zero address

	return ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(typeHash),
		ethutil.PadTo32(toAddr.Bytes()),
		ethutil.PadTo32(valueBig.Bytes()),
		ethutil.PadTo32(dataHash),
		ethutil.PadTo32(big.NewInt(int64(operation)).Bytes()),
		ethutil.PadTo32(nil),              // safeTxGas = 0
		ethutil.PadTo32(nil),              // baseGas = 0
		ethutil.PadTo32(nil),              // gasPrice = 0
		ethutil.PadTo32(zeroAddr.Bytes()), // gasToken = zero address
		ethutil.PadTo32(zeroAddr.Bytes()), // refundReceiver = zero address
		ethutil.PadTo32(nonceBig.Bytes()),
	))
}

func signSafeTx(privateKeyHex, safeAddress string, tx SafeTransaction, nonce string) (string, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("sign safe tx: invalid private key: %w", err)
	}

	domainSep := hashSafeTxDomain(safeAddress)
	structHash := hashSafeTxStruct(tx.To, tx.Value, tx.Data, tx.Operation, nonce)
	eip712Digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)

	// Gnosis Safe uses eth_sign. wrap the EIP-712 digest with the personal message prefix
	ethSignPrefix := []byte("\x19Ethereum Signed Message:\n32")
	signDigest := ethutil.Keccak256(append(ethSignPrefix, eip712Digest...))

	sig, err := crypto.Sign(signDigest, pk)
	if err != nil {
		return "", fmt.Errorf("sign safe tx: sign EIP-712 digest: %w", err)
	}

	// Gnosis Safe v-value adjustment: go-ethereum returns v=0/1 (recovery ID)
	// Safe expects v=31/32 (not the standard 27/28)
	switch sig[64] {
	case 0, 1:
		sig[64] += 31
	case 27, 28:
		sig[64] += 4
	}

	return "0x" + common.Bytes2Hex(sig), nil
}

func hashClobAuthDomain(chainID int) []byte {
	typeHash := ethutil.Keccak256([]byte(eip712AuthDomainType))
	nameHash := ethutil.Keccak256([]byte(eip712AuthDomainName))
	versionHash := ethutil.Keccak256([]byte(eip712Version))
	chainIDBytes := ethutil.PadTo32(new(big.Int).SetInt64(int64(chainID)).Bytes())

	return ethutil.Keccak256(append(append(append(ethutil.PadTo32(typeHash), ethutil.PadTo32(nameHash)...), ethutil.PadTo32(versionHash)...), chainIDBytes...))
}

func hashClobAuthStruct(address, timestamp string, nonce int64) []byte {
	typeHash := ethutil.Keccak256([]byte(eip712ClobAuthType))
	addrBig := new(big.Int)
	if len(address) > 2 {
		addrBig.SetString(address[2:], 16)
	}
	tsHash := ethutil.Keccak256([]byte(timestamp))
	nonceBig := new(big.Int).SetInt64(nonce)
	msgHash := ethutil.Keccak256([]byte(eip712AuthMessage))

	encoded := make([]byte, 0, 160)
	encoded = append(encoded, ethutil.PadTo32(typeHash)...)
	encoded = append(encoded, ethutil.PadTo32(addrBig.Bytes())...)
	encoded = append(encoded, ethutil.PadTo32(tsHash)...)
	encoded = append(encoded, ethutil.PadTo32(nonceBig.Bytes())...)
	encoded = append(encoded, ethutil.PadTo32(msgHash)...)

	return ethutil.Keccak256(encoded)
}

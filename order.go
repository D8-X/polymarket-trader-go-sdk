package polytrade

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/internal/ethutil"
)

type OrderBuilder struct {
	makerAddress       string
	signerAddress      string
	ctfExchangeAddress string
	privateKeyHex      string
	sigType            int
}

func NewOrderBuilder(funderAddress, eoaAddress, ctfExchangeAddress, privateKeyHex string, sigType int) *OrderBuilder {
	return &OrderBuilder{
		makerAddress:       funderAddress,
		signerAddress:      eoaAddress,
		ctfExchangeAddress: ctfExchangeAddress,
		privateKeyHex:      privateKeyHex,
		sigType:            sigType,
	}
}

func (ob *OrderBuilder) PrepareAndSign(tokenID, side, orderType string, price, size float64, apiKey string) (*SignedOrder, error) {
	sideNumeric := SideBuy
	sideStr := "BUY"
	if side == "SELL" {
		sideNumeric = SideSell
		sideStr = "SELL"
	}

	saltBig, err := rand.Int(rand.Reader, big.NewInt(SaltUpperBound))
	if err != nil {
		return nil, fmt.Errorf("prepare order: generate salt: %w", err)
	}
	salt := saltBig.Int64()

	nonce := "0"

	sizeWei := ethutil.RoundTo10k(int64(size * AmountScale))

	var makerAmount, takerAmount int64
	if sideNumeric == SideBuy {
		makerAmount = ethutil.RoundTo10k(int64(float64(sizeWei) * price))
		takerAmount = sizeWei
	} else {
		makerAmount = sizeWei
		takerAmount = ethutil.RoundTo10k(int64(float64(sizeWei) * price))
	}

	// Non-GTD orders must have expiration=0; only GTD uses a real timestamp
	expiration := int64(0)
	if orderType == "GTD" {
		expiration = time.Now().Add(GTDExpiration).Unix()
	}

	order := OrderFields{
		Salt:          salt,
		Maker:         ob.makerAddress,
		Signer:        ob.signerAddress,
		Taker:         ZeroAddress,
		TokenID:       tokenID,
		MakerAmount:   strconv.FormatInt(makerAmount, 10),
		TakerAmount:   strconv.FormatInt(takerAmount, 10),
		Expiration:    strconv.FormatInt(expiration, 10),
		Nonce:         nonce,
		FeeRateBps:    "0",
		Side:          sideStr,
		SignatureType: ob.sigType,
		sideNumeric:   sideNumeric,
	}

	sig, err := ob.signOrder(order)
	if err != nil {
		return nil, fmt.Errorf("prepare order: %w", err)
	}
	order.Signature = sig

	return &SignedOrder{
		Order:     order,
		Owner:     apiKey,
		OrderType: orderType,
	}, nil
}

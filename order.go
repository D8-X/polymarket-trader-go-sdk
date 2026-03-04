package polytrade

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"
)

type OrderBuilder struct {
	makerAddress       string
	signerAddress      string
	ctfExchangeAddress string
	privateKeyHex      string
	sigType            int
}

type OrderOpts struct {
	PostOnly   bool
	DeferExec  bool
	FeeRateBps string
	Expiration time.Duration
	TickSize   string
}

type roundConfig struct {
	price  int
	size   int
	amount int
}

func getRoundConfig(tickSize string) roundConfig {
	switch tickSize {
	case "0.1":
		return roundConfig{price: 1, size: 2, amount: 3}
	case "0.001":
		return roundConfig{price: 3, size: 2, amount: 5}
	case "0.0001":
		return roundConfig{price: 4, size: 2, amount: 6}
	default:
		return roundConfig{price: 2, size: 2, amount: 4}
	}
}

func checkPrecision(v float64, decimals int, label string) error {
	factor := math.Pow(10, float64(decimals))
	rounded := math.Floor(v*factor) / factor
	if math.Abs(v-rounded) > 1e-12 {
		return fmt.Errorf("%s %v exceeds tick-size precision (%d decimals)", label, v, decimals)
	}
	return nil
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

func (ob *OrderBuilder) MakerAddress() string {
	return ob.makerAddress
}

// PrepareAndSign builds and EIP-712-signs an order for the Polymarket CLOB.
//
// Supported order types:
//   - GTC (Good-Till-Cancelled): rests on the book until filled or cancelled.
//   - GTD (Good-Till-Date): rests on the book until filled, cancelled, or expired.
//     Expiration includes a 60-second security threshold required by the API.
//   - FOK (Fill-or-Kill): must fill entirely in one match or be rejected.
//   - FAK (Fill-and-Kill): fills as much as possible immediately, remainder is cancelled.
//
// PostOnly (GTC/GTD only) rejects the order if it would cross the spread.
//
// Sports markets: FOK and FAK orders have a ~3 second placement delay and are
// automatically cancelled at game start.
// See https://docs.polymarket.com/sports/overview#order-types
func (ob *OrderBuilder) PrepareAndSign(tokenID, side, orderType string, price, size float64, apiKey string, opts ...OrderOpts) (*SignedOrder, error) {
	var opt OrderOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

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

	rc := getRoundConfig(opt.TickSize)
	if err := checkPrecision(price, rc.price, "price"); err != nil {
		return nil, fmt.Errorf("prepare order: %w", err)
	}
	if err := checkPrecision(size, rc.size, "size"); err != nil {
		return nil, fmt.Errorf("prepare order: %w", err)
	}
	sizeWei := int64(size * AmountScale)
	amountFactor := math.Pow(10, float64(rc.amount))
	amountWei := int64(math.Floor(size*price*amountFactor) / amountFactor * AmountScale)

	var makerAmount, takerAmount int64
	if sideNumeric == SideBuy {
		makerAmount = amountWei
		takerAmount = sizeWei
	} else {
		makerAmount = sizeWei
		takerAmount = amountWei
	}

	expiration := int64(0)
	if orderType == "GTD" {
		dur := GTDExpiration
		if opt.Expiration > 0 {
			dur = opt.Expiration
		}
		expiration = time.Now().Add(GTDSecurityThreshold + dur).Unix()
	}

	feeRate := "0"
	if opt.FeeRateBps != "" {
		feeRate = opt.FeeRateBps
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
		FeeRateBps:    feeRate,
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
		PostOnly:  opt.PostOnly,
		DeferExec: opt.DeferExec,
	}, nil
}

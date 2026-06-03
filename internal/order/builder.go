package order

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/sign"
)

type Builder struct {
	makerAddress       string
	signerAddress      string
	ctfExchangeAddress string
	privateKeyHex      string

	mu          sync.RWMutex
	builderCode string
}

type Opts struct {
	PostOnly    bool
	DeferExec   bool
	Expiration  time.Duration
	TickSize    string
	BuilderCode string
	Metadata    string
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

func decimalPlaces(s string) int {
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return len(s) - i - 1
	}
	return 0
}

func parseDecimal(s, label string, maxDecimals int) (*big.Rat, int, error) {
	body := strings.TrimPrefix(s, "-")
	dot := -1
	for i := 0; i < len(body); i++ {
		switch c := body[i]; {
		case c >= '0' && c <= '9':
		case c == '.' && dot < 0:
			dot = i
		default:
			return nil, 0, fmt.Errorf("%s %q must be a plain decimal number", label, s)
		}
	}
	if body == "" || dot == 0 || dot == len(body)-1 {
		return nil, 0, fmt.Errorf("%s %q is not a valid decimal", label, s)
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, 0, fmt.Errorf("%s %q is not a valid decimal", label, s)
	}
	if r.Sign() <= 0 {
		return nil, 0, fmt.Errorf("%s must be positive, got %q", label, s)
	}
	dp := decimalPlaces(s)
	if dp > maxDecimals {
		return nil, 0, fmt.Errorf("%s %q exceeds %d decimal places", label, s, maxDecimals)
	}
	return r, dp, nil
}

func ratToMicroFloor(v *big.Rat, decimals int) int64 {
	pow := func(n int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) }
	scaled := new(big.Rat).Mul(v, new(big.Rat).SetInt(pow(decimals)))
	floored := new(big.Int).Quo(scaled.Num(), scaled.Denom())
	return new(big.Int).Mul(floored, pow(6-decimals)).Int64()
}

func ratToMicroCeil(v *big.Rat, decimals int) int64 {
	pow := func(n int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) }
	scaled := new(big.Rat).Mul(v, new(big.Rat).SetInt(pow(decimals)))
	q := new(big.Int).Quo(scaled.Num(), scaled.Denom())
	if new(big.Int).Mul(q, scaled.Denom()).Cmp(scaled.Num()) != 0 {
		q.Add(q, big.NewInt(1))
	}
	return new(big.Int).Mul(q, pow(6-decimals)).Int64()
}

func usdcAmountDecimals(sideNumeric int, orderType string, rc roundConfig) int {
	if sideNumeric == consts.SideBuy && (orderType == consts.OrderTypeFAK || orderType == consts.OrderTypeFOK) {
		return 2
	}
	return rc.amount
}

func NewBuilder(funderAddress, ctfExchangeAddress, privateKeyHex string) *Builder {
	return &Builder{
		makerAddress:       funderAddress,
		signerAddress:      funderAddress,
		ctfExchangeAddress: ctfExchangeAddress,
		privateKeyHex:      privateKeyHex,
	}
}

func (ob *Builder) MakerAddress() string  { return ob.makerAddress }
func (ob *Builder) SignerAddress() string { return ob.signerAddress }

func (ob *Builder) SetBuilderCode(code string) {
	ob.mu.Lock()
	ob.builderCode = code
	ob.mu.Unlock()
}

func (ob *Builder) getBuilderCode() string {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.builderCode
}

func (ob *Builder) PrepareAndSign(tokenID, side, orderType, price, size, apiKey string, opts ...Opts) (*models.SignedOrder, error) {
	var opt Opts
	if len(opts) > 0 {
		opt = opts[0]
	}

	sideNumeric := consts.SideBuy
	sideStr := consts.BUY
	if side == consts.SELL {
		sideNumeric = consts.SideSell
		sideStr = consts.SELL
	}

	saltBig, err := rand.Int(rand.Reader, big.NewInt(consts.SaltUpperBound))
	if err != nil {
		return nil, fmt.Errorf("prepare order: generate salt: %w", err)
	}
	salt := saltBig.Int64()

	rc := getRoundConfig(opt.TickSize)
	priceRat, _, err := parseDecimal(price, "price", rc.price)
	if err != nil {
		return nil, fmt.Errorf("prepare order: %w", err)
	}
	if priceRat.Cmp(new(big.Rat).SetInt64(1)) >= 0 {
		return nil, fmt.Errorf("prepare order: price %s must be below 1", price)
	}
	sizeRat, _, err := parseDecimal(size, "size", rc.size)
	if err != nil {
		return nil, fmt.Errorf("prepare order: %w", err)
	}

	sizeWei := ratToMicroFloor(sizeRat, rc.size)
	usdc := new(big.Rat).Mul(sizeRat, priceRat)
	amountDecimals := usdcAmountDecimals(sideNumeric, orderType, rc)

	var makerAmount, takerAmount int64
	if sideNumeric == consts.SideBuy {
		makerAmount = ratToMicroCeil(usdc, amountDecimals)
		takerAmount = sizeWei
	} else {
		makerAmount = sizeWei
		takerAmount = ratToMicroFloor(usdc, amountDecimals)
	}

	expiration := int64(0)
	if orderType == consts.OrderTypeGTD {
		dur := consts.GTDExpiration
		if opt.Expiration > 0 {
			dur = opt.Expiration
		}
		expiration = time.Now().Add(consts.GTDSecurityThreshold + dur).Unix()
	}

	builder := ob.getBuilderCode()
	if opt.BuilderCode != "" {
		builder = opt.BuilderCode
	}
	if builder == "" {
		builder = consts.ZeroBytes32
	}

	metadata := opt.Metadata
	if metadata == "" {
		metadata = consts.ZeroBytes32
	}

	o := models.OrderFields{
		Salt:          salt,
		Maker:         ob.makerAddress,
		Signer:        ob.signerAddress,
		TokenID:       tokenID,
		MakerAmount:   strconv.FormatInt(makerAmount, 10),
		TakerAmount:   strconv.FormatInt(takerAmount, 10),
		Expiration:    strconv.FormatInt(expiration, 10),
		Timestamp:     strconv.FormatInt(time.Now().UnixMilli(), 10),
		Metadata:      metadata,
		Builder:       builder,
		Side:          sideStr,
		SignatureType: consts.SignatureTypePoly1271,
		SideNumeric:   sideNumeric,
	}

	sig, err := sign.Order(ob.privateKeyHex, ob.ctfExchangeAddress, o)
	if err != nil {
		return nil, fmt.Errorf("prepare order: %w", err)
	}
	o.Signature = sig

	return &models.SignedOrder{
		Order:     o,
		Owner:     apiKey,
		OrderType: orderType,
		PostOnly:  opt.PostOnly,
		DeferExec: opt.DeferExec,
	}, nil
}

package order

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/sign"
	"github.com/ethereum/go-ethereum/crypto"
)

type Builder struct {
	makerAddress       string
	signerAddress      string
	ctfExchangeAddress string
	privateKeyHex      string
	sigType            int

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

func checkPrecision(v float64, decimals int, label string) error {
	factor := math.Pow(10, float64(decimals))
	rounded := math.Floor(v*factor) / factor
	if math.Abs(v-rounded) > 1e-12 {
		return fmt.Errorf("%s %v exceeds tick-size precision (%d decimals)", label, v, decimals)
	}
	return nil
}

func NewBuilder(funderAddress, ctfExchangeAddress, privateKeyHex string, sigType int) *Builder {
	var signer string
	if sigType == consts.SignatureTypePoly1271 {
		signer = funderAddress
	} else if pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex)); err == nil {
		signer = crypto.PubkeyToAddress(pk.PublicKey).Hex()
	}
	return &Builder{
		makerAddress:       funderAddress,
		signerAddress:      signer,
		ctfExchangeAddress: ctfExchangeAddress,
		privateKeyHex:      privateKeyHex,
		sigType:            sigType,
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

func (ob *Builder) PrepareAndSign(tokenID, side, orderType string, price, size float64, apiKey string, opts ...Opts) (*models.SignedOrder, error) {
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
	if opt.TickSize != "" {
		if err := checkPrecision(price, rc.price, "price"); err != nil {
			return nil, fmt.Errorf("prepare order: %w", err)
		}
		if err := checkPrecision(size, rc.size, "size"); err != nil {
			return nil, fmt.Errorf("prepare order: %w", err)
		}
	}
	sizeWei := int64(size * consts.AmountScale)
	amountFactor := math.Pow(10, float64(rc.amount))
	amountWei := int64(math.Floor(size*price*amountFactor) / amountFactor * consts.AmountScale)

	var makerAmount, takerAmount int64
	if sideNumeric == consts.SideBuy {
		makerAmount = amountWei
		takerAmount = sizeWei
	} else {
		makerAmount = sizeWei
		takerAmount = amountWei
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
		SignatureType: ob.sigType,
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

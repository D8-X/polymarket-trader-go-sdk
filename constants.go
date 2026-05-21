package polytrade

import "github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"

const (
	ClobBaseURL    = consts.ClobBaseURL
	DataAPIBaseURL = consts.DataAPIBaseURL
)

const (
	PolygonChainID = consts.PolygonChainID
)

const (
	BUY  = consts.BUY
	SELL = consts.SELL
)

const (
	OrderTypeGTC = consts.OrderTypeGTC
	OrderTypeGTD = consts.OrderTypeGTD
	OrderTypeFOK = consts.OrderTypeFOK
	OrderTypeFAK = consts.OrderTypeFAK
)

const (
	SignatureTypePoly1271 = consts.SignatureTypePoly1271
)

const (
	USDCAddress        = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
	PUSDAddress        = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	CTFExchange        = "0xE111180000d2663C0091e4f400237545B87B996B"
	NegRiskCTFExchange = "0xe2222d279d744050d28e00520010520000310F59"
	NegRiskAdapter     = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"
)

const (
	GTDSecurityThreshold = consts.GTDSecurityThreshold
	SaltUpperBound       = consts.SaltUpperBound
)

const (
	OrderStatusMatched   = consts.OrderStatusMatched
	OrderStatusLive      = consts.OrderStatusLive
	OrderStatusDelayed   = consts.OrderStatusDelayed
	OrderStatusCanceled  = consts.OrderStatusCanceled
	OrderStatusUnmatched = consts.OrderStatusUnmatched
)

const (
	DefaultPollInterval       = consts.DefaultPollInterval
	DefaultDelayedPollTimeout = consts.DefaultDelayedPollTimeout
	DefaultLivePollTimeout    = consts.DefaultLivePollTimeout
)

package polytrade

import "time"

const (
	ClobBaseURL    = "https://clob.polymarket.com"
	DataAPIBaseURL = "https://data-api.polymarket.com"
	SafeAPIBaseURL = "https://safe-transaction-polygon.safe.global/api/v1"
)

const (
	PolygonChainID = 137
)

const (
	BUY  = "BUY"
	SELL = "SELL"
)

const (
	OrderTypeGTC = "GTC"
	OrderTypeGTD = "GTD"
	OrderTypeFOK = "FOK"
	OrderTypeFAK = "FAK"
)

const (
	SignatureTypeEOA        = 0
	SignatureTypePolyProxy  = 1
	SignatureTypeGnosisSafe = 2
	SignatureTypePoly1271   = 3
)

const (
	ZeroAddress = "0x0000000000000000000000000000000000000000"
)

const (
	USDCAddress        = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
	PUSDAddress        = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	CTFExchange        = "0xE111180000d2663C0091e4f400237545B87B996B"
	NegRiskCTFExchange = "0xe2222d279d744050d28e00520010520000310F59"
	NegRiskAdapter     = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"
)

// configuredPUSDAddress is set by SetPUSDAddress; CollateralAddress falls back
// to USDCAddress until callers configure pUSD.
var configuredPUSDAddress = PUSDAddress

// SetPUSDAddress overrides the default Polymarket USD (pUSD) Polygon address.
// CollateralAddress returns this value when non-empty; otherwise it returns
// USDCAddress (USDC.e).
func SetPUSDAddress(addr string) {
	configuredPUSDAddress = addr
}

// CollateralAddress returns the active collateral ERC-20 — pUSD by default,
// falling back to USDC.e if pUSD has been explicitly unset via SetPUSDAddress("").
func CollateralAddress() string {
	if configuredPUSDAddress != "" {
		return configuredPUSDAddress
	}
	return USDCAddress
}

const (
	GTDSecurityThreshold = 60 * time.Second
	SaltUpperBound       = 1 << 62
)

const (
	OrderStatusMatched   = "matched"
	OrderStatusLive      = "live"
	OrderStatusDelayed   = "delayed"
	OrderStatusCanceled  = "canceled"
	OrderStatusUnmatched = "unmatched"
)

const (
	DefaultPollInterval       = 200 * time.Millisecond
	DefaultDelayedPollTimeout = 5 * time.Second
	DefaultLivePollTimeout    = 60 * time.Second
)

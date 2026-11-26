package consts

import "time"

const (
	PolygonChainID = 137
)

const (
	ZeroAddress        = "0x0000000000000000000000000000000000000000"
	USDCAddress        = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
	PUSDAddress        = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
	CTFExchange        = "0xE111180000d2663C0091e4f400237545B87B996B"
	NegRiskCTFExchange = "0xe2222d279d744050d28e00520010520000310F59"
	NegRiskAdapter     = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"

	// Targets for Split / Merge / Redeem. Polymarket's positions live in these
	// adapters, not directly in the ConditionalTokens vault, so calling the
	// vault would happily execute and pay out zero. Pick CollateralAdapter for
	// standard binary markets and NegRiskCollateralAdapter for neg-risk ones.
	CollateralAdapter        = "0xAdA100Db00Ca00073811820692005400218FcE1f"
	NegRiskCollateralAdapter = "0xadA2005600Dec949baf300f4C6120000bDB6eAab"

	CollateralOnramp  = "0x93070a847efEf7F70739046A929D47a521F5B8ee"
	CollateralOfframp = "0x2957922Eb93258b93368531d39fAcCA3B4dC5854"

	DepositWalletFactory = "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07"
	// The factory deploys beacon proxies now. Wallets it deployed before the
	// switch are UUPS proxies whose address is derived from this implementation.
	DepositWalletBeacon         = "0x7A18EDfe055488A3128f01F563e5B479D92ffc3a"
	DepositWalletImplementation = "0x58ca52ebe0dadfdf531cde7062e76746de4db1eb"

	ConditionalTokens = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"

	ExchangeV3       = "0xe3333700cA9d93003F00f0F71f8515005F6c00Aa"
	PositionManager  = "0x006F54F7f9A22e0000CC2AB60031000000ae9fEF"
	ProtocolV2Router = "0x12121212006e4CD160D18e3f00711DA5c3372600"
)

const (
	RelayerBaseURL = "https://relayer-v2.polymarket.com"
	ClobBaseURL    = "https://clob.polymarket.com"
	DataAPIBaseURL = "https://data-api.polymarket.com"
)

const (
	DefaultTimeout       = 10 * time.Second
	CLOBTimeout          = 15 * time.Second
	GTDExpiration        = 5 * time.Minute
	GTDSecurityThreshold = 60 * time.Second
	GTDMinExpiration     = 3 * time.Minute

	DefaultPollInterval       = 200 * time.Millisecond
	DefaultDelayedPollTimeout = 5 * time.Second
	DefaultLivePollTimeout    = 60 * time.Second

	DefaultTradePollInterval   = 250 * time.Millisecond
	DefaultTradeResolveTimeout = 30 * time.Second

	DataAPIRetryAttempts = 4
	DataAPIRetryBackoff  = 500 * time.Millisecond
	DataAPIRetryMaxWait  = 10 * time.Second
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
	AssetTypeCollateral    = "COLLATERAL"
	AssetTypeConditional   = "CONDITIONAL"
	AssetTypeConditionalV2 = "CONDITIONAL-V2"
)

const MarketVersionV2 = "v2"

const MaxCancelOrderIDs = 1000

const SignatureTypePoly1271 = 3

const SaltUpperBound = 1 << 62

const (
	OrderStatusMatched   = "matched"
	OrderStatusLive      = "live"
	OrderStatusDelayed   = "delayed"
	OrderStatusCanceled  = "canceled"
	OrderStatusUnmatched = "unmatched"
	OrderStatusInvalid   = "invalid"
)

const (
	TradeStatusConfirmed = "CONFIRMED"
	TradeStatusFailed    = "FAILED"
)

const (
	AmountScale = 1e6
	SideBuy     = 0
	SideSell    = 1
	ZeroBytes32 = "0x0000000000000000000000000000000000000000000000000000000000000000"
)

const (
	EIP712DepositWalletDomainType = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	EIP712DepositWalletName       = "DepositWallet"
	EIP712DepositWalletVersion    = "1"

	EIP712BatchType = "Batch(address wallet,uint256 nonce,uint256 deadline,Call[] calls)Call(address target,uint256 value,bytes data)"
	EIP712CallType  = "Call(address target,uint256 value,bytes data)"

	EIP712SoladyTypedDataSignType = "TypedDataSign(Order contents,string name,string version,uint256 chainId,address verifyingContract,bytes32 salt)Order(uint256 salt,address maker,address signer,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,uint256 timestamp,bytes32 metadata,bytes32 builder)"

	EIP712DomainType     = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	EIP712AuthDomainType = "EIP712Domain(string name,string version,uint256 chainId)"
	EIP712ClobAuthType   = "ClobAuth(address address,string timestamp,uint256 nonce,string message)"
	EIP712OrderType      = "Order(uint256 salt,address maker,address signer,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint8 side,uint8 signatureType,uint256 timestamp,bytes32 metadata,bytes32 builder)"

	EIP712AuthDomainName  = "ClobAuthDomain"
	EIP712OrderDomainName = "Polymarket CTF Exchange"
	EIP712OrderVersion    = "2"
	EIP712OrderVersionV3  = "3"
	EIP712AuthVersion     = "1"
	EIP712AuthMessage     = "This message attests that I control the given wallet"
)

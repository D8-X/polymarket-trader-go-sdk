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
	SideBuy  = 0
	SideSell = 1
)

const (
	SignatureTypeEOA        = 0
	SignatureTypePolyProxy  = 1
	SignatureTypeGnosisSafe = 2
)

const (
	ProxyFactory      = "0xaB45c5A4B0c941a2F231C04C3f49182e1A254052"
	ProxyInitCodeHash = "0xd21df8dc65880a8606f09fe0ce3df9b8869287ab0b058be05aa9e8af6330a00b"
	ZeroAddress       = "0x0000000000000000000000000000000000000000"
)

const (
	DefaultTimeout = 10 * time.Second
	CLOBTimeout    = 15 * time.Second
)

const (
	AmountScale          = 1e6
	GTDExpiration         = 5 * time.Minute
	GTDSecurityThreshold  = 60 * time.Second
	SaltUpperBound        = 1 << 62
)

const (
	eip712DomainType      = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	eip712AuthDomainType  = "EIP712Domain(string name,string version,uint256 chainId)"
	eip712ClobAuthType    = "ClobAuth(address address,string timestamp,uint256 nonce,string message)"
	eip712OrderType       = "Order(uint256 salt,address maker,address signer,address taker,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint256 expiration,uint256 nonce,uint256 feeRateBps,uint8 side,uint8 signatureType)"
	eip712AuthDomainName  = "ClobAuthDomain"
	eip712OrderDomainName = "Polymarket CTF Exchange"
	eip712Version         = "1"
	eip712AuthMessage     = "This message attests that I control the given wallet"
)

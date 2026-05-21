package consts

import "time"

const (
	PolygonChainID = 137
)

const (
	CollateralOnramp  = "0x93070a847efEf7F70739046A929D47a521F5B8ee"
	CollateralOfframp = "0x2957922Eb93258b93368531d39fAcCA3B4dC5854"

	DepositWalletFactory        = "0x00000000000Fb5C9ADea0298D729A0CB3823Cc07"
	DepositWalletImplementation = "0x58ca52ebe0dadfdf531cde7062e76746de4db1eb"

	ConditionalTokens = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"
)

const (
	RelayerBaseURL = "https://relayer-v2.polymarket.com"
	ClobBaseURL    = "https://clob.polymarket.com"
	DataAPIBaseURL = "https://data-api.polymarket.com"
)

const (
	DefaultTimeout = 10 * time.Second
	CLOBTimeout    = 15 * time.Second
	GTDExpiration  = 5 * time.Minute
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
	EIP712AuthVersion     = "1"
	EIP712AuthMessage     = "This message attests that I control the given wallet"
)

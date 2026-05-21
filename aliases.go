package polytrade

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/auth"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/clob"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/order"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/relayer"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/sweep"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/types"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/wallet"
)

type CLOBClient = clob.Client

func NewCLOBClient() *CLOBClient { return clob.NewClient() }

type OrderBuilder = order.Builder
type OrderOpts = order.Opts

func NewOrderBuilder(funderAddress, ctfExchangeAddress, privateKeyHex string, sigType int) *OrderBuilder {
	return order.NewBuilder(funderAddress, ctfExchangeAddress, privateKeyHex, sigType)
}

type APIError = types.APIError
type L2Credentials = types.L2Credentials
type L2Headers = types.L2Headers
type RelayerCredentials = types.RelayerCredentials
type RelayerResponse = types.RelayerResponse
type RelayerTransaction = types.RelayerTransaction
type WalletCall = types.WalletCall

type OrderFields = models.OrderFields
type SignedOrder = models.SignedOrder
type PlaceOrderResponse = models.PlaceOrderResponse
type OrderStatus = models.OrderStatus
type PollOpts = models.PollOpts
type PollResult = models.PollResult
type BalanceEntry = models.BalanceEntry
type PositionEntry = models.PositionEntry
type CancelResponse = models.CancelResponse
type Trade = models.Trade
type MakerOrder = models.MakerOrder
type OrderBook = models.OrderBook
type OrderBookLevel = models.OrderBookLevel
type BalanceAllowanceResponse = models.BalanceAllowanceResponse
type ClobMarketInfo = models.ClobMarketInfo
type ClobMarketFeeDetails = models.ClobMarketFeeDetails
type ClobMarketInfoToken = models.ClobMarketInfoToken

type MarketRewardsRate = models.MarketRewardsRate
type MarketRewards = models.MarketRewards
type MarketToken = models.MarketToken
type MarketInfo = models.MarketInfo
type SimplifiedMarketInfo = models.SimplifiedMarketInfo
type MarketByTokenInfo = models.MarketByTokenInfo
type MarketLiveActivity = models.MarketLiveActivity

type PriceRequest = models.PriceRequest
type SpreadRequest = models.SpreadRequest
type LastTradePrice = models.LastTradePrice
type PriceHistoryEntry = models.PriceHistoryEntry
type PricesHistoryParams = models.PricesHistoryParams

type MarketRewardConfig = models.MarketRewardConfig
type CurrentRewardMarket = models.CurrentRewardMarket
type OrderScoringResult = models.OrderScoringResult

type PaginatedResponse[T any] = models.PaginatedResponse[T]

type PriceLevel = models.PriceLevel
type SweepLevel = models.SweepLevel
type SweepEstimate = models.SweepEstimate

func DeriveL2Credentials(privateKeyHex string, chainID int) (*L2Credentials, error) {
	return auth.DeriveCredentials(privateKeyHex, chainID)
}

func CreateL2Credentials(privateKeyHex string, chainID int) (*L2Credentials, error) {
	return auth.CreateCredentials(privateKeyHex, chainID)
}

func SignL2Request(creds *L2Credentials, method, path string, body []byte) (*L2Headers, error) {
	return auth.SignRequest(creds, method, path, body)
}

func signL2Message(secret, ts, method, path string, body []byte) (string, error) {
	return auth.SignMessage(secret, ts, method, path, body)
}

func ApplyL2Headers(req *http.Request, h *L2Headers) {
	auth.ApplyHeaders(req, h)
}

func signRelayerHMAC(secret string, timestamp int64, method, path string, body []byte) string {
	return relayer.SignHMAC(secret, timestamp, method, path, body)
}

func applyRelayerHeaders(req *http.Request, creds *RelayerCredentials, method, path string, body []byte) {
	relayer.ApplyHeaders(req, creds, method, path, body)
}

func getRelayerTransaction(ctx context.Context, transactionID string) (*RelayerTransaction, error) {
	return relayer.GetTransaction(ctx, transactionID)
}

func waitForRelayerTransaction(ctx context.Context, transactionID string) (*RelayerTransaction, error) {
	return relayer.WaitForTransaction(ctx, transactionID)
}

func EstimateSweep(book *OrderBook, side string, refPrice, size, maxSlippage float64) (*SweepEstimate, error) {
	return sweep.Estimate(book, side, refPrice, size, maxSlippage)
}

func EstimateSweepFromLevels(levels []PriceLevel, side string, refPrice, size, maxSlippage float64) (*SweepEstimate, error) {
	return sweep.EstimateFromLevels(levels, side, refPrice, size, maxSlippage)
}

type ReceiptFetcher = wallet.ReceiptFetcher

func ExecuteDepositWalletBatch(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, calls []WalletCall, deadline time.Duration, creds *RelayerCredentials) (*RelayerResponse, error) {
	return wallet.ExecuteBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, deadline, creds)
}

func deployAndResolveDepositWallet(ctx context.Context, eth ReceiptFetcher, eoaAddress string, creds *RelayerCredentials) (string, *RelayerResponse, *RelayerTransaction, error) {
	return wallet.DeployAndResolve(ctx, eth, eoaAddress, creds)
}

func wrapAndApproveDepositWallet(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	return wallet.WrapAndApprove(ctx, eoaAddress, privateKeyHex, depositWalletAddress, amount, creds)
}

func approveDepositWalletForBuyOrders(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, creds *RelayerCredentials) (*RelayerResponse, error) {
	return wallet.ApproveForBuy(ctx, eoaAddress, privateKeyHex, depositWalletAddress, creds)
}

func approveDepositWalletForSellOrders(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, creds *RelayerCredentials) (*RelayerResponse, error) {
	return wallet.ApproveForSell(ctx, eoaAddress, privateKeyHex, depositWalletAddress, creds)
}

func transferFromDepositWallet(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress, assetAddress, recipientAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	return wallet.TransferOut(ctx, eoaAddress, privateKeyHex, depositWalletAddress, assetAddress, recipientAddress, amount, creds)
}

func wrapToPUSD(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	return wallet.WrapToPUSD(ctx, eoaAddress, privateKeyHex, depositWalletAddress, amount, creds)
}

func unwrapToUSDC(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	return wallet.UnwrapToUSDC(ctx, eoaAddress, privateKeyHex, depositWalletAddress, amount, creds)
}

func collateralBalanceOf(ctx context.Context, creds *L2Credentials) (*big.Int, error) {
	c := clob.NewClient()
	resp, err := c.GetBalanceAllowance(ctx, "COLLATERAL", "", SignatureTypeGnosisSafe, creds)
	if err != nil {
		return nil, err
	}
	return parseCollateralBalance(resp.Balance), nil
}

func refreshCollateralBalance(ctx context.Context, creds *L2Credentials) error {
	c := clob.NewClient()
	return c.UpdateBalanceAllowance(ctx, "COLLATERAL", "", SignatureTypeGnosisSafe, creds)
}

func parseCollateralBalance(s string) *big.Int {
	bal, ok := new(big.Int).SetString(s, 10)
	if ok {
		return bal
	}
	f, _, err := new(big.Float).Parse(s, 10)
	if err != nil {
		return big.NewInt(0)
	}
	f.Mul(f, new(big.Float).SetFloat64(amountScaleForParse))
	raw, _ := f.Int(nil)
	return raw
}

const amountScaleForParse = 1e6

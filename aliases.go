package polytrade

import (
	"context"
	"math/big"
	"net/http"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/auth"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/clob"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/order"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/sweep"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/wallet"
)

type CLOBClient = clob.Client

func NewCLOBClient() *CLOBClient { return clob.NewClient() }

type OrderBuilder = order.Builder
type OrderOpts = order.Opts

func NewOrderBuilder(funderAddress, ctfExchangeAddress, privateKeyHex string) *OrderBuilder {
	return order.NewBuilder(funderAddress, ctfExchangeAddress, privateKeyHex)
}

type APIError = models.APIError
type L2Credentials = models.L2Credentials
type L2Headers = models.L2Headers
type RelayerCredentials = models.RelayerCredentials
type RelayerResponse = models.RelayerResponse
type RelayerTransaction = models.RelayerTransaction
type WalletCall = models.WalletCall

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
type ClosePositionOpts = models.ClosePositionOpts

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

type ContractCaller = onchain.ContractCaller
type ReceiptFetcher = wallet.ReceiptFetcher

func DeriveL2Credentials(ctx context.Context, privateKeyHex string, chainID int) (*L2Credentials, error) {
	return auth.DeriveCredentials(ctx, privateKeyHex, chainID)
}

func CreateL2Credentials(ctx context.Context, privateKeyHex string, chainID int) (*L2Credentials, error) {
	return auth.CreateCredentials(ctx, privateKeyHex, chainID)
}

func SignL2Request(creds *L2Credentials, method, path string, body []byte) (*L2Headers, error) {
	return auth.SignRequest(creds, method, path, body)
}

func ApplyL2Headers(req *http.Request, h *L2Headers) {
	auth.ApplyHeaders(req, h)
}

func EstimateSweep(book *OrderBook, side string, maxSlippage float64) (*SweepEstimate, error) {
	return sweep.Estimate(book, side, maxSlippage)
}

func EstimateSweepFromLevels(levels []PriceLevel, side string, maxSlippage float64) (*SweepEstimate, error) {
	return sweep.EstimateFromLevels(levels, side, maxSlippage)
}

func GetOutcomeTokenBalance(ctx context.Context, eth ContractCaller, ownerAddress, tokenID string) (*big.Int, error) {
	return onchain.GetOutcomeTokenBalance(ctx, eth, ownerAddress, tokenID)
}

func ExecuteDepositWalletBatch(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, calls []WalletCall, deadline time.Duration, creds *RelayerCredentials) (*RelayerResponse, error) {
	return wallet.ExecuteBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, deadline, creds)
}

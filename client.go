package polytrade

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/relayer"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/wallet"
	"github.com/ethereum/go-ethereum/crypto"
)

type EthClient interface {
	ContractCaller
	ReceiptFetcher
}

type Config struct {
	PrivateKeyHex string
	DepositWallet string
	Creds         *L2Credentials
	RelayerCreds  *RelayerCredentials
	Eth           EthClient
}

type Client struct {
	cfg           Config
	eoa           string
	clob          *CLOBClient
	mu            sync.RWMutex
	creds         *L2Credentials
	depositWallet string
	builder       *OrderBuilder
	heartbeatID   string
}

var (
	errNoCreds         = errors.New("client: no L2 credentials; call Bootstrap or set Config.Creds")
	errNoRelayerCreds  = errors.New("client: no relayer credentials; set Config.RelayerCreds")
	errNoEth           = errors.New("client: no EthClient; set Config.Eth")
	errNoDepositWallet = errors.New("client: no deposit wallet; call Bootstrap or set Config.DepositWallet")
)

func NewClient(cfg Config) (*Client, error) {
	if cfg.PrivateKeyHex == "" {
		return nil, fmt.Errorf("client: PrivateKeyHex is required")
	}
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(cfg.PrivateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("client: invalid PrivateKeyHex: %w", err)
	}
	c := &Client{
		cfg:           cfg,
		eoa:           crypto.PubkeyToAddress(pk.PublicKey).Hex(),
		clob:          NewCLOBClient(),
		creds:         cfg.Creds,
		depositWallet: cfg.DepositWallet,
	}
	if c.depositWallet != "" {
		c.builder = NewOrderBuilder(c.depositWallet, CTFExchange, cfg.PrivateKeyHex)
	}
	return c, nil
}

func (c *Client) EOA() string {
	return c.eoa
}

func (c *Client) DepositWallet() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.depositWallet
}

func (c *Client) Creds() *L2Credentials {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.creds
}

func (c *Client) snapshot() (*L2Credentials, *OrderBuilder, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.creds == nil {
		return nil, nil, "", errNoCreds
	}
	if c.builder == nil {
		return nil, nil, "", errNoDepositWallet
	}
	return c.creds, c.builder, c.depositWallet, nil
}

func (c *Client) DeriveCreds(ctx context.Context) error {
	creds, err := DeriveL2Credentials(ctx, c.cfg.PrivateKeyHex, PolygonChainID)
	if err != nil {
		creds, err = CreateL2Credentials(ctx, c.cfg.PrivateKeyHex, PolygonChainID)
		if err != nil {
			return fmt.Errorf("client: derive/create L2 credentials: %w", err)
		}
	}
	c.mu.Lock()
	c.creds = creds
	c.mu.Unlock()
	return nil
}

func (c *Client) Bootstrap(ctx context.Context, wrapAmount *big.Int) error {
	if c.cfg.RelayerCreds == nil {
		return errNoRelayerCreds
	}
	if c.cfg.Eth == nil {
		return errNoEth
	}
	if err := c.DeriveCreds(ctx); err != nil {
		return err
	}
	addr, _, _, err := wallet.DeployAndResolve(ctx, c.cfg.Eth, c.eoa, c.cfg.RelayerCreds)
	if err != nil {
		return fmt.Errorf("client: bootstrap: %w", err)
	}
	c.mu.Lock()
	c.depositWallet = addr
	c.builder = NewOrderBuilder(addr, CTFExchange, c.cfg.PrivateKeyHex)
	c.mu.Unlock()
	time.Sleep(2 * time.Second)
	if wrapAmount != nil {
		if _, err := wallet.WrapAndApprove(ctx, c.eoa, c.cfg.PrivateKeyHex, addr, wrapAmount, c.cfg.RelayerCreds); err != nil {
			return fmt.Errorf("client: bootstrap: wrap+approve: %w", err)
		}
	}
	return nil
}

func (c *Client) PrepareAndSign(tokenID, side, orderType string, price, size float64, opts ...OrderOpts) (*SignedOrder, error) {
	creds, builder, _, err := c.snapshot()
	if err != nil {
		return nil, err
	}
	return builder.PrepareAndSign(tokenID, side, orderType, price, size, creds.APIKey, opts...)
}

func (c *Client) PlaceOrder(ctx context.Context, signed *SignedOrder) (*PlaceOrderResponse, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.PlaceOrder(ctx, signed, creds)
}

func (c *Client) PlaceOrders(ctx context.Context, orders []*SignedOrder) ([]PlaceOrderResponse, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.PlaceOrders(ctx, orders, creds)
}

func (c *Client) ClosePosition(ctx context.Context, tokenID string, price float64, opts ClosePositionOpts) (*PlaceOrderResponse, error) {
	creds, builder, _, err := c.snapshot()
	if err != nil {
		return nil, err
	}
	if c.cfg.Eth == nil {
		return nil, errNoEth
	}
	balance, err := GetOutcomeTokenBalance(ctx, c.cfg.Eth, builder.MakerAddress(), tokenID)
	if err != nil {
		return nil, fmt.Errorf("close position: %w", err)
	}
	if balance.Sign() <= 0 {
		return nil, fmt.Errorf("close position: no position to close for tokenID %s", tokenID)
	}
	size := onchain.RawBalanceToSize(balance)
	orderType := opts.OrderType
	if orderType == "" {
		orderType = OrderTypeFOK
	}
	tickSize := opts.TickSize
	if tickSize == "" {
		tickSize = "0.01"
	}
	signed, err := builder.PrepareAndSign(tokenID, SELL, orderType, price, size, creds.APIKey, OrderOpts{
		TickSize:  tickSize,
		PostOnly:  opts.PostOnly,
		DeferExec: opts.DeferExec,
	})
	if err != nil {
		return nil, fmt.Errorf("close position: prepare: %w", err)
	}
	return c.clob.PlaceOrder(ctx, signed, creds)
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*OrderStatus, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.GetOrder(ctx, orderID, creds)
}

func (c *Client) GetOpenOrders(ctx context.Context, market, assetID string) ([]OrderStatus, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.GetOpenOrders(ctx, market, assetID, creds)
}

func (c *Client) GetTrades(ctx context.Context, makerAddress, market, assetID string) ([]Trade, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.GetTrades(ctx, makerAddress, market, assetID, creds)
}

func (c *Client) GetPreMigrationOrders(ctx context.Context) ([]OrderStatus, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.GetPreMigrationOrders(ctx, creds)
}

func (c *Client) CancelOrder(ctx context.Context, orderID string) (*CancelResponse, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.CancelOrder(ctx, orderID, creds)
}

func (c *Client) ReplaceOrder(ctx context.Context, oldOrderID string, newOrder *SignedOrder) (*CancelResponse, *PlaceOrderResponse, error) {
	if newOrder == nil {
		return nil, nil, fmt.Errorf("replace order: nil new order")
	}
	cancelResp, cancelErr := c.CancelOrder(ctx, oldOrderID)
	if cancelErr != nil {
		return cancelResp, nil, fmt.Errorf("replace order: cancel %s: %w", oldOrderID, cancelErr)
	}
	placeResp, placeErr := c.PlaceOrder(ctx, newOrder)
	if placeErr != nil {
		return cancelResp, placeResp, fmt.Errorf("replace order: place new: %w", placeErr)
	}
	return cancelResp, placeResp, nil
}

func (c *Client) CancelOrders(ctx context.Context, orderIDs []string) (*CancelResponse, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.CancelOrders(ctx, orderIDs, creds)
}

func (c *Client) CancelAll(ctx context.Context) (*CancelResponse, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.CancelAll(ctx, creds)
}

func (c *Client) CancelMarketOrders(ctx context.Context, market, assetID string) (*CancelResponse, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.CancelMarketOrders(ctx, market, assetID, creds)
}

func (c *Client) AwaitOrder(ctx context.Context, resp *PlaceOrderResponse, opts *PollOpts) (*PollResult, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.AwaitOrder(ctx, resp, creds, opts)
}

func (c *Client) AwaitOrders(ctx context.Context, responses []PlaceOrderResponse, opts *PollOpts) []PollResult {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil
	}
	return c.clob.AwaitOrders(ctx, responses, creds, opts)
}

func (c *Client) AwaitOrderAsync(ctx context.Context, resp *PlaceOrderResponse, opts *PollOpts) <-chan PollResult {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	return c.clob.AwaitOrderAsync(ctx, resp, creds, opts)
}

func (c *Client) AwaitOrdersAsync(ctx context.Context, responses []PlaceOrderResponse, opts *PollOpts) <-chan PollResult {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	return c.clob.AwaitOrdersAsync(ctx, responses, creds, opts)
}

func (c *Client) GetServerTime(ctx context.Context) (int64, error) {
	return c.clob.GetServerTime(ctx)
}

func (c *Client) GetOrderBook(ctx context.Context, tokenID string) (*OrderBook, error) {
	return c.clob.GetOrderBook(ctx, tokenID)
}

func (c *Client) GetPrice(ctx context.Context, tokenID, side string) (string, error) {
	return c.clob.GetPrice(ctx, tokenID, side)
}

func (c *Client) GetMidpoint(ctx context.Context, tokenID string) (string, error) {
	return c.clob.GetMidpoint(ctx, tokenID)
}

func (c *Client) GetSpread(ctx context.Context, tokenID string) (string, error) {
	return c.clob.GetSpread(ctx, tokenID)
}

func (c *Client) GetPrices(ctx context.Context, params []PriceRequest) (map[string]map[string]string, error) {
	return c.clob.GetPrices(ctx, params)
}

func (c *Client) GetSpreads(ctx context.Context, params []SpreadRequest) (map[string]string, error) {
	return c.clob.GetSpreads(ctx, params)
}

func (c *Client) GetLastTradePrice(ctx context.Context, tokenID string) (*LastTradePrice, error) {
	return c.clob.GetLastTradePrice(ctx, tokenID)
}

func (c *Client) GetLastTradePrices(ctx context.Context, params []SpreadRequest) ([]LastTradePrice, error) {
	return c.clob.GetLastTradePrices(ctx, params)
}

func (c *Client) GetPricesHistory(ctx context.Context, p PricesHistoryParams) ([]PriceHistoryEntry, error) {
	return c.clob.GetPricesHistory(ctx, p)
}

func (c *Client) GetTickSize(ctx context.Context, tokenID string) (string, error) {
	return c.clob.GetTickSize(ctx, tokenID)
}

func (c *Client) GetClobMarketInfo(ctx context.Context, conditionID string) (*ClobMarketInfo, error) {
	return c.clob.GetClobMarketInfo(ctx, conditionID)
}

func (c *Client) GetNegRisk(ctx context.Context, tokenID string) (bool, error) {
	return c.clob.GetNegRisk(ctx, tokenID)
}

func (c *Client) GetBalances(ctx context.Context) ([]BalanceEntry, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.GetBalances(ctx, creds)
}

func (c *Client) GetMarkets(ctx context.Context, nextCursor string) (*PaginatedResponse[MarketInfo], error) {
	return c.clob.GetMarkets(ctx, nextCursor)
}

func (c *Client) GetMarket(ctx context.Context, conditionID string) (*MarketInfo, error) {
	return c.clob.GetMarket(ctx, conditionID)
}

func (c *Client) GetSamplingMarkets(ctx context.Context, nextCursor string) (*PaginatedResponse[MarketInfo], error) {
	return c.clob.GetSamplingMarkets(ctx, nextCursor)
}

func (c *Client) GetSimplifiedMarkets(ctx context.Context, nextCursor string) (*PaginatedResponse[SimplifiedMarketInfo], error) {
	return c.clob.GetSimplifiedMarkets(ctx, nextCursor)
}

func (c *Client) GetSamplingSimplifiedMarkets(ctx context.Context, nextCursor string) (*PaginatedResponse[SimplifiedMarketInfo], error) {
	return c.clob.GetSamplingSimplifiedMarkets(ctx, nextCursor)
}

func (c *Client) GetMarketByToken(ctx context.Context, tokenID string) (*MarketByTokenInfo, error) {
	return c.clob.GetMarketByToken(ctx, tokenID)
}

func (c *Client) GetMarketLiveActivity(ctx context.Context, conditionID string) (*MarketLiveActivity, error) {
	return c.clob.GetMarketLiveActivity(ctx, conditionID)
}

func (c *Client) GetPositions(ctx context.Context) ([]PositionEntry, error) {
	c.mu.RLock()
	dw := c.depositWallet
	c.mu.RUnlock()
	if dw == "" {
		return nil, errNoDepositWallet
	}
	return c.clob.GetPositions(ctx, dw)
}

func (c *Client) GetPositionsOf(ctx context.Context, walletAddress string) ([]PositionEntry, error) {
	return c.clob.GetPositions(ctx, walletAddress)
}

func (c *Client) GetBalanceAllowance(ctx context.Context, assetType, tokenID string) (*BalanceAllowanceResponse, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.GetBalanceAllowance(ctx, assetType, tokenID, creds)
}

func (c *Client) UpdateBalanceAllowance(ctx context.Context, assetType, tokenID string) error {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return errNoCreds
	}
	return c.clob.UpdateBalanceAllowance(ctx, assetType, tokenID, creds)
}

func (c *Client) CollateralBalanceOf(ctx context.Context) (*big.Int, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return wallet.CollateralBalance(ctx, creds)
}

func (c *Client) RefreshCollateralBalance(ctx context.Context) error {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return errNoCreds
	}
	return wallet.RefreshCollateralBalance(ctx, creds)
}

func (c *Client) GetCurrentRewards(ctx context.Context) ([]CurrentRewardMarket, error) {
	return c.clob.GetCurrentRewards(ctx)
}

func (c *Client) GetEarningsForUserForDay(ctx context.Context, date string) ([]map[string]any, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.GetEarningsForUserForDay(ctx, date, creds)
}

func (c *Client) GetRewardPercentages(ctx context.Context) (map[string]any, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.GetRewardPercentages(ctx, creds)
}

func (c *Client) IsOrderScoring(ctx context.Context, orderID string) (*OrderScoringResult, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.IsOrderScoring(ctx, orderID, creds)
}

func (c *Client) AreOrdersScoring(ctx context.Context, orderIDs []string) (map[string]bool, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return c.clob.AreOrdersScoring(ctx, orderIDs, creds)
}

func (c *Client) WrapToPUSD(ctx context.Context, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.WrapToPUSD(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, amount, c.cfg.RelayerCreds)
}

func (c *Client) UnwrapToUSDC(ctx context.Context, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.UnwrapToUSDC(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, amount, c.cfg.RelayerCreds)
}

func (c *Client) TransferOut(ctx context.Context, asset, recipient string, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.TransferOut(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, asset, recipient, amount, c.cfg.RelayerCreds)
}

func (c *Client) GetRelayerTransaction(ctx context.Context, transactionID string) (*RelayerTransaction, error) {
	return relayer.GetTransaction(ctx, transactionID)
}

func (c *Client) WaitForRelayerTransaction(ctx context.Context, transactionID string) (*RelayerTransaction, error) {
	return relayer.WaitForTransaction(ctx, transactionID)
}

func (c *Client) SetupWalletForTrading(ctx context.Context, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.WrapAndApprove(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, amount, c.cfg.RelayerCreds)
}

func (c *Client) ApproveForBuy(ctx context.Context) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.ApproveForBuy(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, c.cfg.RelayerCreds)
}

func (c *Client) ApproveForSell(ctx context.Context) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.ApproveForSell(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, c.cfg.RelayerCreds)
}

func (c *Client) SplitPosition(ctx context.Context, conditionID string, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.SplitPosition(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, conditionID, amount, c.cfg.RelayerCreds)
}

func (c *Client) MergePositions(ctx context.Context, conditionID string, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.MergePositions(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, conditionID, amount, c.cfg.RelayerCreds)
}

func (c *Client) RedeemPositions(ctx context.Context, conditionID string) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.RedeemPositions(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, conditionID, c.cfg.RelayerCreds)
}

func (c *Client) requireDepositWalletOps() (string, error) {
	if c.cfg.RelayerCreds == nil {
		return "", errNoRelayerCreds
	}
	c.mu.RLock()
	dw := c.depositWallet
	c.mu.RUnlock()
	if dw == "" {
		return "", errNoDepositWallet
	}
	return dw, nil
}

func (c *Client) PostHeartbeat(ctx context.Context) (string, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return "", errNoCreds
	}
	c.mu.Lock()
	prev := c.heartbeatID
	c.mu.Unlock()
	id, err := c.clob.PostHeartbeat(ctx, prev, creds)
	if id != "" {
		c.mu.Lock()
		c.heartbeatID = id
		c.mu.Unlock()
	}
	return id, err
}

func (c *Client) RunHeartbeat(ctx context.Context, interval time.Duration) <-chan error {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	return c.clob.RunHeartbeat(ctx, interval, creds)
}

func (c *Client) SetBuilderCode(code string) {
	c.mu.RLock()
	b := c.builder
	c.mu.RUnlock()
	if b != nil {
		b.SetBuilderCode(code)
	}
}

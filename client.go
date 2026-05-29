package polytrade

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/relayer"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/wallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const DefaultPolygonRPCURL = "https://polygon.publicnode.com"

type EthClient interface {
	ContractCaller
	ReceiptFetcher
}

type Option func(*clientConfig)

type clientConfig struct {
	rpcURL string
	eth    EthClient
}

func WithRPCURL(url string) Option { // default if neither this nor WithEthClient is set
	return func(cc *clientConfig) { cc.rpcURL = url }
}

func WithEthClient(eth EthClient) Option { // takes precedence over WithRPCURL
	return func(cc *clientConfig) { cc.eth = eth }
}

type Client struct {
	privateKeyHex        string
	eoa                  string
	clob                 *CLOBClient
	relayerCreds         *RelayerCredentials
	eth                  EthClient
	mu                   sync.RWMutex
	l2Creds              *L2Credentials
	depositWalletAddress string
	builder              *OrderBuilder
	heartbeatID          string
	lastRecoveryErr      error
	bootstrapMu          sync.Mutex
}

var (
	errNoCreds         = errors.New("client: no L2 credentials; NewClient should derive them, check CLOB reachability")
	errNoRelayerCreds  = errors.New("client: no relayer credentials; pass them as the third positional arg to NewClient")
	errNoEth           = errors.New("client: no EthClient; pass WithRPCURL or WithEthClient")
	errNoDepositWallet = errors.New("client: no deposit wallet; call Bootstrap to deploy or attach an EthClient so NewClient can auto-recover")
	errNotApproved     = errors.New("client: deposit wallet approvals not set; call Bootstrap or EnsureApprovals")

	ErrWalletAlreadyDeployed = errors.New("client: deposit wallet already deployed for this EOA, attach an EthClient so NewClient can auto-recover it")
)

func NewClient(ctx context.Context, privateKeyHex string, relayerCreds *RelayerCredentials, opts ...Option) (*Client, error) {
	if privateKeyHex == "" {
		return nil, errors.New("client: privateKeyHex is required")
	}
	if relayerCreds == nil {
		return nil, errors.New("client: relayerCreds is required")
	}
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("client: invalid privateKeyHex: %w", err)
	}

	cc := &clientConfig{rpcURL: DefaultPolygonRPCURL}
	for _, opt := range opts {
		opt(cc)
	}

	c := &Client{
		privateKeyHex: privateKeyHex,
		eoa:           crypto.PubkeyToAddress(pk.PublicKey).Hex(),
		clob:          NewCLOBClient(),
		relayerCreds:  relayerCreds,
		eth:           cc.eth,
	}
	if c.eth == nil && cc.rpcURL != "" {
		dialed, err := ethclient.DialContext(ctx, cc.rpcURL)
		if err != nil {
			return nil, fmt.Errorf("client: dial RPC %q: %w", cc.rpcURL, err)
		}
		c.eth = dialed
	}

	if err := c.DeriveCreds(ctx); err != nil {
		return nil, fmt.Errorf("client: derive credentials: %w", err)
	}

	c.recoverDepositWallet(ctx)
	return c, nil
}

func (c *Client) recoverDepositWallet(ctx context.Context) {
	c.mu.RLock()
	already := c.depositWalletAddress != ""
	eth := c.eth
	c.mu.RUnlock()
	if already || eth == nil {
		return
	}
	reader, ok := eth.(CodeReader)
	if !ok {
		return
	}
	addr, deployed, err := LookupDepositWallet(ctx, reader, c.eoa)
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.lastRecoveryErr = err
		return
	}
	if !deployed {
		return
	}
	c.depositWalletAddress = addr
	c.builder = NewOrderBuilder(addr, CTFExchange, c.privateKeyHex)
	c.lastRecoveryErr = nil
}

func (c *Client) EOA() string {
	return c.eoa
}

func (c *Client) DepositWallet() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.depositWalletAddress
}

func (c *Client) Creds() *L2Credentials {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.l2Creds
}

func (c *Client) IsReady(ctx context.Context) (bool, error) {
	c.mu.RLock()
	creds := c.l2Creds
	relayerCreds := c.relayerCreds
	eth := c.eth
	dw := c.depositWalletAddress
	recoveryErr := c.lastRecoveryErr
	c.mu.RUnlock()
	var errs []error
	if creds == nil {
		errs = append(errs, errNoCreds)
	}
	if relayerCreds == nil {
		errs = append(errs, errNoRelayerCreds)
	}
	if eth == nil {
		errs = append(errs, errNoEth)
	}
	if dw == "" {
		errs = append(errs, errNoDepositWallet)
		if recoveryErr != nil {
			errs = append(errs, fmt.Errorf("client: deposit wallet recovery failed: %w", recoveryErr))
		}
	}
	if eth != nil && dw != "" {
		ok, err := onchain.IsFullyApproved(ctx, eth, common.HexToAddress(dw))
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("client: check approvals: %w", err))
		case !ok:
			errs = append(errs, errNotApproved)
		}
	}
	if len(errs) == 0 {
		return true, nil
	}
	return false, errors.Join(errs...)
}

func (c *Client) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	setStr := func(b bool) string {
		if b {
			return "set"
		}
		return "unset"
	}
	dw := c.depositWalletAddress
	if dw == "" {
		dw = "unset"
	}
	return fmt.Sprintf("polytrade.Client{eoa:%s, deposit:%s, creds:%s, relayer:%s, eth:%s}",
		c.eoa, dw, setStr(c.l2Creds != nil), setStr(c.relayerCreds != nil), setStr(c.eth != nil))
}

func (c *Client) snapshot() (*L2Credentials, *OrderBuilder, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.l2Creds == nil {
		return nil, nil, "", errNoCreds
	}
	if c.builder == nil {
		return nil, nil, "", errNoDepositWallet
	}
	return c.l2Creds, c.builder, c.depositWalletAddress, nil
}

func (c *Client) requireCreds() (*L2Credentials, error) {
	c.mu.RLock()
	creds := c.l2Creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	return creds, nil
}

func (c *Client) DeriveCreds(ctx context.Context) error {
	creds, err := DeriveL2Credentials(ctx, c.privateKeyHex, PolygonChainID)
	if err != nil {
		// CLOB returns 400 "Could not derive api key!" when no key exists for this wallet yet 
		// anything else is a real failure and should surface as is to be properly handled by the caller
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
			return fmt.Errorf("client: derive L2 credentials: %w", err)
		}
		creds, err = CreateL2Credentials(ctx, c.privateKeyHex, PolygonChainID)
		if err != nil {
			return fmt.Errorf("client: create L2 credentials: %w", err)
		}
	}
	c.mu.Lock()
	c.l2Creds = creds
	c.mu.Unlock()
	return nil
}

func (c *Client) Deploy(ctx context.Context) (string, error) {
	if c.relayerCreds == nil {
		return "", errNoRelayerCreds
	}
	if c.eth == nil {
		return "", errNoEth
	}
	c.bootstrapMu.Lock()
	defer c.bootstrapMu.Unlock()
	c.mu.RLock()
	addr := c.depositWalletAddress
	c.mu.RUnlock()
	if addr != "" {
		return addr, nil
	}
	deployed, _, tx, err := wallet.DeployAndResolve(ctx, c.eth, c.eoa, c.relayerCreds)
	if err != nil {
		if tx != nil && tx.TransactionHash == "" {
			return "", fmt.Errorf("%w: %v", ErrWalletAlreadyDeployed, err)
		}
		return "", fmt.Errorf("client: deploy: %w", err)
	}
	c.mu.Lock()
	c.depositWalletAddress = deployed
	c.builder = NewOrderBuilder(deployed, CTFExchange, c.privateKeyHex)
	c.mu.Unlock()
	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Done():
		return "", ctx.Err()
	}
	return deployed, nil
}

func (c *Client) EnsureApprovals(ctx context.Context) error {
	if c.relayerCreds == nil {
		return errNoRelayerCreds
	}
	addr := c.DepositWallet()
	if addr == "" {
		return errNoDepositWallet
	}
	c.bootstrapMu.Lock()
	defer c.bootstrapMu.Unlock()
	if c.eth != nil {
		ok, err := onchain.IsFullyApproved(ctx, c.eth, common.HexToAddress(addr))
		if err != nil {
			return fmt.Errorf("client: ensure approvals: check on-chain: %w", err)
		}
		if ok {
			return nil
		}
	}
	resp, err := wallet.ApproveAll(ctx, c.eoa, c.privateKeyHex, addr, c.relayerCreds)
	if err != nil {
		return fmt.Errorf("client: ensure approvals: %w", err)
	}
	if _, err := relayer.WaitForTransaction(ctx, resp.TransactionID); err != nil {
		return fmt.Errorf("client: ensure approvals: wait: %w", err)
	}
	return nil
}

func (c *Client) Bootstrap(ctx context.Context) error {
	c.recoverDepositWallet(ctx)
	if c.DepositWallet() == "" {
		if _, err := c.Deploy(ctx); err != nil {
			return err
		}
	}
	return c.EnsureApprovals(ctx)
}

func (c *Client) PrepareAndSign(tokenID, side, orderType string, price, size float64, opts ...OrderOpts) (*SignedOrder, error) {
	creds, builder, _, err := c.snapshot()
	if err != nil {
		return nil, err
	}
	return builder.PrepareAndSign(tokenID, side, orderType, price, size, creds.APIKey, opts...)
}

func (c *Client) PlaceOrder(ctx context.Context, signed *SignedOrder) (*PlaceOrderResponse, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.PlaceOrder(ctx, signed, creds)
}

func (c *Client) PlaceOrders(ctx context.Context, orders []*SignedOrder) ([]PlaceOrderResponse, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.PlaceOrders(ctx, orders, creds)
}

func (c *Client) ClosePosition(ctx context.Context, tokenID string, price float64, opts ClosePositionOpts) (*PlaceOrderResponse, error) {
	creds, builder, _, err := c.snapshot()
	if err != nil {
		return nil, err
	}
	if c.eth == nil {
		return nil, errNoEth
	}
	balance, err := GetOutcomeTokenBalance(ctx, c.eth, builder.MakerAddress(), tokenID)
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
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.GetOrder(ctx, orderID, creds)
}

func (c *Client) GetOpenOrders(ctx context.Context, market, assetID string) ([]OrderStatus, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.GetOpenOrders(ctx, market, assetID, creds)
}

func (c *Client) GetTrades(ctx context.Context, makerAddress, market, assetID string) ([]Trade, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.GetTrades(ctx, makerAddress, market, assetID, creds)
}

func (c *Client) GetPreMigrationOrders(ctx context.Context) ([]OrderStatus, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.GetPreMigrationOrders(ctx, creds)
}

func (c *Client) CancelOrder(ctx context.Context, orderID string) (*CancelResponse, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.CancelOrder(ctx, orderID, creds)
}

func (c *Client) ERC20Balance(ctx context.Context, tokenAddress string) (*big.Int, error) {
	if c.eth == nil {
		return nil, errNoEth
	}
	dw := c.DepositWallet()
	if dw == "" {
		return nil, errNoDepositWallet
	}
	return onchain.ERC20BalanceOf(ctx, c.eth, tokenAddress, dw)
}

func (c *Client) USDCBalance(ctx context.Context) (*big.Int, error) {
	return c.ERC20Balance(ctx, USDCAddress)
}

func (c *Client) PUSDBalance(ctx context.Context) (*big.Int, error) {
	return c.ERC20Balance(ctx, PUSDAddress)
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
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.CancelOrders(ctx, orderIDs, creds)
}

func (c *Client) CancelAll(ctx context.Context) (*CancelResponse, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.CancelAll(ctx, creds)
}

func (c *Client) CancelMarketOrders(ctx context.Context, market, assetID string) (*CancelResponse, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.CancelMarketOrders(ctx, market, assetID, creds)
}

func (c *Client) AwaitOrder(ctx context.Context, resp *PlaceOrderResponse, opts *PollOpts) (*PollResult, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.AwaitOrder(ctx, resp, creds, opts)
}

func (c *Client) AwaitOrders(ctx context.Context, responses []PlaceOrderResponse, opts *PollOpts) []PollResult {
	c.mu.RLock()
	creds := c.l2Creds
	c.mu.RUnlock()
	if creds == nil {
		return nil
	}
	return c.clob.AwaitOrders(ctx, responses, creds, opts)
}

func (c *Client) AwaitOrderAsync(ctx context.Context, resp *PlaceOrderResponse, opts *PollOpts) <-chan PollResult {
	c.mu.RLock()
	creds := c.l2Creds
	c.mu.RUnlock()
	return c.clob.AwaitOrderAsync(ctx, resp, creds, opts)
}

func (c *Client) AwaitOrdersAsync(ctx context.Context, responses []PlaceOrderResponse, opts *PollOpts) <-chan PollResult {
	c.mu.RLock()
	creds := c.l2Creds
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
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
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
	dw := c.depositWalletAddress
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
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.GetBalanceAllowance(ctx, assetType, tokenID, creds)
}

func (c *Client) UpdateBalanceAllowance(ctx context.Context, assetType, tokenID string) error {
	creds, err := c.requireCreds()
	if err != nil {
		return err
	}
	return c.clob.UpdateBalanceAllowance(ctx, assetType, tokenID, creds)
}

func (c *Client) CollateralBalanceOf(ctx context.Context) (*big.Int, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return wallet.CollateralBalance(ctx, creds)
}

func (c *Client) RefreshCollateralBalance(ctx context.Context) error {
	creds, err := c.requireCreds()
	if err != nil {
		return err
	}
	return wallet.RefreshCollateralBalance(ctx, creds)
}

func (c *Client) GetCurrentRewards(ctx context.Context) ([]CurrentRewardMarket, error) {
	return c.clob.GetCurrentRewards(ctx)
}

func (c *Client) GetEarningsForUserForDay(ctx context.Context, date string) ([]map[string]any, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.GetEarningsForUserForDay(ctx, date, creds)
}

func (c *Client) GetRewardPercentages(ctx context.Context) (map[string]any, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.GetRewardPercentages(ctx, creds)
}

func (c *Client) IsOrderScoring(ctx context.Context, orderID string) (*OrderScoringResult, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.IsOrderScoring(ctx, orderID, creds)
}

func (c *Client) AreOrdersScoring(ctx context.Context, orderIDs []string) (map[string]bool, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return nil, err
	}
	return c.clob.AreOrdersScoring(ctx, orderIDs, creds)
}

func (c *Client) WrapToPUSD(ctx context.Context, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.WrapToPUSD(ctx, c.eoa, c.privateKeyHex, dw, amount, c.relayerCreds)
}

func (c *Client) UnwrapToUSDC(ctx context.Context, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.UnwrapToUSDC(ctx, c.eoa, c.privateKeyHex, dw, amount, c.relayerCreds)
}

func (c *Client) TransferOut(ctx context.Context, asset, recipient string, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.TransferOut(ctx, c.eoa, c.privateKeyHex, dw, asset, recipient, amount, c.relayerCreds)
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
	return wallet.WrapAndApprove(ctx, c.eoa, c.privateKeyHex, dw, amount, c.relayerCreds)
}

func (c *Client) ApproveForBuy(ctx context.Context) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.ApproveForBuy(ctx, c.eoa, c.privateKeyHex, dw, c.relayerCreds)
}

func (c *Client) ApproveForSell(ctx context.Context) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	return wallet.ApproveForSell(ctx, c.eoa, c.privateKeyHex, dw, c.relayerCreds)
}

func (c *Client) SplitPosition(ctx context.Context, conditionID string, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("client: split position: amount must be positive")
	}
	negRisk, err := c.resolveNegRisk(ctx, conditionID)
	if err != nil {
		return nil, err
	}
	return wallet.SplitPosition(ctx, c.eoa, c.privateKeyHex, dw, conditionID, amount, negRisk, c.relayerCreds)
}

func (c *Client) MergePositions(ctx context.Context, conditionID string, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("client: merge positions: amount must be positive")
	}
	negRisk, err := c.resolveNegRisk(ctx, conditionID)
	if err != nil {
		return nil, err
	}
	return wallet.MergePositions(ctx, c.eoa, c.privateKeyHex, dw, conditionID, amount, negRisk, c.relayerCreds)
}

func (c *Client) RedeemPositions(ctx context.Context, conditionID string) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	negRisk, err := c.resolveNegRisk(ctx, conditionID)
	if err != nil {
		return nil, err
	}
	return wallet.RedeemPositions(ctx, c.eoa, c.privateKeyHex, dw, conditionID, negRisk, c.relayerCreds)
}

func (c *Client) resolveNegRisk(ctx context.Context, conditionID string) (bool, error) {
	mkt, err := c.GetMarket(ctx, conditionID)
	if err != nil {
		return false, fmt.Errorf("client: resolve neg risk: %w", err)
	}
	return mkt.NegRisk, nil
}

func (c *Client) requireDepositWalletOps() (string, error) {
	if c.relayerCreds == nil {
		return "", errNoRelayerCreds
	}
	c.mu.RLock()
	dw := c.depositWalletAddress
	c.mu.RUnlock()
	if dw == "" {
		return "", errNoDepositWallet
	}
	return dw, nil
}

func (c *Client) PostHeartbeat(ctx context.Context) (string, error) {
	creds, err := c.requireCreds()
	if err != nil {
		return "", err
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
	creds := c.l2Creds
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

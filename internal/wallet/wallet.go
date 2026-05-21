package wallet

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/relayer"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type ReceiptFetcher interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
}

func AddressFromReceipt(ctx context.Context, eth ReceiptFetcher, txHash string) (string, error) {
	receipt, err := eth.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return "", fmt.Errorf("deposit wallet address from receipt: %w", err)
	}
	if receipt == nil {
		return "", fmt.Errorf("deposit wallet address from receipt: nil receipt")
	}
	factory := common.HexToAddress(consts.DepositWalletFactory)
	for _, lg := range receipt.Logs {
		if lg.Address != factory {
			return lg.Address.Hex(), nil
		}
	}
	return "", fmt.Errorf("deposit wallet address from receipt: no log emitter other than the factory")
}

func DeployAndResolve(ctx context.Context, eth ReceiptFetcher, eoaAddress string, creds *types.RelayerCredentials) (string, *types.RelayerResponse, *types.RelayerTransaction, error) {
	deployResp, err := Deploy(ctx, eoaAddress, creds)
	if err != nil {
		return "", nil, nil, err
	}
	tx, err := relayer.WaitForTransaction(ctx, deployResp.TransactionID)
	if err != nil {
		return "", deployResp, tx, err
	}
	addr, err := AddressFromReceipt(ctx, eth, tx.TransactionHash)
	if err != nil {
		return "", deployResp, tx, err
	}
	return addr, deployResp, tx, nil
}

func Deploy(ctx context.Context, eoaAddress string, creds *types.RelayerCredentials) (*types.RelayerResponse, error) {
	body, err := json.Marshal(map[string]string{
		"type": "WALLET-CREATE",
		"from": eoaAddress,
		"to":   consts.DepositWalletFactory,
	})
	if err != nil {
		return nil, fmt.Errorf("deploy deposit wallet: marshal: %w", err)
	}
	endpoint := "/submit"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, consts.RelayerBaseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deploy deposit wallet: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if creds != nil {
		relayer.ApplyHeaders(req, creds, http.MethodPost, endpoint, body)
	}
	httpClient := &http.Client{Timeout: consts.CLOBTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deploy deposit wallet: http request: %w", err)
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("deploy deposit wallet: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, &types.APIError{StatusCode: resp.StatusCode, Endpoint: "POST " + endpoint, Body: string(out)}
	}
	var r types.RelayerResponse
	if err := json.Unmarshal(out, &r); err != nil {
		return nil, fmt.Errorf("deploy deposit wallet: unmarshal response: %w", err)
	}
	return &r, nil
}

func Nonce(ctx context.Context, eoaAddress string, creds *types.RelayerCredentials) (string, error) {
	signPath := "/nonce"
	fullURL := fmt.Sprintf("%s/nonce?address=%s&type=WALLET", consts.RelayerBaseURL, eoaAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("deposit wallet nonce: build request: %w", err)
	}
	if creds != nil {
		relayer.ApplyHeaders(req, creds, http.MethodGet, signPath, nil)
	}
	httpClient := &http.Client{Timeout: consts.DefaultTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("deposit wallet nonce: http request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("deposit wallet nonce: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", &types.APIError{StatusCode: resp.StatusCode, Endpoint: "GET /nonce", Body: string(body)}
	}
	var r models.NonceResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("deposit wallet nonce: unmarshal response: %w", err)
	}
	return r.Nonce, nil
}

func ExecuteBatch(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, calls []types.WalletCall, deadline time.Duration, creds *types.RelayerCredentials) (*types.RelayerResponse, error) {
	if len(calls) == 0 {
		return nil, fmt.Errorf("execute deposit wallet batch: no calls provided")
	}
	if deadline <= 0 {
		deadline = 15 * time.Minute
	}
	deadlineUnix := time.Now().Add(deadline).Unix()

	nonceStr, err := Nonce(ctx, eoaAddress, creds)
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: %w", err)
	}
	nonce, err := strconv.ParseInt(nonceStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: parse nonce: %w", err)
	}
	signature, err := onchain.SignBatch(privateKeyHex, depositWalletAddress, nonce, deadlineUnix, calls)
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: %w", err)
	}

	callsJSON := make([]map[string]string, 0, len(calls))
	for _, c := range calls {
		value := c.Value
		if value == nil {
			value = new(big.Int)
		}
		callsJSON = append(callsJSON, map[string]string{
			"target": c.Target,
			"value":  value.String(),
			"data":   "0x" + hex.EncodeToString(c.Data),
		})
	}

	body, err := json.Marshal(map[string]any{
		"type":      "WALLET",
		"from":      eoaAddress,
		"to":        consts.DepositWalletFactory,
		"nonce":     strconv.FormatInt(nonce, 10),
		"signature": signature,
		"depositWalletParams": map[string]any{
			"depositWallet": depositWalletAddress,
			"deadline":      strconv.FormatInt(deadlineUnix, 10),
			"calls":         callsJSON,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: marshal: %w", err)
	}
	endpoint := "/submit"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, consts.RelayerBaseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if creds != nil {
		relayer.ApplyHeaders(req, creds, http.MethodPost, endpoint, body)
	}
	httpClient := &http.Client{Timeout: consts.CLOBTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: http request: %w", err)
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, &types.APIError{StatusCode: resp.StatusCode, Endpoint: "POST " + endpoint, Body: string(out)}
	}
	var r types.RelayerResponse
	if err := json.Unmarshal(out, &r); err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: unmarshal response: %w", err)
	}
	return &r, nil
}

func ApproveForBuy(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, creds *types.RelayerCredentials) (*types.RelayerResponse, error) {
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []types.WalletCall{
		{Target: consts.PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.CTFExchange, maxU)},
		{Target: consts.PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.NegRiskCTFExchange, maxU)},
	}
	return ExecuteBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func ApproveForSell(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, creds *types.RelayerCredentials) (*types.RelayerResponse, error) {
	calls := []types.WalletCall{
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeSetApprovalForAllCalldata(consts.CTFExchange, true)},
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeSetApprovalForAllCalldata(consts.NegRiskCTFExchange, true)},
	}
	return ExecuteBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func TransferOut(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress, assetAddress, recipientAddress string, amount *big.Int, creds *types.RelayerCredentials) (*types.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("transfer from deposit wallet: amount must be positive")
	}
	calls := []types.WalletCall{
		{Target: assetAddress, Value: new(big.Int), Data: onchain.EncodeTransferCalldata(recipientAddress, amount)},
	}
	return ExecuteBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func WrapAndApprove(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *types.RelayerCredentials) (*types.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("wrap and approve deposit wallet: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []types.WalletCall{
		{Target: consts.USDCAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.CollateralOnramp, maxU)},
		{Target: consts.CollateralOnramp, Value: new(big.Int), Data: onchain.EncodeOnrampWrapCalldata(consts.USDCAddress, depositWalletAddress, amount)},
		{Target: consts.PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.CTFExchange, maxU)},
		{Target: consts.PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.NegRiskCTFExchange, maxU)},
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeSetApprovalForAllCalldata(consts.CTFExchange, true)},
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeSetApprovalForAllCalldata(consts.NegRiskCTFExchange, true)},
	}
	return ExecuteBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func WrapToPUSD(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *types.RelayerCredentials) (*types.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("wrap to pusd: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []types.WalletCall{
		{Target: consts.USDCAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.CollateralOnramp, maxU)},
		{Target: consts.CollateralOnramp, Value: new(big.Int), Data: onchain.EncodeOnrampWrapCalldata(consts.USDCAddress, depositWalletAddress, amount)},
	}
	return ExecuteBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func UnwrapToUSDC(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *types.RelayerCredentials) (*types.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("unwrap to usdc: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []types.WalletCall{
		{Target: consts.PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.CollateralOfframp, maxU)},
		{Target: consts.CollateralOfframp, Value: new(big.Int), Data: onchain.EncodeOfframpUnwrapCalldata(consts.USDCAddress, depositWalletAddress, amount)},
	}
	return ExecuteBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

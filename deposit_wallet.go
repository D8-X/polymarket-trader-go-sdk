package polytrade

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

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type WalletCall struct {
	Target string
	Value  *big.Int
	Data   []byte
}

type ReceiptFetcher interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
}

func depositWalletAddressFromReceipt(ctx context.Context, eth ReceiptFetcher, txHash string) (string, error) {
	receipt, err := eth.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return "", fmt.Errorf("deposit wallet address from receipt: %w", err)
	}
	if receipt == nil {
		return "", fmt.Errorf("deposit wallet address from receipt: nil receipt")
	}
	factory := common.HexToAddress(depositWalletFactory)
	for _, lg := range receipt.Logs {
		if lg.Address != factory {
			return lg.Address.Hex(), nil
		}
	}
	return "", fmt.Errorf("deposit wallet address from receipt: no log emitter other than the factory")
}

func deployAndResolveDepositWallet(ctx context.Context, eth ReceiptFetcher, eoaAddress string, creds *RelayerCredentials) (string, *RelayerResponse, *RelayerTransaction, error) {
	deployResp, err := deployDepositWallet(ctx, eoaAddress, creds)
	if err != nil {
		return "", nil, nil, err
	}
	tx, err := waitForRelayerTransaction(ctx, deployResp.TransactionID)
	if err != nil {
		return "", deployResp, tx, err
	}
	addr, err := depositWalletAddressFromReceipt(ctx, eth, tx.TransactionHash)
	if err != nil {
		return "", deployResp, tx, err
	}
	return addr, deployResp, tx, nil
}

func deployDepositWallet(ctx context.Context, eoaAddress string, creds *RelayerCredentials) (*RelayerResponse, error) {
	body, err := json.Marshal(map[string]string{
		"type": "WALLET-CREATE",
		"from": eoaAddress,
		"to":   depositWalletFactory,
	})
	if err != nil {
		return nil, fmt.Errorf("deploy deposit wallet: marshal: %w", err)
	}
	endpoint := "/submit"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, relayerBaseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deploy deposit wallet: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if creds != nil {
		applyRelayerHeaders(req, creds, http.MethodPost, endpoint, body)
	}
	httpClient := &http.Client{Timeout: clobTimeout}
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
		return nil, &APIError{StatusCode: resp.StatusCode, Endpoint: "POST " + endpoint, Body: string(out)}
	}
	var r RelayerResponse
	if err := json.Unmarshal(out, &r); err != nil {
		return nil, fmt.Errorf("deploy deposit wallet: unmarshal response: %w", err)
	}
	return &r, nil
}

func depositWalletNonce(ctx context.Context, eoaAddress string, creds *RelayerCredentials) (string, error) {
	signPath := "/nonce"
	fullURL := fmt.Sprintf("%s/nonce?address=%s&type=WALLET", relayerBaseURL, eoaAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("deposit wallet nonce: build request: %w", err)
	}
	if creds != nil {
		applyRelayerHeaders(req, creds, http.MethodGet, signPath, nil)
	}
	httpClient := &http.Client{Timeout: defaultTimeout}
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
		return "", &APIError{StatusCode: resp.StatusCode, Endpoint: "GET /nonce", Body: string(body)}
	}
	var r nonceResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("deposit wallet nonce: unmarshal response: %w", err)
	}
	return r.Nonce, nil
}

func depositWalletDomainSeparator(walletAddress string) []byte {
	domainTypeHash := ethutil.Keccak256([]byte(eip712DepositWalletDomainType))
	return ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(domainTypeHash),
		ethutil.PadTo32(ethutil.Keccak256([]byte(eip712DepositWalletName))),
		ethutil.PadTo32(ethutil.Keccak256([]byte(eip712DepositWalletVersion))),
		ethutil.PadTo32(big.NewInt(PolygonChainID).Bytes()),
		ethutil.PadTo32(common.HexToAddress(walletAddress).Bytes()),
	))
}

func callStructHash(c WalletCall) []byte {
	value := c.Value
	if value == nil {
		value = new(big.Int)
	}
	return ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(ethutil.Keccak256([]byte(eip712CallType))),
		ethutil.PadTo32(common.HexToAddress(c.Target).Bytes()),
		ethutil.PadTo32(value.Bytes()),
		ethutil.Keccak256(c.Data),
	))
}

func signBatch(privateKeyHex, walletAddress string, nonce, deadline int64, calls []WalletCall) (string, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("sign batch: invalid private key: %w", err)
	}
	domainSep := depositWalletDomainSeparator(walletAddress)
	var callHashes []byte
	for _, c := range calls {
		callHashes = append(callHashes, callStructHash(c)...)
	}
	callsArrayHash := ethutil.Keccak256(callHashes)
	structHash := ethutil.Keccak256(ethutil.Concat(
		ethutil.PadTo32(ethutil.Keccak256([]byte(eip712BatchType))),
		ethutil.PadTo32(common.HexToAddress(walletAddress).Bytes()),
		ethutil.PadTo32(big.NewInt(nonce).Bytes()),
		ethutil.PadTo32(big.NewInt(deadline).Bytes()),
		callsArrayHash,
	))
	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)
	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		return "", fmt.Errorf("sign batch: sign EIP-712 digest: %w", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}
	return "0x" + common.Bytes2Hex(sig), nil
}

func ExecuteDepositWalletBatch(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, calls []WalletCall, deadline time.Duration, creds *RelayerCredentials) (*RelayerResponse, error) {
	if len(calls) == 0 {
		return nil, fmt.Errorf("execute deposit wallet batch: no calls provided")
	}
	if deadline <= 0 {
		deadline = 15 * time.Minute
	}
	deadlineUnix := time.Now().Add(deadline).Unix()

	nonceStr, err := depositWalletNonce(ctx, eoaAddress, creds)
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: %w", err)
	}
	nonce, err := strconv.ParseInt(nonceStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: parse nonce: %w", err)
	}
	signature, err := signBatch(privateKeyHex, depositWalletAddress, nonce, deadlineUnix, calls)
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
		"to":        depositWalletFactory,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, relayerBaseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if creds != nil {
		applyRelayerHeaders(req, creds, http.MethodPost, endpoint, body)
	}
	httpClient := &http.Client{Timeout: clobTimeout}
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
		return nil, &APIError{StatusCode: resp.StatusCode, Endpoint: "POST " + endpoint, Body: string(out)}
	}
	var r RelayerResponse
	if err := json.Unmarshal(out, &r); err != nil {
		return nil, fmt.Errorf("execute deposit wallet batch: unmarshal response: %w", err)
	}
	return &r, nil
}

func approveDepositWalletForBuyOrders(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, creds *RelayerCredentials) (*RelayerResponse, error) {
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []WalletCall{
		{Target: PUSDAddress, Value: new(big.Int), Data: encodeApproveCalldataAmount(CTFExchange, maxU)},
		{Target: PUSDAddress, Value: new(big.Int), Data: encodeApproveCalldataAmount(NegRiskCTFExchange, maxU)},
	}
	return ExecuteDepositWalletBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func approveDepositWalletForSellOrders(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, creds *RelayerCredentials) (*RelayerResponse, error) {
	calls := []WalletCall{
		{Target: conditionalTokens, Value: new(big.Int), Data: encodeSetApprovalForAllCalldata(CTFExchange, true)},
		{Target: conditionalTokens, Value: new(big.Int), Data: encodeSetApprovalForAllCalldata(NegRiskCTFExchange, true)},
	}
	return ExecuteDepositWalletBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func transferFromDepositWallet(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress, assetAddress, recipientAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("transfer from deposit wallet: amount must be positive")
	}
	calls := []WalletCall{
		{Target: assetAddress, Value: new(big.Int), Data: encodeTransferCalldata(recipientAddress, amount)},
	}
	return ExecuteDepositWalletBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func wrapAndApproveDepositWallet(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("wrap and approve deposit wallet: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []WalletCall{
		{Target: USDCAddress, Value: new(big.Int), Data: encodeApproveCalldataAmount(collateralOnramp, maxU)},
		{Target: collateralOnramp, Value: new(big.Int), Data: encodeOnrampWrapCalldata(USDCAddress, depositWalletAddress, amount)},
		{Target: PUSDAddress, Value: new(big.Int), Data: encodeApproveCalldataAmount(CTFExchange, maxU)},
		{Target: PUSDAddress, Value: new(big.Int), Data: encodeApproveCalldataAmount(NegRiskCTFExchange, maxU)},
		{Target: conditionalTokens, Value: new(big.Int), Data: encodeSetApprovalForAllCalldata(CTFExchange, true)},
		{Target: conditionalTokens, Value: new(big.Int), Data: encodeSetApprovalForAllCalldata(NegRiskCTFExchange, true)},
	}
	return ExecuteDepositWalletBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func encodeApproveCalldataAmount(spender string, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("approve(address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(spender).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return data
}

func encodeTransferCalldata(to string, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("transfer(address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(to).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return data
}

func encodeOnrampWrapCalldata(asset, to string, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("wrap(address,address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(asset).Bytes())...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(to).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return data
}

func encodeSetApprovalForAllCalldata(operator string, approved bool) []byte {
	selector := ethutil.Keccak256([]byte("setApprovalForAll(address,bool)"))[:4]
	flag := big.NewInt(0)
	if approved {
		flag = big.NewInt(1)
	}
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(operator).Bytes())...)
	data = append(data, ethutil.PadTo32(flag.Bytes())...)
	return data
}

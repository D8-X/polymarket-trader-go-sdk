package polytrade

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/D8-X/polymarket-trader-go-sdk/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// maxUint256 is 2^256 - 1, used for unlimited token approvals.
var maxUint256 = func() *big.Int {
	n := new(big.Int).Lsh(big.NewInt(1), 256)
	return n.Sub(n, big.NewInt(1))
}()

// getSafeNonce fetches the current Safe nonce for the given EOA from the relayer
func getSafeNonce(ctx context.Context, eoaAddress string, creds *BuilderCredentials) (string, error) {
	endpoint := fmt.Sprintf("/nonce?address=%s&type=SAFE", eoaAddress)
	url := RelayerBaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("get safe nonce: build request: %w", err)
	}

	if creds != nil {
		applyBuilderHeaders(req, creds, http.MethodGet, endpoint, nil)
	}

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get safe nonce: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("get safe nonce: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   "GET /nonce",
			Body:       string(body),
		}
	}

	var result nonceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("get safe nonce: unmarshal response: %w", err)
	}
	return result.Nonce, nil
}

// encodeApproveCalldata ABI-encodes an ERC20 approve(address,uint256) call with MaxUint256.
func encodeApproveCalldata(spender string) string {
	// selector = keccak256("approve(address,uint256)")[:4]
	selector := ethutil.Keccak256([]byte("approve(address,uint256)"))[:4]
	spenderAddr := common.HexToAddress(spender)
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(spenderAddr.Bytes())...)
	data = append(data, ethutil.PadTo32(maxUint256.Bytes())...)
	return "0x" + hex.EncodeToString(data)
}

func encodeTransferCalldata(to string, amount *big.Int) string {
	selector := ethutil.Keccak256([]byte("transfer(address,uint256)"))[:4]
	toAddr := common.HexToAddress(to)
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(toAddr.Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return "0x" + hex.EncodeToString(data)
}

// encodeSetApprovalForAllCalldata ABI-encodes a setApprovalForAll(address,bool) call with true.
func encodeSetApprovalForAllCalldata(operator string) string {
	// selector = keccak256("setApprovalForAll(address,bool)")[:4]
	selector := ethutil.Keccak256([]byte("setApprovalForAll(address,bool)"))[:4]
	operatorAddr := common.HexToAddress(operator)
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(operatorAddr.Bytes())...)
	data = append(data, ethutil.PadTo32(big.NewInt(1).Bytes())...) // true
	return "0x" + hex.EncodeToString(data)
}

// encodeMultiSend encodes multiple SafeTransactions into a single multiSend(bytes) call.
func encodeMultiSend(txns []SafeTransaction) string {
	// Encode each inner transaction: encodePacked(uint8, address, uint256, uint256, bytes)
	var packed []byte
	for _, tx := range txns {
		dataBytes := common.FromHex(tx.Data)

		valueBig := new(big.Int)
		valueBig.SetString(tx.Value, 10)

		// 1 byte operation
		packed = append(packed, byte(tx.Operation))
		// 20 bytes address
		packed = append(packed, common.HexToAddress(tx.To).Bytes()...)
		// 32 bytes value
		packed = append(packed, ethutil.PadTo32(valueBig.Bytes())...)
		// 32 bytes data length
		dataLenBytes := make([]byte, 32)
		binary.BigEndian.PutUint64(dataLenBytes[24:], uint64(len(dataBytes)))
		packed = append(packed, dataLenBytes...)
		// N bytes data
		packed = append(packed, dataBytes...)
	}

	// Wrap in multiSend(bytes) ABI call
	// selector = keccak256("multiSend(bytes)")[:4]
	selector := ethutil.Keccak256([]byte("multiSend(bytes)"))[:4]

	// ABI encode dynamic bytes: offset (32) + length (32) + data (padded to 32-byte boundary)
	offset := ethutil.PadTo32(big.NewInt(32).Bytes())
	length := ethutil.PadTo32(big.NewInt(int64(len(packed))).Bytes())

	// Pad data to 32-byte boundary
	paddedData := packed
	if remainder := len(packed) % 32; remainder != 0 {
		paddedData = append(paddedData, make([]byte, 32-remainder)...)
	}

	result := make([]byte, 0, 4+32+32+len(paddedData))
	result = append(result, selector...)
	result = append(result, offset...)
	result = append(result, length...)
	result = append(result, paddedData...)

	return "0x" + hex.EncodeToString(result)
}

// ExecuteSafeTransaction executes one or more transactions through a Gnosis Safe
// via the Polymarket relayer.
func ExecuteSafeTransaction(ctx context.Context, eoaAddress, privateKeyHex string, txns []SafeTransaction, creds *BuilderCredentials) (*RelayerResponse, error) {
	if len(txns) == 0 {
		return nil, fmt.Errorf("execute safe tx: no transactions provided")
	}

	safeAddress := DeriveSafeAddress(eoaAddress)

	nonce, err := getSafeNonce(ctx, eoaAddress, creds)
	if err != nil {
		return nil, fmt.Errorf("execute safe tx: %w", err)
	}

	// Aggregate transactions
	var aggregated SafeTransaction
	if len(txns) == 1 {
		aggregated = txns[0]
	} else {
		//  encode all txns and wrap as a DelegateCall to the multisend contract
		aggregated = SafeTransaction{
			To:        SafeMultisend,
			Value:     "0",
			Data:      encodeMultiSend(txns),
			Operation: OperationDelegateCall,
		}
	}

	sig, err := signSafeTx(privateKeyHex, safeAddress, aggregated, nonce)
	if err != nil {
		return nil, fmt.Errorf("execute safe tx: %w", err)
	}

	reqBody := safeTxRequest{
		Type:        "SAFE",
		From:        eoaAddress,
		To:          aggregated.To,
		ProxyWallet: safeAddress,
		Data:        aggregated.Data,
		Nonce:       nonce,
		Signature:   sig,
		SignatureParams: safeTxSignatureParams{
			Operation:      fmt.Sprintf("%d", aggregated.Operation),
			SafeTxnGas:     "0",
			BaseGas:        "0",
			GasPrice:       "0",
			GasToken:       ZeroAddress,
			RefundReceiver: ZeroAddress,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("execute safe tx: marshal request: %w", err)
	}

	endpoint := "/submit"
	url := RelayerBaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("execute safe tx: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if creds != nil {
		applyBuilderHeaders(req, creds, http.MethodPost, endpoint, jsonBody)
	}

	httpClient := &http.Client{Timeout: CLOBTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute safe tx: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("execute safe tx: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   "POST " + endpoint,
			Body:       string(body),
		}
	}

	var relayResp RelayerResponse
	if err := json.Unmarshal(body, &relayResp); err != nil {
		return nil, fmt.Errorf("execute safe tx: unmarshal response: %w", err)
	}

	return &relayResp, nil
}

// TransferUSDCViaSafe transfers USDC from the sender's Safe to a recipient address
// via the Polymarket relayer (gasless). Amount is in raw units (6 decimals, e.g. 1_000_000 = 1 USDC).
func TransferUSDCViaSafe(ctx context.Context, eoaAddress, privateKeyHex string, to string, amount *big.Int, creds *BuilderCredentials) (*RelayerResponse, error) {
	tx := SafeTransaction{
		To:        USDCAddress,
		Value:     "0",
		Data:      encodeTransferCalldata(to, amount),
		Operation: OperationCall,
	}
	return ExecuteSafeTransaction(ctx, eoaAddress, privateKeyHex, []SafeTransaction{tx}, creds)
}

// ApproveSafeTokens submits a batched Safe transaction that approves USDC and
// CTFExchange for both the CTF exchange and the NegRisk CTF exchange.
func ApproveSafeTokens(ctx context.Context, eoaAddress, privateKeyHex, ctfExchangeAddress string, creds *BuilderCredentials) (*RelayerResponse, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("approve safe tokens: invalid private key: %w", err)
	}
	derivedEOA := crypto.PubkeyToAddress(pk.PublicKey).Hex()
	if eoaAddress == "" {
		eoaAddress = derivedEOA
	}

	txns := []SafeTransaction{
		//  USDC approve for CTF exchange
		{
			To:        USDCAddress,
			Value:     "0",
			Data:      encodeApproveCalldata(ctfExchangeAddress),
			Operation: OperationCall,
		},
		// USDC approve for NegRisk CTF exchange
		{
			To:        USDCAddress,
			Value:     "0",
			Data:      encodeApproveCalldata(NegRiskCTFExchange),
			Operation: OperationCall,
		},
		//  CTFExchange setApprovalForAll for CTF exchange
		{
			To:        CTFExchange,
			Value:     "0",
			Data:      encodeSetApprovalForAllCalldata(ctfExchangeAddress),
			Operation: OperationCall,
		},
		//CTFExchange setApprovalForAll for NegRisk CTF exchange
		{
			To:        CTFExchange,
			Value:     "0",
			Data:      encodeSetApprovalForAllCalldata(NegRiskCTFExchange),
			Operation: OperationCall,
		},
	}

	return ExecuteSafeTransaction(ctx, eoaAddress, privateKeyHex, txns, creds)
}

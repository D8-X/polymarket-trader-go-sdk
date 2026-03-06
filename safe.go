package polytrade

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func DerivePolymarketProxy(eoaAddress string) string {
	addrBytes := common.HexToAddress(eoaAddress).Bytes()
	salt := ethutil.Keccak256(addrBytes)

	factoryBytes := common.HexToAddress(ProxyFactory).Bytes()
	initHash := common.FromHex(ProxyInitCodeHash)

	data := make([]byte, 0, 1+20+32+32)
	data = append(data, 0xff)
	data = append(data, factoryBytes...)
	data = append(data, salt...)
	data = append(data, initHash...)

	hash := ethutil.Keccak256(data)
	return common.BytesToAddress(hash[12:]).Hex()
}

// DeriveSafeAddress computes the deterministic CREATE2 address for a Gnosis Safe
// given an EOA address, using the Polymarket Safe factory and init code hash.
func DeriveSafeAddress(eoaAddress string) string {
	// ABI-encoded salt: keccak256(abi.encode(address)) = keccak256(PadTo32(address))
	addrBytes := common.HexToAddress(eoaAddress).Bytes()
	salt := ethutil.Keccak256(ethutil.PadTo32(addrBytes))

	factoryBytes := common.HexToAddress(SafeFactory).Bytes()
	initHash := common.FromHex(SafeInitCodeHash)

	data := make([]byte, 0, 1+20+32+32)
	data = append(data, 0xff)
	data = append(data, factoryBytes...)
	data = append(data, salt...)
	data = append(data, initHash...)

	hash := ethutil.Keccak256(data)
	return common.BytesToAddress(hash[12:]).Hex()
}

// IsSafeDeployed checks if a Safe has been deployed at the given address via the relayer api
func IsSafeDeployed(ctx context.Context, safeAddress string) (bool, error) {
	url := fmt.Sprintf("%s/deployed?address=%s", RelayerBaseURL, safeAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("is safe deployed: build request: %w", err)
	}

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("is safe deployed: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("is safe deployed: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return false, &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   "GET /deployed",
			Body:       string(body),
		}
	}

	var result deployedResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("is safe deployed: unmarshal response: %w", err)
	}
	return result.Deployed, nil
}

// DeploySafe derives the deterministic Safe address for the given private key's EOA,
// checks if it is already deployed, and if not, submits a deployment request to the
// Polymarket relayer. Returns the Safe address.
func DeploySafe(ctx context.Context, privateKeyHex string, creds *BuilderCredentials) (string, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("deploy safe: invalid private key: %w", err)
	}
	eoaAddress := crypto.PubkeyToAddress(pk.PublicKey).Hex()
	safeAddr := DeriveSafeAddress(eoaAddress)

	deployed, err := IsSafeDeployed(ctx, safeAddr)
	if err != nil {
		return "", fmt.Errorf("deploy safe: check deployed: %w", err)
	}
	if deployed {
		return safeAddr, nil
	}

	sig, err := signSafeCreate(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("deploy safe: %w", err)
	}

	reqBody := safeCreateRequest{
		Type:        "SAFE-CREATE",
		From:        eoaAddress,
		To:          SafeFactory,
		ProxyWallet: safeAddr,
		Data:        "0x",
		Signature:   sig,
		SignatureParams: safeCreateSignatureParams{
			PaymentToken:    ZeroAddress,
			Payment:         "0",
			PaymentReceiver: ZeroAddress,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("deploy safe: marshal request: %w", err)
	}

	endpoint := "/submit"
	url := RelayerBaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("deploy safe: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if creds != nil {
		applyBuilderHeaders(req, creds, http.MethodPost, endpoint, jsonBody)
	}

	client := &http.Client{Timeout: CLOBTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("deploy safe: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("deploy safe: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   "POST " + endpoint,
			Body:       string(body),
		}
	}

	return safeAddr, nil
}

// EnsureSafeAddress first tries to look up an existing Safe for the EOA via the
// Safe Transaction Service. If none is found, it derives and deploys a new one
// via the Polymarket relayer
func EnsureSafeAddress(ctx context.Context, eoaAddress, privateKeyHex string, creds *BuilderCredentials) (string, error) {
	safeAddr, err := LookupSafeAddress(ctx, eoaAddress)
	if err == nil {
		return safeAddr, nil
	}
	return DeploySafe(ctx, privateKeyHex, creds)
}

func signBuilderHMAC(secret string, timestamp int64, method, path string, body []byte) string {
	message := strconv.FormatInt(timestamp, 10) + method + path
	if len(body) > 0 {
		message += string(body)
	}

	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		secretBytes = []byte(secret)
	}

	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(message))
	sig := base64.StdEncoding.EncodeToString(h.Sum(nil))

	sig = strings.ReplaceAll(sig, "+", "-")
	sig = strings.ReplaceAll(sig, "/", "_")
	return sig
}

func applyBuilderHeaders(req *http.Request, creds *BuilderCredentials, method, path string, body []byte) {
	ts := time.Now().Unix()
	req.Header.Set("POLY_BUILDER_API_KEY", creds.APIKey)
	req.Header.Set("POLY_BUILDER_PASSPHRASE", creds.Passphrase)
	req.Header.Set("POLY_BUILDER_TIMESTAMP", strconv.FormatInt(ts, 10))
	req.Header.Set("POLY_BUILDER_SIGNATURE", signBuilderHMAC(creds.Secret, ts, method, path, body))
}

func LookupSafeAddress(ctx context.Context, eoaAddress string) (string, error) {
	endpoint := fmt.Sprintf("/owners/%s/safes/", eoaAddress)
	url := SafeAPIBaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("lookup safe: build request: %w", err)
	}

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("lookup safe: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("lookup safe: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   "GET " + endpoint,
			Body:       string(body),
		}
	}

	var result struct {
		Safes []string `json:"safes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("lookup safe: unmarshal response: %w", err)
	}
	if len(result.Safes) == 0 {
		return "", fmt.Errorf("lookup safe: no Gnosis Safe found for address %s", eoaAddress)
	}
	return result.Safes[0], nil
}

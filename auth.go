package polytrade

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/d8x/polymarket-sports-sdk-go/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func DeriveL2Credentials(privateKeyHex string, chainID int) (*L2Credentials, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("derive credentials: invalid private key: %w", err)
	}
	address := crypto.PubkeyToAddress(pk.PublicKey)

	now := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := int64(0)

	domainSep := hashClobAuthDomain(chainID)
	structHash := hashClobAuthStruct(address.Hex(), now, nonce)
	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)

	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		return nil, fmt.Errorf("derive credentials: sign EIP-712: %w", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}

	endpoint := "/auth/derive-api-key"
	url := fmt.Sprintf("%s%s", ClobBaseURL, endpoint)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("derive credentials: build request: %w", err)
	}
	req.Header.Set("POLY_ADDRESS", address.Hex())
	req.Header.Set("POLY_SIGNATURE", "0x"+common.Bytes2Hex(sig))
	req.Header.Set("POLY_TIMESTAMP", now)
	req.Header.Set("POLY_NONCE", strconv.FormatInt(nonce, 10))

	client := &http.Client{Timeout: CLOBTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("derive credentials: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("derive credentials: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   "GET " + endpoint,
			Body:       string(body),
		}
	}

	var creds L2Credentials
	if err := json.Unmarshal(body, &creds); err != nil {
		return nil, fmt.Errorf("derive credentials: unmarshal response: %w", err)
	}
	creds.Address = address.Hex()

	return &creds, nil
}

func CreateL2Credentials(privateKeyHex string, chainID int) (*L2Credentials, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("create credentials: invalid private key: %w", err)
	}
	address := crypto.PubkeyToAddress(pk.PublicKey)

	now := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := int64(0)

	domainSep := hashClobAuthDomain(chainID)
	structHash := hashClobAuthStruct(address.Hex(), now, nonce)
	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)

	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		return nil, fmt.Errorf("create credentials: sign EIP-712: %w", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}

	endpoint := "/auth/api-key"
	url := fmt.Sprintf("%s%s", ClobBaseURL, endpoint)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create credentials: build request: %w", err)
	}
	req.Header.Set("POLY_ADDRESS", address.Hex())
	req.Header.Set("POLY_SIGNATURE", "0x"+common.Bytes2Hex(sig))
	req.Header.Set("POLY_TIMESTAMP", now)
	req.Header.Set("POLY_NONCE", strconv.FormatInt(nonce, 10))

	client := &http.Client{Timeout: CLOBTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create credentials: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("create credentials: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   "POST " + endpoint,
			Body:       string(body),
		}
	}

	var creds L2Credentials
	if err := json.Unmarshal(body, &creds); err != nil {
		return nil, fmt.Errorf("create credentials: unmarshal response: %w", err)
	}
	creds.Address = address.Hex()

	return &creds, nil
}

func SignL2Request(creds *L2Credentials, method, path string, body []byte) (*L2Headers, error) {
	if creds == nil {
		return nil, fmt.Errorf("sign request: credentials are nil")
	}
	if creds.Secret == "" {
		return nil, fmt.Errorf("sign request: credentials secret is empty")
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)

	message := ts + method + path
	if len(body) > 0 {
		message += string(body)
	}

	secretBytes, err := base64.URLEncoding.DecodeString(creds.Secret)
	if err != nil {
		secretBytes, err = base64.RawURLEncoding.DecodeString(creds.Secret)
		if err != nil {
			secretBytes, err = base64.StdEncoding.DecodeString(creds.Secret)
			if err != nil {
				return nil, fmt.Errorf("sign request: decode secret: %w", err)
			}
		}
	}

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(message))
	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	return &L2Headers{
		Address:    creds.Address,
		APIKey:     creds.APIKey,
		Passphrase: creds.Passphrase,
		Signature:  signature,
		Timestamp:  ts,
	}, nil
}

func ApplyL2Headers(req *http.Request, h *L2Headers) {
	req.Header.Set("POLY_ADDRESS", h.Address)
	req.Header.Set("POLY_API_KEY", h.APIKey)
	req.Header.Set("POLY_PASSPHRASE", h.Passphrase)
	req.Header.Set("POLY_SIGNATURE", h.Signature)
	req.Header.Set("POLY_TIMESTAMP", h.Timestamp)
}

package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func DeriveCredentials(privateKeyHex string, chainID int) (*types.L2Credentials, error) {
	return fetchCredentials(privateKeyHex, chainID, http.MethodGet, "/auth/derive-api-key")
}

func CreateCredentials(privateKeyHex string, chainID int) (*types.L2Credentials, error) {
	return fetchCredentials(privateKeyHex, chainID, http.MethodPost, "/auth/api-key")
}

func fetchCredentials(privateKeyHex string, chainID int, method, endpoint string) (*types.L2Credentials, error) {
	pk, err := crypto.HexToECDSA(ethutil.StripHexPrefix(privateKeyHex))
	if err != nil {
		return nil, fmt.Errorf("auth credentials: invalid private key: %w", err)
	}
	address := crypto.PubkeyToAddress(pk.PublicKey)

	now := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := int64(0)

	domainSep := HashClobAuthDomain(chainID)
	structHash := HashClobAuthStruct(address.Hex(), now, nonce)
	digest := ethutil.Keccak256Pack([]byte{0x19, 0x01}, domainSep, structHash)

	sig, err := crypto.Sign(digest, pk)
	if err != nil {
		return nil, fmt.Errorf("auth credentials: sign EIP-712: %w", err)
	}
	if sig[64] < 27 {
		sig[64] += 27
	}

	url := consts.ClobBaseURL + endpoint
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("auth credentials: build request: %w", err)
	}
	req.Header.Set("POLY_ADDRESS", address.Hex())
	req.Header.Set("POLY_SIGNATURE", "0x"+common.Bytes2Hex(sig))
	req.Header.Set("POLY_TIMESTAMP", now)
	req.Header.Set("POLY_NONCE", strconv.FormatInt(nonce, 10))

	client := &http.Client{Timeout: consts.CLOBTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth credentials: http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("auth credentials: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &types.APIError{
			StatusCode: resp.StatusCode,
			Endpoint:   method + " " + endpoint,
			Body:       string(body),
		}
	}

	var creds types.L2Credentials
	if err := json.Unmarshal(body, &creds); err != nil {
		return nil, fmt.Errorf("auth credentials: unmarshal response: %w", err)
	}
	creds.Address = address.Hex()
	return &creds, nil
}

func SignRequest(creds *types.L2Credentials, method, path string, body []byte) (*types.L2Headers, error) {
	if creds == nil {
		return nil, fmt.Errorf("sign request: credentials are nil")
	}
	if creds.Secret == "" {
		return nil, fmt.Errorf("sign request: credentials secret is empty")
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := SignMessage(creds.Secret, ts, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}
	return &types.L2Headers{
		Address:    creds.Address,
		APIKey:     creds.APIKey,
		Passphrase: creds.Passphrase,
		Signature:  signature,
		Timestamp:  ts,
	}, nil
}

func SignMessage(secret, ts, method, path string, body []byte) (string, error) {
	message := ts + method + path
	if len(body) > 0 {
		message += string(body)
	}

	secretBytes, err := base64.URLEncoding.DecodeString(secret)
	if err != nil {
		secretBytes, err = base64.RawURLEncoding.DecodeString(secret)
		if err != nil {
			secretBytes, err = base64.StdEncoding.DecodeString(secret)
			if err != nil {
				return "", fmt.Errorf("decode secret: %w", err)
			}
		}
	}

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(message))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func ApplyHeaders(req *http.Request, h *types.L2Headers) {
	req.Header.Set("POLY_ADDRESS", h.Address)
	req.Header.Set("POLY_API_KEY", h.APIKey)
	req.Header.Set("POLY_PASSPHRASE", h.Passphrase)
	req.Header.Set("POLY_SIGNATURE", h.Signature)
	req.Header.Set("POLY_TIMESTAMP", h.Timestamp)
}

func HashClobAuthDomain(chainID int) []byte {
	typeHash := ethutil.Keccak256([]byte(consts.EIP712AuthDomainType))
	nameHash := ethutil.Keccak256([]byte(consts.EIP712AuthDomainName))
	versionHash := ethutil.Keccak256([]byte(consts.EIP712AuthVersion))
	chainIDBytes := ethutil.PadTo32(new(big.Int).SetInt64(int64(chainID)).Bytes())

	return ethutil.Keccak256(append(append(append(ethutil.PadTo32(typeHash), ethutil.PadTo32(nameHash)...), ethutil.PadTo32(versionHash)...), chainIDBytes...))
}

func HashClobAuthStruct(address, timestamp string, nonce int64) []byte {
	typeHash := ethutil.Keccak256([]byte(consts.EIP712ClobAuthType))
	addrBig := new(big.Int)
	if len(address) > 2 {
		addrBig.SetString(address[2:], 16)
	}
	tsHash := ethutil.Keccak256([]byte(timestamp))
	nonceBig := new(big.Int).SetInt64(nonce)
	msgHash := ethutil.Keccak256([]byte(consts.EIP712AuthMessage))

	encoded := make([]byte, 0, 160)
	encoded = append(encoded, ethutil.PadTo32(typeHash)...)
	encoded = append(encoded, ethutil.PadTo32(addrBig.Bytes())...)
	encoded = append(encoded, ethutil.PadTo32(tsHash)...)
	encoded = append(encoded, ethutil.PadTo32(nonceBig.Bytes())...)
	encoded = append(encoded, ethutil.PadTo32(msgHash)...)

	return ethutil.Keccak256(encoded)
}

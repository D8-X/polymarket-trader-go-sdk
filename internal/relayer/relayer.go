package relayer

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

const (
	StateNew       = "STATE_NEW"
	StateExecuted  = "STATE_EXECUTED"
	StateMined     = "STATE_MINED"
	StateConfirmed = "STATE_CONFIRMED"
	StateInvalid   = "STATE_INVALID"
	StateFailed    = "STATE_FAILED"
)

func SignHMAC(secret string, timestamp int64, method, path string, body []byte) string {
	message := strconv.FormatInt(timestamp, 10) + method + path
	if len(body) > 0 {
		message += string(body)
	}
	secretBytes, err := base64.URLEncoding.DecodeString(secret)
	if err != nil {
		secretBytes, err = base64.RawURLEncoding.DecodeString(secret)
		if err != nil {
			secretBytes, err = base64.StdEncoding.DecodeString(secret)
			if err != nil {
				secretBytes = []byte(secret)
			}
		}
	}
	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(message))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func ApplyHeaders(req *http.Request, creds *models.RelayerCredentials, method, path string, body []byte) {
	ts := time.Now().Unix()
	req.Header.Set("POLY_BUILDER_API_KEY", creds.APIKey)
	req.Header.Set("POLY_BUILDER_PASSPHRASE", creds.Passphrase)
	req.Header.Set("POLY_BUILDER_TIMESTAMP", strconv.FormatInt(ts, 10))
	req.Header.Set("POLY_BUILDER_SIGNATURE", SignHMAC(creds.Secret, ts, method, path, body))
}

func GetTransaction(ctx context.Context, transactionID string) (*models.RelayerTransaction, error) {
	endpoint := fmt.Sprintf("/transaction?id=%s", transactionID)
	url := consts.RelayerBaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("get relayer tx: build request: %w", err)
	}
	client := &http.Client{Timeout: consts.DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get relayer tx: http request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("get relayer tx: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &models.APIError{StatusCode: resp.StatusCode, Endpoint: "GET /transaction", Body: string(body)}
	}
	var txns []models.RelayerTransaction
	if err := json.Unmarshal(body, &txns); err != nil {
		return nil, fmt.Errorf("get relayer tx: unmarshal response: %w", err)
	}
	if len(txns) == 0 {
		return nil, fmt.Errorf("get relayer tx: transaction %s not found", transactionID)
	}
	return &txns[0], nil
}

func WaitForTransaction(ctx context.Context, transactionID string) (*models.RelayerTransaction, error) {
	pollInterval := 2 * time.Second
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		tx, err := GetTransaction(ctx, transactionID)
		if err != nil {
			return nil, fmt.Errorf("wait for relayer tx: %w", err)
		}
		switch tx.State {
		case StateMined, StateConfirmed:
			return tx, nil
		case StateFailed, StateInvalid:
			return tx, fmt.Errorf("wait for relayer tx: transaction %s: %s", transactionID, tx.State)
		}
		select {
		case <-ctx.Done():
			return tx, fmt.Errorf("wait for relayer tx: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

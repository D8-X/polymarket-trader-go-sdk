package polytrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/D8-X/polymarket-trader-go-sdk/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
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

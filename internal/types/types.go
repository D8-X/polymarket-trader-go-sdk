package types

import (
	"fmt"
	"math/big"
)

type APIError struct {
	StatusCode int
	Endpoint   string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("polymarket api %s returned status %d: %s", e.Endpoint, e.StatusCode, e.Body)
}

type RelayerCredentials struct {
	APIKey     string
	Secret     string
	Passphrase string
}

type RelayerResponse struct {
	TransactionID   string `json:"transactionID"`
	State           string `json:"state"`
	TransactionHash string `json:"transactionHash"`
}

type RelayerTransaction struct {
	TransactionID   string `json:"transactionID"`
	TransactionHash string `json:"transactionHash"`
	From            string `json:"from"`
	To              string `json:"to"`
	ProxyAddress    string `json:"proxyAddress"`
	Data            string `json:"data"`
	Nonce           string `json:"nonce"`
	Value           string `json:"value"`
	State           string `json:"state"`
	Type            string `json:"type"`
	Metadata        string `json:"metadata"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type WalletCall struct {
	Target string
	Value  *big.Int
	Data   []byte
}

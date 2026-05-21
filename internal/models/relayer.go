package models

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

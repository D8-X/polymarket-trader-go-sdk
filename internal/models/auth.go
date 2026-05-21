package models

type L2Credentials struct {
	Address    string `json:"address,omitempty"`
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

type L2Headers struct {
	Address    string
	APIKey     string
	Passphrase string
	Signature  string
	Timestamp  string
}

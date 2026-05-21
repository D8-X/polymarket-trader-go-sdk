package models

import "math/big"

type WalletCall struct {
	Target string
	Value  *big.Int
	Data   []byte
}

type NonceResponse struct {
	Nonce string `json:"nonce"`
}

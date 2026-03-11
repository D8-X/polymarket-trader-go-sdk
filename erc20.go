package polytrade

import (
	"context"
	"math/big"
	"strings"
)

// USDCBalanceOf returns the USDC collateral balance for the wallet
// associated with the given private key, via the Polymarket CLOB API.
// Returns the balance in raw USDC units (6 decimals).
func USDCBalanceOf(ctx context.Context, privateKeyHex string) (*big.Int, error) {
	pk := strings.TrimPrefix(privateKeyHex, "0x")
	creds, err := DeriveL2Credentials(pk, PolygonChainID)
	if err != nil {
		return big.NewInt(0), nil
	}
	clob := NewCLOBClient()
	resp, err := clob.GetBalanceAllowance(ctx, "COLLATERAL", "", SignatureTypeGnosisSafe, creds)
	if err != nil {
		return nil, err
	}
	return parseUSDCBalance(resp.Balance), nil
}

// RefreshUSDCBalance triggers Polymarket to re-scan the on-chain USDC balance
// for the wallet associated with the given private key and deposit it into the
// exchange. Call this after transferring USDC to a Safe so the funds become
// available for trading.
func RefreshUSDCBalance(ctx context.Context, privateKeyHex string) error {
	pk := strings.TrimPrefix(privateKeyHex, "0x")
	creds, err := DeriveL2Credentials(pk, PolygonChainID)
	if err != nil {
		return err
	}
	clob := NewCLOBClient()
	return clob.UpdateBalanceAllowance(ctx, "COLLATERAL", "", SignatureTypeGnosisSafe, creds)
}

func parseUSDCBalance(s string) *big.Int {
	bal, ok := new(big.Int).SetString(s, 10)
	if ok {
		return bal
	}
	f, _, err := new(big.Float).Parse(s, 10)
	if err != nil {
		return big.NewInt(0)
	}
	f.Mul(f, new(big.Float).SetFloat64(AmountScale))
	raw, _ := f.Int(nil)
	return raw
}

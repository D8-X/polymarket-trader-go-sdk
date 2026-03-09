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

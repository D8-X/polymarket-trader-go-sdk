package polytrade

import (
	"context"
	"math/big"
)

// collateralBalanceOf returns the collateral balance available for trading on
// the Polymarket CLOB for the wallet identified by the provided L2 credentials.
// Returns the balance in raw units (6 decimals). Post-V2 the underlying asset
// is pUSD; pre-V2 it was USDC.e.
func collateralBalanceOf(ctx context.Context, creds *L2Credentials) (*big.Int, error) {
	clob := NewCLOBClient()
	resp, err := clob.GetBalanceAllowance(ctx, "COLLATERAL", "", SignatureTypeGnosisSafe, creds)
	if err != nil {
		return nil, err
	}
	return parseCollateralBalance(resp.Balance), nil
}

// refreshCollateralBalance triggers Polymarket to re-scan the on-chain
// collateral balance for the Safe associated with the provided L2 credentials
// and deposit it into the exchange. Call this after transferring collateral
// to a Safe so the funds become available for trading.
func refreshCollateralBalance(ctx context.Context, creds *L2Credentials) error {
	clob := NewCLOBClient()
	return clob.UpdateBalanceAllowance(ctx, "COLLATERAL", "", SignatureTypeGnosisSafe, creds)
}

func parseCollateralBalance(s string) *big.Int {
	bal, ok := new(big.Int).SetString(s, 10)
	if ok {
		return bal
	}
	f, _, err := new(big.Float).Parse(s, 10)
	if err != nil {
		return big.NewInt(0)
	}
	f.Mul(f, new(big.Float).SetFloat64(amountScale))
	raw, _ := f.Int(nil)
	return raw
}

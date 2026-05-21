package wallet

import (
	"context"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/clob"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/types"
)

func CollateralBalance(ctx context.Context, creds *types.L2Credentials) (*big.Int, error) {
	c := clob.NewClient()
	resp, err := c.GetBalanceAllowance(ctx, "COLLATERAL", "", creds)
	if err != nil {
		return nil, err
	}
	return parseCollateralBalance(resp.Balance), nil
}

func RefreshCollateralBalance(ctx context.Context, creds *types.L2Credentials) error {
	c := clob.NewClient()
	return c.UpdateBalanceAllowance(ctx, "COLLATERAL", "", creds)
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
	f.Mul(f, new(big.Float).SetFloat64(consts.AmountScale))
	raw, _ := f.Int(nil)
	return raw
}

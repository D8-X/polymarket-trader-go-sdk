package polytrade

import (
	"context"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
)

func (c *Client) SplitPosition(ctx context.Context, conditionID string, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("split position: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []WalletCall{
		{Target: PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.ConditionalTokens, maxU)},
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeSplitPositionCalldata(PUSDAddress, conditionID, []int64{1, 2}, amount)},
	}
	return ExecuteDepositWalletBatch(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, calls, 0, c.cfg.RelayerCreds)
}

func (c *Client) MergePositions(ctx context.Context, conditionID string, amount *big.Int) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("merge positions: amount must be positive")
	}
	calls := []WalletCall{
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeMergePositionsCalldata(PUSDAddress, conditionID, []int64{1, 2}, amount)},
	}
	return ExecuteDepositWalletBatch(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, calls, 0, c.cfg.RelayerCreds)
}

func (c *Client) RedeemPositions(ctx context.Context, conditionID string) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	calls := []WalletCall{
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeRedeemPositionsCalldata(PUSDAddress, conditionID, []int64{1, 2})},
	}
	return ExecuteDepositWalletBatch(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, calls, 0, c.cfg.RelayerCreds)
}

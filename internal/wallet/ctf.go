package wallet

import (
	"context"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
)

func SplitPosition(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, amount *big.Int, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("split position: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []models.WalletCall{
		{Target: consts.PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.ConditionalTokens, maxU)},
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeSplitPositionCalldata(consts.PUSDAddress, conditionID, []int64{1, 2}, amount)},
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

func MergePositions(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, amount *big.Int, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("merge positions: amount must be positive")
	}
	calls := []models.WalletCall{
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeMergePositionsCalldata(consts.PUSDAddress, conditionID, []int64{1, 2}, amount)},
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

func RedeemPositions(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	calls := []models.WalletCall{
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: onchain.EncodeRedeemPositionsCalldata(consts.PUSDAddress, conditionID, []int64{1, 2})},
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

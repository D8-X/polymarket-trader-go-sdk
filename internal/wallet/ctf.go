package wallet

import (
	"context"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
)

func ctfTarget(negRisk bool) string {
	if negRisk {
		return consts.NegRiskCollateralAdapter
	}
	return consts.CollateralAdapter
}

func SplitPosition(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, amount *big.Int, negRisk bool, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("split position: amount must be positive")
	}
	target := ctfTarget(negRisk)
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []models.WalletCall{
		{Target: consts.PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(target, maxU)},
		{Target: target, Value: new(big.Int), Data: onchain.EncodeSplitPositionCalldata(consts.PUSDAddress, conditionID, []int64{1, 2}, amount)},
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

func MergePositions(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, amount *big.Int, negRisk bool, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("merge positions: amount must be positive")
	}
	target := ctfTarget(negRisk)
	calls := []models.WalletCall{
		{Target: target, Value: new(big.Int), Data: onchain.EncodeMergePositionsCalldata(consts.PUSDAddress, conditionID, []int64{1, 2}, amount)},
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

func RedeemPositions(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, negRisk bool, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	target := ctfTarget(negRisk)
	calls := []models.WalletCall{
		{Target: target, Value: new(big.Int), Data: onchain.EncodeRedeemPositionsCalldata(consts.PUSDAddress, conditionID, []int64{1, 2})},
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

func SplitPositionV2(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, amount *big.Int, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("split position: amount must be positive")
	}
	word, err := ethutil.V2ConditionWord(conditionID)
	if err != nil {
		return nil, fmt.Errorf("split position: %w", err)
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []models.WalletCall{
		{Target: consts.PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.ProtocolV2Router, maxU)},
		{Target: consts.ProtocolV2Router, Value: new(big.Int), Data: onchain.EncodeRouterSplitCalldata(word, amount)},
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

func MergePositionsV2(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, amount *big.Int, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("merge positions: amount must be positive")
	}
	word, err := ethutil.V2ConditionWord(conditionID)
	if err != nil {
		return nil, fmt.Errorf("merge positions: %w", err)
	}
	calls := []models.WalletCall{
		{Target: consts.PositionManager, Value: new(big.Int), Data: onchain.EncodeSetApprovalForAllCalldata(consts.ProtocolV2Router, true)},
		{Target: consts.ProtocolV2Router, Value: new(big.Int), Data: onchain.EncodeRouterMergeCalldata(word, amount)},
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

// RedeemPositionsV2 redeems the YES and NO balances given, skipping empty ones.
func RedeemPositionsV2(ctx context.Context, eoa, privateKeyHex, depositWallet, conditionID string, balances [2]*big.Int, creds *models.RelayerCredentials) (*models.RelayerResponse, error) {
	word, err := ethutil.V2ConditionWord(conditionID)
	if err != nil {
		return nil, fmt.Errorf("redeem positions: %w", err)
	}
	calls := []models.WalletCall{
		{Target: consts.PositionManager, Value: new(big.Int), Data: onchain.EncodeSetApprovalForAllCalldata(consts.ProtocolV2Router, true)},
	}
	for i, b := range balances {
		if b != nil && b.Sign() > 0 {
			calls = append(calls, models.WalletCall{Target: consts.ProtocolV2Router, Value: new(big.Int), Data: onchain.EncodeRouterRedeemCalldata(word, int64(i), b)})
		}
	}
	if len(calls) == 1 {
		return nil, fmt.Errorf("redeem positions: no balance to redeem for condition %s", conditionID)
	}
	return ExecuteBatch(ctx, eoa, privateKeyHex, depositWallet, calls, 0, creds)
}

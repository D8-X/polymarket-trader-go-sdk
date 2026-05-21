package polytrade

import (
	"context"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
)

func wrapToPUSD(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("wrap to pusd: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []WalletCall{
		{Target: USDCAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.CollateralOnramp, maxU)},
		{Target: consts.CollateralOnramp, Value: new(big.Int), Data: onchain.EncodeOnrampWrapCalldata(USDCAddress, depositWalletAddress, amount)},
	}
	return ExecuteDepositWalletBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func unwrapToUSDC(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("unwrap to usdc: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []WalletCall{
		{Target: PUSDAddress, Value: new(big.Int), Data: onchain.EncodeApproveCalldata(consts.CollateralOfframp, maxU)},
		{Target: consts.CollateralOfframp, Value: new(big.Int), Data: onchain.EncodeOfframpUnwrapCalldata(USDCAddress, depositWalletAddress, amount)},
	}
	return ExecuteDepositWalletBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

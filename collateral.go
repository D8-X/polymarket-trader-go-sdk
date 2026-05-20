package polytrade

import (
	"context"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
)

func WrapToPUSD(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("wrap to pusd: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []WalletCall{
		{Target: USDCAddress, Value: new(big.Int), Data: encodeApproveCalldataAmount(CollateralOnramp, maxU)},
		{Target: CollateralOnramp, Value: new(big.Int), Data: encodeOnrampWrapCalldata(USDCAddress, depositWalletAddress, amount)},
	}
	return ExecuteDepositWalletBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func UnwrapToUSDC(ctx context.Context, eoaAddress, privateKeyHex, depositWalletAddress string, amount *big.Int, creds *RelayerCredentials) (*RelayerResponse, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, fmt.Errorf("unwrap to usdc: amount must be positive")
	}
	maxU := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	calls := []WalletCall{
		{Target: PUSDAddress, Value: new(big.Int), Data: encodeApproveCalldataAmount(CollateralOfframp, maxU)},
		{Target: CollateralOfframp, Value: new(big.Int), Data: encodeOfframpUnwrapCalldata(USDCAddress, depositWalletAddress, amount)},
	}
	return ExecuteDepositWalletBatch(ctx, eoaAddress, privateKeyHex, depositWalletAddress, calls, 0, creds)
}

func encodeOfframpUnwrapCalldata(asset, to string, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("unwrap(address,address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(asset).Bytes())...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(to).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return data
}

package polytrade

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
)

func encodeWrapCalldata(selector []byte, asset, to string, amount *big.Int) string {
	data := make([]byte, 0, 4+32+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(asset).Bytes())...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(to).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	return "0x" + hex.EncodeToString(data)
}

// WrapToPUSD wraps USDC.e held in the Safe into pUSD via the Collateral Onramp.
// The minted pUSD lands back in the Safe.
func WrapToPUSD(ctx context.Context, eoaAddress, privateKeyHex string, amount *big.Int, creds *BuilderCredentials) (*RelayerResponse, error) {
	safeAddress := DeriveSafeAddress(eoaAddress)
	selector := ethutil.Keccak256([]byte("wrap(address,address,uint256)"))[:4]

	txns := []SafeTransaction{
		{
			To:        USDCAddress,
			Value:     "0",
			Data:      encodeApproveCalldata(CollateralOnramp),
			Operation: OperationCall,
		},
		{
			To:        CollateralOnramp,
			Value:     "0",
			Data:      encodeWrapCalldata(selector, USDCAddress, safeAddress, amount),
			Operation: OperationCall,
		},
	}
	return ExecuteSafeTransaction(ctx, eoaAddress, privateKeyHex, txns, creds)
}

// UnwrapToUSDC unwraps pUSD held in the Safe back to USDC.e via the Collateral Offramp.
// The recovered USDC.e lands back in the Safe.
func UnwrapToUSDC(ctx context.Context, eoaAddress, privateKeyHex string, amount *big.Int, creds *BuilderCredentials) (*RelayerResponse, error) {
	safeAddress := DeriveSafeAddress(eoaAddress)
	selector := ethutil.Keccak256([]byte("unwrap(address,address,uint256)"))[:4]

	txns := []SafeTransaction{
		{
			To:        PUSDAddress,
			Value:     "0",
			Data:      encodeApproveCalldata(CollateralOfframp),
			Operation: OperationCall,
		},
		{
			To:        CollateralOfframp,
			Value:     "0",
			Data:      encodeWrapCalldata(selector, USDCAddress, safeAddress, amount),
			Operation: OperationCall,
		},
	}
	return ExecuteSafeTransaction(ctx, eoaAddress, privateKeyHex, txns, creds)
}

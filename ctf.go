package polytrade

import (
	"context"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
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
		{Target: PUSDAddress, Value: new(big.Int), Data: encodeApproveCalldataAmount(consts.ConditionalTokens, maxU)},
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: encodeSplitPositionCalldata(PUSDAddress, conditionID, []int64{1, 2}, amount)},
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
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: encodeMergePositionsCalldata(PUSDAddress, conditionID, []int64{1, 2}, amount)},
	}
	return ExecuteDepositWalletBatch(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, calls, 0, c.cfg.RelayerCreds)
}

func (c *Client) RedeemPositions(ctx context.Context, conditionID string) (*RelayerResponse, error) {
	dw, err := c.requireDepositWalletOps()
	if err != nil {
		return nil, err
	}
	calls := []WalletCall{
		{Target: consts.ConditionalTokens, Value: new(big.Int), Data: encodeRedeemPositionsCalldata(PUSDAddress, conditionID, []int64{1, 2})},
	}
	return ExecuteDepositWalletBatch(ctx, c.eoa, c.cfg.PrivateKeyHex, dw, calls, 0, c.cfg.RelayerCreds)
}

func encodeSplitPositionCalldata(collateral, conditionID string, partition []int64, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("splitPosition(address,bytes32,bytes32,uint256[],uint256)"))[:4]
	data := make([]byte, 0, 4+5*32+32+len(partition)*32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(collateral).Bytes())...)
	data = append(data, make([]byte, 32)...) // parentCollectionId = bytes32(0)
	data = append(data, parseBytes32Hex(conditionID)...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(5*32)).Bytes())...) // offset to partition data
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(len(partition))).Bytes())...)
	for _, p := range partition {
		data = append(data, ethutil.PadTo32(big.NewInt(p).Bytes())...)
	}
	return data
}

func encodeMergePositionsCalldata(collateral, conditionID string, partition []int64, amount *big.Int) []byte {
	selector := ethutil.Keccak256([]byte("mergePositions(address,bytes32,bytes32,uint256[],uint256)"))[:4]
	data := make([]byte, 0, 4+5*32+32+len(partition)*32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(collateral).Bytes())...)
	data = append(data, make([]byte, 32)...)
	data = append(data, parseBytes32Hex(conditionID)...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(5*32)).Bytes())...)
	data = append(data, ethutil.PadTo32(amount.Bytes())...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(len(partition))).Bytes())...)
	for _, p := range partition {
		data = append(data, ethutil.PadTo32(big.NewInt(p).Bytes())...)
	}
	return data
}

func encodeRedeemPositionsCalldata(collateral, conditionID string, indexSets []int64) []byte {
	selector := ethutil.Keccak256([]byte("redeemPositions(address,bytes32,bytes32,uint256[])"))[:4]
	data := make([]byte, 0, 4+4*32+32+len(indexSets)*32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(collateral).Bytes())...)
	data = append(data, make([]byte, 32)...)
	data = append(data, parseBytes32Hex(conditionID)...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(4*32)).Bytes())...)
	data = append(data, ethutil.PadTo32(big.NewInt(int64(len(indexSets))).Bytes())...)
	for _, i := range indexSets {
		data = append(data, ethutil.PadTo32(big.NewInt(i).Bytes())...)
	}
	return data
}

func parseBytes32Hex(s string) []byte {
	stripped := ethutil.StripHexPrefix(s)
	b := common.FromHex("0x" + stripped)
	out := make([]byte, 32)
	if len(b) >= 32 {
		copy(out, b[len(b)-32:])
	} else {
		copy(out[32-len(b):], b)
	}
	return out
}

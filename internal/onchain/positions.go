package onchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type ContractCaller interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

func GetOutcomeTokenBalance(ctx context.Context, eth ContractCaller, ownerAddress, tokenID string) (*big.Int, error) {
	if eth == nil {
		return nil, fmt.Errorf("get outcome token balance: nil ContractCaller")
	}
	tokenIDBig, ok := new(big.Int).SetString(tokenID, 10)
	if !ok {
		return nil, fmt.Errorf("get outcome token balance: invalid tokenID %q", tokenID)
	}
	selector := ethutil.Keccak256([]byte("balanceOf(address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(ownerAddress).Bytes())...)
	data = append(data, ethutil.PadTo32(tokenIDBig.Bytes())...)
	addr := common.HexToAddress(consts.ConditionalTokens)
	msg := ethereum.CallMsg{To: &addr, Data: data}
	result, err := eth.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("get outcome token balance: %w", err)
	}
	return new(big.Int).SetBytes(result), nil
}

func RawBalanceToSize(balance *big.Int) float64 {
	if balance == nil || balance.Sign() == 0 {
		return 0
	}
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(consts.AmountScale)).Float64()
	return f
}

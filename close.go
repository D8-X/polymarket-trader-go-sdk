package polytrade

import (
	"context"
	"fmt"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type ContractCaller interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

type ClosePositionOpts struct {
	OrderType string
	TickSize  string
	PostOnly  bool
	DeferExec bool
}

func GetOutcomeTokenBalance(ctx context.Context, eth ContractCaller, ownerAddress, tokenID string) (*big.Int, error) {
	tokenIDBig, ok := new(big.Int).SetString(tokenID, 10)
	if !ok {
		return nil, fmt.Errorf("get outcome token balance: invalid tokenID %q", tokenID)
	}
	selector := ethutil.Keccak256([]byte("balanceOf(address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32)
	data = append(data, selector...)
	data = append(data, ethutil.PadTo32(common.HexToAddress(ownerAddress).Bytes())...)
	data = append(data, ethutil.PadTo32(tokenIDBig.Bytes())...)
	addr := common.HexToAddress(ConditionalTokens)
	msg := ethereum.CallMsg{To: &addr, Data: data}
	result, err := eth.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("get outcome token balance: %w", err)
	}
	return new(big.Int).SetBytes(result), nil
}

func rawBalanceToSize(balance *big.Int) float64 {
	if balance == nil || balance.Sign() == 0 {
		return 0
	}
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(AmountScale)).Float64()
	return f
}

func (c *CLOBClient) ClosePosition(ctx context.Context, eth ContractCaller, builder *OrderBuilder, tokenID string, price float64, creds *L2Credentials, opts ClosePositionOpts) (*PlaceOrderResponse, error) {
	if builder == nil {
		return nil, fmt.Errorf("close position: nil builder")
	}
	if eth == nil {
		return nil, fmt.Errorf("close position: nil contract caller")
	}
	balance, err := GetOutcomeTokenBalance(ctx, eth, builder.makerAddress, tokenID)
	if err != nil {
		return nil, fmt.Errorf("close position: %w", err)
	}
	if balance.Sign() <= 0 {
		return nil, fmt.Errorf("close position: no position to close for tokenID %s", tokenID)
	}
	size := rawBalanceToSize(balance)

	orderType := opts.OrderType
	if orderType == "" {
		orderType = OrderTypeFOK
	}
	tickSize := opts.TickSize
	if tickSize == "" {
		tickSize = "0.01"
	}
	signed, err := builder.PrepareAndSign(tokenID, SELL, orderType, price, size, creds.APIKey, OrderOpts{
		TickSize:  tickSize,
		PostOnly:  opts.PostOnly,
		DeferExec: opts.DeferExec,
	})
	if err != nil {
		return nil, fmt.Errorf("close position: prepare: %w", err)
	}
	return c.PlaceOrder(ctx, signed, creds)
}

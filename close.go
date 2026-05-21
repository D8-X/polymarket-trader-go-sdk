package polytrade

import (
	"context"
	"math/big"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/onchain"
)

type ContractCaller = onchain.ContractCaller

type ClosePositionOpts struct {
	OrderType string
	TickSize  string
	PostOnly  bool
	DeferExec bool
}

func GetOutcomeTokenBalance(ctx context.Context, eth ContractCaller, ownerAddress, tokenID string) (*big.Int, error) {
	return onchain.GetOutcomeTokenBalance(ctx, eth, ownerAddress, tokenID)
}

func rawBalanceToSize(balance *big.Int) float64 {
	return onchain.RawBalanceToSize(balance)
}

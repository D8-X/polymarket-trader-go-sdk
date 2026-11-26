package onchain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ethutil"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type ReceiptFetcher interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
}

var (
	payoutRedemptionTopic = common.BytesToHash(ethutil.Keccak256([]byte("PayoutRedemption(address,address,bytes32,bytes32,uint256[],uint256)")))
	erc20TransferTopic    = common.BytesToHash(ethutil.Keccak256([]byte("Transfer(address,address,uint256)")))
)

// PayoutFromRedeemReceipt sums the CTF PayoutRedemption logs for redeemer. Protocol V2
// redemptions emit none, so it then falls back to the pUSD transferred to redeemer.

func PayoutFromRedeemReceipt(ctx context.Context, eth ReceiptFetcher, txHash, redeemer string) (*big.Int, error) {
	if !strings.HasPrefix(txHash, "0x") {
		txHash = "0x" + txHash
	}
	receipt, err := eth.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, fmt.Errorf("redeem payout: fetch receipt: %w", err)
	}
	want := common.HexToAddress(redeemer)
	payout := new(big.Int)
	matched := false
	for _, lg := range receipt.Logs {
		if len(lg.Topics) < 2 || lg.Topics[0] != payoutRedemptionTopic {
			continue
		}
		matched = true
		if common.BytesToAddress(lg.Topics[1].Bytes()) != want {
			continue
		}
		if len(lg.Data) < 32 {
			continue
		}
		payout.Add(payout, new(big.Int).SetBytes(lg.Data[len(lg.Data)-32:]))
	}
	if matched {
		return payout, nil
	}
	pusd := common.HexToAddress(consts.PUSDAddress)
	for _, lg := range receipt.Logs {
		if lg.Address != pusd || len(lg.Topics) != 3 || lg.Topics[0] != erc20TransferTopic || len(lg.Data) != 32 {
			continue
		}
		if common.BytesToAddress(lg.Topics[2].Bytes()) != want || common.BytesToAddress(lg.Topics[1].Bytes()) == want {
			continue
		}
		payout.Add(payout, new(big.Int).SetBytes(lg.Data))
	}
	return payout, nil
}

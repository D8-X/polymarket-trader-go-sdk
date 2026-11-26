package onchain

import (
	"context"
	"math/big"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type stubReceiptFetcher struct{ receipt *ethtypes.Receipt }

func (s stubReceiptFetcher) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	return s.receipt, nil
}

func TestPayoutFromRedeemReceiptSumsMatchingLogs(t *testing.T) {
	redeemer := common.HexToAddress("0xF7d4a87A96C1D4B2BB05b2750ed6A6f9c7eb5E62")
	other := common.HexToAddress("0x1111111111111111111111111111111111111111")
	mkPayout := func(addr common.Address, amount int64) *ethtypes.Log {
		data := make([]byte, 32*5)
		new(big.Int).SetInt64(amount).FillBytes(data[len(data)-32:])
		return &ethtypes.Log{
			Topics: []common.Hash{payoutRedemptionTopic, common.BytesToHash(addr.Bytes())},
			Data:   data,
		}
	}
	rcpt := &ethtypes.Receipt{Logs: []*ethtypes.Log{
		mkPayout(redeemer, 500_000),
		mkPayout(other, 999_999),
		mkPayout(redeemer, 1_500_000),
	}}
	got, err := PayoutFromRedeemReceipt(context.Background(), stubReceiptFetcher{receipt: rcpt}, "0xabc", redeemer.Hex())
	if err != nil {
		t.Fatal(err)
	}
	want := big.NewInt(2_000_000)
	if got.Cmp(want) != 0 {
		t.Errorf("payout: got %v want %v", got, want)
	}
}

func TestPayoutFromRedeemReceiptZeroWhenNoMatch(t *testing.T) {
	rcpt := &ethtypes.Receipt{Logs: []*ethtypes.Log{}}
	got, err := PayoutFromRedeemReceipt(context.Background(), stubReceiptFetcher{receipt: rcpt}, "0xabc", "0xF7d4a87A96C1D4B2BB05b2750ed6A6f9c7eb5E62")
	if err != nil {
		t.Fatal(err)
	}
	if got.Sign() != 0 {
		t.Errorf("expected 0 payout from empty receipt, got %v", got)
	}
}

func TestPayoutFromRedeemReceiptFallsBackToPUSDTransfers(t *testing.T) {
	redeemer := common.HexToAddress("0xF7d4a87A96C1D4B2BB05b2750ed6A6f9c7eb5E62")
	module := common.HexToAddress("0x1000008dD9001B968442c1000017eaE6E0dA00Ba")
	pusd := common.HexToAddress(consts.PUSDAddress)
	mkTransfer := func(token, from, to common.Address, amount int64) *ethtypes.Log {
		data := make([]byte, 32)
		big.NewInt(amount).FillBytes(data)
		return &ethtypes.Log{
			Address: token,
			Topics:  []common.Hash{erc20TransferTopic, common.BytesToHash(from.Bytes()), common.BytesToHash(to.Bytes())},
			Data:    data,
		}
	}
	rcpt := &ethtypes.Receipt{Logs: []*ethtypes.Log{
		mkTransfer(pusd, module, redeemer, 3_000_000),
		mkTransfer(pusd, redeemer, module, 7),
		mkTransfer(common.HexToAddress(consts.USDCAddress), module, redeemer, 9),
	}}
	got, err := PayoutFromRedeemReceipt(context.Background(), stubReceiptFetcher{receipt: rcpt}, "0xabc", redeemer.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if got.Cmp(big.NewInt(3_000_000)) != 0 {
		t.Errorf("payout: got %v want 3000000", got)
	}
}

package polytrade

import (
	"context"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type mockReceiptFetcher struct {
	receipt *ethtypes.Receipt
	err     error
}

func (m *mockReceiptFetcher) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	return m.receipt, m.err
}

func TestDepositWalletAddressFromReceiptPicksNonFactoryEmitter(t *testing.T) {
	factory := common.HexToAddress(consts.DepositWalletFactory)
	wallet := common.HexToAddress("0xf7d4a87a96c1d4b2bb05b2750ed6a6f9c7eb5e62")
	mf := &mockReceiptFetcher{
		receipt: &ethtypes.Receipt{
			Logs: []*ethtypes.Log{
				{Address: wallet},
				{Address: wallet},
				{Address: factory},
			},
		},
	}
	got, err := depositWalletAddressFromReceipt(context.Background(), mf, "0x00")
	if err != nil {
		t.Fatal(err)
	}
	if !common.IsHexAddress(got) || common.HexToAddress(got) != wallet {
		t.Errorf("got %s want %s", got, wallet.Hex())
	}
}

func TestDepositWalletAddressFromReceiptErrorsWhenOnlyFactoryEmits(t *testing.T) {
	factory := common.HexToAddress(consts.DepositWalletFactory)
	mf := &mockReceiptFetcher{
		receipt: &ethtypes.Receipt{
			Logs: []*ethtypes.Log{
				{Address: factory},
			},
		},
	}
	_, err := depositWalletAddressFromReceipt(context.Background(), mf, "0x00")
	if err == nil {
		t.Fatal("expected error when only factory emits")
	}
}

func TestDepositWalletAddressFromReceiptErrorsOnNilReceipt(t *testing.T) {
	mf := &mockReceiptFetcher{receipt: nil}
	_, err := depositWalletAddressFromReceipt(context.Background(), mf, "0x00")
	if err == nil {
		t.Fatal("expected error on nil receipt")
	}
}

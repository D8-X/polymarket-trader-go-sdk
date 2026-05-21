package wallet

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

func TestAddressFromReceiptPicksNonFactoryEmitter(t *testing.T) {
	factory := common.HexToAddress(consts.DepositWalletFactory)
	deposit := common.HexToAddress("0xf7d4a87a96c1d4b2bb05b2750ed6a6f9c7eb5e62")
	mf := &mockReceiptFetcher{
		receipt: &ethtypes.Receipt{
			Logs: []*ethtypes.Log{
				{Address: deposit},
				{Address: deposit},
				{Address: factory},
			},
		},
	}
	got, err := AddressFromReceipt(context.Background(), mf, "0x00")
	if err != nil {
		t.Fatal(err)
	}
	if !common.IsHexAddress(got) || common.HexToAddress(got) != deposit {
		t.Errorf("got %s want %s", got, deposit.Hex())
	}
}

func TestAddressFromReceiptErrorsWhenOnlyFactoryEmits(t *testing.T) {
	factory := common.HexToAddress(consts.DepositWalletFactory)
	mf := &mockReceiptFetcher{
		receipt: &ethtypes.Receipt{
			Logs: []*ethtypes.Log{
				{Address: factory},
			},
		},
	}
	_, err := AddressFromReceipt(context.Background(), mf, "0x00")
	if err == nil {
		t.Fatal("expected error when only factory emits")
	}
}

func TestAddressFromReceiptErrorsOnNilReceipt(t *testing.T) {
	mf := &mockReceiptFetcher{receipt: nil}
	_, err := AddressFromReceipt(context.Background(), mf, "0x00")
	if err == nil {
		t.Fatal("expected error on nil receipt")
	}
}

//go:build smoke

package smoke

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	polytrade "github.com/D8-X/polymarket-trader-go-sdk/v2"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestSetupWalletForTradingSmoke(t *testing.T) {
	pk := os.Getenv("POLYMARKET_TEST_PK")
	dw := os.Getenv("POLYMARKET_TEST_DEPOSIT_WALLET")
	rpc := os.Getenv("POLYMARKET_TEST_RPC")
	rk := os.Getenv("POLYMARKET_TEST_RELAYER_API_KEY")
	rs := os.Getenv("POLYMARKET_TEST_RELAYER_SECRET")
	rp := os.Getenv("POLYMARKET_TEST_RELAYER_PASSPHRASE")
	if pk == "" || dw == "" || rpc == "" || rk == "" || rs == "" || rp == "" {
		t.Skip("requires POLYMARKET_TEST_PK, POLYMARKET_TEST_DEPOSIT_WALLET, POLYMARKET_TEST_RPC, and POLYMARKET_TEST_RELAYER_{API_KEY,SECRET,PASSPHRASE}")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	eth, err := ethclient.DialContext(ctx, rpc)
	if err != nil {
		t.Fatalf("dial rpc: %v", err)
	}
	defer eth.Close()

	cli, err := polytrade.NewClient(ctx, polytrade.Config{
		PrivateKeyHex: pk,
		DepositWallet: dw,
		Eth:           eth,
		RelayerCreds:  &polytrade.RelayerCredentials{APIKey: rk, Secret: rs, Passphrase: rp},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	dust := big.NewInt(1000) // 0.001 pUSD

	t.Log("unwrap dust pUSD -> USDC.e (creates input for setup)")
	unwrap, err := cli.UnwrapToUSDC(ctx, dust)
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	t.Logf("  txID=%s state=%s", unwrap.TransactionID, unwrap.State)
	tx, err := cli.WaitForRelayerTransaction(ctx, unwrap.TransactionID)
	if err != nil {
		t.Fatalf("unwrap mine: %v", err)
	}
	t.Logf("  unwrap mined: %s", tx.TransactionHash)
	time.Sleep(5 * time.Second) // relayer holds wallet-busy briefly after a tx mines

	t.Log("SetupWalletForTrading: re-wraps dust + reapplies approvals in one Batch")
	setup, err := cli.SetupWalletForTrading(ctx, dust)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Logf("  txID=%s state=%s", setup.TransactionID, setup.State)
	tx2, err := cli.WaitForRelayerTransaction(ctx, setup.TransactionID)
	if err != nil {
		t.Fatalf("setup mine: %v", err)
	}
	t.Logf("  setup mined: %s", tx2.TransactionHash)
}

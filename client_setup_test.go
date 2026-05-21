package polytrade

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"
)

func TestSetupWalletForTradingRequiresRelayerCreds(t *testing.T) {
	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet})
	_, err := cli.SetupWalletForTrading(context.Background(), big.NewInt(1_000_000))
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

func TestSetupWalletForTradingRequiresDepositWallet(t *testing.T) {
	cli, _ := NewClient(Config{
		PrivateKeyHex: testPrivateKey,
		RelayerCreds:  &RelayerCredentials{APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	_, err := cli.SetupWalletForTrading(context.Background(), big.NewInt(1_000_000))
	if !errors.Is(err, errNoDepositWallet) {
		t.Errorf("expected errNoDepositWallet, got %v", err)
	}
}

func TestSetupWalletForTradingRejectsBadAmount(t *testing.T) {
	cli, _ := NewClient(Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		RelayerCreds:  &RelayerCredentials{APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	_, err := cli.SetupWalletForTrading(context.Background(), big.NewInt(0))
	if err == nil || !strings.Contains(err.Error(), "amount must be positive") {
		t.Errorf("got %v", err)
	}
}

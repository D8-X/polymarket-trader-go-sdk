package polytrade

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"
)

func TestClientSplitPositionRequiresRelayerCreds(t *testing.T) {
	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet})
	_, err := cli.SplitPosition(context.Background(), "0xc0", big.NewInt(1))
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

func TestClientSplitPositionRejectsBadAmount(t *testing.T) {
	cli, _ := NewClient(Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		RelayerCreds:  &RelayerCredentials{APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	_, err := cli.SplitPosition(context.Background(), "0xc0", big.NewInt(0))
	if err == nil || !strings.Contains(err.Error(), "amount must be positive") {
		t.Errorf("got %v", err)
	}
}

package polytrade

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewClientRequiresPrivateKey(t *testing.T) {
	_, err := NewClient(Config{})
	if err == nil {
		t.Fatal("expected error for missing PrivateKeyHex")
	}
}

func TestNewClientRejectsInvalidPrivateKey(t *testing.T) {
	_, err := NewClient(Config{PrivateKeyHex: "0xnot-hex"})
	if err == nil {
		t.Fatal("expected error for invalid PrivateKeyHex")
	}
}

func TestNewClientPopulatesEOA(t *testing.T) {
	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cli.EOA(), "0x") || len(cli.EOA()) != 42 {
		t.Errorf("EOA: got %s", cli.EOA())
	}
	if cli.DepositWallet() != "" {
		t.Errorf("DepositWallet should be empty without Bootstrap, got %s", cli.DepositWallet())
	}
	if cli.Creds() != nil {
		t.Errorf("Creds should be nil without Bootstrap, got %+v", cli.Creds())
	}
}

func TestNewClientHonoursConfigDepositWallet(t *testing.T) {
	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet})
	if err != nil {
		t.Fatal(err)
	}
	if cli.DepositWallet() != testDepositWallet {
		t.Errorf("got %s want %s", cli.DepositWallet(), testDepositWallet)
	}
}

func TestClientPrepareAndSignRequiresCreds(t *testing.T) {
	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.PrepareAndSign("100", BUY, OrderTypeGTC, 0.5, 10)
	if !errors.Is(err, errNoCreds) {
		t.Errorf("expected errNoCreds, got %v", err)
	}
}

func TestClientPrepareAndSignRequiresDepositWallet(t *testing.T) {
	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.PrepareAndSign("100", BUY, OrderTypeGTC, 0.5, 10)
	if !errors.Is(err, errNoDepositWallet) {
		t.Errorf("expected errNoDepositWallet, got %v", err)
	}
}

func TestClientPrepareAndSignHappyPath(t *testing.T) {
	cli, err := NewClient(Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		Creds:         &L2Credentials{APIKey: "k"},
	})
	if err != nil {
		t.Fatal(err)
	}
	signed, err := cli.PrepareAndSign("100", BUY, OrderTypeGTC, 0.55, 10, OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatal(err)
	}
	if signed.Order.Side != BUY {
		t.Errorf("side: got %s want BUY", signed.Order.Side)
	}
	if signed.Order.Maker != testDepositWallet {
		t.Errorf("maker: got %s want %s", signed.Order.Maker, testDepositWallet)
	}
	if signed.Order.SignatureType != SignatureTypePoly1271 {
		t.Errorf("sigType: got %d want %d", signed.Order.SignatureType, SignatureTypePoly1271)
	}
}

func TestClientBootstrapRequiresEth(t *testing.T) {
	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey, RelayerCreds: &RelayerCredentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	err = cli.Bootstrap(context.Background(), nil)
	if !errors.Is(err, errNoEth) {
		t.Errorf("expected errNoEth, got %v", err)
	}
}

func TestClientBootstrapRequiresRelayerCreds(t *testing.T) {
	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey})
	if err != nil {
		t.Fatal(err)
	}
	err = cli.Bootstrap(context.Background(), nil)
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

func TestClientWrapRequiresRelayerCreds(t *testing.T) {
	cli, err := NewClient(Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.WrapToPUSD(context.Background(), nil)
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

func TestClientSetBuilderCodeIsSafeWithoutDepositWallet(t *testing.T) {
	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey})
	cli.SetBuilderCode("0xabc") // must not panic
}

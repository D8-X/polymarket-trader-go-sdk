package polytrade

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ws"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/websocket"
)

const (
	testPrivateKey    = "0x2222222222222222222222222222222222222222222222222222222222222222"
	testDepositWallet = "0x000000000000000000000000000000000000d077"
)

// --- construction ---

func TestNewClientRequiresPrivateKey(t *testing.T) {
	_, err := NewClient(context.Background(), Config{})
	if err == nil {
		t.Fatal("expected error for missing PrivateKeyHex")
	}
}

func TestNewClientRejectsInvalidPrivateKey(t *testing.T) {
	_, err := NewClient(context.Background(), Config{PrivateKeyHex: "0xnot-hex"})
	if err == nil {
		t.Fatal("expected error for invalid PrivateKeyHex")
	}
}

func TestNewClientPopulatesEOA(t *testing.T) {
	cli, err := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cli.EOA(), "0x") || len(cli.EOA()) != 42 {
		t.Errorf("EOA: got %s", cli.EOA())
	}
	if cli.DepositWallet() != "" {
		t.Errorf("DepositWallet should be empty without Bootstrap, got %s", cli.DepositWallet())
	}
	if cli.Creds() == nil {
		t.Errorf("Creds should be populated after NewClient")
	}
}

func TestNewClientHonoursConfigDepositWallet(t *testing.T) {
	cli, err := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet, Creds: &L2Credentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	if cli.DepositWallet() != testDepositWallet {
		t.Errorf("got %s want %s", cli.DepositWallet(), testDepositWallet)
	}
}

type stubEth struct {
	code []byte
}

func (s stubEth) CallContract(ctx context.Context, msg ethereum.CallMsg, blk *big.Int) ([]byte, error) {
	return nil, nil
}

func (s stubEth) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	return nil, nil
}

func (s stubEth) CodeAt(ctx context.Context, addr common.Address, blk *big.Int) ([]byte, error) {
	return s.code, nil
}

func TestNewClientAutoRecoversExistingDepositWallet(t *testing.T) {
	eth := stubEth{code: []byte{0x36, 0x3d, 0x3d}}
	cli, err := NewClient(context.Background(), Config{
		PrivateKeyHex:   testPrivateKey,
		Eth:             eth,
		Creds:           &L2Credentials{APIKey: "k"},
		SkipCredsDerive: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cli.DepositWallet() == "" {
		t.Error("expected DepositWallet to be auto-recovered when Eth reports code at derived address")
	}
}

func TestNewClientSkipsRecoveryWhenNoCode(t *testing.T) {
	eth := stubEth{code: nil}
	cli, err := NewClient(context.Background(), Config{
		PrivateKeyHex:   testPrivateKey,
		Eth:             eth,
		Creds:           &L2Credentials{APIKey: "k"},
		SkipCredsDerive: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cli.DepositWallet() != "" {
		t.Errorf("expected empty DepositWallet when no code at derived address; got %s", cli.DepositWallet())
	}
}

func TestClientSetBuilderCodeIsSafeWithoutDepositWallet(t *testing.T) {
	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	cli.SetBuilderCode("0xabc")
}

// --- prepare and sign ---

func TestClientPrepareAndSignRequiresDepositWallet(t *testing.T) {
	cli, err := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.PrepareAndSign("100", BUY, OrderTypeGTC, 0.5, 10)
	if !errors.Is(err, errNoDepositWallet) {
		t.Errorf("expected errNoDepositWallet, got %v", err)
	}
}

func TestClientPrepareAndSignHappyPath(t *testing.T) {
	cli, err := NewClient(context.Background(), Config{
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

// --- bootstrap ---

func TestClientBootstrapRequiresEth(t *testing.T) {
	cli, err := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, RelayerCreds: &RelayerCredentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	err = cli.Bootstrap(context.Background())
	if !errors.Is(err, errNoEth) {
		t.Errorf("expected errNoEth, got %v", err)
	}
}

func TestClientBootstrapRequiresRelayerCreds(t *testing.T) {
	cli, err := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	err = cli.Bootstrap(context.Background())
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

func TestClientWrapRequiresRelayerCreds(t *testing.T) {
	cli, err := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet, Creds: &L2Credentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.WrapToPUSD(context.Background(), nil)
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

// --- setup wallet ---

func TestSetupWalletForTradingRequiresRelayerCreds(t *testing.T) {
	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet, Creds: &L2Credentials{APIKey: "k"}})
	_, err := cli.SetupWalletForTrading(context.Background(), big.NewInt(1_000_000))
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

func TestSetupWalletForTradingRequiresDepositWallet(t *testing.T) {
	cli, _ := NewClient(context.Background(), Config{
		PrivateKeyHex: testPrivateKey,
		RelayerCreds:  &RelayerCredentials{APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	_, err := cli.SetupWalletForTrading(context.Background(), big.NewInt(1_000_000))
	if !errors.Is(err, errNoDepositWallet) {
		t.Errorf("expected errNoDepositWallet, got %v", err)
	}
}

func TestSetupWalletForTradingRejectsBadAmount(t *testing.T) {
	cli, _ := NewClient(context.Background(), Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		RelayerCreds:  &RelayerCredentials{APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	_, err := cli.SetupWalletForTrading(context.Background(), big.NewInt(0))
	if err == nil || !strings.Contains(err.Error(), "amount must be positive") {
		t.Errorf("got %v", err)
	}
}

// --- CTF ---

func TestClientSplitPositionRequiresRelayerCreds(t *testing.T) {
	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet, Creds: &L2Credentials{APIKey: "k"}})
	_, err := cli.SplitPosition(context.Background(), "0xc0", big.NewInt(1))
	if !errors.Is(err, errNoRelayerCreds) {
		t.Errorf("expected errNoRelayerCreds, got %v", err)
	}
}

func TestClientSplitPositionRejectsBadAmount(t *testing.T) {
	cli, _ := NewClient(context.Background(), Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		RelayerCreds:  &RelayerCredentials{APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	_, err := cli.SplitPosition(context.Background(), "0xc0", big.NewInt(0))
	if err == nil || !strings.Contains(err.Error(), "amount must be positive") {
		t.Errorf("got %v", err)
	}
}

// --- markets and positions ---

func TestClientWrappersHitClob(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/markets/0xabc":
			_ = json.NewEncoder(w).Encode(MarketInfo{ConditionID: "0xabc"})
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	cli.clob.SetBaseURL(srv.URL)
	mkt, err := cli.GetMarket(context.Background(), "0xabc")
	if err != nil {
		t.Fatal(err)
	}
	if mkt.ConditionID != "0xabc" {
		t.Errorf("got %s", mkt.ConditionID)
	}
}

func TestClientGetPositionsRequiresDepositWallet(t *testing.T) {
	cli, err := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cli.GetPositions(context.Background())
	if !errors.Is(err, errNoDepositWallet) {
		t.Errorf("expected errNoDepositWallet, got %v", err)
	}
}

func TestClientGetPositionsCallsDepositWalletAddress(t *testing.T) {
	var queried string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queried = r.URL.Query().Get("user")
		_ = json.NewEncoder(w).Encode([]PositionEntry{
			{Asset: "tok-1", ConditionID: "0x01", Size: 5, AvgPrice: 0.20, CurPrice: 0.22, Outcome: "Yes", Title: "Test market"},
		})
	}))
	defer srv.Close()

	cli, err := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, DepositWallet: testDepositWallet, Creds: &L2Credentials{APIKey: "k"}})
	if err != nil {
		t.Fatal(err)
	}
	cli.clob.SetDataAPIBaseURL(srv.URL)

	positions, err := cli.GetPositions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(queried, testDepositWallet) {
		t.Errorf("queried user: got %s want %s", queried, testDepositWallet)
	}
	if len(positions) != 1 {
		t.Fatalf("positions: got %d want 1", len(positions))
	}
	if positions[0].Size != 5 {
		t.Errorf("size: got %v want 5", positions[0].Size)
	}
}

func TestClientGetPositionsOfQueriesArbitraryAddress(t *testing.T) {
	var queried string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queried = r.URL.Query().Get("user")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	cli.clob.SetDataAPIBaseURL(srv.URL)

	other := "0x000000000000000000000000000000000000beef"
	_, err := cli.GetPositionsOf(context.Background(), other)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(queried, other) {
		t.Errorf("queried user: got %s want %s", queried, other)
	}
}

// --- replace order ---

func TestReplaceOrderRejectsNilOrder(t *testing.T) {
	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	_, _, err := cli.ReplaceOrder(context.Background(), "0x1", nil)
	if err == nil || !strings.Contains(err.Error(), "nil new order") {
		t.Errorf("got %v", err)
	}
}

func TestReplaceOrderCancelsThenPlaces(t *testing.T) {
	var sequence []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/order":
			sequence = append(sequence, "cancel")
			_ = json.NewEncoder(w).Encode(CancelResponse{Canceled: []string{"old-1"}})
		case r.Method == http.MethodPost && r.URL.Path == "/order":
			sequence = append(sequence, "place")
			_ = json.NewEncoder(w).Encode(PlaceOrderResponse{OrderID: "new-1", Success: true, Status: "live"})
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	cli, _ := NewClient(context.Background(), Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		Creds:         &L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	cli.clob.SetBaseURL(srv.URL)

	signed, err := cli.PrepareAndSign("100", BUY, OrderTypeGTC, 0.5, 5, OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatal(err)
	}
	cancelResp, placeResp, err := cli.ReplaceOrder(context.Background(), "old-1", signed)
	if err != nil {
		t.Fatal(err)
	}
	if len(sequence) != 2 || sequence[0] != "cancel" || sequence[1] != "place" {
		t.Errorf("call sequence: got %v want [cancel place]", sequence)
	}
	if cancelResp == nil || len(cancelResp.Canceled) != 1 || cancelResp.Canceled[0] != "old-1" {
		t.Errorf("cancel response: %+v", cancelResp)
	}
	if placeResp == nil || placeResp.OrderID != "new-1" {
		t.Errorf("place response: %+v", placeResp)
	}
}

func TestReplaceOrderStopsOnCancelFailure(t *testing.T) {
	var sequence []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/order":
			sequence = append(sequence, "cancel")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"order not found"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/order":
			sequence = append(sequence, "place")
			_ = json.NewEncoder(w).Encode(PlaceOrderResponse{OrderID: "should-not-happen"})
		}
	}))
	defer srv.Close()

	cli, _ := NewClient(context.Background(), Config{
		PrivateKeyHex: testPrivateKey,
		DepositWallet: testDepositWallet,
		Creds:         &L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"},
	})
	cli.clob.SetBaseURL(srv.URL)

	signed, err := cli.PrepareAndSign("100", BUY, OrderTypeGTC, 0.5, 5, OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatal(err)
	}
	_, placeResp, err := cli.ReplaceOrder(context.Background(), "old-1", signed)
	if err == nil {
		t.Fatal("expected error on cancel failure")
	}
	if placeResp != nil {
		t.Errorf("place should not have run, got %+v", placeResp)
	}
	if len(sequence) != 1 || sequence[0] != "cancel" {
		t.Errorf("call sequence: got %v want [cancel]", sequence)
	}
}

// --- websocket ---

func mockWSServer(t *testing.T, handler func(c *websocket.Conn, subscribeMsg []byte)) (string, func()) {
	t.Helper()
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, sub, err := conn.ReadMessage()
		if err != nil && !errors.Is(err, io.EOF) {
			return
		}
		handler(conn, sub)
	}))
	return "ws" + strings.TrimPrefix(srv.URL, "http"), srv.Close
}

func TestSubscribeMarketReceivesEvents(t *testing.T) {
	url, cleanup := mockWSServer(t, func(c *websocket.Conn, sub []byte) {
		var got struct {
			AssetsIDs []string `json:"assets_ids"`
			Type      string   `json:"type"`
		}
		_ = json.Unmarshal(sub, &got)
		if got.Type != "market" || len(got.AssetsIDs) != 1 || got.AssetsIDs[0] != "tok1" {
			return
		}
		_ = c.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"book","asset_id":"tok1"}`))
		_ = c.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"price_change","asset_id":"tok1","price":"0.55"}`))
		time.Sleep(50 * time.Millisecond)
	})
	defer cleanup()

	prev := ws.MarketURL
	ws.MarketURL = url
	defer func() { ws.MarketURL = prev }()

	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := cli.SubscribeMarket(ctx, []string{"tok1"})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	got := 0
	timeout := time.After(2 * time.Second)
	for got < 2 {
		select {
		case e := <-sub.Events():
			if e.Type != "book" && e.Type != "price_change" {
				t.Errorf("unexpected event type %s", e.Type)
			}
			got++
		case err := <-sub.Errs():
			t.Fatalf("unexpected error: %v", err)
		case <-timeout:
			t.Fatalf("timeout: got %d events", got)
		}
	}
}

func TestSubscribeUserSendsAuth(t *testing.T) {
	type authBlob struct {
		APIKey     string `json:"apiKey"`
		Secret     string `json:"secret"`
		Passphrase string `json:"passphrase"`
	}
	url, cleanup := mockWSServer(t, func(c *websocket.Conn, sub []byte) {
		var got struct {
			Auth    authBlob `json:"auth"`
			Markets []string `json:"markets"`
			Type    string   `json:"type"`
		}
		_ = json.Unmarshal(sub, &got)
		if got.Type != "user" || got.Auth.APIKey != "k" || got.Auth.Secret != "s" || got.Auth.Passphrase != "p" {
			return
		}
		_ = c.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"trade"}`))
		time.Sleep(50 * time.Millisecond)
	})
	defer cleanup()

	prev := ws.UserURL
	ws.UserURL = url
	defer func() { ws.UserURL = prev }()

	cli, _ := NewClient(context.Background(), Config{
		PrivateKeyHex: testPrivateKey,
		Creds:         &L2Credentials{Address: "0x0", APIKey: "k", Secret: "s", Passphrase: "p"},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := cli.SubscribeUser(ctx, []string{"0xcond"})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	select {
	case e := <-sub.Events():
		if e.Type != "trade" {
			t.Errorf("got type %s want trade", e.Type)
		}
	case err := <-sub.Errs():
		t.Fatalf("error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestSubscribeMarketReconnectingReplaysSubscribeAfterDisconnect(t *testing.T) {
	var (
		mu            sync.Mutex
		connectCount  int
		subscribeMsgs [][]byte
	)
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, sub, err := conn.ReadMessage()
		if err != nil {
			return
		}
		mu.Lock()
		connectCount++
		attempt := connectCount
		subscribeMsgs = append(subscribeMsgs, append([]byte{}, sub...))
		mu.Unlock()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"event_type":"book","attempt":`+strconv.Itoa(attempt)+`}`))
		if attempt == 1 {
			time.Sleep(50 * time.Millisecond)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}))
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	defer srv.Close()

	prev := ws.MarketURL
	ws.MarketURL = url
	defer func() { ws.MarketURL = prev }()

	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := cli.SubscribeMarketReconnecting(ctx, []string{"tok1"})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	got := 0
	timeout := time.After(3 * time.Second)
	for got < 2 {
		select {
		case e, ok := <-sub.Events():
			if !ok {
				t.Fatal("events channel closed before two events")
			}
			if e.Type != "book" {
				t.Errorf("unexpected type %s", e.Type)
			}
			got++
		case <-sub.Errs():
		case <-timeout:
			mu.Lock()
			t.Fatalf("timeout: got %d events, %d connects", got, connectCount)
		}
	}
	mu.Lock()
	if connectCount < 2 {
		t.Errorf("expected at least 2 server connects, got %d", connectCount)
	}
	if len(subscribeMsgs) < 2 || string(subscribeMsgs[0]) != string(subscribeMsgs[1]) {
		t.Errorf("subscribe messages differed across reconnects:\n  first: %s\n  second: %s", subscribeMsgs[0], subscribeMsgs[1])
	}
	mu.Unlock()
}

func TestSubscriptionCloseIdempotent(t *testing.T) {
	url, cleanup := mockWSServer(t, func(c *websocket.Conn, sub []byte) {
		time.Sleep(time.Second)
	})
	defer cleanup()
	prev := ws.MarketURL
	ws.MarketURL = url
	defer func() { ws.MarketURL = prev }()

	cli, _ := NewClient(context.Background(), Config{PrivateKeyHex: testPrivateKey, Creds: &L2Credentials{APIKey: "k"}})
	sub, err := cli.SubscribeMarket(context.Background(), []string{"tok1"})
	if err != nil {
		t.Fatal(err)
	}
	_ = sub.Close()
	_ = sub.Close()
}

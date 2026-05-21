package polytrade

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/ws"
	"github.com/gorilla/websocket"
)

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

	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey})
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

func TestSubscribeUserRequiresCreds(t *testing.T) {
	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey})
	_, err := cli.SubscribeUser(context.Background(), []string{"0xc0"})
	if !errors.Is(err, errNoCreds) {
		t.Errorf("expected errNoCreds, got %v", err)
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

	cli, _ := NewClient(Config{
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

	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey})
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

	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey})
	sub, err := cli.SubscribeMarket(context.Background(), []string{"tok1"})
	if err != nil {
		t.Fatal(err)
	}
	_ = sub.Close()
	_ = sub.Close()
}

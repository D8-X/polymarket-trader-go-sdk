package polytrade

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	prev := wsMarketURL
	wsMarketURL = url
	defer func() { wsMarketURL = prev }()

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

	prev := wsUserURL
	wsUserURL = url
	defer func() { wsUserURL = prev }()

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

func TestParseWSEventsHandlesBatch(t *testing.T) {
	msg := []byte(`[{"event_type":"a"},{"event_type":"b"}]`)
	out := parseWSEvents(msg)
	if len(out) != 2 || out[0].Type != "a" || out[1].Type != "b" {
		t.Errorf("got %+v", out)
	}
}

func TestSubscriptionCloseIdempotent(t *testing.T) {
	url, cleanup := mockWSServer(t, func(c *websocket.Conn, sub []byte) {
		time.Sleep(time.Second)
	})
	defer cleanup()
	prev := wsMarketURL
	wsMarketURL = url
	defer func() { wsMarketURL = prev }()

	cli, _ := NewClient(Config{PrivateKeyHex: testPrivateKey})
	sub, err := cli.SubscribeMarket(context.Background(), []string{"tok1"})
	if err != nil {
		t.Fatal(err)
	}
	_ = sub.Close()
	_ = sub.Close()
}

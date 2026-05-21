package polytrade

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	wsMarketURL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"
	wsUserURL   = "wss://ws-subscriptions-clob.polymarket.com/ws/user"
)

const wsPingEvery = 10 * time.Second

type WSEvent struct {
	Type string
	Raw  json.RawMessage
}

type wsSubscription struct {
	events chan WSEvent
	errs   chan error
	cancel context.CancelFunc
	closed chan struct{}
	once   sync.Once

	mu   sync.Mutex
	conn *websocket.Conn
}

func (s *wsSubscription) Events() <-chan WSEvent { return s.events }
func (s *wsSubscription) Errs() <-chan error     { return s.errs }

func (s *wsSubscription) setConn(c *websocket.Conn) {
	s.mu.Lock()
	s.conn = c
	s.mu.Unlock()
}

func (s *wsSubscription) Close() error {
	s.once.Do(func() {
		s.cancel()
		s.mu.Lock()
		if s.conn != nil {
			_ = s.conn.Close()
		}
		s.mu.Unlock()
		<-s.closed
	})
	return nil
}

type MarketSubscription = wsSubscription
type UserSubscription = wsSubscription

var wsDialer = websocket.DefaultDialer

func (c *Client) SubscribeMarket(ctx context.Context, assetIDs []string) (*MarketSubscription, error) {
	body, err := json.Marshal(map[string]any{
		"assets_ids":             assetIDs,
		"type":                   "market",
		"custom_feature_enabled": true,
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe market: marshal: %w", err)
	}
	return c.openWS(ctx, wsMarketURL, body)
}

func (c *Client) SubscribeUser(ctx context.Context, conditionIDs []string) (*UserSubscription, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	body, err := json.Marshal(map[string]any{
		"auth": map[string]string{
			"apiKey":     creds.APIKey,
			"secret":     creds.Secret,
			"passphrase": creds.Passphrase,
		},
		"markets": conditionIDs,
		"type":    "user",
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe user: marshal: %w", err)
	}
	return c.openWS(ctx, wsUserURL, body)
}

func (c *Client) openWS(ctx context.Context, url string, subscribeBody []byte) (*wsSubscription, error) {
	conn, err := dialAndSubscribe(ctx, url, subscribeBody)
	if err != nil {
		return nil, err
	}

	subCtx, cancel := context.WithCancel(ctx)
	sub := &wsSubscription{
		events: make(chan WSEvent, 64),
		errs:   make(chan error, 4),
		cancel: cancel,
		closed: make(chan struct{}),
	}
	sub.setConn(conn)

	go func() {
		defer close(sub.closed)
		defer close(sub.events)
		defer close(sub.errs)
		runConnection(subCtx, conn, sub)
	}()

	return sub, nil
}

func (c *Client) SubscribeMarketReconnecting(ctx context.Context, assetIDs []string) (*MarketSubscription, error) {
	body, err := json.Marshal(map[string]any{
		"assets_ids":             assetIDs,
		"type":                   "market",
		"custom_feature_enabled": true,
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe market reconnecting: marshal: %w", err)
	}
	return c.openWSReconnecting(ctx, wsMarketURL, body)
}

func (c *Client) SubscribeUserReconnecting(ctx context.Context, conditionIDs []string) (*UserSubscription, error) {
	c.mu.RLock()
	creds := c.creds
	c.mu.RUnlock()
	if creds == nil {
		return nil, errNoCreds
	}
	body, err := json.Marshal(map[string]any{
		"auth": map[string]string{
			"apiKey":     creds.APIKey,
			"secret":     creds.Secret,
			"passphrase": creds.Passphrase,
		},
		"markets": conditionIDs,
		"type":    "user",
	})
	if err != nil {
		return nil, fmt.Errorf("subscribe user reconnecting: marshal: %w", err)
	}
	return c.openWSReconnecting(ctx, wsUserURL, body)
}

func (c *Client) openWSReconnecting(ctx context.Context, url string, subscribeBody []byte) (*wsSubscription, error) {
	initial, err := dialAndSubscribe(ctx, url, subscribeBody)
	if err != nil {
		return nil, err
	}

	subCtx, cancel := context.WithCancel(ctx)
	sub := &wsSubscription{
		events: make(chan WSEvent, 64),
		errs:   make(chan error, 8),
		cancel: cancel,
		closed: make(chan struct{}),
	}
	sub.setConn(initial)

	go func() {
		defer close(sub.closed)
		defer close(sub.events)
		defer close(sub.errs)
		conn := initial
		backoff := time.Second
		const maxBackoff = 30 * time.Second
		for {
			runConnection(subCtx, conn, sub)
			if subCtx.Err() != nil {
				return
			}
			select {
			case sub.errs <- fmt.Errorf("websocket: disconnected, reconnecting"):
			default:
			}
			select {
			case <-subCtx.Done():
				return
			case <-time.After(backoff):
			}
			newConn, err := dialAndSubscribe(subCtx, url, subscribeBody)
			if err != nil {
				select {
				case sub.errs <- fmt.Errorf("websocket: reconnect failed: %w", err):
				default:
				}
				if backoff < maxBackoff {
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
				}
				continue
			}
			backoff = time.Second
			conn = newConn
			sub.setConn(newConn)
		}
	}()

	return sub, nil
}

func dialAndSubscribe(ctx context.Context, url string, subscribeBody []byte) (*websocket.Conn, error) {
	conn, _, err := wsDialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket: dial: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, subscribeBody); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("websocket: send subscribe: %w", err)
	}
	return conn, nil
}

func runConnection(ctx context.Context, conn *websocket.Conn, sub *wsSubscription) {
	pingCtx, pingCancel := context.WithCancel(ctx)
	defer pingCancel()
	go func() {
		ticker := time.NewTicker(wsPingEvery)
		defer ticker.Stop()
		for {
			select {
			case <-pingCtx.Done():
				return
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte("PING"), time.Now().Add(5*time.Second)); err != nil {
					return
				}
			}
		}
	}()
	for {
		if ctx.Err() != nil {
			return
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			_ = conn.Close()
			return
		}
		for _, evt := range parseWSEvents(msg) {
			select {
			case sub.events <- evt:
			case <-ctx.Done():
				_ = conn.Close()
				return
			}
		}
	}
}

func parseWSEvents(msg []byte) []WSEvent {
	var single struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(msg, &single); err == nil && single.EventType != "" {
		return []WSEvent{{Type: single.EventType, Raw: json.RawMessage(msg)}}
	}
	var batch []json.RawMessage
	if err := json.Unmarshal(msg, &batch); err == nil {
		out := make([]WSEvent, 0, len(batch))
		for _, raw := range batch {
			var item struct {
				EventType string `json:"event_type"`
			}
			_ = json.Unmarshal(raw, &item)
			out = append(out, WSEvent{Type: item.EventType, Raw: raw})
		}
		return out
	}
	return []WSEvent{{Type: "", Raw: json.RawMessage(msg)}}
}

package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	MarketURL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"
	UserURL   = "wss://ws-subscriptions-clob.polymarket.com/ws/user"
)

const (
	PingEvery = 10 * time.Second
	PongWait  = 25 * time.Second
)

type Event struct {
	Type string
	Raw  json.RawMessage
}

type Subscription struct {
	events chan Event
	errs   chan error
	cancel context.CancelFunc
	closed chan struct{}
	once   sync.Once

	mu   sync.Mutex
	conn *websocket.Conn
}

func (s *Subscription) Events() <-chan Event { return s.events }
func (s *Subscription) Errs() <-chan error   { return s.errs }

func (s *Subscription) setConn(c *websocket.Conn) {
	s.mu.Lock()
	s.conn = c
	s.mu.Unlock()
}

func (s *Subscription) Close() error {
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

var Dialer = websocket.DefaultDialer

func Open(ctx context.Context, url string, subscribeBody []byte) (*Subscription, error) {
	conn, err := dialAndSubscribe(ctx, url, subscribeBody)
	if err != nil {
		return nil, err
	}

	subCtx, cancel := context.WithCancel(ctx)
	sub := &Subscription{
		events: make(chan Event, 64),
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

func OpenReconnecting(ctx context.Context, url string, subscribeBody []byte) (*Subscription, error) {
	initial, err := dialAndSubscribe(ctx, url, subscribeBody)
	if err != nil {
		return nil, err
	}

	subCtx, cancel := context.WithCancel(ctx)
	sub := &Subscription{
		events: make(chan Event, 64),
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
	conn, _, err := Dialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket: dial: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, subscribeBody); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("websocket: send subscribe: %w", err)
	}
	return conn, nil
}

func runConnection(ctx context.Context, conn *websocket.Conn, sub *Subscription) {
	_ = conn.SetReadDeadline(time.Now().Add(PongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(PongWait))
	})
	pingCtx, pingCancel := context.WithCancel(ctx)
	defer pingCancel()
	go func() {
		ticker := time.NewTicker(PingEvery)
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
		for _, evt := range ParseEvents(msg) {
			select {
			case sub.events <- evt:
			case <-ctx.Done():
				_ = conn.Close()
				return
			}
		}
	}
}

func ParseEvents(msg []byte) []Event {
	var single struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(msg, &single); err == nil && single.EventType != "" {
		return []Event{{Type: single.EventType, Raw: json.RawMessage(msg)}}
	}
	var batch []json.RawMessage
	if err := json.Unmarshal(msg, &batch); err == nil {
		out := make([]Event, 0, len(batch))
		for _, raw := range batch {
			var item struct {
				EventType string `json:"event_type"`
			}
			_ = json.Unmarshal(raw, &item)
			out = append(out, Event{Type: item.EventType, Raw: raw})
		}
		return out
	}
	return []Event{{Type: "", Raw: json.RawMessage(msg)}}
}

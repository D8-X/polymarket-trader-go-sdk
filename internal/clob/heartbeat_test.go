package clob

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func TestPostHeartbeatRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/heartbeats" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var in map[string]string
		_ = json.Unmarshal(body, &in)
		if _, ok := in["heartbeat_id"]; !ok {
			t.Errorf("expected heartbeat_id field, got body %s", string(body))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"heartbeat_id": "session-xyz"})
	}))
	defer srv.Close()

	clob := NewClient()
	clob.SetBaseURL(srv.URL)
	creds := &models.L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"}

	got, err := clob.PostHeartbeat(context.Background(), "", creds)
	if err != nil {
		t.Fatalf("post heartbeat: %v", err)
	}
	if got != "session-xyz" {
		t.Errorf("got %q want %q", got, "session-xyz")
	}
}

func TestPostHeartbeatPropagatesErrorMsg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"heartbeat_id": "stale-id",
			"error_msg":    "Invalid Heartbeat ID",
		})
	}))
	defer srv.Close()

	clob := NewClient()
	clob.SetBaseURL(srv.URL)
	creds := &models.L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"}

	_, err := clob.PostHeartbeat(context.Background(), "foo", creds)
	if err == nil || !strings.Contains(err.Error(), "Invalid Heartbeat ID") {
		t.Errorf("expected error containing the server msg, got %v", err)
	}
}

func TestPostHeartbeatReturnsServerIdOn400(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"heartbeat_id": "active-from-server",
			"error_msg":    "Invalid Heartbeat ID",
		})
	}))
	defer srv.Close()

	clob := NewClient()
	clob.SetBaseURL(srv.URL)
	creds := &models.L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"}

	id, err := clob.PostHeartbeat(context.Background(), "", creds)
	if id != "active-from-server" {
		t.Errorf("expected server id to be returned alongside the error, got %q", id)
	}
	if err == nil {
		t.Fatal("expected error on 400")
	}
	if !strings.Contains(err.Error(), "Invalid Heartbeat ID") {
		t.Errorf("err: got %v", err)
	}
}

func TestPostHeartbeatRetriesWithExpectedID(t *testing.T) {
	var sent []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		sent = append(sent, body["heartbeat_id"])
		if body["heartbeat_id"] != "expected" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"heartbeat_id": "expected", "error_msg": "Invalid Heartbeat ID"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"heartbeat_id": "next"})
	}))
	defer srv.Close()

	clob := NewClient()
	clob.SetBaseURL(srv.URL)
	creds := &models.L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"}

	id, err := clob.PostHeartbeat(context.Background(), "stale", creds)
	if err != nil {
		t.Fatalf("expected the retry to succeed, got %v", err)
	}
	if id != "next" {
		t.Errorf("id: got %q want next", id)
	}
	if len(sent) != 2 || sent[0] != "stale" || sent[1] != "expected" {
		t.Errorf("sent ids: got %v want [stale expected]", sent)
	}
}

func TestRunHeartbeatStopsOnContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"heartbeat_id": "tick"})
	}))
	defer srv.Close()

	clob := NewClient()
	clob.SetBaseURL(srv.URL)
	creds := &models.L2Credentials{Address: "0x0", APIKey: "k", Secret: "AAAA", Passphrase: "p"}

	ctx, cancel := context.WithCancel(context.Background())
	errs := clob.RunHeartbeat(ctx, 50*time.Millisecond, creds)

	time.Sleep(120 * time.Millisecond)
	cancel()

	select {
	case _, open := <-errs:
		if open {
			drained := 0
			for range errs {
				drained++
				if drained > 10 {
					t.Fatal("error channel never closed")
				}
			}
		}
	case <-time.After(time.Second):
		t.Fatal("error channel did not close after ctx cancel")
	}
}

package polytrade

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type heartbeatResponse struct {
	HeartbeatID string `json:"heartbeat_id"`
	ErrorMsg    string `json:"error_msg,omitempty"`
}

func (c *CLOBClient) PostHeartbeat(ctx context.Context, heartbeatID string, creds *L2Credentials) (string, error) {
	body, err := json.Marshal(map[string]string{"heartbeat_id": heartbeatID})
	if err != nil {
		return "", fmt.Errorf("post heartbeat: marshal: %w", err)
	}
	path := "/v1/heartbeats"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("post heartbeat: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	headers, err := SignL2Request(creds, http.MethodPost, path, body)
	if err != nil {
		return "", fmt.Errorf("post heartbeat: sign: %w", err)
	}
	ApplyL2Headers(req, headers)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("post heartbeat: http: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("post heartbeat: read body: %w", err)
	}
	var out heartbeatResponse
	_ = json.Unmarshal(respBody, &out)
	if resp.StatusCode != http.StatusOK {
		msg := out.ErrorMsg
		if msg == "" {
			msg = string(respBody)
		}
		return out.HeartbeatID, fmt.Errorf("post heartbeat: status %d: %s", resp.StatusCode, msg)
	}
	if out.ErrorMsg != "" {
		return out.HeartbeatID, fmt.Errorf("post heartbeat: %s", out.ErrorMsg)
	}
	return out.HeartbeatID, nil
}

func (c *CLOBClient) RunHeartbeat(ctx context.Context, interval time.Duration, creds *L2Credentials) <-chan error {
	errs := make(chan error, 1)
	go func() {
		defer close(errs)
		var hbID string
		send := func(initial bool) {
			id, err := c.PostHeartbeat(ctx, hbID, creds)
			adopted := id != "" && id != hbID
			if id != "" {
				hbID = id
			}
			if err != nil {
				if initial && adopted {
					return
				}
				select {
				case errs <- err:
				default:
				}
			}
		}
		send(true)
		if interval <= 0 {
			interval = 5 * time.Second
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				send(false)
			}
		}
	}()
	return errs
}

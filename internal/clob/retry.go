package clob

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

type retryPolicy struct {
	attempts int
	backoff  time.Duration
	maxWait  time.Duration
}


func (c *Client) getWithRetry(ctx context.Context, fullURL, endpoint string) ([]byte, error) {
	attempts := max(c.retry.attempts, 1)
	backoff := c.retry.backoff
	for attempt := 1; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		body, err := c.doRequest(req, endpoint)
		var apiErr *models.APIError
		if err == nil || attempt == attempts || !errors.As(err, &apiErr) ||
			(apiErr.StatusCode != http.StatusTooManyRequests && apiErr.StatusCode != http.StatusServiceUnavailable) {
			return body, err
		}

		wait := backoff
		if apiErr.RetryAfter > 0 {
			wait = apiErr.RetryAfter
		}
		if wait > c.retry.maxWait {
			return nil, err
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("%w (while waiting to retry after: %v)", ctx.Err(), err)
		case <-timer.C:
		}
		backoff *= 2
	}
}

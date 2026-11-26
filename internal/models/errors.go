package models

import (
	"errors"
	"fmt"
	"time"
)

var ErrOrderNotFound = errors.New("order not found")

type APIError struct {
	StatusCode int
	Endpoint   string
	Body       string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("polymarket api %s returned status %d: %s", e.Endpoint, e.StatusCode, e.Body)
}

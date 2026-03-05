package polytrade

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AwaitOrder polls a single placed order until it reaches a terminal status
// (matched, canceled) or the timeout expires. Pass nil for opts to use
// defaults (200ms interval, 10s timeout for delayed orders).
func (c *CLOBClient) AwaitOrder(ctx context.Context, resp *PlaceOrderResponse, creds *L2Credentials, opts *PollOpts) (*PollResult, error) {
	if resp == nil {
		return nil, fmt.Errorf("await order: nil response")
	}
	results := c.awaitMany(ctx, []PlaceOrderResponse{*resp}, creds, opts)
	r := results[0]
	if r.Err != nil {
		return &r, r.Err
	}
	return &r, nil
}

// AwaitOrders polls multiple placed orders concurrently (in a single loop)
// until all reach a terminal status or the timeout expires.
func (c *CLOBClient) AwaitOrders(ctx context.Context, responses []PlaceOrderResponse, creds *L2Credentials, opts *PollOpts) []PollResult {
	return c.awaitMany(ctx, responses, creds, opts)
}

func (c *CLOBClient) awaitMany(ctx context.Context, responses []PlaceOrderResponse, creds *L2Credentials, opts *PollOpts) []PollResult {
	results := make([]PollResult, len(responses))

	var pending []int

	for i, r := range responses {
		results[i] = PollResult{
			OrderID:     r.OrderID,
			PlaceStatus: r.Status,
		}
		if !r.Success {
			results[i].Err = fmt.Errorf("order %s placement failed: %s", r.OrderID, r.ErrorMsg)
			continue
		}
		if isTerminalStatus(r.Status) {
			continue
		}
		pending = append(pending, i)
	}

	if len(pending) == 0 {
		return results
	}

	interval := DefaultPollInterval
	timeout := autoTimeout(responses, pending)
	if opts != nil {
		if opts.Interval > 0 {
			interval = opts.Interval
		}
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			for _, i := range pending {
				results[i].Err = fmt.Errorf("order %s: %w", results[i].OrderID, ctx.Err())
			}
			return results
		case <-ticker.C:
			var stillPending []int
			for _, i := range pending {
				status, err := c.GetOrder(ctx, results[i].OrderID, creds)
				if err != nil {
					stillPending = append(stillPending, i)
					continue
				}
				if isTerminalStatus(status.Status) {
					results[i].Status = status
				} else {
					stillPending = append(stillPending, i)
				}
			}
			pending = stillPending
			if len(pending) == 0 {
				return results
			}
		}
	}
}

// awaitOne polls a single order until it reaches a terminal status or the
// context is cancelled. It is used by the async variants; the sync path
// continues to use awaitMany directly.
func (c *CLOBClient) awaitOne(ctx context.Context, resp PlaceOrderResponse, creds *L2Credentials, interval time.Duration) PollResult {
	result := PollResult{
		OrderID:     resp.OrderID,
		PlaceStatus: resp.Status,
	}

	if !resp.Success {
		result.Err = fmt.Errorf("order %s placement failed: %s", resp.OrderID, resp.ErrorMsg)
		return result
	}

	if isTerminalStatus(resp.Status) {
		return result
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			result.Err = fmt.Errorf("order %s: %w", result.OrderID, ctx.Err())
			return result
		case <-ticker.C:
			status, err := c.GetOrder(ctx, result.OrderID, creds)
			if err != nil {
				// Transient error — retry next tick.
				continue
			}
			if isTerminalStatus(status.Status) {
				result.Status = status
				return result
			}
		}
	}
}

// AwaitOrderAsync is the channel-based variant of AwaitOrder. It returns a
// channel that will receive exactly one PollResult and then be closed. The
// caller can continue doing other work while the order is being polled.
func (c *CLOBClient) AwaitOrderAsync(ctx context.Context, resp *PlaceOrderResponse, creds *L2Credentials, opts *PollOpts) <-chan PollResult {
	ch := make(chan PollResult, 1)

	if resp == nil {
		ch <- PollResult{Err: fmt.Errorf("await order async: nil response")}
		close(ch)
		return ch
	}

	interval := DefaultPollInterval
	timeout := DefaultDelayedPollTimeout
	if resp.Status == OrderStatusLive {
		timeout = DefaultLivePollTimeout
	}
	if opts != nil {
		if opts.Interval > 0 {
			interval = opts.Interval
		}
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
	}

	go func() {
		defer close(ch)
		tctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ch <- c.awaitOne(tctx, *resp, creds, interval)
	}()

	return ch
}

// AwaitOrdersAsync is the channel-based variant of AwaitOrders. It returns a
// channel that streams PollResult values as each order independently reaches
// a terminal status. The channel is closed once all results have been sent.
func (c *CLOBClient) AwaitOrdersAsync(ctx context.Context, responses []PlaceOrderResponse, creds *L2Credentials, opts *PollOpts) <-chan PollResult {
	ch := make(chan PollResult, len(responses))

	if len(responses) == 0 {
		close(ch)
		return ch
	}

	pending := make([]int, 0, len(responses))
	for i, r := range responses {
		if r.Success && !isTerminalStatus(r.Status) {
			pending = append(pending, i)
		}
	}

	interval := DefaultPollInterval
	timeout := autoTimeout(responses, pending)
	if opts != nil {
		if opts.Interval > 0 {
			interval = opts.Interval
		}
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
	}

	tctx, cancel := context.WithTimeout(ctx, timeout)

	var wg sync.WaitGroup
	wg.Add(len(responses))

	for _, r := range responses {
		go func(resp PlaceOrderResponse) {
			defer wg.Done()
			ch <- c.awaitOne(tctx, resp, creds, interval)
		}(r)
	}

	go func() {
		wg.Wait()
		close(ch)
		cancel()
	}()

	return ch
}

// autoTimeout picks a default timeout based on the place statuses of the
// pending orders: 10s if all are delayed, 60s if any are live.
func autoTimeout(responses []PlaceOrderResponse, pending []int) time.Duration {
	for _, i := range pending {
		if responses[i].Status == OrderStatusLive {
			return DefaultLivePollTimeout
		}
	}
	return DefaultDelayedPollTimeout
}

func isTerminalStatus(status string) bool {
	switch status {
	case OrderStatusMatched, OrderStatusCanceled:
		return true
	}
	return false
}

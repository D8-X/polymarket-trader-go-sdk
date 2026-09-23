package models

import "time"

type PollOpts struct {
	Interval time.Duration
	Timeout  time.Duration
}

type PollResult struct {
	OrderID      string
	Status       *OrderStatus
	PlaceStatus  string
	MakingAmount string
	TakingAmount string
	ErrorMsg     string
	Err          error
}

type ResolveOpts struct {
	Interval time.Duration
	Timeout  time.Duration
}

// TradeResolution is the outcome of resolving a set of trade IDs. Pending
// holds the IDs that still had no hash and no failure when the timeout hit.
type TradeResolution struct {
	Hashes  []string
	Trades  []Trade
	Failed  []string
	Pending []string
}

package models

import "time"

type PollOpts struct {
	Interval time.Duration
	Timeout  time.Duration
}

type PollResult struct {
	OrderID     string
	Status      *OrderStatus
	PlaceStatus string
	Err         error
}

package polytrade

import (
	"context"
	"fmt"
)

type ClosePositionOpts struct {
	OrderType string
	TickSize  string
	PostOnly  bool
	DeferExec bool
}

func (ob *OrderBuilder) PrepareClose(tokenID string, price, size float64, apiKey string, opts ClosePositionOpts) (*SignedOrder, error) {
	orderType := opts.OrderType
	if orderType == "" {
		orderType = OrderTypeFOK
	}
	tickSize := opts.TickSize
	if tickSize == "" {
		tickSize = "0.01"
	}
	return ob.PrepareAndSign(tokenID, SELL, orderType, price, size, apiKey, OrderOpts{
		TickSize:  tickSize,
		PostOnly:  opts.PostOnly,
		DeferExec: opts.DeferExec,
	})
}

func (c *CLOBClient) ClosePosition(ctx context.Context, builder *OrderBuilder, tokenID string, price, size float64, creds *L2Credentials, opts ClosePositionOpts) (*PlaceOrderResponse, error) {
	signed, err := builder.PrepareClose(tokenID, price, size, creds.APIKey, opts)
	if err != nil {
		return nil, fmt.Errorf("close position: %w", err)
	}
	return c.PlaceOrder(ctx, signed, creds)
}

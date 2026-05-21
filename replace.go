package polytrade

import (
	"context"
	"fmt"
)

func (c *Client) ReplaceOrder(ctx context.Context, oldOrderID string, newOrder *SignedOrder) (*CancelResponse, *PlaceOrderResponse, error) {
	if newOrder == nil {
		return nil, nil, fmt.Errorf("replace order: nil new order")
	}
	cancelResp, cancelErr := c.CancelOrder(ctx, oldOrderID)
	if cancelErr != nil {
		return cancelResp, nil, fmt.Errorf("replace order: cancel %s: %w", oldOrderID, cancelErr)
	}
	placeResp, placeErr := c.PlaceOrder(ctx, newOrder)
	if placeErr != nil {
		return cancelResp, placeResp, fmt.Errorf("replace order: place new: %w", placeErr)
	}
	return cancelResp, placeResp, nil
}

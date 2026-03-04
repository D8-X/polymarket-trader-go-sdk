# Polymarket CLOB SDK (Go)

Go SDK for the [Polymarket CLOB API](https://docs.polymarket.com).

## Usage example

```go
package main

import (
	"context"
	"fmt"
	"log"

	polytrade "github.com/D8-X/polymarket-trader-go-sdk"
)

func main() {
	ctx := context.Background()
	privateKey := "your-private-key-hex"

	// 1. Create the CLOB client
	clob := polytrade.NewCLOBClient()

	// 2. Derive L2 credentials (API key, secret, passphrase)
	creds, err := polytrade.DeriveL2Credentials(privateKey, polytrade.PolygonChainID)
	if err != nil {
		// First time: create credentials
		creds, err = polytrade.CreateL2Credentials(privateKey, polytrade.PolygonChainID)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 3. Resolve your Gnosis Safe address
	safeAddr, err := polytrade.LookupSafeAddress(ctx, "0xYourEOA")
	if err != nil {
		log.Fatal(err)
	}

	// 4. Create an order builder
	//    For neg-risk markets, use the Neg Risk CTF Exchange address instead.
	ctfExchange := "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"
	builder := polytrade.NewOrderBuilder(
		safeAddr,    // funder (Safe address)
		"0xYourEOA", // signer (EOA)
		ctfExchange,
		privateKey,
		polytrade.SignatureTypeGnosisSafe,
	)

	// 5. Get market data
	tokenID := "your-token-id"

	book, _ := clob.GetOrderBook(ctx, tokenID)
	fmt.Printf("best bid: %s  best ask: %s\n", book.Bids[0].Price, book.Asks[0].Price)

	tickSize, _ := clob.GetTickSize(ctx, tokenID)
	fmt.Printf("tick size: %s\n", tickSize)

	// 6. Build and sign an order
	order, err := builder.PrepareAndSign(
		tokenID,
		"BUY",
		"FOK", // GTC, GTD, FOK, or FAK
		0.55,  // price
		10.0,  // size
		creds.APIKey,
		polytrade.OrderOpts{TickSize: tickSize}, // optional: enables precision validation
	)
	if err != nil {
		log.Fatal(err)
	}

	// 7. Place the order
	resp, err := clob.PlaceOrder(ctx, order, creds)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("order %s success=%v\n", resp.OrderID, resp.Success)

	// 8. Check order status
	status, _ := clob.GetOrder(ctx, resp.OrderID, creds)
	fmt.Printf("status: %s  matched: %s/%s\n", status.Status, status.SizeMatched, status.OriginalSize)

	// 9. Cancel
	cancelResp, _ := clob.CancelOrder(ctx, resp.OrderID, creds)
	fmt.Printf("canceled: %v\n", cancelResp.Canceled)
}
```

## Order Types

| Type | Behavior |
|------|----------|
| GTC  | Rests on the book until filled or cancelled |
| GTD  | Like GTC but with an expiration (60s security threshold added automatically) |
| FOK  | Must fill entirely in one match or is rejected |
| FAK  | Fills as much as possible immediately, remainder is cancelled |

`PostOnly` (GTC/GTD only) rejects the order if it would trade immediately.

**Sports markets:** FOK and FAK orders have a ~3 second placement delay and are automatically cancelled at game start. See [Polymarket sports docs](https://docs.polymarket.com/sports/overview#order-types).

## Tick Size

Pass `TickSize` in `OrderOpts` to enable precision validation. If omitted, defaults to `"0.01"` rounding without validation.

| TickSize | Price decimals | Size decimals | Amount decimals |
|----------|---------------|---------------|-----------------|
| 0.1      | 1             | 2             | 3               |
| 0.01     | 2             | 2             | 4               |
| 0.001    | 3             | 2             | 5               |
| 0.0001   | 4             | 2             | 6               |

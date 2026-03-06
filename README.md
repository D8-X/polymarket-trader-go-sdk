# Polymarket Trader Go SDK

Go SDK for trading on [Polymarket](https://docs.polymarket.com). Covers [Safe wallet provisioning](#gnosis-safe-provisioning) via the relayer (no browser login needed), L2 authentication, order building with [order book sweep](#order-book-sweep) and slippage control, and CLOB interaction.

## Gnosis Safe Provisioning

The SDK derives and deploys Gnosis Safe wallets via Polymarket's relayer, enabling fully automated wallet provisioning for trading bots and hedgers.

```go
ctx := context.Background()
privateKey := "your-private-key-hex"

// Builder credentials from your Polymarket Builder Profile
// See https://docs.polymarket.com/developers/builders/relayer-client
builderCreds := &polytrade.BuilderCredentials{
    APIKey:     "your-builder-api-key",
    Secret:     "your-builder-secret",
    Passphrase: "your-builder-passphrase",
}

// Derive Safe address, deploy if needed, return address + relayer response
safeAddr, relayResp, err := polytrade.EnsureSafeAddress(ctx, "0xYourEOA", privateKey, builderCreds)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Safe address:", safeAddr)
if relayResp != nil {
    fmt.Printf("Deployed: txID=%s state=%s\n", relayResp.TransactionID, relayResp.State)
}
```

Individual functions are also available:

```go
safeAddr := polytrade.DeriveSafeAddress("0xYourEOA")

// Check deployment status via Polymarket relayer
deployed, err := polytrade.IsSafeDeployed(ctx, safeAddr)

// Deploy via relayer. It checks deployment firsta nd skips if already deployed
safeAddr, relayResp, err := polytrade.DeploySafe(ctx, privateKey, builderCreds)

// Lookup via Safe Transaction Service (safe.global). To be used for a second independent validation
safeAddr, err := polytrade.LookupSafeAddress(ctx, "0xYourEOA")
```

After deployment, fund the Safe address with USDC on Polygon to start trading.

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

	// 3. Ensure Gnosis Safe is deployed
	builderCreds := &polytrade.BuilderCredentials{
		APIKey:     "your-builder-api-key",
		Secret:     "your-builder-secret",
		Passphrase: "your-builder-passphrase",
	}
	safeAddr, _, err := polytrade.EnsureSafeAddress(ctx, "0xYourEOA", privateKey, builderCreds)
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

## Order Book Sweep

Sweep multiple order book levels with slippage control:

```go
book, _ := clob.GetOrderBook(ctx, tokenID)
sweep, err := builder.PrepareSweep(book, "BUY", "FOK", 0, 100, 0.02, creds.APIKey) // 0 = use best book price
if err != nil {
	log.Fatal(err)
}
for _, lvl := range sweep.Levels {
	fmt.Printf("level price=%.4f size=%.2f slippage=%.4f\n", lvl.Price, lvl.Size, lvl.Slippage)
}
responses, _ := clob.PlaceOrders(ctx, sweep.Orders, creds)
```

`PrepareSweep` walks the book from best price (or a caller-provided `refPrice`), signs one order per level with the given order type, and stops when the requested size is filled or slippage exceeds the threshold (2% in the example above). The book's `TickSize` is used automatically for rounding.

## Order Polling

For FOK/FAK orders on sports markets, orders go through a `"delayed"` state (~3s) before being matched. Use `AwaitOrder` / `AwaitOrders` to poll until a terminal status is reached:

```go
// Place and await a single order
resp, _ := clob.PlaceOrder(ctx, signedOrder, creds)
result, _ := clob.AwaitOrder(ctx, resp, creds, nil) // default: 200ms poll, 10s timeout
fmt.Printf("order %s: %s matched=%s/%s\n",
    result.OrderID, result.Status.Status, result.Status.SizeMatched, result.Status.OriginalSize)

// Place sweep and await all orders
responses, _ := clob.PlaceOrders(ctx, sweep.Orders, creds)
results := clob.AwaitOrders(ctx, responses, creds, nil)
for _, r := range results {
    if r.Status != nil {
        fmt.Printf("order %s: %s matched=%s/%s\n",
            r.OrderID, r.Status.Status, r.Status.SizeMatched, r.Status.OriginalSize)
    }
}

// Custom poll options
results = clob.AwaitOrders(ctx, responses, creds, &polytrade.PollOpts{
    Interval: 1 * time.Second,
    Timeout:  30 * time.Second,
})
```

Default timeouts: 10s if all orders are delayed, 60s if any are live (GTC/GTD). Override via `PollOpts`.

### Channel based async polling

`AwaitOrderAsync` and `AwaitOrdersAsync` return channels so you can continue doing other work while orders are being polled:

```go
// Async: place and continue working
ch := clob.AwaitOrderAsync(ctx, resp, creds, nil)
// ... do other work ...
result := <-ch

// Async: stream results as orders complete
ch := clob.AwaitOrdersAsync(ctx, responses, creds, nil)
for result := range ch {
    fmt.Printf("order %s: %s\n", result.OrderID, result.Status.Status)
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

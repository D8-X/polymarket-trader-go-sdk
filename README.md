# Polymarket Trader Go SDK

Go SDK for trading on [Polymarket](https://docs.polymarket.com). Designed for fully programmatic, UI-free trading from a Go process. The SDK owns the whole flow: deriving the Gnosis Safe, deploying it via the Polymarket relayer, approving the CTF exchanges, wrapping USDC.e into pUSD, signing orders, and placing them on the CLOB. No browser, no MetaMask, no polymarket.com signup.

Compatible with the new Polymarket CLOB ([V2 migration](https://docs.polymarket.com/v2-migration)): orders use the V2 EIP-712 domain with the `timestamp` / `metadata` / `builder` fields, builder attribution is plumbed through `OrderOpts.BuilderCode`, and dynamic fees are queryable via `CLOBClient.GetClobMarketInfo`.

The SDK targets Gnosis-Safe-backed accounts (`SignatureTypeGnosisSafe`). Use a fresh EOA that has never been used to sign up on polymarket.com, so the Polymarket account is created by the SDK as a Safe.

## Compared to the official clients

Covers the standard auth, order, and market-data operations the official [py-clob-client-v2](https://github.com/Polymarket/py-clob-client-v2) and [clob-client-v2](https://github.com/Polymarket/clob-client-v2) provide. Adds bot-focused extras on top: Gnosis Safe provisioning, gasless transactions through the Polymarket relayer, order book sweep with slippage control, async polling for FOK/FAK delayed states, and Collateral Onramp wrap/unwrap helpers. Pre-migration orders, rewards, order-scoring, WebSocket streams, market discovery, CTF split/merge/redeem, and RFQ are not yet implemented.

## Installation

```bash
go get github.com/D8-X/polymarket-trader-go-sdk/v2
```

## Gnosis Safe Provisioning

The SDK derives and deploys Gnosis Safe wallets via Polymarket's relayer, enabling fully automated wallet provisioning for trading bots and hedgers.

```go
ctx := context.Background()
privateKey := "your-private-key-hex"

// HMAC credentials for the Polymarket Safe relayer (gasless tx submission).
// Distinct from L2 CLOB credentials. See https://docs.polymarket.com/developers/builders/relayer-client
relayerCreds := &polytrade.RelayerCredentials{
    APIKey:     "your-relayer-api-key",
    Secret:     "your-relayer-secret",
    Passphrase: "your-relayer-passphrase",
}

// One-shot onboarding: derives/creates L2 creds, deploys the Safe if needed,
// and approves the CTF exchanges. Idempotent.
boot, err := polytrade.Bootstrap(ctx, privateKey, relayerCreds)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Safe address:", boot.SafeAddress)
```

Lower-level steps are also available:

```go
// Derive Safe address, deploy if needed, return address + relayer response
safeAddr, relayResp, err := polytrade.EnsureSafeAddress(ctx, "0xYourEOA", privateKey, relayerCreds)
```

Individual functions are also available:

```go
safeAddr := polytrade.DeriveSafeAddress("0xYourEOA")

// Check deployment status via Polymarket relayer
deployed, err := polytrade.IsSafeDeployed(ctx, safeAddr)

// Deploy via relayer. Checks deployment first and skips if already deployed.
safeAddr, relayResp, err := polytrade.DeploySafe(ctx, privateKey, relayerCreds)

// Lookup via Safe Transaction Service (safe.global). Useful as a second independent validation.
safeAddr, err := polytrade.LookupSafeAddress(ctx, "0xYourEOA")
```

After deployment, fund the Safe with the active collateral token on Polygon. Use `polytrade.CollateralAddress()`; it returns pUSD by default and falls back to USDC.e only if `polytrade.SetPUSDAddress("")` is called.

## Collateral Balance

Query and refresh the collateral balance available for trading:

```go
// Reuse L2 creds derived once at startup.
creds, _ := polytrade.DeriveL2Credentials(privateKey, polytrade.PolygonChainID)

// Current balance (raw units, 6 decimals)
balance, err := polytrade.CollateralBalanceOf(ctx, creds)
fmt.Printf("balance: %s\n", balance)

// After transferring collateral to the Safe, refresh so Polymarket picks up the new balance
err = polytrade.RefreshCollateralBalance(ctx, creds)
```

`RefreshCollateralBalance` calls `UpdateBalanceAllowance` under the hood. The underlying CLOB API uses the `COLLATERAL` asset type, so these helpers work for both USDC.e and pUSD.

### Wrapping and Unwrapping pUSD

API traders moving between USDC.e and pUSD can use the Collateral Onramp / Offramp via the relayer:

```go
amount := big.NewInt(5_000_000) // 5 pUSD (6 decimals)

// USDC.e in the Safe -> pUSD in the Safe
relayResp, err := polytrade.WrapToPUSD(ctx, eoaAddress, privateKey, amount, relayerCreds)

// pUSD in the Safe -> USDC.e in the Safe
relayResp, err = polytrade.UnwrapToUSDC(ctx, eoaAddress, privateKey, amount, relayerCreds)
```

Both helpers batch the required ERC-20 approval and the on/offramp call into a single gasless Safe transaction.

## Usage example

```go
package main

import (
	"context"
	"fmt"
	"log"

	polytrade "github.com/D8-X/polymarket-trader-go-sdk/v2"
)

func main() {
	ctx := context.Background()
	privateKey := "your-private-key-hex"
	clob := polytrade.NewCLOBClient()

	// 1. One-shot onboarding: L2 creds, Safe deploy, CTF approvals. Idempotent.
	relayerCreds := &polytrade.RelayerCredentials{
		APIKey:     "your-relayer-api-key",
		Secret:     "your-relayer-secret",
		Passphrase: "your-relayer-passphrase",
	}
	boot, err := polytrade.Bootstrap(ctx, privateKey, relayerCreds)
	if err != nil {
		log.Fatal(err)
	}
	creds := boot.Creds

	// 2. Create an order builder.
	//    For neg-risk markets, pass polytrade.NegRiskCTFExchange instead.
	builder := polytrade.NewOrderBuilder(
		boot.SafeAddress,
		polytrade.CTFExchange,
		privateKey,
		polytrade.SignatureTypeGnosisSafe,
	)

	// Optional: attach a builder code from your Polymarket Builder Profile to every order
	// builder.SetBuilderCode("0x...")

	// 3. Get market data
	tokenID := "your-token-id"

	book, _ := clob.GetOrderBook(ctx, tokenID)
	fmt.Printf("best bid: %s  best ask: %s\n", book.Bids[0].Price, book.Asks[0].Price)

	// Per-market metadata: tick size, min order size, fee details, tokens
	info, _ := clob.GetClobMarketInfo(ctx, "your-condition-id")
	tickSize := info.MinTickSize.String()
	fmt.Printf("tick: %s  fee r=%g e=%d takerOnly=%v\n",
		tickSize, info.FeeDetails.Rate, info.FeeDetails.Exponent, info.FeeDetails.TakerOnly)

	// 4. Build and sign an order
	order, err := builder.PrepareAndSign(
		tokenID,
		polytrade.BUY,
		polytrade.OrderTypeFOK, // or OrderTypeGTC / GTD / FAK
		0.55,                   // price
		10.0,                   // size
		creds.APIKey,
		polytrade.OrderOpts{TickSize: tickSize}, // optional: enables precision validation
	)
	if err != nil {
		log.Fatal(err)
	}

	// 5. Place the order
	resp, err := clob.PlaceOrder(ctx, order, creds)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("order %s success=%v\n", resp.OrderID, resp.Success)

	// 6. Check order status
	status, _ := clob.GetOrder(ctx, resp.OrderID, creds)
	fmt.Printf("status: %s  matched: %s/%s\n", status.Status, status.SizeMatched, status.OriginalSize)

	// 7. Cancel
	cancelResp, _ := clob.CancelOrder(ctx, resp.OrderID, creds)
	fmt.Printf("canceled: %v\n", cancelResp.Canceled)
}
```

## Order Book Sweep

Sweep multiple order book levels with slippage control:

```go
book, _ := clob.GetOrderBook(ctx, tokenID)
sweep, err := builder.PrepareSweep(book, polytrade.BUY, polytrade.OrderTypeFOK, 0, 100, 0.02, creds.APIKey) // 0 = use best book price
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

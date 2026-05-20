# Polymarket Trader Go SDK

Go SDK for trading on [Polymarket](https://docs.polymarket.com). Designed for fully programmatic, UI-free trading from a Go process. The SDK owns the whole onboarding and order path. It deploys the V2 deposit wallet via the Polymarket relayer, derives L2 CLOB credentials, wraps USDC.e into pUSD, approves the CTF exchanges, signs orders, and places them on the CLOB. No browser, no MetaMask, no polymarket.com signup.

The SDK targets Polymarket V2 deposit-wallet accounts (`SignatureTypePoly1271`, EIP-1271 + ERC-7739). Use a fresh EOA so the Polymarket account is created by the SDK from scratch.

## Compared to the official clients

Covers the standard auth, order, and market-data operations the official [py-clob-client-v2](https://github.com/Polymarket/py-clob-client-v2) and [clob-client-v2](https://github.com/Polymarket/clob-client-v2) provide. Adds bot-focused extras on top such as deposit-wallet onboarding through the Polymarket relayer (gasless), order book sweep with slippage control, async polling for FOK/FAK delayed states, and Collateral Onramp wrap/unwrap helpers. Pre-migration orders, rewards, order-scoring, WebSocket streams, market discovery, CTF split/merge/redeem, and RFQ are not yet implemented.

## Installation

```bash
go get github.com/D8-X/polymarket-trader-go-sdk/v2
```

## How an order reaches the CLOB

Three phases. The first two are one-time setup per EOA. The third runs for every trade.

**1. Onboard the EOA.** Polymarket V2 doesn't accept orders signed by a raw EOA. Each account trades through a deposit wallet, a smart contract deployed per EOA at a deterministic CREATE2 address. The relayer deploys it on demand via a `WALLET-CREATE` request (no EOA signature needed in the body, just relayer HMAC credentials). Once on-chain, the contract is permanently bound to your EOA and verifies signatures via EIP-1271. You also derive L2 CLOB credentials (`apiKey` / `secret` / `passphrase`) by signing a one-time EIP-712 ClobAuth payload with your EOA. These are HMAC creds used to authenticate every request to `clob.polymarket.com`.

**2. Fund and approve the deposit wallet.** Trades settle in pUSD, not USDC. Send USDC.e to the deposit wallet, then submit a signed Batch through the relayer that approves the Collateral Onramp, wraps USDC.e to pUSD, and approves the V2 CTF exchanges to pull pUSD on match. The Batch is one EIP-712 signature over the `DepositWallet/1` domain (verifying contract is the deposit wallet itself). The relayer executes it on-chain and pays the gas.

**3. Build, sign, and post the order.** A V2 Order is a standard EIP-712 struct with `maker = signer = deposit wallet` and `signatureType = 3` (POLY_1271). The signature is not a raw ECDSA over the order digest. Because the maker is a contract, the wallet's EIP-1271 verifier expects an ERC-7739 wrap. The EOA signs a `TypedDataSign(...)` digest that nests the order inside the deposit-wallet domain, and the wire signature is `inner_sig || appDomainSep || contentsHash || utf8(orderTypeStr) || uint16be(len)` (around 318 bytes). The signed order is `POST`ed to `/order` with L2 HMAC headers. The CLOB checks the signature via EIP-1271, verifies funder allowance and balance on-chain, then rests or matches it.

The EOA only ever signs. The relayer pays every gas cost. The order doesn't touch the chain until a match settles.

## Doing it with the SDK

The SDK collapses each phase into a small set of calls. The pieces below mirror the three phases above.

```go
ctx := context.Background()
privateKey := "your-private-key-hex"

relayerCreds := &polytrade.RelayerCredentials{
    APIKey:     "your-relayer-api-key",
    Secret:     "your-relayer-secret",
    Passphrase: "your-relayer-passphrase",
}
```

Phase 1 deploys the deposit wallet (idempotent) and grabs L2 creds. The deployment tx receipt's `contractAddress` field is the deterministic CREATE2 address. Store it. Subsequent calls use it as the funder address.

```go
deployResp, err := polytrade.DeployDepositWallet(ctx, eoaAddress, relayerCreds)
depositWallet := "0x..." // from polygonscan or the deploy tx receipt

creds, err := polytrade.DeriveL2Credentials(privateKey, polytrade.PolygonChainID)
if err != nil {
    creds, err = polytrade.CreateL2Credentials(privateKey, polytrade.PolygonChainID)
}
```

Phase 2 sends USDC.e to `depositWallet` from any wallet you control (the SDK does not move funds in for you), then wraps + approves in one signed Batch. The approvals cover both BUY and SELL flows. They grant the V2 exchanges spend on the deposit wallet's pUSD and `setApprovalForAll` over the conditional-token ERC-1155 so SELL orders can transfer outcome tokens out.

```go
amount := big.NewInt(1_000_000) // 1.0 USDC.e
_, err = polytrade.WrapAndApproveDepositWallet(ctx, eoaAddress, privateKey, depositWallet, amount, relayerCreds)
```

Both phases at once:

```go
boot, err := polytrade.BootstrapDepositWallet(ctx, privateKey, depositWallet, amount, relayerCreds)
creds := boot.Creds
```

Phase 3 builds, signs, and places an order. The builder handles `maker = signer = depositWallet` and the ERC-7739 wrap automatically when you pass `SignatureTypePoly1271`.

```go
builder := polytrade.NewOrderBuilder(
    depositWallet,
    polytrade.CTFExchange,
    privateKey,
    polytrade.SignatureTypePoly1271,
)
signed, err := builder.PrepareAndSign(
    tokenID,
    polytrade.BUY,
    polytrade.OrderTypeGTC,
    0.55, 10,
    creds.APIKey,
    polytrade.OrderOpts{TickSize: "0.01"},
)
resp, err := clob.PlaceOrder(ctx, signed, creds)
```

To close (sell) a position, pass an `ethclient.Client` and the SDK reads the held quantity on-chain before signing a SELL for the full balance. No size argument.

```go
eth, _ := ethclient.DialContext(ctx, polygonRPCURL)
resp, err := clob.ClosePosition(ctx, eth, builder, tokenID, 0.50, creds, polytrade.ClosePositionOpts{})
```

## Collateral Balance

Query and refresh the collateral balance available for trading:

```go
// Reuse L2 creds derived once at startup.
creds, _ := polytrade.DeriveL2Credentials(privateKey, polytrade.PolygonChainID)

// Current balance (raw units, 6 decimals)
balance, err := polytrade.CollateralBalanceOf(ctx, creds)
fmt.Printf("balance: %s\n", balance)

// After transferring collateral to the deposit wallet, refresh so Polymarket picks up the new balance
err = polytrade.RefreshCollateralBalance(ctx, creds)
```

`RefreshCollateralBalance` calls `UpdateBalanceAllowance` under the hood. The underlying CLOB API uses the `COLLATERAL` asset type, so these helpers work for both USDC.e and pUSD.

### Wrapping and Unwrapping pUSD

API traders moving between USDC.e and pUSD can use the Collateral Onramp / Offramp through the deposit wallet.

```go
amount := big.NewInt(5_000_000) // 5 pUSD (6 decimals)

// USDC.e in the deposit wallet -> pUSD in the deposit wallet
relayResp, err := polytrade.WrapToPUSD(ctx, eoaAddress, privateKey, depositWallet, amount, relayerCreds)

// pUSD in the deposit wallet -> USDC.e in the deposit wallet
relayResp, err = polytrade.UnwrapToUSDC(ctx, eoaAddress, privateKey, depositWallet, amount, relayerCreds)
```

Both helpers batch the required ERC-20 approval and the on/offramp call into a single gasless relayer batch signed over the `DepositWallet/1` EIP-712 domain.

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

	// 1. One-shot onboarding (deposit wallet deploy, L2 creds, wrap + approvals).
	relayerCreds := &polytrade.RelayerCredentials{
		APIKey:     "your-relayer-api-key",
		Secret:     "your-relayer-secret",
		Passphrase: "your-relayer-passphrase",
	}
	depositWallet := "0xYourDepositWallet" // from the WALLET-CREATE receipt
	boot, err := polytrade.BootstrapDepositWallet(ctx, privateKey, depositWallet, big.NewInt(1_000_000), relayerCreds)
	if err != nil {
		log.Fatal(err)
	}
	creds := boot.Creds

	// 2. Create an order builder.
	//    For neg-risk markets, pass polytrade.NegRiskCTFExchange instead.
	builder := polytrade.NewOrderBuilder(
		boot.DepositWalletAddress,
		polytrade.CTFExchange,
		privateKey,
		polytrade.SignatureTypePoly1271,
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
result, _ := clob.AwaitOrder(ctx, resp, creds, nil) // default: 200ms poll, 5s timeout
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

Default timeouts are 5s if all orders are delayed, 60s if any are live (GTC/GTD). Override via `PollOpts`.

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

## Heartbeat

For long-running market makers, Polymarket auto-cancels all open orders if no heartbeat is received within ~15 seconds. Send one explicitly or run a background loop. The server rotates the `heartbeat_id` per call; the helpers track it for you.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
errs := clob.RunHeartbeat(ctx, 5*time.Second, creds)
go func() {
    for e := range errs {
        log.Printf("heartbeat: %v", e)
    }
}()
```

## Order Types

| Type | Behavior |
|------|----------|
| GTC  | Rests on the book until filled or cancelled |
| GTD  | Like GTC but with an expiration (60s security threshold added automatically) |
| FOK  | Must fill entirely in one match or is rejected |
| FAK  | Fills as much as possible immediately, remainder is cancelled |

`PostOnly` (GTC/GTD only) rejects the order if it would trade immediately.

**Sports markets.** Marketable orders (FOK / FAK) have a ~1 second placement delay before matching. Outstanding limit orders are automatically cancelled when the game begins. See [Polymarket order docs](https://docs.polymarket.com/trading/orders/create#sports-markets).

## Tick Size

Pass `TickSize` in `OrderOpts` to enable precision validation. If omitted, defaults to `"0.01"` rounding without validation.

| TickSize | Price decimals | Size decimals | Amount decimals |
|----------|---------------|---------------|-----------------|
| 0.1      | 1             | 2             | 3               |
| 0.01     | 2             | 2             | 4               |
| 0.001    | 3             | 2             | 5               |
| 0.0001   | 4             | 2             | 6               |

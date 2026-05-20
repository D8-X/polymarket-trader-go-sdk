# Polymarket Trader Go SDK

Go SDK for trading on [Polymarket](https://docs.polymarket.com). Designed for fully programmatic, UI-free trading from a Go process. The SDK owns the whole onboarding and order path. It deploys the V2 deposit wallet via the Polymarket relayer, derives L2 CLOB credentials, wraps USDC.e into pUSD, approves the CTF exchanges, signs orders, and places them on the CLOB. No browser, no MetaMask, no polymarket.com signup.

The SDK targets Polymarket V2 deposit-wallet accounts (`SignatureTypePoly1271`, EIP-1271 + ERC-7739). Use a fresh EOA so the Polymarket account is created by the SDK from scratch.

## Compared to the official clients

Covers the standard auth, order, and market-data operations the official [py-clob-client-v2](https://github.com/Polymarket/py-clob-client-v2) and [clob-client-v2](https://github.com/Polymarket/clob-client-v2) provide. Adds bot-focused extras on top such as deposit-wallet onboarding through the Polymarket relayer (gasless), pre-trade sweep estimation with slippage control, async polling for FOK/FAK delayed states, and Collateral Onramp wrap/unwrap helpers. Pre-migration orders, rewards, order-scoring, WebSocket streams, market discovery, CTF split/merge/redeem, and RFQ are not yet implemented.

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

A single `Client` owns the private key, the L2 credentials, the deposit-wallet address, and the relayer + RPC handles. Everything else is a method on it.

```go
ctx := context.Background()
eth, _ := ethclient.DialContext(ctx, polygonRPCURL)

cli, err := polytrade.NewClient(polytrade.Config{
    PrivateKeyHex: "your-private-key-hex",
    Eth:           eth,
    RelayerCreds: &polytrade.RelayerCredentials{
        APIKey:     "your-relayer-api-key",
        Secret:     "your-relayer-secret",
        Passphrase: "your-relayer-passphrase",
    },
})
if err != nil {
    log.Fatal(err)
}

// One-shot onboarding. Derives L2 creds, deploys the deposit wallet,
// resolves its address from the receipt, wraps + approves in one signed
// Batch (both BUY and SELL flows).
if err := cli.Bootstrap(ctx, big.NewInt(1_000_000)); err != nil {
    log.Fatal(err)
}
fmt.Println("Deposit wallet:", cli.DepositWallet())

// Place an order.
signed, _ := cli.PrepareAndSign(tokenID, polytrade.BUY, polytrade.OrderTypeGTC, 0.55, 10, polytrade.OrderOpts{TickSize: "0.01"})
resp, _ := cli.PlaceOrder(ctx, signed)

// Close a position (auto-fetches held quantity on-chain).
closed, _ := cli.ClosePosition(ctx, tokenID, 0.50, polytrade.ClosePositionOpts{})
```

## Collateral

```go
balance, _ := cli.CollateralBalanceOf(ctx)               // current pUSD balance
_ = cli.RefreshCollateralBalance(ctx)                    // tell the CLOB to re-index

// Wrap USDC.e in the deposit wallet to pUSD (gasless, one signed Batch).
_, _ = cli.WrapToPUSD(ctx, big.NewInt(5_000_000))

// Reverse direction.
_, _ = cli.UnwrapToUSDC(ctx, big.NewInt(5_000_000))

// Withdraw out of the deposit wallet to any address.
_, _ = cli.TransferOut(ctx, polytrade.PUSDAddress, recipientAddress, big.NewInt(5_000_000))
```

## Positions

```go
positions, _ := cli.GetPositions(ctx)                                  // for the deposit wallet
others, _ := cli.GetPositionsOf(ctx, "0xAnyWallet")                    // for any address
for _, p := range positions {
    fmt.Printf("%s %s size=%g avg=%g cur=%g\n", p.Title, p.Outcome, p.Size, p.AvgPrice, p.CurPrice)
}
```

## Usage example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	polytrade "github.com/D8-X/polymarket-trader-go-sdk/v2"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	ctx := context.Background()
	eth, _ := ethclient.DialContext(ctx, "https://polygon-rpc.com")

	cli, err := polytrade.NewClient(polytrade.Config{
		PrivateKeyHex: "your-private-key-hex",
		Eth:           eth,
		RelayerCreds: &polytrade.RelayerCredentials{
			APIKey:     "your-relayer-api-key",
			Secret:     "your-relayer-secret",
			Passphrase: "your-relayer-passphrase",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := cli.Bootstrap(ctx, big.NewInt(1_000_000)); err != nil {
		log.Fatal(err)
	}

	// Optional: builder attribution.
	// cli.SetBuilderCode("0x...")

	tokenID := "your-token-id"

	book, _ := cli.GetOrderBook(ctx, tokenID)
	fmt.Printf("best bid: %s  best ask: %s\n", book.Bids[0].Price, book.Asks[0].Price)

	info, _ := cli.GetClobMarketInfo(ctx, "your-condition-id")
	tickSize := info.MinTickSize.String()

	signed, err := cli.PrepareAndSign(
		tokenID, polytrade.BUY, polytrade.OrderTypeFOK,
		0.55, 10,
		polytrade.OrderOpts{TickSize: tickSize},
	)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := cli.PlaceOrder(ctx, signed)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("order %s success=%v\n", resp.OrderID, resp.Success)

	status, _ := cli.GetOrder(ctx, resp.OrderID)
	fmt.Printf("status: %s  matched: %s/%s\n", status.Status, status.SizeMatched, status.OriginalSize)

	cancelResp, _ := cli.CancelOrder(ctx, resp.OrderID)
	fmt.Printf("canceled: %v\n", cancelResp.Canceled)
}
```

## Sweep estimation

`EstimateSweep` walks the order book from the best price (or a caller-supplied `refPrice`) until your requested size is filled or slippage exceeds the threshold. It returns the deepest price touched, the total fillable size, the size-weighted average price, and the per-level breakdown. It does NOT sign orders. The matching engine already walks levels for you, so a single limit order at the worst price fills the same way as N orders at each level.

```go
book, _ := cli.GetOrderBook(ctx, tokenID)
est, err := polytrade.EstimateSweep(book, polytrade.BUY, 0, 100, 0.02)
if err != nil {
	log.Fatal(err)
}
for _, lvl := range est.Levels {
	fmt.Printf("level price=%.4f size=%.2f slippage=%.4f\n", lvl.Price, lvl.Size, lvl.Slippage)
}

signed, _ := cli.PrepareAndSign(
	tokenID, polytrade.BUY, polytrade.OrderTypeFAK,
	est.WorstPrice, est.TotalSize,
	polytrade.OrderOpts{TickSize: book.TickSize},
)
resp, _ := cli.PlaceOrder(ctx, signed)
```

## Order Polling

For FOK/FAK orders on sports markets, orders go through a `"delayed"` state (~3s) before being matched. Use `AwaitOrder` / `AwaitOrders` to poll until a terminal status is reached:

```go
// Place and await a single order
resp, _ := cli.PlaceOrder(ctx, signedOrder)
result, _ := cli.AwaitOrder(ctx, resp, nil) // default: 200ms poll, 5s timeout
fmt.Printf("order %s: %s matched=%s/%s\n",
    result.OrderID, result.Status.Status, result.Status.SizeMatched, result.Status.OriginalSize)

// Place a batch of orders and await all of them
responses, _ := cli.PlaceOrders(ctx, signedOrders)
results := cli.AwaitOrders(ctx, responses, nil)
for _, r := range results {
    if r.Status != nil {
        fmt.Printf("order %s: %s matched=%s/%s\n",
            r.OrderID, r.Status.Status, r.Status.SizeMatched, r.Status.OriginalSize)
    }
}

// Custom poll options
results = cli.AwaitOrders(ctx, responses, &polytrade.PollOpts{
    Interval: 1 * time.Second,
    Timeout:  30 * time.Second,
})
```

Default timeouts are 5s if all orders are delayed, 60s if any are live (GTC/GTD). Override via `PollOpts`.

### Channel based async polling

`AwaitOrderAsync` and `AwaitOrdersAsync` return channels so you can continue doing other work while orders are being polled:

```go
// Async: place and continue working
ch := cli.AwaitOrderAsync(ctx, resp, nil)
// ... do other work ...
result := <-ch

// Async: stream results as orders complete
ch := cli.AwaitOrdersAsync(ctx, responses, nil)
for result := range ch {
    fmt.Printf("order %s: %s\n", result.OrderID, result.Status.Status)
}
```

## Heartbeat

For long-running market makers, Polymarket auto-cancels all open orders if no heartbeat is received within ~15 seconds. Send one explicitly or run a background loop. The server rotates the `heartbeat_id` per call; the helpers track it for you.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
errs := cli.RunHeartbeat(ctx, 5*time.Second)
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

package polytrade

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

// PrepareSweep walks the order book from a reference price and builds one signed
// order per level until size is filled or slippage exceeds maxSlippage.
//
// If refPrice <= 0, the best price from the book is used as reference.
// For BUY orders it walks the asks (ascending); for SELL it walks the bids
// (descending). The book's TickSize is used automatically for amount rounding.
func (ob *OrderBuilder) PrepareSweep(
	book *OrderBook,
	side, orderType string,
	refPrice, size, maxSlippage float64, apiKey string, opts ...OrderOpts) (*SweepResult, error) {
	var opt OrderOpts
	if len(opts) > 0 {
		opt = opts[0]
	}
	// take TickSize from the book if the caller didn't set one.
	if opt.TickSize == "" {
		opt.TickSize = book.TickSize
	}

	levels, err := parseLevels(book, side)
	if err != nil {
		return nil, err
	}

	return ob.PrepareSweepFromLevels(book.AssetID, levels, side, orderType, refPrice, size, maxSlippage, apiKey, opt)
}

// PrepareSweepFromLevels is the core sweep method. It takes preparsed price
// levels (already sorted from best to worst) and builds one signed order per level
// until size is filled or slippage exceeds maxSlippage.
//
// If refPrice <= 0, the first level's price is used as reference.
func (ob *OrderBuilder) PrepareSweepFromLevels(assetID string, levels []PriceLevel, side, orderType string, refPrice, size, maxSlippage float64, apiKey string, opts ...OrderOpts) (*SweepResult, error) {
	var opt OrderOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	if len(levels) == 0 {
		return nil, fmt.Errorf("prepare sweep: no levels provided")
	}

	bestPrice := levels[0].Price
	if refPrice > 0 {
		bestPrice = refPrice
	}

	result := &SweepResult{BestPrice: bestPrice, Side: side}
	remaining := size
	var costSum float64

	for _, lvl := range levels {
		if remaining <= 0 {
			break
		}
		slippage := math.Abs(lvl.Price-bestPrice) / bestPrice
		if slippage > maxSlippage {
			break
		}
		fillSize := math.Min(remaining, lvl.Size)
		if fillSize <= 0 {
			continue
		}

		signed, err := ob.PrepareAndSign(assetID, side, orderType, lvl.Price, fillSize, apiKey, opt)
		if err != nil {
			return nil, fmt.Errorf("prepare sweep level price=%v size=%v: %w", lvl.Price, fillSize, err)
		}

		result.Orders = append(result.Orders, signed)
		result.Levels = append(result.Levels, SweepLevel{
			Price:    lvl.Price,
			Size:     fillSize,
			Slippage: slippage,
		})
		result.TotalSize += fillSize
		result.WorstPrice = lvl.Price
		costSum += lvl.Price * fillSize
		remaining -= fillSize
	}

	if len(result.Orders) == 0 {
		return nil, fmt.Errorf("prepare sweep: no levels within %.2f%% slippage", maxSlippage*100)
	}
	if result.TotalSize > 0 {
		result.AvgPrice = costSum / result.TotalSize
	}
	return result, nil
}

func parseLevels(book *OrderBook, side string) ([]PriceLevel, error) {
	var raw []OrderBookLevel
	switch side {
	case "BUY":
		raw = book.Asks
	case "SELL":
		raw = book.Bids
	default:
		return nil, fmt.Errorf("prepare sweep: invalid side %q (want BUY or SELL)", side)
	}

	out := make([]PriceLevel, 0, len(raw))
	for _, lvl := range raw {
		p, err := strconv.ParseFloat(lvl.Price, 64)
		if err != nil {
			return nil, fmt.Errorf("prepare sweep: parse price %q: %w", lvl.Price, err)
		}
		s, err := strconv.ParseFloat(lvl.Size, 64)
		if err != nil {
			return nil, fmt.Errorf("prepare sweep: parse size %q: %w", lvl.Size, err)
		}
		out = append(out, PriceLevel{Price: p, Size: s})
	}

	// Ensure correct ordering: ascending for BUY (asks), descending for SELL (bids).
	if side == "BUY" {
		sort.Slice(out, func(i, j int) bool { return out[i].Price < out[j].Price })
	} else {
		sort.Slice(out, func(i, j int) bool { return out[i].Price > out[j].Price })
	}
	return out, nil
}

package sweep

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

const (
	Buy  = "BUY"
	Sell = "SELL"
)

func Estimate(book *models.OrderBook, side string, maxSlippage float64) (*models.SweepEstimate, error) {
	if book == nil {
		return nil, fmt.Errorf("estimate sweep: nil order book")
	}
	levels, err := parseLevels(book, side)
	if err != nil {
		return nil, err
	}
	refPrice := bookMid(book)
	return EstimateFromLevels(levels, side, refPrice, maxSlippage)
}

func EstimateFromLevels(levels []models.PriceLevel, side string, refPrice, maxSlippage float64) (*models.SweepEstimate, error) {
	if side != Buy && side != Sell {
		return nil, fmt.Errorf("estimate sweep: invalid side %q (want BUY or SELL)", side)
	}
	if maxSlippage < 0 {
		return nil, fmt.Errorf("estimate sweep: maxSlippage must be non-negative")
	}

	est := &models.SweepEstimate{Side: side}
	if len(levels) == 0 {
		return est, nil
	}
	best := levels[0].Price
	if refPrice <= 0 {
		refPrice = best
	}
	est.BestPrice = best

	var limit float64
	if side == Buy {
		limit = refPrice * (1 + maxSlippage)
	} else {
		limit = refPrice * (1 - maxSlippage)
	}

	var vol, size float64
	for _, lvl := range levels {
		newVol := vol + lvl.Price*lvl.Size
		newSize := size + lvl.Size
		avg := newVol / newSize

		crosses := (side == Buy && avg > limit) || (side == Sell && avg < limit)
		if !crosses {
			vol, size = newVol, newSize
			est.Levels = append(est.Levels, models.SweepLevel{
				Price:    lvl.Price,
				Size:     lvl.Size,
				Slippage: relSlippage(lvl.Price, refPrice),
			})
			est.WorstPrice = lvl.Price
			continue
		}
		d := (limit*size - vol) / (lvl.Price - limit)
		if d > 0 && d <= lvl.Size {
			vol += lvl.Price * d
			size += d
			est.Levels = append(est.Levels, models.SweepLevel{
				Price:    lvl.Price,
				Size:     d,
				Slippage: relSlippage(lvl.Price, refPrice),
			})
			est.WorstPrice = lvl.Price
		}
		break
	}

	if size > 0 {
		est.TotalSize = size
		est.AvgPrice = vol / size
	}
	return est, nil
}

func BestFillablePrice(book *models.OrderBook, side string, size float64) (float64, error) {
	if book == nil {
		return 0, fmt.Errorf("best fillable price: nil order book")
	}
	if size <= 0 {
		return 0, fmt.Errorf("best fillable price: size must be positive")
	}
	levels, err := parseLevels(book, side)
	if err != nil {
		return 0, err
	}
	remaining := size
	for _, lvl := range levels {
		if lvl.Size >= remaining {
			return lvl.Price, nil
		}
		remaining -= lvl.Size
	}
	return 0, fmt.Errorf("best fillable price: book has only %.4f size available for size %.4f", size-remaining, size)
}

func relSlippage(price, ref float64) float64 {
	if ref == 0 {
		return 0
	}
	diff := price - ref
	if diff < 0 {
		diff = -diff
	}
	return diff / ref
}

func bookMid(book *models.OrderBook) float64 {
	if book == nil || len(book.Bids) == 0 || len(book.Asks) == 0 {
		return 0
	}
	bestBid, bidOK := bestPrice(book.Bids, true)
	bestAsk, askOK := bestPrice(book.Asks, false)
	if !bidOK || !askOK {
		return 0
	}
	return (bestBid + bestAsk) / 2
}

func bestPrice(raw []models.OrderBookLevel, highest bool) (float64, bool) {
	var best float64
	first := true
	for _, lvl := range raw {
		p, err := strconv.ParseFloat(lvl.Price, 64)
		if err != nil {
			continue
		}
		if first || (highest && p > best) || (!highest && p < best) {
			best = p
			first = false
		}
	}
	return best, !first
}

func parseLevels(book *models.OrderBook, side string) ([]models.PriceLevel, error) {
	var raw []models.OrderBookLevel
	switch side {
	case Buy:
		raw = book.Asks
	case Sell:
		raw = book.Bids
	default:
		return nil, fmt.Errorf("estimate sweep: invalid side %q (want BUY or SELL)", side)
	}

	out := make([]models.PriceLevel, 0, len(raw))
	for _, lvl := range raw {
		p, err := strconv.ParseFloat(lvl.Price, 64)
		if err != nil {
			return nil, fmt.Errorf("estimate sweep: parse price %q: %w", lvl.Price, err)
		}
		s, err := strconv.ParseFloat(lvl.Size, 64)
		if err != nil {
			return nil, fmt.Errorf("estimate sweep: parse size %q: %w", lvl.Size, err)
		}
		out = append(out, models.PriceLevel{Price: p, Size: s})
	}

	if side == Buy {
		sort.Slice(out, func(i, j int) bool { return out[i].Price < out[j].Price })
	} else {
		sort.Slice(out, func(i, j int) bool { return out[i].Price > out[j].Price })
	}
	return out, nil
}

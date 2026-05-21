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
	return EstimateFromLevels(levels, side, maxSlippage)
}

func EstimateFromLevels(levels []models.PriceLevel, side string, maxSlippage float64) (*models.SweepEstimate, error) {
	if len(levels) == 0 {
		return nil, fmt.Errorf("estimate sweep: no levels provided")
	}
	if side != Buy && side != Sell {
		return nil, fmt.Errorf("estimate sweep: invalid side %q (want BUY or SELL)", side)
	}
	if maxSlippage < 0 {
		return nil, fmt.Errorf("estimate sweep: maxSlippage must be non-negative")
	}

	best := levels[0].Price
	var limit float64
	if side == Buy {
		limit = best * (1 + maxSlippage)
	} else {
		limit = best * (1 - maxSlippage)
	}

	est := &models.SweepEstimate{BestPrice: best, Side: side}
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
				Slippage: relSlippage(lvl.Price, best),
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
				Slippage: relSlippage(lvl.Price, best),
			})
			est.WorstPrice = lvl.Price
		}
		break
	}

	if len(est.Levels) == 0 {
		return nil, fmt.Errorf("estimate sweep: no levels within %.4f%% slippage", maxSlippage*100)
	}
	est.TotalSize = size
	est.AvgPrice = vol / size
	return est, nil
}

func relSlippage(price, best float64) float64 {
	if best == 0 {
		return 0
	}
	diff := price - best
	if diff < 0 {
		diff = -diff
	}
	return diff / best
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

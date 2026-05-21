package sweep

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

const (
	Buy  = "BUY"
	Sell = "SELL"
)

func Estimate(book *models.OrderBook, side string, refPrice, size, maxSlippage float64) (*models.SweepEstimate, error) {
	if book == nil {
		return nil, fmt.Errorf("estimate sweep: nil order book")
	}
	levels, err := parseLevels(book, side)
	if err != nil {
		return nil, err
	}
	return EstimateFromLevels(levels, side, refPrice, size, maxSlippage)
}

func EstimateFromLevels(levels []models.PriceLevel, side string, refPrice, size, maxSlippage float64) (*models.SweepEstimate, error) {
	if len(levels) == 0 {
		return nil, fmt.Errorf("estimate sweep: no levels provided")
	}
	if side != Buy && side != Sell {
		return nil, fmt.Errorf("estimate sweep: invalid side %q (want BUY or SELL)", side)
	}

	bestPrice := levels[0].Price
	if refPrice > 0 {
		bestPrice = refPrice
	}

	est := &models.SweepEstimate{BestPrice: bestPrice, Side: side}
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
		est.Levels = append(est.Levels, models.SweepLevel{
			Price:    lvl.Price,
			Size:     fillSize,
			Slippage: slippage,
		})
		est.TotalSize += fillSize
		est.WorstPrice = lvl.Price
		costSum += lvl.Price * fillSize
		remaining -= fillSize
	}

	if len(est.Levels) == 0 {
		return nil, fmt.Errorf("estimate sweep: no levels within %.2f%% slippage", maxSlippage*100)
	}
	if est.TotalSize > 0 {
		est.AvgPrice = costSum / est.TotalSize
	}
	return est, nil
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

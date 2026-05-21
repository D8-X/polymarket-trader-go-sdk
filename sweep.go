package polytrade

import "github.com/D8-X/polymarket-trader-go-sdk/v2/internal/sweep"

func EstimateSweep(book *OrderBook, side string, refPrice, size, maxSlippage float64) (*SweepEstimate, error) {
	return sweep.Estimate(book, side, refPrice, size, maxSlippage)
}

func EstimateSweepFromLevels(levels []PriceLevel, side string, refPrice, size, maxSlippage float64) (*SweepEstimate, error) {
	return sweep.EstimateFromLevels(levels, side, refPrice, size, maxSlippage)
}

package sweep

import (
	"math"
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func approxEqual(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

func TestEstimateFromLevelsBuyWalksWholeBookWithinBudget(t *testing.T) {
	levels := []models.PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.52, Size: 10},
		{Price: 0.55, Size: 20},
	}
	est, err := EstimateFromLevels(levels, Buy, 0.5)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if est.TotalSize != 35 {
		t.Errorf("total size: got %v want 35", est.TotalSize)
	}
	if len(est.Levels) != 3 {
		t.Errorf("levels: got %d want 3", len(est.Levels))
	}
	if est.BestPrice != 0.50 || est.WorstPrice != 0.55 {
		t.Errorf("best/worst: got %v/%v want 0.50/0.55", est.BestPrice, est.WorstPrice)
	}
}

func TestEstimateFromLevelsSellWalksWholeBookWithinBudget(t *testing.T) {
	levels := []models.PriceLevel{
		{Price: 0.55, Size: 8},
		{Price: 0.52, Size: 10},
		{Price: 0.50, Size: 20},
	}
	est, err := EstimateFromLevels(levels, Sell, 0.5)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if est.TotalSize != 38 {
		t.Errorf("total size: got %v want 38", est.TotalSize)
	}
	if est.BestPrice != 0.55 || est.WorstPrice != 0.50 {
		t.Errorf("best/worst: got %v/%v want 0.55/0.50", est.BestPrice, est.WorstPrice)
	}
}

func TestEstimateFromLevelsPartialFillsCrossingLevel(t *testing.T) {
	levels := []models.PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.52, Size: 5},
		{Price: 0.60, Size: 50},
	}
	est, err := EstimateFromLevels(levels, Buy, 0.05)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	// d = (0.525*10 - 5.1) / (0.60 - 0.525) = 0.15 / 0.075 = 2
	if !approxEqual(est.TotalSize, 12, 1e-9) {
		t.Errorf("total size: got %v want 12", est.TotalSize)
	}
	if len(est.Levels) != 3 {
		t.Errorf("levels: got %d want 3", len(est.Levels))
	}
	if !approxEqual(est.Levels[2].Size, 2, 1e-9) {
		t.Errorf("partial fill size: got %v want 2", est.Levels[2].Size)
	}
	if !approxEqual(est.AvgPrice, 0.525, 1e-9) {
		t.Errorf("avg price: got %v want 0.525", est.AvgPrice)
	}
}

func TestEstimateFromLevelsBudgetTooTightForAnyLevel(t *testing.T) {
	levels := []models.PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.51, Size: 100},
	}
	est, err := EstimateFromLevels(levels, Buy, 0)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if est.TotalSize != 5 {
		t.Errorf("total size: got %v want 5", est.TotalSize)
	}
}

func TestEstimateFromLevelsNoLevels(t *testing.T) {
	if _, err := EstimateFromLevels(nil, Buy, 0.05); err == nil {
		t.Fatal("expected error for empty levels")
	}
}

func TestEstimateFromLevelsInvalidSide(t *testing.T) {
	levels := []models.PriceLevel{{Price: 0.50, Size: 5}}
	if _, err := EstimateFromLevels(levels, "BAD", 0.05); err == nil {
		t.Fatal("expected error for invalid side")
	}
}

func TestEstimateFromLevelsNegativeSlippage(t *testing.T) {
	levels := []models.PriceLevel{{Price: 0.50, Size: 5}}
	if _, err := EstimateFromLevels(levels, Buy, -0.01); err == nil {
		t.Fatal("expected error for negative slippage")
	}
}

func TestEstimateFromLevelsAvgPriceWhenFullyConsumed(t *testing.T) {
	levels := []models.PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.52, Size: 5},
	}
	est, err := EstimateFromLevels(levels, Buy, 0.5)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	want := (0.50*5 + 0.52*5) / 10
	if !approxEqual(est.AvgPrice, want, 1e-9) {
		t.Errorf("avg price: got %v want %v", est.AvgPrice, want)
	}
}

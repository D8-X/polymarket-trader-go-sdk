package polytrade

import "testing"

func TestEstimateSweepFromLevelsBuyFills(t *testing.T) {
	levels := []PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.52, Size: 10},
		{Price: 0.55, Size: 20},
	}
	est, err := EstimateSweepFromLevels(levels, BUY, 0, 12, 0.5)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if est.TotalSize != 12 {
		t.Errorf("total size: got %v want 12", est.TotalSize)
	}
	if len(est.Levels) != 2 {
		t.Errorf("levels: got %d want 2", len(est.Levels))
	}
	if est.Levels[0].Price != 0.50 || est.Levels[0].Size != 5 {
		t.Errorf("level 0: got %+v want price=0.50 size=5", est.Levels[0])
	}
	if est.Levels[1].Price != 0.52 || est.Levels[1].Size != 7 {
		t.Errorf("level 1: got %+v want price=0.52 size=7", est.Levels[1])
	}
	if est.BestPrice != 0.50 || est.WorstPrice != 0.52 {
		t.Errorf("best/worst: got %v/%v want 0.50/0.52", est.BestPrice, est.WorstPrice)
	}
}

func TestEstimateSweepFromLevelsSellFills(t *testing.T) {
	levels := []PriceLevel{
		{Price: 0.55, Size: 8},
		{Price: 0.52, Size: 10},
		{Price: 0.50, Size: 20},
	}
	est, err := EstimateSweepFromLevels(levels, SELL, 0, 12, 0.5)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if est.TotalSize != 12 {
		t.Errorf("total size: got %v want 12", est.TotalSize)
	}
	if len(est.Levels) != 2 {
		t.Errorf("levels: got %d want 2", len(est.Levels))
	}
	if est.BestPrice != 0.55 || est.WorstPrice != 0.52 {
		t.Errorf("best/worst: got %v/%v want 0.55/0.52", est.BestPrice, est.WorstPrice)
	}
}

func TestEstimateSweepFromLevelsStopsAtSlippage(t *testing.T) {
	levels := []PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.52, Size: 5},
		{Price: 0.60, Size: 50},
	}
	est, err := EstimateSweepFromLevels(levels, BUY, 0, 100, 0.05)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if est.TotalSize != 10 {
		t.Errorf("total size: got %v want 10 (only first two levels within 5%% slippage)", est.TotalSize)
	}
	if len(est.Levels) != 2 {
		t.Errorf("levels: got %d want 2", len(est.Levels))
	}
}

func TestEstimateSweepFromLevelsRefPriceOverride(t *testing.T) {
	levels := []PriceLevel{
		{Price: 0.51, Size: 5},
	}
	est, err := EstimateSweepFromLevels(levels, BUY, 0.50, 10, 0.05)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if est.BestPrice != 0.50 {
		t.Errorf("best price: got %v want 0.50 (refPrice override)", est.BestPrice)
	}
	if est.TotalSize != 5 {
		t.Errorf("total size: got %v want 5", est.TotalSize)
	}
}

func TestEstimateSweepFromLevelsNoLevels(t *testing.T) {
	if _, err := EstimateSweepFromLevels(nil, BUY, 0, 10, 0.05); err == nil {
		t.Fatal("expected error for empty levels")
	}
}

func TestEstimateSweepFromLevelsAvgPrice(t *testing.T) {
	levels := []PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.52, Size: 5},
	}
	est, err := EstimateSweepFromLevels(levels, BUY, 0, 10, 0.5)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	want := (0.50*5 + 0.52*5) / 10
	if est.AvgPrice != want {
		t.Errorf("avg price: got %v want %v", est.AvgPrice, want)
	}
}

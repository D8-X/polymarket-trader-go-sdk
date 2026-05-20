package polytrade

import "testing"

func newSweepBuilder(t *testing.T) *OrderBuilder {
	t.Helper()
	_ = testEOA(t)
	return NewOrderBuilder(testDepositWallet, CTFExchange, testPrivateKey, SignatureTypePoly1271)
}

func TestPrepareSweepFromLevelsBuyFills(t *testing.T) {
	ob := newSweepBuilder(t)
	levels := []PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.52, Size: 10},
		{Price: 0.55, Size: 20},
	}
	res, err := ob.PrepareSweepFromLevels("100", levels, BUY, OrderTypeFAK, 0, 12, 0.5, "k", OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if res.TotalSize != 12 {
		t.Errorf("total size: got %v want 12", res.TotalSize)
	}
	if len(res.Levels) != 2 {
		t.Errorf("levels: got %d want 2", len(res.Levels))
	}
	if res.Levels[0].Price != 0.50 || res.Levels[0].Size != 5 {
		t.Errorf("level 0: got %+v want price=0.50 size=5", res.Levels[0])
	}
	if res.Levels[1].Price != 0.52 || res.Levels[1].Size != 7 {
		t.Errorf("level 1: got %+v want price=0.52 size=7", res.Levels[1])
	}
	if res.BestPrice != 0.50 || res.WorstPrice != 0.52 {
		t.Errorf("best/worst: got %v/%v want 0.50/0.52", res.BestPrice, res.WorstPrice)
	}
}

func TestPrepareSweepFromLevelsSellFills(t *testing.T) {
	ob := newSweepBuilder(t)
	levels := []PriceLevel{
		{Price: 0.55, Size: 8},
		{Price: 0.52, Size: 10},
		{Price: 0.50, Size: 20},
	}
	res, err := ob.PrepareSweepFromLevels("100", levels, SELL, OrderTypeFAK, 0, 12, 0.5, "k", OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if res.TotalSize != 12 {
		t.Errorf("total size: got %v want 12", res.TotalSize)
	}
	if len(res.Levels) != 2 {
		t.Errorf("levels: got %d want 2", len(res.Levels))
	}
	if res.BestPrice != 0.55 || res.WorstPrice != 0.52 {
		t.Errorf("best/worst: got %v/%v want 0.55/0.52", res.BestPrice, res.WorstPrice)
	}
}

func TestPrepareSweepFromLevelsStopsAtSlippage(t *testing.T) {
	ob := newSweepBuilder(t)
	levels := []PriceLevel{
		{Price: 0.50, Size: 5},
		{Price: 0.52, Size: 5}, // 4% slippage
		{Price: 0.60, Size: 50}, // 20% slippage, must be skipped
	}
	res, err := ob.PrepareSweepFromLevels("100", levels, BUY, OrderTypeFAK, 0, 100, 0.05, "k", OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if res.TotalSize != 10 {
		t.Errorf("total size: got %v want 10 (only first two levels within 5%% slippage)", res.TotalSize)
	}
	if len(res.Levels) != 2 {
		t.Errorf("levels: got %d want 2", len(res.Levels))
	}
}

func TestPrepareSweepFromLevelsRefPriceOverride(t *testing.T) {
	ob := newSweepBuilder(t)
	levels := []PriceLevel{
		{Price: 0.51, Size: 5},
	}
	res, err := ob.PrepareSweepFromLevels("100", levels, BUY, OrderTypeFAK, 0.50, 10, 0.05, "k", OrderOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if res.BestPrice != 0.50 {
		t.Errorf("best price: got %v want 0.50 (refPrice override)", res.BestPrice)
	}
	if res.TotalSize != 5 {
		t.Errorf("total size: got %v want 5", res.TotalSize)
	}
}

func TestPrepareSweepFromLevelsNoLevels(t *testing.T) {
	ob := newSweepBuilder(t)
	if _, err := ob.PrepareSweepFromLevels("100", nil, BUY, OrderTypeFAK, 0, 10, 0.05, "k"); err == nil {
		t.Fatal("expected error for empty levels")
	}
}


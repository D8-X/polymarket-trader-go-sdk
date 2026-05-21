package models

type PriceLevel struct {
	Price float64
	Size  float64
}

type SweepLevel struct {
	Price    float64
	Size     float64
	Slippage float64
}

type SweepEstimate struct {
	Levels     []SweepLevel
	Side       string
	BestPrice  float64
	WorstPrice float64
	AvgPrice   float64
	TotalSize  float64
}

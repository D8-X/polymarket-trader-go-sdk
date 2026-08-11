package models

type PositionEntry struct {
	Asset             string  `json:"asset"`
	ConditionID       string  `json:"conditionId"`
	Size              float64 `json:"size"`
	AvgPrice          float64 `json:"avgPrice"`
	CurPrice          float64 `json:"curPrice"`
	Outcome           string  `json:"outcome"`
	Title             string  `json:"title"`
	InitialValue      float64 `json:"initialValue"`      // excludes fees
	GrossInitialValue float64 `json:"grossInitialValue"` // includes fees, ~1e-4 rounding
	EntryFeesUsdc     float64 `json:"entryFeesUsdc"`
}

type PositionsOpts struct {
	IncludeArchived bool     // excluded by default
	SizeThreshold   *float64 // nil or neg means 0
	Limit           *int     // nil, 0 or neg leaves the server default of 100, max is 500.
	Offset          *int     // nil or neg leaves the server default of 0.
}

type BalanceEntry struct {
	AssetID string  `json:"asset_id"`
	Balance float64 `json:"balance"`
}

type BalanceAllowanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances"`
}

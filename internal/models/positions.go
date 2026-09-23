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
	Redeemable        bool    `json:"redeemable"` // market resolved
	Mergeable         bool    `json:"mergeable"`  // both outcomes held
}

type PositionsOpts struct {
	Limit           int     // required, 1 to 500
	Offset          int     // 0 to 10000
	SizeThreshold   float64 // neg means 0
	IncludeArchived bool    // excluded by default
	Redeemable      *bool   // nil means both, false means open positions only
	Mergeable       *bool   // nil means both
}

type BalanceEntry struct {
	AssetID string  `json:"asset_id"`
	Balance float64 `json:"balance"`
}

type BalanceAllowanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances"`
}

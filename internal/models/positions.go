package models

type PositionEntry struct {
	Asset       string  `json:"asset"`
	ConditionID string  `json:"conditionId"`
	Size        float64 `json:"size"`
	AvgPrice    float64 `json:"avgPrice"`
	CurPrice    float64 `json:"curPrice"`
	Outcome     string  `json:"outcome"`
	Title       string  `json:"title"`
}

type BalanceEntry struct {
	AssetID string  `json:"asset_id"`
	Balance float64 `json:"balance"`
}

type BalanceAllowanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances"`
}

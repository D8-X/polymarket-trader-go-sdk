package models

type PositionStatus string

const (
	PositionStatusOpen           PositionStatus = "OPEN"
	PositionStatusRedeemable     PositionStatus = "REDEEMABLE"      // resolved, winners and losers
	PositionStatusRedeemableLost PositionStatus = "REDEEMABLE_LOST" // resolved losers only, rows report REDEEMABLE
	PositionStatusMergeable      PositionStatus = "MERGEABLE"       // both outcomes held
	PositionStatusClosed         PositionStatus = "CLOSED"          // exited
)

type PositionsSort string

const (
	PositionsSortCurrentValue  PositionsSort = "CURRENT_VALUE"
	PositionsSortPrice         PositionsSort = "PRICE"
	PositionsSortTokens        PositionsSort = "TOKENS"
	PositionsSortUnrealizedPnl PositionsSort = "UNREALIZED_PNL"
	PositionsSortRealizedPnl   PositionsSort = "REALIZED_PNL"
	PositionsSortTotalPnl      PositionsSort = "TOTAL_PNL"
	PositionsSortTimestamp     PositionsSort = "TIMESTAMP"
)

type PositionEntry struct {
	Asset             string         `json:"token_id"`
	ConditionID       string         `json:"condition_id"`
	Size              float64        `json:"current_size"`
	AvgPrice          float64        `json:"avg_price"`
	CurPrice          float64        `json:"current_price"`
	CurrentValue      float64        `json:"current_value"`
	Outcome           string         `json:"outcome"`
	OutcomeIndex      int            `json:"outcome_index"` // 999 when unlabeled
	OppositeAsset     string         `json:"opposite_token_id"`
	Title             string         `json:"title"`
	Slug              string         `json:"slug"`
	EventID           string         `json:"event_id"`
	EndDate           string         `json:"end_date"`
	InitialValue      float64        `json:"entry_cost_usdc"` // excludes fees
	GrossInitialValue float64        `json:"total_cost_usdc"` // includes fees
	EntryFeesUsdc     float64        `json:"entry_fees_usdc"`
	RealizedPnl       float64        `json:"realized_pnl"`
	UnrealizedPnl     float64        `json:"unrealized_pnl"`
	Status            PositionStatus `json:"status"`
	Redeemable        bool           `json:"redeemable"` // resolved, not necessarily won
	Mergeable         bool           `json:"mergeable"`
	NegativeRisk      bool           `json:"negative_risk"`
	Archived          bool           `json:"archived"`
}

type PositionsOpts struct {
	Limit           int            // up to 1000, 0 means the server default of 100
	Cursor          string         // NextCursor of the previous page, empty for the first
	Status          PositionStatus // empty means OPEN
	ConditionIDs    []string       // at most 20
	Title           string         // case insensitive substring
	MinSize         float64        // shares, the server never goes below 0.1
	IncludeArchived bool           // not allowed with CLOSED
	SortBy          PositionsSort  // empty follows the status
	Ascending       bool
}

type PositionsPage struct {
	Positions  []PositionEntry     `json:"data"`
	Pagination PositionsPagination `json:"pagination"`
}

type PositionsPagination struct {
	NextCursor string `json:"next_cursor"` // empty on the last page
	HasMore    bool   `json:"has_more"`
}

type BalanceEntry struct {
	AssetID string  `json:"asset_id"`
	Balance float64 `json:"balance"`
}

type BalanceAllowanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances"`
}

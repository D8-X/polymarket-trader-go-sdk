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

// PositionEntry keeps the v1 /positions JSON names so stored or relayed entries stay compatible.
type PositionEntry struct {
	Asset             string         `json:"asset"`
	ConditionID       string         `json:"conditionId"`
	Size              float64        `json:"size"`
	AvgPrice          float64        `json:"avgPrice"`
	CurPrice          float64        `json:"curPrice"`
	Outcome           string         `json:"outcome"`
	Title             string         `json:"title"`
	InitialValue      float64        `json:"initialValue"`      // excludes fees
	GrossInitialValue float64        `json:"grossInitialValue"` // includes fees
	EntryFeesUsdc     float64        `json:"entryFeesUsdc"`
	CurrentValue      float64        `json:"currentValue"`
	RealizedPnl       float64        `json:"realizedPnl"`
	UnrealizedPnl     float64        `json:"unrealizedPnl"`
	OutcomeIndex      int            `json:"outcomeIndex"` // 999 when unlabeled
	OppositeAsset     string         `json:"oppositeAsset"`
	Slug              string         `json:"slug"`
	EventID           string         `json:"eventId"`
	EndDate           string         `json:"endDate"`
	Status            PositionStatus `json:"status"`
	Redeemable        bool           `json:"redeemable"` // resolved, not necessarily won
	Mergeable         bool           `json:"mergeable"`
	NegativeRisk      bool           `json:"negativeRisk"`
	Archived          bool           `json:"archived"`
}

type PositionsOpts struct {
	Limit           int            // page size up to 1000, 0 means 100 for a page and 1000 for a full walk
	Cursor          string         // NextCursor of the previous page, empty to start from the top
	Status          PositionStatus // empty means OPEN
	ConditionIDs    []string       // at most 20
	Title           string         // case insensitive substring
	MinSize         float64        // shares, the server never goes below 0.1
	IncludeArchived bool           // not allowed with CLOSED
	SortBy          PositionsSort  // empty follows the status
	Ascending       bool
}

type PositionsPage struct {
	Positions  []PositionEntry     `json:"positions"`
	Pagination PositionsPagination `json:"pagination"`
}

type PositionsPagination struct {
	NextCursor string `json:"nextCursor"` // empty on the last page
	HasMore    bool   `json:"hasMore"`
}

type BalanceEntry struct {
	AssetID string  `json:"asset_id"`
	Balance float64 `json:"balance"`
}

type BalanceAllowanceResponse struct {
	Balance    string            `json:"balance"`
	Allowances map[string]string `json:"allowances"`
}

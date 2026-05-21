package models

type MarketRewardConfig struct {
	AssetAddress string  `json:"asset_address"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	RatePerDay   float64 `json:"rate_per_day"`
	TotalRewards float64 `json:"total_rewards"`
	ID           int     `json:"id"`
}

type CurrentRewardMarket struct {
	ConditionID      string               `json:"condition_id"`
	RewardsConfig    []MarketRewardConfig `json:"rewards_config"`
	RewardsMaxSpread float64              `json:"rewards_max_spread"`
	RewardsMinSize   float64              `json:"rewards_min_size"`
	NativeDailyRate  float64              `json:"native_daily_rate"`
	TotalDailyRate   float64              `json:"total_daily_rate"`
}

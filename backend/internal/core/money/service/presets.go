package service

// ContributionPreset is a ready-made Canadian registered account setup.
// Figures are CRA dollar limits as of 2026; adjust them in the UI if they
// change.
type ContributionPreset struct {
	Key                    string           `json:"key"`
	Name                   string           `json:"name"`
	Description            string           `json:"description"`
	EligibilityAge         *int             `json:"eligibility_age"`
	CarryForwardCap        *float64         `json:"carry_forward_cap"`
	LifetimeLimit          *float64         `json:"lifetime_limit"`
	WithdrawalsRestoreRoom bool             `json:"withdrawals_restore_room"`
	NeedsStartYear         bool             `json:"needs_start_year"`
	SuggestOpeningRoom     bool             `json:"suggest_opening_room"`
	Limits                 []PresetLimitRow `json:"limits"`
}

type PresetLimitRow struct {
	Year        int     `json:"year"`
	AnnualLimit float64 `json:"annual_limit"`
}

func intPtr(v int) *int           { return &v }
func floatPtr(v float64) *float64 { return &v }

var contributionPresets = []ContributionPreset{
	{
		Key:  "tfsa",
		Name: "TFSA",
		Description: "Room accrues every year from 2009 (or the year you turn 18). " +
			"Unused room carries forward forever and withdrawals are added back the next January.",
		EligibilityAge:         intPtr(18),
		WithdrawalsRestoreRoom: true,
		Limits: []PresetLimitRow{
			{Year: 2009, AnnualLimit: 5_000},
			{Year: 2013, AnnualLimit: 5_500},
			{Year: 2015, AnnualLimit: 10_000},
			{Year: 2016, AnnualLimit: 5_500},
			{Year: 2019, AnnualLimit: 6_000},
			{Year: 2023, AnnualLimit: 6_500},
			{Year: 2024, AnnualLimit: 7_000},
		},
	},
	{
		Key:  "fhsa",
		Name: "FHSA",
		Description: "$8,000 of room per year starting the year you open your first FHSA. " +
			"Up to $8,000 of unused room carries forward, with a $40,000 lifetime limit.",
		EligibilityAge:  intPtr(18),
		CarryForwardCap: floatPtr(8_000),
		LifetimeLimit:   floatPtr(40_000),
		NeedsStartYear:  true,
		Limits: []PresetLimitRow{
			{Year: 2023, AnnualLimit: 8_000},
		},
	},
	{
		Key:  "rrsp",
		Name: "RRSP",
		Description: "Room depends on your income. Enter the deduction limit from your latest " +
			"notice of assessment as opening room, then add each year's new room as it's assessed.",
		SuggestOpeningRoom: true,
		Limits:             []PresetLimitRow{},
	},
}

func ListContributionPresets() []ContributionPreset {
	return contributionPresets
}

func findContributionPreset(key string) (ContributionPreset, bool) {
	for _, preset := range contributionPresets {
		if preset.Key == key {
			return preset, true
		}
	}
	return ContributionPreset{}, false
}

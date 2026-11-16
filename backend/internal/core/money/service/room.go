package service

import (
	"math"
	"sort"

	"be-simpletracker/internal/core/money/models"
)

// RoomRules is everything needed to calculate contribution room for a
// contribution type. It is kept independent of the database so the
// calculation can be tested directly.
type RoomRules struct {
	StartYear              int
	Limits                 []models.ContributionLimit
	CarryForwardCap        *float64
	LifetimeLimit          *float64
	WithdrawalsRestoreRoom bool
	OpeningRoomYear        *int
	OpeningRoom            *float64
}

// YearFlow is the money moved into and out of all accounts of a contribution
// type during one calendar year.
type YearFlow struct {
	Contributed float64
	Withdrawn   float64
}

// RoomYear explains how the room for one year was built up.
type RoomYear struct {
	Year int `json:"year"`
	// NewRoom is the room granted this year (the annual limit, or the opening
	// room in the opening year).
	NewRoom float64 `json:"new_room"`
	// LimitYear is the year of the limit rule that produced NewRoom, so the
	// UI can show when a limit is inherited from an earlier year.
	LimitYear       *int    `json:"limit_year"`
	IsOpening       bool    `json:"is_opening"`
	CarriedIn       float64 `json:"carried_in"`
	Restored        float64 `json:"restored"`
	LifetimeCapped  float64 `json:"lifetime_capped"`
	Available       float64 `json:"available"`
	Contributed     float64 `json:"contributed"`
	Withdrawn       float64 `json:"withdrawn"`
	Remaining       float64 `json:"remaining"`
	CarriedOut      float64 `json:"carried_out"`
	OverContributed bool    `json:"over_contributed"`
}

type RoomSummary struct {
	StartYear int `json:"start_year"`
	// Available is the room left to contribute in the current year.
	Available           float64    `json:"available"`
	ContributedThisYear float64    `json:"contributed_this_year"`
	WithdrawnThisYear   float64    `json:"withdrawn_this_year"`
	LifetimeContributed float64    `json:"lifetime_contributed"`
	LifetimeRemaining   *float64   `json:"lifetime_remaining"`
	OverContributed     bool       `json:"over_contributed"`
	Years               []RoomYear `json:"years"`
}

// limitForYear returns the most recent limit at or before year.
func limitForYear(limits []models.ContributionLimit, year int) (models.ContributionLimit, bool) {
	var best models.ContributionLimit
	found := false
	for _, limit := range limits {
		if limit.Year <= year && (!found || limit.Year > best.Year) {
			best = limit
			found = true
		}
	}
	return best, found
}

func roundCents(value float64) float64 {
	return math.Round(value*100) / 100
}

// CalculateRoom walks every year from the start year through currentYear and
// returns the year-by-year breakdown (newest first).
//
// Each year: available = carried in + new room + restored withdrawals,
// capped so lifetime contributions cannot exceed the lifetime limit. Unused
// room carries forward (subject to the carry-forward cap); an
// over-contribution carries forward as negative room.
func CalculateRoom(rules RoomRules, flows map[int]YearFlow, currentYear int) RoomSummary {
	startYear := rules.StartYear
	opening := rules.OpeningRoomYear != nil && rules.OpeningRoom != nil
	if opening {
		startYear = *rules.OpeningRoomYear
	}
	summary := RoomSummary{StartYear: startYear, Years: []RoomYear{}}

	years := make([]int, 0, len(flows))
	for year := range flows {
		years = append(years, year)
	}
	sort.Ints(years)
	for _, year := range years {
		if year < startYear {
			summary.LifetimeContributed += flows[year].Contributed
		}
	}

	carried := 0.0
	for year := startYear; year <= currentYear; year++ {
		row := RoomYear{Year: year, CarriedIn: carried}
		if opening && year == startYear {
			row.NewRoom = *rules.OpeningRoom
			row.IsOpening = true
		} else if limit, ok := limitForYear(rules.Limits, year); ok {
			row.NewRoom = limit.AnnualLimit
			limitYear := limit.Year
			row.LimitYear = &limitYear
		}
		if rules.WithdrawalsRestoreRoom && year > startYear {
			row.Restored = flows[year-1].Withdrawn
		}
		available := row.CarriedIn + row.NewRoom + row.Restored
		if rules.LifetimeLimit != nil {
			lifetimeLeft := *rules.LifetimeLimit - summary.LifetimeContributed
			if available > lifetimeLeft {
				row.LifetimeCapped = available - lifetimeLeft
				available = lifetimeLeft
			}
		}
		flow := flows[year]
		row.Available = roundCents(available)
		row.Contributed = flow.Contributed
		row.Withdrawn = flow.Withdrawn
		row.Remaining = roundCents(available - flow.Contributed)
		row.OverContributed = row.Remaining < 0

		switch {
		case row.Remaining < 0:
			carried = row.Remaining
		case rules.CarryForwardCap != nil:
			carried = math.Min(row.Remaining, *rules.CarryForwardCap)
		default:
			carried = row.Remaining
		}
		row.CarriedOut = carried
		summary.LifetimeContributed += flow.Contributed
		summary.Years = append(summary.Years, row)
	}

	if n := len(summary.Years); n > 0 {
		current := summary.Years[n-1]
		if current.Year == currentYear {
			summary.Available = current.Remaining
			summary.ContributedThisYear = current.Contributed
			summary.WithdrawnThisYear = current.Withdrawn
			summary.OverContributed = current.OverContributed
		}
	}
	if rules.LifetimeLimit != nil {
		remaining := roundCents(*rules.LifetimeLimit - summary.LifetimeContributed)
		summary.LifetimeRemaining = &remaining
	}
	summary.LifetimeContributed = roundCents(summary.LifetimeContributed)

	// Newest year first reads better in the UI.
	for i, j := 0, len(summary.Years)-1; i < j; i, j = i+1, j-1 {
		summary.Years[i], summary.Years[j] = summary.Years[j], summary.Years[i]
	}
	return summary
}

// ResolveStartYear picks the first year room accrues for a contribution type.
// The opening-room year wins when set; otherwise the latest of the explicit
// start year and the year the user reaches the eligibility age; otherwise the
// first limit year. Returns false when room cannot be calculated, including
// when an eligibility age is set but the birth year is unknown.
func ResolveStartYear(contributionType models.ContributionType, birthYear *int) (int, bool) {
	if contributionType.OpeningRoomYear != nil && contributionType.OpeningRoom != nil {
		return *contributionType.OpeningRoomYear, true
	}
	// Guessing without the birth year could overstate room by years.
	if contributionType.EligibilityAge != nil && birthYear == nil && contributionType.StartYear == nil {
		return 0, false
	}
	start, ok := 0, false
	if contributionType.StartYear != nil {
		start, ok = *contributionType.StartYear, true
	}
	if contributionType.EligibilityAge != nil && birthYear != nil {
		eligible := *birthYear + *contributionType.EligibilityAge
		if !ok || eligible > start {
			start, ok = eligible, true
		}
	}
	firstLimit, hasLimit := 0, false
	for _, limit := range contributionType.Limits {
		if !hasLimit || limit.Year < firstLimit {
			firstLimit, hasLimit = limit.Year, true
		}
	}
	if !ok {
		return firstLimit, hasLimit
	}
	// Room cannot accrue before the first limit exists.
	if hasLimit && firstLimit > start {
		start = firstLimit
	}
	return start, true
}

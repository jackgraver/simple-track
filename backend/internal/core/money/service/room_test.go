package service

import (
	"testing"

	"be-simpletracker/internal/core/money/models"
)

func limits(rows ...[2]float64) []models.ContributionLimit {
	out := make([]models.ContributionLimit, 0, len(rows))
	for _, row := range rows {
		out = append(out, models.ContributionLimit{Year: int(row[0]), AnnualLimit: row[1]})
	}
	return out
}

func yearRow(t *testing.T, summary RoomSummary, year int) RoomYear {
	t.Helper()
	for _, row := range summary.Years {
		if row.Year == year {
			return row
		}
	}
	t.Fatalf("no row for year %d", year)
	return RoomYear{}
}

func TestTFSAWithdrawalsRestoreNextYear(t *testing.T) {
	summary := CalculateRoom(RoomRules{
		StartYear:              2024,
		Limits:                 limits([2]float64{2024, 7000}),
		WithdrawalsRestoreRoom: true,
	}, map[int]YearFlow{
		2024: {Contributed: 5000, Withdrawn: 1000},
	}, 2026)

	if got := yearRow(t, summary, 2024).Remaining; got != 2000 {
		t.Fatalf("2024 remaining = %v, want 2000", got)
	}
	y2025 := yearRow(t, summary, 2025)
	if y2025.Restored != 1000 || y2025.Available != 10000 {
		t.Fatalf("2025 = %+v, want restored 1000 and available 10000", y2025)
	}
	if summary.Available != 17000 {
		t.Fatalf("available = %v, want 17000", summary.Available)
	}
	if summary.Years[0].Year != 2026 {
		t.Fatalf("years should be newest first, got %d", summary.Years[0].Year)
	}
}

func TestWithdrawalsDoNotRestoreWhenDisabled(t *testing.T) {
	summary := CalculateRoom(RoomRules{
		StartYear: 2025,
		Limits:    limits([2]float64{2025, 7000}),
	}, map[int]YearFlow{2025: {Contributed: 7000, Withdrawn: 7000}}, 2026)
	if summary.Available != 7000 {
		t.Fatalf("available = %v, want 7000", summary.Available)
	}
}

func TestFHSACarryForwardCapAndLifetime(t *testing.T) {
	cap := 8000.0
	lifetime := 40000.0
	summary := CalculateRoom(RoomRules{
		StartYear:       2023,
		Limits:          limits([2]float64{2023, 8000}),
		CarryForwardCap: &cap,
		LifetimeLimit:   &lifetime,
	}, map[int]YearFlow{
		2025: {Contributed: 16000},
	}, 2026)

	if got := yearRow(t, summary, 2024).Available; got != 16000 {
		t.Fatalf("2024 available = %v, want 16000", got)
	}
	y2024 := yearRow(t, summary, 2024)
	if y2024.CarriedOut != 8000 {
		t.Fatalf("2024 carried out = %v, want 8000 (capped)", y2024.CarriedOut)
	}
	if got := yearRow(t, summary, 2025).Available; got != 16000 {
		t.Fatalf("2025 available = %v, want 16000", got)
	}
	if summary.Available != 8000 {
		t.Fatalf("2026 available = %v, want 8000", summary.Available)
	}
	if summary.LifetimeRemaining == nil || *summary.LifetimeRemaining != 24000 {
		t.Fatalf("lifetime remaining = %v, want 24000", summary.LifetimeRemaining)
	}
}

func TestLifetimeLimitCapsAvailableRoom(t *testing.T) {
	lifetime := 10000.0
	summary := CalculateRoom(RoomRules{
		StartYear:     2025,
		Limits:        limits([2]float64{2025, 8000}),
		LifetimeLimit: &lifetime,
	}, map[int]YearFlow{2025: {Contributed: 8000}}, 2026)
	current := yearRow(t, summary, 2026)
	if current.Available != 2000 || current.LifetimeCapped != 6000 {
		t.Fatalf("2026 = %+v, want available 2000 capped 6000", current)
	}
}

func TestOverContributionCarriesForwardAsNegativeRoom(t *testing.T) {
	summary := CalculateRoom(RoomRules{
		StartYear: 2025,
		Limits:    limits([2]float64{2025, 7000}),
	}, map[int]YearFlow{2025: {Contributed: 8000}}, 2026)
	y2025 := yearRow(t, summary, 2025)
	if !y2025.OverContributed || y2025.Remaining != -1000 {
		t.Fatalf("2025 = %+v, want over-contributed by 1000", y2025)
	}
	if summary.Available != 6000 || summary.OverContributed {
		t.Fatalf("summary = %+v, want 6000 available and no current over-contribution", summary)
	}
}

func TestOpeningRoomAnchorsCalculation(t *testing.T) {
	year := 2025
	room := 20000.0
	summary := CalculateRoom(RoomRules{
		StartYear:       2009,
		Limits:          limits([2]float64{2009, 5000}, [2]float64{2024, 7000}),
		OpeningRoomYear: &year,
		OpeningRoom:     &room,
	}, map[int]YearFlow{
		2020: {Contributed: 50000},
		2025: {Contributed: 5000},
	}, 2026)
	if summary.StartYear != 2025 || len(summary.Years) != 2 {
		t.Fatalf("summary should start at the opening year, got %+v", summary)
	}
	opening := yearRow(t, summary, 2025)
	if !opening.IsOpening || opening.Remaining != 15000 {
		t.Fatalf("2025 = %+v, want opening row with 15000 remaining", opening)
	}
	if summary.Available != 22000 {
		t.Fatalf("available = %v, want 22000", summary.Available)
	}
	if summary.LifetimeContributed != 55000 {
		t.Fatalf("lifetime contributed = %v, want 55000 (includes pre-opening)", summary.LifetimeContributed)
	}
}

func TestLimitAppliesUntilReplaced(t *testing.T) {
	summary := CalculateRoom(RoomRules{
		StartYear: 2022,
		Limits:    limits([2]float64{2024, 7000}, [2]float64{2022, 6000}),
	}, nil, 2026)
	y2023 := yearRow(t, summary, 2023)
	if y2023.NewRoom != 6000 || y2023.LimitYear == nil || *y2023.LimitYear != 2022 {
		t.Fatalf("2023 = %+v, want 6000 inherited from 2022", y2023)
	}
	if summary.Available != 6000*2+7000*3 {
		t.Fatalf("available = %v, want 33000", summary.Available)
	}
}

func TestResolveStartYear(t *testing.T) {
	tfsaLimits := limits([2]float64{2009, 5000})
	birth := 2000
	age := 18
	start := 2020
	openingYear := 2025
	openingRoom := 1000.0

	cases := []struct {
		name      string
		input     models.ContributionType
		birthYear *int
		want      int
		ok        bool
	}{
		{"eligibility from birth year", models.ContributionType{EligibilityAge: &age, Limits: tfsaLimits}, &birth, 2018, true},
		{"eligibility age needs birth year", models.ContributionType{EligibilityAge: &age, Limits: tfsaLimits}, nil, 0, false},
		{"explicit start without birth year", models.ContributionType{StartYear: &start, EligibilityAge: &age, Limits: tfsaLimits}, nil, 2020, true},
		{"no eligibility uses first limit", models.ContributionType{Limits: tfsaLimits}, nil, 2009, true},
		{"explicit start later than eligibility", models.ContributionType{StartYear: &start, EligibilityAge: &age, Limits: tfsaLimits}, &birth, 2020, true},
		{"cannot start before first limit", models.ContributionType{EligibilityAge: &age, Limits: tfsaLimits}, intPtr(1980), 2009, true},
		{"opening room wins", models.ContributionType{StartYear: &start, OpeningRoomYear: &openingYear, OpeningRoom: &openingRoom}, nil, 2025, true},
		{"nothing configured", models.ContributionType{}, nil, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ResolveStartYear(tc.input, tc.birthYear)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("got (%d, %v), want (%d, %v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestAnnualLimitReportedIncludingOpeningYear(t *testing.T) {
	openingYear, openingRoom := 2025, 30000.0
	summary := CalculateRoom(RoomRules{
		Limits:          limits([2]float64{2009, 5000}, [2]float64{2024, 7000}),
		OpeningRoomYear: &openingYear,
		OpeningRoom:     &openingRoom,
	}, nil, 2026)
	for _, year := range []int{2025, 2026} {
		row := yearRow(t, summary, year)
		if row.AnnualLimit == nil || *row.AnnualLimit != 7000 {
			t.Fatalf("%d annual limit = %v, want 7000", year, row.AnnualLimit)
		}
	}
	if row := yearRow(t, summary, 2025); row.NewRoom != openingRoom {
		t.Fatalf("2025 new room = %v, want opening room %v", row.NewRoom, openingRoom)
	}

	noLimits := CalculateRoom(RoomRules{OpeningRoomYear: &openingYear, OpeningRoom: &openingRoom}, nil, 2026)
	if row := yearRow(t, noLimits, 2026); row.AnnualLimit != nil {
		t.Fatalf("annual limit = %v, want nil without limit rules", *row.AnnualLimit)
	}
}

package weight

import (
	"be-simpletracker/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const renphoHeader = "No.,Date,Time,Weight(lb),BMI,Body Fat Percentage(%),Body Fat Mass(lb),Muscle Percentage(%),Muscle Mass(lb),Skeletal Muscle Percentage(%),Skeletal Muscle Mass(lb),Bone Percentage(%),Bone Mass(lb),Protein Percentage(%),Protein Mass(lb),Body Water Percentage(%),Body Water Mass(lb),Fat-Free Mass(lb),Subcutaneous Fat(%),Visceral Fat,BMR(kcal),Metabolic Age,WHR (Waist-to-Hip Ratio),Optimal Weight(lb),Weight Level,Body Type,Target to optimal weight(lb),Target to optimal muscle mass(lb),Target to optimal fat mass(lb),Remarks\n"

const renphoRow = "1,2026.11.19,09:01:08,164.2,24.3,16.6,27.2,79.2,130.0,53.9,88.4,4.2,6.8,19.0,31.2,60.2,98.8,137.0,14.4,7,1711,22,--,--,--,--,--,--,--,--\n"

func TestParseRenphoCSV(t *testing.T) {
	logs, err := ParseRenphoCSV(renphoHeader + renphoRow)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("got %d logs", len(logs))
	}
	l := logs[0]
	if l.Date.Format("2006-01-02") != "2026-11-19" || l.WeightLbs != 164.2 {
		t.Fatalf("bad date/weight: %v %v", l.Date, l.WeightLbs)
	}
	if l.MeasuredAt == nil || l.MeasuredAt.Format("15:04:05") != "09:01:08" {
		t.Fatalf("bad measured_at: %v", l.MeasuredAt)
	}
	if l.BodyFatPct == nil || *l.BodyFatPct != 16.6 || l.BMRKcal == nil || *l.BMRKcal != 1711 ||
		l.MetabolicAge == nil || *l.MetabolicAge != 22 || l.VisceralFat == nil || *l.VisceralFat != 7 {
		t.Fatalf("metrics not parsed: %+v", l)
	}
	if l.WHR != nil || l.WeightLevel != nil || l.Remarks != nil || l.OptimalWeightLbs != nil {
		t.Fatalf("`--` should be nil: %+v", l)
	}
}

func TestParseRenphoCSVTabsAndLatestPerDay(t *testing.T) {
	text := renphoHeader + renphoRow + "2,2026.11.19,18:00:00,165.0,24.4,,,,,,,,,,,,,,,,,,,,,,,,,\n"
	text = strings.ReplaceAll(text, ",", "\t")
	logs, err := ParseRenphoCSV(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].WeightLbs != 165.0 {
		t.Fatalf("expected latest-of-day 165.0, got %+v", logs)
	}
}

func TestParseRenphoCSVErrors(t *testing.T) {
	for name, text := range map[string]string{
		"empty":    "",
		"no rows":  renphoHeader,
		"kg":       "Date,Time,Weight(kg)\n2026.11.19,09:00:00,70\n",
		"bad date": "Date,Weight(lb)\n11/19/2026,160\n",
		"bad wt":   "Date,Weight(lb)\n2026.11.19,abc\n",
	} {
		if _, err := ParseRenphoCSV(text); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestImportBodyWeightsUpserts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&BodyWeightLog{}); err != nil {
		t.Fatal(err)
	}
	logs, _ := ParseRenphoCSV(renphoHeader + renphoRow)
	if c, u, err := ImportBodyWeights(database.WithDB(context.Background(), db), logs); err != nil || c != 1 || u != 0 {
		t.Fatalf("first import: %d %d %v", c, u, err)
	}
	// Re-import with a later weight and without body fat: row is replaced, not duplicated.
	logs2, _ := ParseRenphoCSV("Date,Weight(lb)\n2026.11.19,163.0\n")
	if c, u, err := ImportBodyWeights(database.WithDB(context.Background(), db), logs2); err != nil || c != 0 || u != 1 {
		t.Fatalf("second import: %d %d %v", c, u, err)
	}
	rows, _ := ListBodyWeights(database.WithDB(context.Background(), db), 0)
	if len(rows) != 1 || rows[0].WeightLbs != 163.0 || rows[0].BodyFatPct != nil {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestParseRenphoCSVDayCap(t *testing.T) {
	days := func(n, perDay int) string {
		var b strings.Builder
		b.WriteString("Date,Time,Weight(lb)\n")
		start := time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local)
		for d := 0; d < n; d++ {
			for k := 0; k < perDay; k++ {
				fmt.Fprintf(&b, "%s,0%d:00:00,164.2\n", start.AddDate(0, 0, d).Format("2006.01.02"), k)
			}
		}
		return b.String()
	}
	// Several weigh-ins a day collapse first, so only days count.
	if logs, err := ParseRenphoCSV(days(MaxImportRows, 3)); err != nil || len(logs) != MaxImportRows {
		t.Fatalf("%d days x 3: %d logs, err %v", MaxImportRows, len(logs), err)
	}
	_, err := ParseRenphoCSV(days(MaxImportRows+1, 1))
	if err == nil || !strings.Contains(err.Error(), "too many days") {
		t.Fatalf("%d days: err = %v, want too many days", MaxImportRows+1, err)
	}
}

// A full Renpho export at the cap, pasted tab-separated (each tab is 2 bytes in JSON), fits the
// route's 5 MiB body limit (cmd/server bodyLimits).
func TestRenphoExportAtCapFitsBodyLimit(t *testing.T) {
	tsv := strings.ReplaceAll(renphoHeader+strings.Repeat(renphoRow, MaxImportRows), ",", "\t")
	body, err := json.Marshal(importWeightsBody{CSV: tsv})
	if err != nil {
		t.Fatal(err)
	}
	if len(body) > 5<<20 {
		t.Fatalf("request body for %d rows is %d bytes, over 5 MiB", MaxImportRows, len(body))
	}
}

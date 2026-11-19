package weight

import (
	"strings"
	"testing"

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
	if c, u, err := ImportBodyWeights(db, logs); err != nil || c != 1 || u != 0 {
		t.Fatalf("first import: %d %d %v", c, u, err)
	}
	// Re-import with a later weight and without body fat: row is replaced, not duplicated.
	logs2, _ := ParseRenphoCSV("Date,Weight(lb)\n2026.11.19,163.0\n")
	if c, u, err := ImportBodyWeights(db, logs2); err != nil || c != 0 || u != 1 {
		t.Fatalf("second import: %d %d %v", c, u, err)
	}
	rows, _ := ListBodyWeights(db, 0)
	if len(rows) != 1 || rows[0].WeightLbs != 163.0 || rows[0].BodyFatPct != nil {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

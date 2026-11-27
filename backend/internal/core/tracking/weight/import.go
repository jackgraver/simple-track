package weight

import (
	"be-simpletracker/internal/database"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type numField func(*BodyWeightLog) **float64
type textField func(*BodyWeightLog) **string

// Renpho CSV header -> destination field. Headers are matched case-insensitively.
var numColumns = map[string]numField{
	"bmi":                               func(l *BodyWeightLog) **float64 { return &l.BMI },
	"body fat percentage(%)":            func(l *BodyWeightLog) **float64 { return &l.BodyFatPct },
	"body fat mass(lb)":                 func(l *BodyWeightLog) **float64 { return &l.BodyFatMassLbs },
	"muscle percentage(%)":              func(l *BodyWeightLog) **float64 { return &l.MusclePct },
	"muscle mass(lb)":                   func(l *BodyWeightLog) **float64 { return &l.MuscleMassLbs },
	"skeletal muscle percentage(%)":     func(l *BodyWeightLog) **float64 { return &l.SkeletalMusclePct },
	"skeletal muscle mass(lb)":          func(l *BodyWeightLog) **float64 { return &l.SkeletalMuscleMassLbs },
	"bone percentage(%)":                func(l *BodyWeightLog) **float64 { return &l.BonePct },
	"bone mass(lb)":                     func(l *BodyWeightLog) **float64 { return &l.BoneMassLbs },
	"protein percentage(%)":             func(l *BodyWeightLog) **float64 { return &l.ProteinPct },
	"protein mass(lb)":                  func(l *BodyWeightLog) **float64 { return &l.ProteinMassLbs },
	"body water percentage(%)":          func(l *BodyWeightLog) **float64 { return &l.BodyWaterPct },
	"body water mass(lb)":               func(l *BodyWeightLog) **float64 { return &l.BodyWaterMassLbs },
	"fat-free mass(lb)":                 func(l *BodyWeightLog) **float64 { return &l.FatFreeMassLbs },
	"subcutaneous fat(%)":               func(l *BodyWeightLog) **float64 { return &l.SubcutaneousFatPct },
	"visceral fat":                      func(l *BodyWeightLog) **float64 { return &l.VisceralFat },
	"bmr(kcal)":                         func(l *BodyWeightLog) **float64 { return &l.BMRKcal },
	"metabolic age":                     func(l *BodyWeightLog) **float64 { return &l.MetabolicAge },
	"whr (waist-to-hip ratio)":          func(l *BodyWeightLog) **float64 { return &l.WHR },
	"optimal weight(lb)":                func(l *BodyWeightLog) **float64 { return &l.OptimalWeightLbs },
	"target to optimal weight(lb)":      func(l *BodyWeightLog) **float64 { return &l.TargetOptimalWeightLbs },
	"target to optimal muscle mass(lb)": func(l *BodyWeightLog) **float64 { return &l.TargetOptimalMuscleMassLbs },
	"target to optimal fat mass(lb)":    func(l *BodyWeightLog) **float64 { return &l.TargetOptimalFatMassLbs },
}

var textColumns = map[string]textField{
	"weight level": func(l *BodyWeightLog) **string { return &l.WeightLevel },
	"body type":    func(l *BodyWeightLog) **string { return &l.BodyType },
	"remarks":      func(l *BodyWeightLog) **string { return &l.Remarks },
}

func isBlank(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == "--"
}

// ParseRenphoCSV parses a Renpho CSV export (comma or tab separated, as copied to
// the clipboard). Rows sharing a date collapse to the latest measurement.
func ParseRenphoCSV(text string) ([]BodyWeightLog, error) {
	text = strings.TrimPrefix(strings.TrimSpace(text), "\ufeff")
	if text == "" {
		return nil, errors.New("no data")
	}
	r := csv.NewReader(strings.NewReader(text))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	first := text
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		first = text[:i]
	}
	if strings.Contains(first, "\t") && !strings.Contains(first, ",") {
		r.Comma = '\t'
	}
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid csv: %w", err)
	}
	if len(records) < 2 {
		return nil, errors.New("expected a header row and at least one data row")
	}

	header := map[string]int{}
	for i, h := range records[0] {
		header[strings.ToLower(strings.TrimSpace(h))] = i
	}
	dateIdx, hasDate := header["date"]
	weightIdx, hasWeight := header["weight(lb)"]
	if !hasDate {
		return nil, errors.New("missing Date column")
	}
	if !hasWeight {
		if _, kg := header["weight(kg)"]; kg {
			return nil, errors.New("export is in kg; switch Renpho to lb and export again")
		}
		return nil, errors.New("missing Weight(lb) column")
	}
	timeIdx, hasTime := header["time"]

	byDate := map[string]BodyWeightLog{}
	var order []string
	for n, rec := range records[1:] {
		line := n + 2
		get := func(i int) string {
			if i < len(rec) {
				return strings.TrimSpace(rec[i])
			}
			return ""
		}
		if len(rec) == 1 && get(0) == "" {
			continue
		}
		day, err := time.ParseInLocation("2006.01.02", get(dateIdx), time.Local)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid date %q", line, get(dateIdx))
		}
		w, err := strconv.ParseFloat(get(weightIdx), 64)
		if err != nil || w <= 0 {
			return nil, fmt.Errorf("row %d: invalid weight %q", line, get(weightIdx))
		}
		log := BodyWeightLog{Date: day, WeightLbs: w}
		if hasTime && !isBlank(get(timeIdx)) {
			if tm, err := time.ParseInLocation("2006.01.02 15:04:05", get(dateIdx)+" "+get(timeIdx), time.Local); err == nil {
				log.MeasuredAt = &tm
			}
		}
		for name, idx := range header {
			if f, ok := numColumns[name]; ok && !isBlank(get(idx)) {
				v, err := strconv.ParseFloat(get(idx), 64)
				if err != nil {
					return nil, fmt.Errorf("row %d: invalid %s %q", line, name, get(idx))
				}
				*f(&log) = &v
			} else if f, ok := textColumns[name]; ok && !isBlank(get(idx)) {
				v := get(idx)
				*f(&log) = &v
			}
		}
		key := day.Format("2006-01-02")
		prev, seen := byDate[key]
		if !seen {
			order = append(order, key)
		}
		if !seen || prev.MeasuredAt == nil || (log.MeasuredAt != nil && log.MeasuredAt.After(*prev.MeasuredAt)) {
			byDate[key] = log
		}
	}
	if len(order) == 0 {
		return nil, errors.New("no data rows found")
	}
	out := make([]BodyWeightLog, 0, len(order))
	for _, k := range order {
		out = append(out, byDate[k])
	}
	return out, nil
}

// ImportBodyWeights upserts every log by date in one transaction, replacing all
// metric columns for matching days. It returns how many rows were created/updated.
func ImportBodyWeights(ctx context.Context, logs []BodyWeightLog) (created, updated int, err error) {
	db := database.From(ctx)
	err = db.Transaction(func(tx *gorm.DB) error {
		for i := range logs {
			l := logs[i]
			var existing BodyWeightLog
			e := tx.Where("date = ?", l.Date).First(&existing).Error
			switch {
			case errors.Is(e, gorm.ErrRecordNotFound):
				if e := tx.Create(&l).Error; e != nil {
					return e
				}
				created++
			case e != nil:
				return e
			default:
				l.ID, l.CreatedAt = existing.ID, existing.CreatedAt
				if e := tx.Select("*").Omit(clause.Associations).Save(&l).Error; e != nil {
					return e
				}
				updated++
			}
		}
		return nil
	})
	return
}

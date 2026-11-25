package missed

import (
	"be-simpletracker/internal/clock"
	"time"

	"be-simpletracker/internal/core/tracking/steps"
	"be-simpletracker/internal/core/tracking/weight"

	"gorm.io/gorm"
)

func GetMissedYesterday(db *gorm.DB) (date time.Time, missingWeight bool, missingSteps bool, err error) {
	date = clock.ZerodTime(1)
	var weightCount int64
	if err = db.Model(&weight.BodyWeightLog{}).Where("date = ?", date).Count(&weightCount).Error; err != nil {
		return time.Time{}, false, false, err
	}
	var stepsCount int64
	if err = db.Model(&steps.StepLog{}).Where("date = ?", date).Count(&stepsCount).Error; err != nil {
		return time.Time{}, false, false, err
	}
	missingWeight = weightCount == 0
	missingSteps = stepsCount == 0
	return date, missingWeight, missingSteps, nil
}

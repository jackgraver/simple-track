package missed

import (
	"time"

	"be-simpletracker/internal/core/tracking/steps"
	"be-simpletracker/internal/core/tracking/weight"
	"be-simpletracker/internal/utils"

	"gorm.io/gorm"
)

func GetMissedYesterday(db *gorm.DB) (date time.Time, missingWeight bool, missingSteps bool, err error) {
	date = utils.ZerodTime(1)
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

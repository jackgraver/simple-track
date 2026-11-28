package missed

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/database"
	"context"
	"time"

	"be-simpletracker/internal/core/tracking/steps"
	"be-simpletracker/internal/core/tracking/weight"
)

func GetMissedYesterday(ctx context.Context) (date time.Time, missingWeight bool, missingSteps bool, err error) {
	db := database.From(ctx)
	date = clock.ZerodTime(1)
	var weightCount int64
	if err = db.Model(&weight.BodyWeightLog{}).Scopes(database.Owned(ctx)).Where("date = ?", date).Count(&weightCount).Error; err != nil {
		return time.Time{}, false, false, err
	}
	var stepsCount int64
	if err = db.Model(&steps.StepLog{}).Scopes(database.Owned(ctx)).Where("date = ?", date).Count(&stepsCount).Error; err != nil {
		return time.Time{}, false, false, err
	}
	missingWeight = weightCount == 0
	missingSteps = stepsCount == 0
	return date, missingWeight, missingSteps, nil
}

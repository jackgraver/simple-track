package service

import (
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"be-simpletracker/internal/core/media/models"

	"gorm.io/gorm"
)

// MediaInput is used for both create and update. Omitted fields are left
// untouched on update; explicit nulls clear nullable fields.
type MediaInput struct {
	Type          Optional[models.MediaType]   `json:"type"`
	Title         Optional[string]             `json:"title"`
	Creator       Optional[string]             `json:"creator"`
	ReleaseYear   Optional[int]                `json:"release_year"`
	Genres        Optional[string]             `json:"genres"`
	CoverURL      Optional[string]             `json:"cover_url"`
	Synopsis      Optional[string]             `json:"synopsis"`
	Status        Optional[models.MediaStatus] `json:"status"`
	Rating        Optional[float64]            `json:"rating"`
	Review        Optional[string]             `json:"review"`
	Favorite      Optional[bool]               `json:"favorite"`
	Progress      Optional[int]                `json:"progress"`
	Length        Optional[int]                `json:"length"`
	StartedAt     Optional[string]             `json:"started_at"`  // YYYY-MM-DD
	FinishedAt    Optional[string]             `json:"finished_at"` // YYYY-MM-DD
	TimesConsumed Optional[int]                `json:"times_consumed"`
	Source        Optional[string]             `json:"source"`
	ExternalID    Optional[string]             `json:"external_id"`
}

type ListFilter struct {
	Type     string
	Status   string
	Query    string
	Favorite *bool
	Sort     string
}

var sortOrders = map[string]string{
	"":         "updated_at DESC, id DESC",
	"updated":  "updated_at DESC, id DESC",
	"created":  "created_at DESC, id DESC",
	"title":    "LOWER(title) ASC, id ASC",
	"rating":   "rating IS NULL, rating DESC, LOWER(title) ASC",
	"finished": "finished_at IS NULL, finished_at DESC, id DESC",
	"release":  "release_year IS NULL, release_year DESC, LOWER(title) ASC",
}

var ErrInvalid = errors.New("invalid media item")

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

func ListMedia(ctx context.Context, f ListFilter) ([]models.MediaItem, error) {
	db := database.From(ctx)
	order, ok := sortOrders[f.Sort]
	if !ok {
		return nil, invalid("unknown sort %q", f.Sort)
	}
	q := db.Model(&models.MediaItem{})
	if f.Type != "" {
		if !slices.Contains(models.MediaTypes, models.MediaType(f.Type)) {
			return nil, invalid("unknown type %q", f.Type)
		}
		q = q.Where("type = ?", f.Type)
	}
	if f.Status != "" {
		if !slices.Contains(models.MediaStatuses, models.MediaStatus(f.Status)) {
			return nil, invalid("unknown status %q", f.Status)
		}
		q = q.Where("status = ?", f.Status)
	}
	if f.Favorite != nil {
		q = q.Where("favorite = ?", *f.Favorite)
	}
	if s := strings.TrimSpace(f.Query); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("LOWER(title) LIKE ? OR LOWER(creator) LIKE ?", like, like)
	}
	items := []models.MediaItem{}
	if err := q.Order(order).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func GetMedia(ctx context.Context, id uint) (*models.MediaItem, error) {
	db := database.From(ctx)
	var item models.MediaItem
	if err := db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func CreateMedia(ctx context.Context, in MediaInput, now time.Time) (*models.MediaItem, error) {
	db := database.From(ctx)
	item := models.MediaItem{Status: models.MediaStatusPlanned}
	if !in.Type.Set || in.Type.Value == nil {
		return nil, invalid("type is required")
	}
	if !in.Title.Set || in.Title.Value == nil {
		return nil, invalid("title is required")
	}
	if err := applyInput(&item, in, now); err != nil {
		return nil, err
	}
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func UpdateMedia(ctx context.Context, id uint, in MediaInput, now time.Time) (*models.MediaItem, error) {
	db := database.From(ctx)
	item, err := GetMedia(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := applyInput(item, in, now); err != nil {
		return nil, err
	}
	if err := db.Save(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func DeleteMedia(ctx context.Context, id uint) error {
	db := database.From(ctx)
	result := db.Delete(&models.MediaItem{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func applyInput(item *models.MediaItem, in MediaInput, now time.Time) error {
	prevStatus := item.Status

	if in.Type.Set {
		if in.Type.Value == nil {
			return invalid("type cannot be null")
		}
		item.Type = *in.Type.Value
	}
	if in.Title.Set {
		if in.Title.Value == nil {
			return invalid("title cannot be null")
		}
		item.Title = strings.TrimSpace(*in.Title.Value)
	}
	setString(&item.Creator, in.Creator)
	setString(&item.Genres, in.Genres)
	setString(&item.CoverURL, in.CoverURL)
	setString(&item.Synopsis, in.Synopsis)
	setString(&item.Review, in.Review)
	setString(&item.Source, in.Source)
	setString(&item.ExternalID, in.ExternalID)
	setPtr(&item.ReleaseYear, in.ReleaseYear)
	setPtr(&item.Rating, in.Rating)
	setPtr(&item.Progress, in.Progress)
	setPtr(&item.Length, in.Length)
	if in.Status.Set {
		if in.Status.Value == nil {
			return invalid("status cannot be null")
		}
		item.Status = *in.Status.Value
	}
	if in.Favorite.Set {
		item.Favorite = in.Favorite.Value != nil && *in.Favorite.Value
	}
	if in.TimesConsumed.Set {
		item.TimesConsumed = 0
		if in.TimesConsumed.Value != nil {
			item.TimesConsumed = *in.TimesConsumed.Value
		}
	}
	if err := setDate(&item.StartedAt, in.StartedAt, "started_at"); err != nil {
		return err
	}
	if err := setDate(&item.FinishedAt, in.FinishedAt, "finished_at"); err != nil {
		return err
	}

	// Fill in sensible defaults when the status moves forward, unless the
	// caller set the field explicitly in this same request.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if item.Status != prevStatus {
		switch item.Status {
		case models.MediaStatusInProgress:
			if !in.StartedAt.Set && item.StartedAt == nil {
				item.StartedAt = &today
			}
		case models.MediaStatusCompleted:
			if !in.FinishedAt.Set && item.FinishedAt == nil {
				item.FinishedAt = &today
			}
			if !in.TimesConsumed.Set && item.TimesConsumed == 0 {
				item.TimesConsumed = 1
			}
			if !in.Progress.Set && item.Length != nil {
				length := *item.Length
				item.Progress = &length
			}
		}
	}

	return validate(item)
}

func validate(item *models.MediaItem) error {
	if !slices.Contains(models.MediaTypes, item.Type) {
		return invalid("unknown type %q", item.Type)
	}
	if !slices.Contains(models.MediaStatuses, item.Status) {
		return invalid("unknown status %q", item.Status)
	}
	if item.Title == "" {
		return invalid("title is required")
	}
	if r := item.Rating; r != nil {
		if *r < 0.5 || *r > 5 || math.Mod(*r*2, 1) != 0 {
			return invalid("rating must be between 0.5 and 5 in steps of 0.5")
		}
	}
	if y := item.ReleaseYear; y != nil && (*y < 0 || *y > 9999) {
		return invalid("release_year out of range")
	}
	if p := item.Progress; p != nil && *p < 0 {
		return invalid("progress cannot be negative")
	}
	if l := item.Length; l != nil && *l < 0 {
		return invalid("length cannot be negative")
	}
	if len(item.Source) > 32 || len(item.ExternalID) > 64 {
		return invalid("source/external_id too long")
	}
	if item.TimesConsumed < 0 {
		return invalid("times_consumed cannot be negative")
	}
	if item.StartedAt != nil && item.FinishedAt != nil && item.FinishedAt.Before(*item.StartedAt) {
		return invalid("finished_at is before started_at")
	}
	return nil
}

func setString(dst *string, o Optional[string]) {
	if !o.Set {
		return
	}
	*dst = ""
	if o.Value != nil {
		*dst = strings.TrimSpace(*o.Value)
	}
}

func setPtr[T any](dst **T, o Optional[T]) {
	if o.Set {
		*dst = o.Value
	}
}

func setDate(dst **time.Time, o Optional[string], field string) error {
	if !o.Set {
		return nil
	}
	if o.Value == nil || strings.TrimSpace(*o.Value) == "" {
		*dst = nil
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(*o.Value), time.Local)
	if err != nil {
		return invalid("%s must be YYYY-MM-DD", field)
	}
	*dst = &t
	return nil
}

type MediaStats struct {
	Total             int64            `json:"total"`
	ByType            map[string]int64 `json:"by_type"`
	ByStatus          map[string]int64 `json:"by_status"`
	AverageRating     *float64         `json:"average_rating"`
	CompletedThisYear int64            `json:"completed_this_year"`
}

func GetMediaStats(ctx context.Context, now time.Time) (*MediaStats, error) {
	db := database.From(ctx)
	stats := MediaStats{ByType: map[string]int64{}, ByStatus: map[string]int64{}}
	type row struct {
		Key   string
		Count int64
	}
	for _, g := range []struct {
		col string
		dst map[string]int64
	}{{"type", stats.ByType}, {"status", stats.ByStatus}} {
		var rows []row
		if err := db.Model(&models.MediaItem{}).
			Select(g.col + " AS key, COUNT(*) AS count").
			Group(g.col).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			g.dst[r.Key] = r.Count
			if g.col == "type" {
				stats.Total += r.Count
			}
		}
	}
	if err := db.Model(&models.MediaItem{}).
		Where("rating IS NOT NULL").
		Select("AVG(rating)").Scan(&stats.AverageRating).Error; err != nil {
		return nil, err
	}
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	if err := db.Model(&models.MediaItem{}).
		Where("status = ? AND finished_at >= ?", models.MediaStatusCompleted, yearStart).
		Count(&stats.CompletedThisYear).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}

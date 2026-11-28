package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"be-simpletracker/internal/core/media/models"
	"be-simpletracker/internal/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Media is two tables (docs/multi-user/DECISIONS.md "C5 media"): the shared catalog (media_items:
// what the thing is) and the caller's own saves of it (user_media: status, rating, notes, dates).
// The API shows a save with its catalog entry side by side (Entry). All DTOs below are request
// shapes only: none carries user_id or created_by, so a forged value in a body is never read.

// CatalogInput holds the catalog fields of a create or an edit. Omitted fields are left
// untouched on edit; explicit nulls clear nullable fields.
type CatalogInput struct {
	Type        Optional[models.MediaType] `json:"type"`
	Title       Optional[string]           `json:"title"`
	Creator     Optional[string]           `json:"creator"`
	ReleaseYear Optional[int]              `json:"release_year"`
	Genres      Optional[string]           `json:"genres"`
	CoverURL    Optional[string]           `json:"cover_url"`
	Synopsis    Optional[string]           `json:"synopsis"`
	Length      Optional[int]              `json:"length"`
	Source      Optional[string]           `json:"source"`
	ExternalID  Optional[string]           `json:"external_id"`
}

// PersonalInput holds the caller's own fields of a save.
type PersonalInput struct {
	Status        Optional[models.MediaStatus] `json:"status"`
	Rating        Optional[float64]            `json:"rating"`
	Review        Optional[string]             `json:"review"`
	Favorite      Optional[bool]               `json:"favorite"`
	Progress      Optional[int]                `json:"progress"`
	StartedAt     Optional[string]             `json:"started_at"`  // YYYY-MM-DD
	FinishedAt    Optional[string]             `json:"finished_at"` // YYYY-MM-DD
	TimesConsumed Optional[int]                `json:"times_consumed"`
}

// MediaInput is the body of POST /media: a catalog entry (found by source + external_id, else
// created) and the caller's save of it.
type MediaInput struct {
	CatalogInput
	PersonalInput
}

// Entry is one of the caller's saves with its catalog entry, flat. ID is the user_media id.
type Entry struct {
	ID             uint      `json:"ID"`
	CreatedAt      time.Time `json:"CreatedAt"`
	UpdatedAt      time.Time `json:"UpdatedAt"`
	MediaItemID    uint      `json:"media_item_id"`
	AddedBy        uint      `json:"created_by" gorm:"column:created_by"` // the catalog entry's creator
	CanEditCatalog bool      `json:"can_edit_catalog" gorm:"-"`

	Type        models.MediaType `json:"type"`
	Title       string           `json:"title"`
	Creator     string           `json:"creator"`
	ReleaseYear *int             `json:"release_year"`
	Genres      string           `json:"genres"`
	CoverURL    string           `json:"cover_url"`
	Synopsis    string           `json:"synopsis"`
	Length      *int             `json:"length"`
	Source      string           `json:"source"`
	ExternalID  string           `json:"external_id"`

	Status        models.MediaStatus `json:"status"`
	Rating        *float64           `json:"rating"`
	Review        string             `json:"review"`
	Favorite      bool               `json:"favorite"`
	Progress      *int               `json:"progress"`
	StartedAt     *time.Time         `json:"started_at"`
	FinishedAt    *time.Time         `json:"finished_at"`
	TimesConsumed int                `json:"times_consumed"`
}

// CatalogEntry is a catalog item as the add picker shows it: SavedID is the caller's own save of it
// (user_media id), nil if they haven't saved it.
type CatalogEntry struct {
	ID             uint             `json:"ID"`
	CreatedAt      time.Time        `json:"CreatedAt"`
	UpdatedAt      time.Time        `json:"UpdatedAt"`
	AddedBy        uint             `json:"created_by"`
	CanEditCatalog bool             `json:"can_edit_catalog"`
	SavedID        *uint            `json:"saved_id"`
	Type           models.MediaType `json:"type"`
	Title          string           `json:"title"`
	Creator        string           `json:"creator"`
	ReleaseYear    *int             `json:"release_year"`
	Genres         string           `json:"genres"`
	CoverURL       string           `json:"cover_url"`
	Synopsis       string           `json:"synopsis"`
	Length         *int             `json:"length"`
	Source         string           `json:"source"`
	ExternalID     string           `json:"external_id"`
}

type ListFilter struct {
	Type     string
	Status   string
	Query    string
	Favorite *bool
	Sort     string
}

var sortOrders = map[string]string{
	"":         "user_media.updated_at DESC, user_media.id DESC",
	"updated":  "user_media.updated_at DESC, user_media.id DESC",
	"created":  "user_media.created_at DESC, user_media.id DESC",
	"title":    "LOWER(media_items.title) ASC, user_media.id ASC",
	"rating":   "user_media.rating IS NULL, user_media.rating DESC, LOWER(media_items.title) ASC",
	"finished": "user_media.finished_at IS NULL, user_media.finished_at DESC, user_media.id DESC",
	"release":  "media_items.release_year IS NULL, media_items.release_year DESC, LOWER(media_items.title) ASC",
}

const entryColumns = `user_media.id AS id, user_media.created_at AS created_at, user_media.updated_at AS updated_at,
	user_media.media_item_id AS media_item_id, media_items.created_by AS created_by,
	media_items.type AS type, media_items.title AS title, media_items.creator AS creator,
	media_items.release_year AS release_year, media_items.genres AS genres, media_items.cover_url AS cover_url,
	media_items.synopsis AS synopsis, media_items.length AS length, media_items.source AS source,
	media_items.external_id AS external_id,
	user_media.status AS status, user_media.rating AS rating, user_media.review AS review,
	user_media.favorite AS favorite, user_media.progress AS progress, user_media.started_at AS started_at,
	user_media.finished_at AS finished_at, user_media.times_consumed AS times_consumed`

var (
	ErrInvalid = errors.New("invalid media item")
	// ErrForbidden: the catalog entry exists but the caller may not change it (not its creator,
	// not an admin). Catalog rows are visible to everyone, so this is a 403, not a 404.
	ErrForbidden = errors.New("not allowed to change this catalog entry")
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// saves is the caller's live saves joined to their live catalog entries. The scope on user_media
// is the ownership filter; media_items is a readable catalog.
func saves(ctx context.Context) *gorm.DB {
	return database.From(ctx).Model(&models.UserMedia{}).Scopes(database.Owned(ctx)).
		Joins("JOIN media_items ON media_items.id = user_media.media_item_id AND media_items.deleted_at IS NULL")
}

func canEditCatalog(ctx context.Context, createdBy uint) bool {
	id, ok := database.CurrentIdentity(ctx)
	return ok && (id.IsAdmin || id.UserID == createdBy)
}

func fillEntries(ctx context.Context, entries []Entry) {
	for i := range entries {
		entries[i].CanEditCatalog = canEditCatalog(ctx, entries[i].AddedBy)
	}
}

func ListMedia(ctx context.Context, f ListFilter) ([]Entry, error) {
	order, ok := sortOrders[f.Sort]
	if !ok {
		return nil, invalid("unknown sort %q", f.Sort)
	}
	q := saves(ctx)
	if f.Type != "" {
		if !slices.Contains(models.MediaTypes, models.MediaType(f.Type)) {
			return nil, invalid("unknown type %q", f.Type)
		}
		q = q.Where("media_items.type = ?", f.Type)
	}
	if f.Status != "" {
		if !slices.Contains(models.MediaStatuses, models.MediaStatus(f.Status)) {
			return nil, invalid("unknown status %q", f.Status)
		}
		q = q.Where("user_media.status = ?", f.Status)
	}
	if f.Favorite != nil {
		q = q.Where("user_media.favorite = ?", *f.Favorite)
	}
	if s := strings.TrimSpace(f.Query); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where("(LOWER(media_items.title) LIKE ? OR LOWER(media_items.creator) LIKE ?)", like, like)
	}
	entries := []Entry{}
	if err := q.Select(entryColumns).Order(order).Scan(&entries).Error; err != nil {
		return nil, err
	}
	fillEntries(ctx, entries)
	return entries, nil
}

// GetMedia returns one of the caller's saves; another user's save and a missing ID are the same
// not-found error.
func GetMedia(ctx context.Context, id uint) (*Entry, error) {
	var entries []Entry
	if err := saves(ctx).Where("user_media.id = ?", id).Select(entryColumns).Scan(&entries).Error; err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, database.ErrNotFound
	}
	fillEntries(ctx, entries)
	return &entries[0], nil
}

// CreateMedia saves a catalog entry for the caller, adding it to the catalog first unless one with
// the same (source, external_id) is there already. created is false when the caller already had
// it saved (the existing save is returned untouched).
func CreateMedia(ctx context.Context, in MediaInput, now time.Time) (entry *Entry, created bool, err error) {
	if !in.Type.Set || in.Type.Value == nil {
		return nil, false, invalid("type is required")
	}
	if !in.Title.Set || in.Title.Value == nil {
		return nil, false, invalid("title is required")
	}
	item, err := findOrCreateCatalog(ctx, in.CatalogInput)
	if err != nil {
		return nil, false, err
	}
	return saveItem(ctx, item, in.PersonalInput, now)
}

// SaveCatalogItem saves an existing catalog entry for the caller (status defaults to planned).
func SaveCatalogItem(ctx context.Context, itemID uint, in PersonalInput, now time.Time) (*Entry, bool, error) {
	item, err := getCatalogItem(ctx, itemID)
	if err != nil {
		return nil, false, err
	}
	return saveItem(ctx, item, in, now)
}

func UpdateMedia(ctx context.Context, id uint, in PersonalInput, now time.Time) (*Entry, error) {
	db := database.From(ctx)
	var save models.UserMedia
	if err := db.Scopes(database.Owned(ctx)).First(&save, id).Error; err != nil {
		return nil, err
	}
	item, err := getCatalogItem(ctx, save.MediaItemID)
	if err != nil {
		return nil, err
	}
	if err := applyPersonal(&save, item, in, now); err != nil {
		return nil, err
	}
	if err := database.Affected(db.Scopes(database.Owned(ctx)).Model(&models.UserMedia{}).Where("id = ?", save.ID).
		Select("Status", "Rating", "Review", "Favorite", "Progress", "StartedAt", "FinishedAt", "TimesConsumed").
		Updates(&save)); err != nil {
		return nil, err
	}
	return GetMedia(ctx, id)
}

// DeleteMedia removes the caller's save. The catalog entry stays.
func DeleteMedia(ctx context.Context, id uint) error {
	return database.Affected(database.From(ctx).Scopes(database.Owned(ctx)).Delete(&models.UserMedia{}, id))
}

func getCatalogItem(ctx context.Context, id uint) (*models.MediaItem, error) {
	var item models.MediaItem
	if err := database.From(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func findOrCreateCatalog(ctx context.Context, in CatalogInput) (*models.MediaItem, error) {
	db := database.From(ctx)
	item := models.MediaItem{}
	if err := applyCatalog(&item, in); err != nil {
		return nil, err
	}
	find := func() (*models.MediaItem, error) {
		if item.Source == "" || item.ExternalID == "" {
			return nil, nil
		}
		var existing models.MediaItem
		err := db.Where("source = ? AND external_id = ?", item.Source, item.ExternalID).First(&existing).Error
		if database.IsNotFound(err) {
			return nil, nil
		}
		return &existing, err
	}
	if existing, err := find(); existing != nil || err != nil {
		return existing, err
	}
	// A savepoint, so losing a race to the unique index doesn't poison the request transaction.
	createErr := db.Transaction(func(tx *gorm.DB) error {
		return tx.Omit(clause.Associations).Create(&item).Error
	})
	if createErr != nil {
		if existing, err := find(); existing != nil && err == nil {
			return existing, nil
		}
		return nil, createErr
	}
	return &item, nil
}

func saveItem(ctx context.Context, item *models.MediaItem, in PersonalInput, now time.Time) (*Entry, bool, error) {
	db := database.From(ctx)
	existing := func() (*Entry, error) {
		var save models.UserMedia
		err := db.Scopes(database.Owned(ctx)).Where("media_item_id = ?", item.ID).First(&save).Error
		if err != nil {
			return nil, err
		}
		return GetMedia(ctx, save.ID)
	}
	if e, err := existing(); err == nil {
		return e, false, nil
	} else if !database.IsNotFound(err) {
		return nil, false, err
	}
	save := models.UserMedia{MediaItemID: item.ID, Status: models.MediaStatusPlanned}
	if err := applyPersonal(&save, item, in, now); err != nil {
		return nil, false, err
	}
	createErr := db.Transaction(func(tx *gorm.DB) error {
		return tx.Omit(clause.Associations).Create(&save).Error
	})
	if createErr != nil {
		if e, err := existing(); err == nil {
			return e, false, nil
		}
		return nil, false, createErr
	}
	e, err := GetMedia(ctx, save.ID)
	return e, true, err
}

// SearchCatalog lists catalog entries matching q (title or creator) and type, with the caller's
// own save of each.
func SearchCatalog(ctx context.Context, q, mediaType string) ([]CatalogEntry, error) {
	db := database.From(ctx)
	tx := db.Model(&models.MediaItem{})
	if mediaType != "" {
		if !slices.Contains(models.MediaTypes, models.MediaType(mediaType)) {
			return nil, invalid("unknown type %q", mediaType)
		}
		tx = tx.Where("type = ?", mediaType)
	}
	if s := strings.TrimSpace(q); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		tx = tx.Where("(LOWER(title) LIKE ? OR LOWER(creator) LIKE ?)", like, like)
	}
	var items []models.MediaItem
	if err := tx.Order("LOWER(title) ASC, id ASC").Limit(50).Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]CatalogEntry, 0, len(items))
	ids := make([]uint, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
		out = append(out, catalogEntry(ctx, it))
	}
	if len(ids) == 0 {
		return out, nil
	}
	var mine []models.UserMedia
	if err := db.Scopes(database.Owned(ctx)).Where("media_item_id IN ?", ids).Find(&mine).Error; err != nil {
		return nil, err
	}
	saved := make(map[uint]uint, len(mine))
	for _, m := range mine {
		saved[m.MediaItemID] = m.ID
	}
	for i := range out {
		if id, ok := saved[out[i].ID]; ok {
			out[i].SavedID = &id
		}
	}
	return out, nil
}

func catalogEntry(ctx context.Context, it models.MediaItem) CatalogEntry {
	return CatalogEntry{
		ID: it.ID, CreatedAt: it.CreatedAt, UpdatedAt: it.UpdatedAt, AddedBy: it.CreatedBy,
		CanEditCatalog: canEditCatalog(ctx, it.CreatedBy),
		Type:           it.Type, Title: it.Title, Creator: it.Author, ReleaseYear: it.ReleaseYear,
		Genres: it.Genres, CoverURL: it.CoverURL, Synopsis: it.Synopsis, Length: it.Length,
		Source: it.Source, ExternalID: it.ExternalID,
	}
}

// UpdateCatalog edits a catalog entry: its creator or an admin only (ErrForbidden otherwise,
// database.ErrNotFound if it doesn't exist).
func UpdateCatalog(ctx context.Context, id uint, in CatalogInput) (*CatalogEntry, error) {
	db := database.From(ctx)
	item, err := getCatalogItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canEditCatalog(ctx, item.CreatedBy) {
		return nil, ErrForbidden
	}
	if err := applyCatalog(item, in); err != nil {
		return nil, err
	}
	// source + external_id can't move onto another entry's identity.
	if in.Source.Set || in.ExternalID.Set {
		var clash int64
		if err := db.Model(&models.MediaItem{}).
			Where("id <> ? AND source = ? AND external_id = ? AND source <> '' AND external_id <> ''", item.ID, item.Source, item.ExternalID).
			Count(&clash).Error; err != nil {
			return nil, err
		}
		if clash > 0 {
			return nil, invalid("another catalog entry already has that source and external_id")
		}
	}
	if err := database.Affected(db.Model(&models.MediaItem{}).Where("id = ?", item.ID).
		Select("Type", "Title", "Author", "ReleaseYear", "Genres", "CoverURL", "Synopsis", "Length", "Source", "ExternalID").
		Updates(item)); err != nil {
		return nil, err
	}
	e := catalogEntry(ctx, *item)
	return &e, nil
}

// DeleteCatalog soft-deletes a catalog entry for everyone: admin only. People's saves of it are
// kept but no longer listed.
func DeleteCatalog(ctx context.Context, id uint) error {
	if i, ok := database.CurrentIdentity(ctx); !ok || !i.IsAdmin {
		return ErrForbidden
	}
	return database.Affected(database.From(ctx).Delete(&models.MediaItem{}, id))
}

func applyCatalog(item *models.MediaItem, in CatalogInput) error {
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
	setString(&item.Author, in.Creator)
	setString(&item.Genres, in.Genres)
	setString(&item.CoverURL, in.CoverURL)
	setString(&item.Synopsis, in.Synopsis)
	setString(&item.Source, in.Source)
	setString(&item.ExternalID, in.ExternalID)
	setPtr(&item.ReleaseYear, in.ReleaseYear)
	setPtr(&item.Length, in.Length)
	if err := validateCatalog(item); err != nil {
		return err
	}
	// Only judged when the caller sets one of them, so a legacy row with half a pair stays editable.
	if (in.Source.Set || in.ExternalID.Set) && (item.Source == "") != (item.ExternalID == "") {
		return invalid("source and external_id go together")
	}
	return nil
}

func validateCatalog(item *models.MediaItem) error {
	if !slices.Contains(models.MediaTypes, item.Type) {
		return invalid("unknown type %q", item.Type)
	}
	if item.Title == "" {
		return invalid("title is required")
	}
	if y := item.ReleaseYear; y != nil && (*y < 0 || *y > 9999) {
		return invalid("release_year out of range")
	}
	if l := item.Length; l != nil && *l < 0 {
		return invalid("length cannot be negative")
	}
	if len(item.Source) > 32 || len(item.ExternalID) > 64 {
		return invalid("source/external_id too long")
	}
	return nil
}

func applyPersonal(save *models.UserMedia, item *models.MediaItem, in PersonalInput, now time.Time) error {
	prevStatus := save.Status
	setString(&save.Review, in.Review)
	setPtr(&save.Rating, in.Rating)
	setPtr(&save.Progress, in.Progress)
	if in.Status.Set {
		if in.Status.Value == nil {
			return invalid("status cannot be null")
		}
		save.Status = *in.Status.Value
	}
	if in.Favorite.Set {
		save.Favorite = in.Favorite.Value != nil && *in.Favorite.Value
	}
	if in.TimesConsumed.Set {
		save.TimesConsumed = 0
		if in.TimesConsumed.Value != nil {
			save.TimesConsumed = *in.TimesConsumed.Value
		}
	}
	if err := setDate(&save.StartedAt, in.StartedAt, "started_at"); err != nil {
		return err
	}
	if err := setDate(&save.FinishedAt, in.FinishedAt, "finished_at"); err != nil {
		return err
	}

	// Fill in sensible defaults when the status moves forward, unless the
	// caller set the field explicitly in this same request.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if save.Status != prevStatus {
		switch save.Status {
		case models.MediaStatusInProgress:
			if !in.StartedAt.Set && save.StartedAt == nil {
				save.StartedAt = &today
			}
		case models.MediaStatusCompleted:
			if !in.FinishedAt.Set && save.FinishedAt == nil {
				save.FinishedAt = &today
			}
			if !in.TimesConsumed.Set && save.TimesConsumed == 0 {
				save.TimesConsumed = 1
			}
			if !in.Progress.Set && item.Length != nil {
				length := *item.Length
				save.Progress = &length
			}
		}
	}
	return validatePersonal(save)
}

func validatePersonal(save *models.UserMedia) error {
	if !slices.Contains(models.MediaStatuses, save.Status) {
		return invalid("unknown status %q", save.Status)
	}
	if r := save.Rating; r != nil {
		if *r < 0.5 || *r > 5 || math.Mod(*r*2, 1) != 0 {
			return invalid("rating must be between 0.5 and 5 in steps of 0.5")
		}
	}
	if p := save.Progress; p != nil && *p < 0 {
		return invalid("progress cannot be negative")
	}
	if save.TimesConsumed < 0 {
		return invalid("times_consumed cannot be negative")
	}
	if save.StartedAt != nil && save.FinishedAt != nil && save.FinishedAt.Before(*save.StartedAt) {
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

// MediaStats counts the caller's saves.
type MediaStats struct {
	Total             int64            `json:"total"`
	ByType            map[string]int64 `json:"by_type"`
	ByStatus          map[string]int64 `json:"by_status"`
	AverageRating     *float64         `json:"average_rating"`
	CompletedThisYear int64            `json:"completed_this_year"`
}

func GetMediaStats(ctx context.Context, now time.Time) (*MediaStats, error) {
	stats := MediaStats{ByType: map[string]int64{}, ByStatus: map[string]int64{}}
	type row struct {
		Key   string
		Count int64
	}
	for _, g := range []struct {
		col string
		dst map[string]int64
	}{{"media_items.type", stats.ByType}, {"user_media.status", stats.ByStatus}} {
		var rows []row
		if err := saves(ctx).
			Select(g.col + " AS key, COUNT(*) AS count").
			Group(g.col).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			g.dst[r.Key] = r.Count
			if g.col == "media_items.type" {
				stats.Total += r.Count
			}
		}
	}
	if err := saves(ctx).
		Where("user_media.rating IS NOT NULL").
		Select("AVG(user_media.rating)").Scan(&stats.AverageRating).Error; err != nil {
		return nil, err
	}
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	if err := saves(ctx).
		Where("user_media.status = ? AND user_media.finished_at >= ?", models.MediaStatusCompleted, yearStart).
		Count(&stats.CompletedThisYear).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}

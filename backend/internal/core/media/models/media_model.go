package models

import (
	"be-simpletracker/internal/database"
	"time"

	"gorm.io/gorm"
)

type MediaType string

const (
	MediaTypeMovie   MediaType = "movie"
	MediaTypeTV      MediaType = "tv"
	MediaTypeBook    MediaType = "book"
	MediaTypeGame    MediaType = "game"
	MediaTypeAnime   MediaType = "anime"
	MediaTypeComic   MediaType = "comic"
	MediaTypeAlbum   MediaType = "album"
	MediaTypePodcast MediaType = "podcast"
	MediaTypeOther   MediaType = "other"
)

var MediaTypes = []MediaType{
	MediaTypeMovie, MediaTypeTV, MediaTypeBook, MediaTypeGame, MediaTypeAnime,
	MediaTypeComic, MediaTypeAlbum, MediaTypePodcast, MediaTypeOther,
}

type MediaStatus string

const (
	MediaStatusPlanned    MediaStatus = "planned"     // want to watch/read/play
	MediaStatusInProgress MediaStatus = "in_progress" // currently consuming
	MediaStatusCompleted  MediaStatus = "completed"
	MediaStatusOnHold     MediaStatus = "on_hold"
	MediaStatusDropped    MediaStatus = "dropped"
)

var MediaStatuses = []MediaStatus{
	MediaStatusPlanned, MediaStatusInProgress, MediaStatusCompleted, MediaStatusOnHold, MediaStatusDropped,
}

// MediaItem is the shared catalog entry (docs/multi-user/DECISIONS.md "C5 media"): what the thing
// is, not what anyone thinks of it. Everyone signed in reads and adds; the creator or an admin
// edits; only an admin deletes. Personal state lives in UserMedia.
type MediaItem struct {
	gorm.Model
	database.Creator
	Type  MediaType `json:"type" gorm:"type:varchar(16);not null;index"`
	Title string    `json:"title" gorm:"not null;index"`
	// Author is the director, author, developer, artist... (the column and JSON key stay
	// "creator"; the Go name differs because database.Creator is embedded).
	Author      string `json:"creator" gorm:"column:creator"`
	ReleaseYear *int   `json:"release_year"`
	Genres      string `json:"genres"` // free text, comma separated
	CoverURL    string `json:"cover_url"`
	Synopsis    string `json:"synopsis"`
	// Length is in the medium's natural unit (pages, episodes, hours...).
	Length *int `json:"length"`
	// Where the metadata was looked up from (e.g. "tmdb" + "movie/949"); empty if entered by hand.
	// One live catalog row per non-empty (source, external_id).
	Source     string `json:"source" gorm:"type:varchar(32)"`
	ExternalID string `json:"external_id" gorm:"type:varchar(64)"`
}

func (MediaItem) TableName() string { return "media_items" }

// UserMedia is one user's save of a catalog item: status, rating, notes, dates. Owned: reads go
// through database.Owned, RLS backs it up. At most one live row per (user, catalog item).
type UserMedia struct {
	gorm.Model
	database.Owner
	MediaItemID uint        `gorm:"not null;index"`
	Status      MediaStatus `gorm:"type:varchar(16);not null"`
	// Rating is 0.5–5.0 in half-star steps; nil means unrated.
	Rating   *float64 `gorm:"type:numeric(2,1)"`
	Review   string
	Favorite bool `gorm:"not null;default:false"`
	// Progress is in the same unit as the catalog item's Length.
	Progress      *int
	StartedAt     *time.Time
	FinishedAt    *time.Time
	TimesConsumed int `gorm:"not null;default:0"` // rewatch/reread count
}

func (UserMedia) TableName() string { return "user_media" }

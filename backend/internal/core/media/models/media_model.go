package models

import (
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

type MediaItem struct {
	gorm.Model
	Type        MediaType   `json:"type" gorm:"type:varchar(16);not null;index"`
	Title       string      `json:"title" gorm:"not null;index"`
	Creator     string      `json:"creator"` // director, author, developer, artist...
	ReleaseYear *int        `json:"release_year"`
	Genres      string      `json:"genres"` // free text, comma separated
	CoverURL    string      `json:"cover_url"`
	Synopsis    string      `json:"synopsis"`
	Status      MediaStatus `json:"status" gorm:"type:varchar(16);not null;index"`
	// Rating is 0.5–5.0 in half-star steps; nil means unrated.
	Rating   *float64 `json:"rating" gorm:"type:numeric(2,1)"`
	Review   string   `json:"review"`
	Favorite bool     `json:"favorite" gorm:"not null;default:false"`
	// Progress/Length are in the medium's natural unit (pages, episodes, hours...).
	Progress      *int       `json:"progress"`
	Length        *int       `json:"length"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	TimesConsumed int        `json:"times_consumed" gorm:"not null;default:0"` // rewatch/reread count
}

func (MediaItem) TableName() string { return "media_items" }

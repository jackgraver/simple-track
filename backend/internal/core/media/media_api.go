package media

import (
	"be-simpletracker/internal/core/media/controller"
	"be-simpletracker/internal/core/media/lookup"
	"be-simpletracker/internal/env"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	lookupService := lookup.NewService(env.OptionalString("TMDB_API_KEY"))
	applyLookupURLOverrides(lookupService)
	controller.RegisterMediaRoutes(router.Group("/media", authMiddleware), h.db, lookupService)
}

// applyLookupURLOverrides points the lookup sources at other servers (the e2e suite's fake
// upstream). Ignored in production so a stray variable can't redirect real lookups.
func applyLookupURLOverrides(s *lookup.Service) {
	if env.IsProduction() {
		return
	}
	for key, dst := range map[string]*string{
		"LOOKUP_TMDB_URL":        &s.TMDBBase,
		"LOOKUP_TMDB_IMAGE_URL":  &s.TMDBImageBase,
		"LOOKUP_OPENLIBRARY_URL": &s.OpenLibBase,
		"LOOKUP_OPENLIB_COVERS":  &s.OpenLibCovers,
		"LOOKUP_ANILIST_URL":     &s.AniListURL,
		"LOOKUP_ITUNES_URL":      &s.ITunesBase,
	} {
		if v := env.OptionalString(key); v != "" {
			*dst = v
		}
	}
}

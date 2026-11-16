package lookup

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// TMDB accepts either a v3 API key (query param) or a v4 read access token
// (bearer JWT); both are shown on the TMDB API settings page.
func (s *Service) tmdbGet(ctx context.Context, path string, params url.Values, out any) error {
	if s.tmdbKey == "" {
		return fmt.Errorf("%w: set TMDB_API_KEY to look up movies and TV", ErrNotConfigured)
	}
	if params == nil {
		params = url.Values{}
	}
	header := http.Header{}
	if strings.HasPrefix(s.tmdbKey, "eyJ") {
		header.Set("Authorization", "Bearer "+s.tmdbKey)
	} else {
		params.Set("api_key", s.tmdbKey)
	}
	return s.getJSON(ctx, withQuery(s.TMDBBase+path, params), header, out)
}

func (s *Service) tmdbImage(size, path string) string {
	if path == "" {
		return ""
	}
	return s.TMDBImageBase + "/" + size + path
}

type tmdbSearchItem struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Name         string `json:"name"`
	ReleaseDate  string `json:"release_date"`
	FirstAirDate string `json:"first_air_date"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
}

func (s *Service) tmdbSearch(ctx context.Context, kind, query string) ([]Result, error) {
	var resp struct {
		Results []tmdbSearchItem `json:"results"`
	}
	params := url.Values{"query": {query}, "include_adult": {"false"}}
	if err := s.tmdbGet(ctx, "/search/"+kind, params, &resp); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(resp.Results))
	for _, r := range resp.Results {
		title, date := r.Title, r.ReleaseDate
		if kind == "tv" {
			title, date = r.Name, r.FirstAirDate
		}
		results = append(results, Result{
			Source:      "tmdb",
			ExternalID:  fmt.Sprintf("%s/%d", kind, r.ID),
			Title:       title,
			ReleaseYear: yearFromDate(date),
			Synopsis:    r.Overview,
			CoverURL:    s.tmdbImage("w185", r.PosterPath),
		})
	}
	return results, nil
}

type tmdbNamed struct {
	Name string `json:"name"`
}

func (s *Service) tmdbDetails(ctx context.Context, externalID string) (*Result, error) {
	kind, id, ok := strings.Cut(externalID, "/")
	if !ok || (kind != "movie" && kind != "tv") || id == "" {
		return nil, fmt.Errorf("%w: bad tmdb id %q", ErrNotFound, externalID)
	}
	var d struct {
		tmdbSearchItem
		Runtime          int         `json:"runtime"`
		NumberOfEpisodes int         `json:"number_of_episodes"`
		Genres           []tmdbNamed `json:"genres"`
		CreatedBy        []tmdbNamed `json:"created_by"`
		Credits          struct {
			Crew []struct {
				Name string `json:"name"`
				Job  string `json:"job"`
			} `json:"crew"`
		} `json:"credits"`
	}
	params := url.Values{}
	if kind == "movie" {
		params.Set("append_to_response", "credits")
	}
	if err := s.tmdbGet(ctx, "/"+kind+"/"+url.PathEscape(id), params, &d); err != nil {
		return nil, err
	}
	genres := make([]string, 0, len(d.Genres))
	for _, g := range d.Genres {
		genres = append(genres, g.Name)
	}
	r := &Result{
		Source:     "tmdb",
		ExternalID: externalID,
		Synopsis:   d.Overview,
		Genres:     joinLimit(genres, 4),
		CoverURL:   s.tmdbImage("w500", d.PosterPath),
	}
	if kind == "movie" {
		var directors []string
		for _, c := range d.Credits.Crew {
			if c.Job == "Director" {
				directors = append(directors, c.Name)
			}
		}
		r.Title = d.Title
		r.ReleaseYear = yearFromDate(d.ReleaseDate)
		r.Creator = joinLimit(directors, 2)
		r.Length = intPtr(d.Runtime)
	} else {
		creators := make([]string, 0, len(d.CreatedBy))
		for _, c := range d.CreatedBy {
			creators = append(creators, c.Name)
		}
		r.Title = d.Name
		r.ReleaseYear = yearFromDate(d.FirstAirDate)
		r.Creator = joinLimit(creators, 2)
		r.Length = intPtr(d.NumberOfEpisodes)
	}
	return r, nil
}

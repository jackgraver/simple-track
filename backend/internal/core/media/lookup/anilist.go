package lookup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const aniListFields = `id type format
  title { romaji english }
  startDate { year }
  coverImage { extraLarge large }
  description(asHtml: false)
  genres episodes chapters volumes
  studios(isMain: true) { nodes { name } }
  staff(perPage: 6, sort: [RELEVANCE]) { edges { role } nodes { name { full } } }`

type aniListMedia struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Format string `json:"format"`
	Title  struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
	} `json:"title"`
	StartDate struct {
		Year int `json:"year"`
	} `json:"startDate"`
	CoverImage struct {
		ExtraLarge string `json:"extraLarge"`
		Large      string `json:"large"`
	} `json:"coverImage"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
	Episodes    int      `json:"episodes"`
	Chapters    int      `json:"chapters"`
	Volumes     int      `json:"volumes"`
	Studios     struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"studios"`
	Staff struct {
		Edges []struct {
			Role string `json:"role"`
		} `json:"edges"`
		Nodes []struct {
			Name struct {
				Full string `json:"full"`
			} `json:"name"`
		} `json:"nodes"`
	} `json:"staff"`
}

func (s *Service) aniListQuery(ctx context.Context, query string, vars map[string]any, out any) error {
	body, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.AniListURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
		} `json:"errors"`
	}
	if err := s.do(req, &resp); err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		if resp.Errors[0].Status == http.StatusNotFound {
			return ErrNotFound
		}
		return newLookupError(ErrUpstream, "anilist: %s", resp.Errors[0].Message)
	}
	return json.Unmarshal(resp.Data, out)
}

func (s *Service) aniListSearch(ctx context.Context, mediaType, query string) ([]Result, error) {
	var data struct {
		Page struct {
			Media []aniListMedia `json:"media"`
		} `json:"Page"`
	}
	q := `query ($search: String, $type: MediaType, $perPage: Int) {
  Page(perPage: $perPage) { media(search: $search, type: $type, isAdult: false) { ` + aniListFields + ` } }
}`
	vars := map[string]any{"search": query, "type": mediaType, "perPage": maxResults}
	if err := s.aniListQuery(ctx, q, vars, &data); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(data.Page.Media))
	for _, m := range data.Page.Media {
		results = append(results, aniListResult(m))
	}
	return results, nil
}

func (s *Service) aniListDetails(ctx context.Context, externalID string) (*Result, error) {
	id, err := strconv.Atoi(externalID)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("%w: bad anilist id %q", ErrNotFound, externalID)
	}
	var data struct {
		Media *aniListMedia `json:"Media"`
	}
	q := `query ($id: Int) { Media(id: $id) { ` + aniListFields + ` } }`
	if err := s.aniListQuery(ctx, q, map[string]any{"id": id}, &data); err != nil {
		return nil, err
	}
	if data.Media == nil {
		return nil, ErrNotFound
	}
	r := aniListResult(*data.Media)
	return &r, nil
}

func aniListResult(m aniListMedia) Result {
	title := m.Title.English
	if title == "" {
		title = m.Title.Romaji
	}
	cover := m.CoverImage.ExtraLarge
	if cover == "" {
		cover = m.CoverImage.Large
	}
	r := Result{
		Source:      "anilist",
		ExternalID:  strconv.Itoa(m.ID),
		Title:       title,
		ReleaseYear: intPtr(m.StartDate.Year),
		Synopsis:    cleanText(m.Description),
		Genres:      joinLimit(m.Genres, 4),
		CoverURL:    cover,
	}
	if m.Type == "MANGA" {
		// Credit the author(s) from staff roles like "Story & Art", "Story".
		var authors []string
		for i, e := range m.Staff.Edges {
			role := strings.ToLower(e.Role)
			if i < len(m.Staff.Nodes) && (strings.Contains(role, "story") || strings.Contains(role, "art")) {
				authors = append(authors, m.Staff.Nodes[i].Name.Full)
			}
		}
		r.Creator = joinLimit(authors, 2)
		r.Length = intPtr(m.Chapters)
		if r.Length == nil {
			r.Length = intPtr(m.Volumes)
		}
	} else {
		studios := make([]string, 0, len(m.Studios.Nodes))
		for _, n := range m.Studios.Nodes {
			studios = append(studios, n.Name)
		}
		r.Creator = joinLimit(studios, 2)
		r.Length = intPtr(m.Episodes)
	}
	return r
}

package lookup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const olSearchFields = "key,title,author_name,first_publish_year,cover_i,number_of_pages_median,subject"

type olDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	FirstPublishYear int      `json:"first_publish_year"`
	CoverID          int      `json:"cover_i"`
	Pages            int      `json:"number_of_pages_median"`
	Subject          []string `json:"subject"`
}

func (s *Service) olCover(id int, size string) string {
	if id <= 0 {
		return ""
	}
	return fmt.Sprintf("%s/b/id/%d-%s.jpg", s.OpenLibCovers, id, size)
}

func (s *Service) olSearch(ctx context.Context, q string, limit int) ([]olDoc, error) {
	var resp struct {
		Docs []olDoc `json:"docs"`
	}
	params := url.Values{"q": {q}, "limit": {fmt.Sprint(limit)}, "fields": {olSearchFields}}
	if err := s.getJSON(ctx, withQuery(s.OpenLibBase+"/search.json", params), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Docs, nil
}

func (s *Service) olResult(d olDoc, coverSize string) Result {
	return Result{
		Source:      "openlibrary",
		ExternalID:  d.Key,
		Title:       d.Title,
		Creator:     joinLimit(d.AuthorName, 2),
		ReleaseYear: intPtr(d.FirstPublishYear),
		Genres:      olGenres(d.Subject),
		CoverURL:    s.olCover(d.CoverID, coverSize),
		Length:      intPtr(d.Pages),
	}
}

func (s *Service) openLibrarySearch(ctx context.Context, query string) ([]Result, error) {
	docs, err := s.olSearch(ctx, query, maxResults)
	if err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(docs))
	for _, d := range docs {
		results = append(results, s.olResult(d, "M"))
	}
	return results, nil
}

var olWorkKeyRe = regexp.MustCompile(`^/works/OL\d+W$`)

func (s *Service) openLibraryDetails(ctx context.Context, key string) (*Result, error) {
	if !olWorkKeyRe.MatchString(key) {
		return nil, fmt.Errorf("%w: bad open library key %q", ErrNotFound, key)
	}
	docs, err := s.olSearch(ctx, "key:"+key, 1)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, ErrNotFound
	}
	r := s.olResult(docs[0], "L")

	// The search index has no description; the work record does.
	var work struct {
		Description json.RawMessage `json:"description"`
	}
	if err := s.getJSON(ctx, s.OpenLibBase+key+".json", nil, &work); err == nil {
		r.Synopsis = olDescription(work.Description)
	}
	return &r, nil
}

// olDescription handles both "text" and {"type": ..., "value": "text"}.
func olDescription(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		var obj struct {
			Value string `json:"value"`
		}
		if json.Unmarshal(raw, &obj) != nil {
			return ""
		}
		text = obj.Value
	}
	// Strip trailing source/link boilerplate like "----------\nAlso contained in: ...".
	if i := strings.Index(text, "\n----"); i > 0 {
		text = text[:i]
	}
	return cleanText(text)
}

// olGenres picks a few readable subjects, skipping index noise like
// "nyt:..." / "award:..." tags and place/character headings.
func olGenres(subjects []string) string {
	keep := make([]string, 0, len(subjects))
	for _, sub := range subjects {
		lower := strings.ToLower(sub)
		if strings.ContainsAny(sub, ":(),") || strings.Contains(lower, "new york times") ||
			strings.Contains(lower, "award") || strings.Contains(lower, "large type") || len(sub) > 30 {
			continue
		}
		keep = append(keep, sub)
	}
	return joinLimit(keep, 3)
}

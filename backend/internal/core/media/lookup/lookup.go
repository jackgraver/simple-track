// Package lookup searches public media databases (TMDB, Open Library,
// AniList, iTunes) so items can be filled in without manual research.
package lookup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"be-simpletracker/internal/core/media/models"
)

var (
	ErrNotConfigured = errors.New("lookup source is not configured")
	ErrUnsupported   = errors.New("no lookup source for this media type")
	ErrNotFound      = errors.New("lookup result not found")
	ErrUpstream      = errors.New("lookup source request failed")
)

// lookupError wraps a sentinel with detail for the server log. Error() returns only the
// sentinel text, so nothing about the upstream request (URL, query, API key) or its
// response body can reach a client that prints the error. Use Detail for logging.
type lookupError struct {
	sentinel error
	detail   string
}

func (e *lookupError) Error() string { return e.sentinel.Error() }
func (e *lookupError) Unwrap() error { return e.sentinel }

func newLookupError(sentinel error, format string, args ...any) error {
	return &lookupError{sentinel: sentinel, detail: fmt.Sprintf(format, args...)}
}

// Detail returns a log-safe description of err: the sentinel plus any recorded detail,
// with key/token query parameters redacted.
func Detail(err error) string {
	if err == nil {
		return ""
	}
	var le *lookupError
	if errors.As(err, &le) && le.detail != "" {
		return le.sentinel.Error() + ": " + redact(le.detail)
	}
	return redact(err.Error())
}

var secretParamRe = regexp.MustCompile(`(?i)\b(api_key|apikey|key|token|access_token|authorization)=[^&\s"']*`)

// redact masks secret-looking query parameters in free text.
func redact(s string) string {
	return secretParamRe.ReplaceAllString(s, "$1=REDACTED")
}

// transportDetail describes a client.Do error without the request URL's query string.
func transportDetail(err error) string {
	var ue *url.Error
	if errors.As(err, &ue) {
		u := ue.URL
		if parsed, perr := url.Parse(u); perr == nil {
			parsed.RawQuery = ""
			parsed.User = nil
			u = parsed.String()
		} else {
			u = redact(u)
		}
		return fmt.Sprintf("%s %q: %s", ue.Op, u, redact(ue.Err.Error()))
	}
	return redact(err.Error())
}

// Result is a provider-neutral match, shaped like the MediaItem fields it fills.
type Result struct {
	Source      string `json:"source"`      // tmdb, openlibrary, anilist, itunes
	ExternalID  string `json:"external_id"` // provider-specific, e.g. "movie/949"
	Title       string `json:"title"`
	Creator     string `json:"creator"`
	ReleaseYear *int   `json:"release_year"`
	Synopsis    string `json:"synopsis"`
	Genres      string `json:"genres"`
	CoverURL    string `json:"cover_url"`
	Length      *int   `json:"length"`
}

type Service struct {
	client  *http.Client
	tmdbKey string

	// Base URLs, overridable in tests.
	TMDBBase      string
	TMDBImageBase string
	OpenLibBase   string
	OpenLibCovers string
	AniListURL    string
	ITunesBase    string
}

func NewService(tmdbKey string) *Service {
	return &Service{
		client:        &http.Client{Timeout: 10 * time.Second},
		tmdbKey:       strings.TrimSpace(tmdbKey),
		TMDBBase:      "https://api.themoviedb.org/3",
		TMDBImageBase: "https://image.tmdb.org/t/p",
		OpenLibBase:   "https://openlibrary.org",
		OpenLibCovers: "https://covers.openlibrary.org",
		AniListURL:    "https://graphql.anilist.co",
		ITunesBase:    "https://itunes.apple.com",
	}
}

const maxResults = 8

// Search returns candidate matches for a title. Results may be partial
// (e.g. no creator for TMDB); call Details on the chosen one.
func (s *Service) Search(ctx context.Context, mediaType models.MediaType, query string) ([]Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []Result{}, nil
	}
	var (
		results []Result
		err     error
	)
	switch mediaType {
	case models.MediaTypeMovie:
		results, err = s.tmdbSearch(ctx, "movie", query)
	case models.MediaTypeTV:
		results, err = s.tmdbSearch(ctx, "tv", query)
	case models.MediaTypeBook:
		results, err = s.openLibrarySearch(ctx, query)
	case models.MediaTypeAnime:
		results, err = s.aniListSearch(ctx, "ANIME", query)
	case models.MediaTypeComic:
		results, err = s.aniListSearch(ctx, "MANGA", query)
	case models.MediaTypeAlbum:
		results, err = s.iTunesSearch(ctx, "album", query)
	case models.MediaTypePodcast:
		results, err = s.iTunesSearch(ctx, "podcast", query)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, mediaType)
	}
	if err != nil {
		return nil, err
	}
	if len(results) > maxResults {
		results = results[:maxResults]
	}
	return results, nil
}

// Details returns the full record for a result previously returned by Search.
func (s *Service) Details(ctx context.Context, source, externalID string) (*Result, error) {
	switch source {
	case "tmdb":
		return s.tmdbDetails(ctx, externalID)
	case "openlibrary":
		return s.openLibraryDetails(ctx, externalID)
	case "anilist":
		return s.aniListDetails(ctx, externalID)
	case "itunes":
		return s.iTunesDetails(ctx, externalID)
	default:
		return nil, fmt.Errorf("%w: unknown source %q", ErrUnsupported, source)
	}
}

// --- shared helpers ---

func (s *Service) getJSON(ctx context.Context, rawURL string, header http.Header, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	for k, v := range header {
		req.Header[k] = v
	}
	return s.do(req, out)
}

func (s *Service) do(req *http.Request, out any) error {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "simple-track/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return newLookupError(ErrUpstream, "%s", transportDetail(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return newLookupError(ErrNotConfigured, "credentials rejected by %s", req.URL.Host)
	}
	if resp.StatusCode >= 300 {
		return newLookupError(ErrUpstream, "%s returned %d", req.URL.Host, resp.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 5<<20)).Decode(out); err != nil {
		return newLookupError(ErrUpstream, "decoding %s response: %v", req.URL.Host, err)
	}
	return nil
}

func withQuery(base string, params url.Values) string {
	return base + "?" + params.Encode()
}

// yearFromDate parses the year from "2006-01-02"-style or RFC3339 strings.
func yearFromDate(s string) *int {
	if len(s) < 4 {
		return nil
	}
	var y int
	if _, err := fmt.Sscanf(s[:4], "%d", &y); err != nil || y <= 0 {
		return nil
	}
	return &y
}

func intPtr(n int) *int {
	if n <= 0 {
		return nil
	}
	return &n
}

var (
	tagRe        = regexp.MustCompile(`<[^>]*>`)
	blankLinesRe = regexp.MustCompile(`\n{3,}`)
)

// cleanText strips HTML tags/entities and tidies whitespace.
func cleanText(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = html.UnescapeString(tagRe.ReplaceAllString(s, ""))
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimSpace(blankLinesRe.ReplaceAllString(s, "\n\n"))
}

func joinLimit(items []string, limit int) string {
	out := make([]string, 0, limit)
	seen := map[string]bool{}
	for _, it := range items {
		it = strings.TrimSpace(it)
		key := strings.ReplaceAll(strings.ToLower(it), "-", " ")
		if it == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, it)
		if len(out) == limit {
			break
		}
	}
	return strings.Join(out, ", ")
}

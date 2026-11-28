package lookup

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"be-simpletracker/internal/core/media/models"
)

// fakeAPI serves canned responses keyed by path and records requests.
func fakeAPI(t *testing.T, routes map[string]string) (*httptest.Server, *[]*http.Request) {
	t.Helper()
	var seen []*http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r)
		body, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv, &seen
}

func testService(key, base string) *Service {
	s := NewService(key)
	s.TMDBBase = base + "/3"
	s.TMDBImageBase = "https://img.test/t/p"
	s.OpenLibBase = base
	s.OpenLibCovers = "https://covers.test"
	s.AniListURL = base + "/graphql"
	s.ITunesBase = base
	return s
}

func TestTMDBMovieSearchAndDetails(t *testing.T) {
	srv, seen := fakeAPI(t, map[string]string{
		"/3/search/movie": `{"results":[{"id":949,"title":"Heat","release_date":"1995-12-15","overview":"Cops and robbers.","poster_path":"/heat.jpg"}]}`,
		"/3/movie/949": `{"id":949,"title":"Heat","release_date":"1995-12-15","overview":"Cops and robbers.","poster_path":"/heat.jpg","runtime":170,
			"genres":[{"name":"Action"},{"name":"Crime"}],"credits":{"crew":[{"name":"Michael Mann","job":"Director"},{"name":"Dante Spinotti","job":"Director of Photography"}]}}`,
	})
	s := testService("v3key", srv.URL)

	results, err := s.Search(context.Background(), models.MediaTypeMovie, "heat")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ExternalID != "movie/949" || *results[0].ReleaseYear != 1995 || results[0].CoverURL != "https://img.test/t/p/w185/heat.jpg" {
		t.Fatalf("unexpected search results: %+v", results)
	}
	if got := (*seen)[0].URL.Query(); got.Get("api_key") != "v3key" || got.Get("query") != "heat" {
		t.Fatalf("v3 key should be sent as api_key param, got %v", got)
	}

	d, err := s.Details(context.Background(), "tmdb", "movie/949")
	if err != nil {
		t.Fatal(err)
	}
	if d.Creator != "Michael Mann" || *d.Length != 170 || d.Genres != "Action, Crime" || d.CoverURL != "https://img.test/t/p/w500/heat.jpg" {
		t.Fatalf("unexpected details: %+v", d)
	}
}

func TestTMDBTVDetailsAndBearerToken(t *testing.T) {
	srv, seen := fakeAPI(t, map[string]string{
		"/3/tv/1438": `{"id":1438,"name":"The Wire","first_air_date":"2002-06-02","number_of_episodes":60,"created_by":[{"name":"David Simon"}],"genres":[{"name":"Crime"}]}`,
	})
	s := testService("eyJhbGciOiJIUzI1NiJ9.token", srv.URL)
	d, err := s.Details(context.Background(), "tmdb", "tv/1438")
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "The Wire" || d.Creator != "David Simon" || *d.Length != 60 || *d.ReleaseYear != 2002 {
		t.Fatalf("unexpected details: %+v", d)
	}
	req := (*seen)[0]
	if req.Header.Get("Authorization") != "Bearer eyJhbGciOiJIUzI1NiJ9.token" || req.URL.Query().Has("api_key") {
		t.Fatalf("v4 token should be sent as bearer header")
	}
}

func TestTMDBWithoutKeyIsNotConfigured(t *testing.T) {
	s := testService("", "http://unused.invalid")
	if _, err := s.Search(context.Background(), models.MediaTypeTV, "wire"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestOpenLibraryDetails(t *testing.T) {
	srv, _ := fakeAPI(t, map[string]string{
		"/search.json": `{"docs":[{"key":"/works/OL893414W","title":"Dune","author_name":["Frank Herbert"],"first_publish_year":1965,"cover_i":11481354,
			"number_of_pages_median":608,"subject":["Dune (Imaginary place)","Fiction","nyt:mass-market=2021","Science fiction","Hugo Award Winner","Ecology"]}]}`,
		"/works/OL893414W.json": `{"description":{"type":"/type/text","value":"Set on the desert planet Arrakis.\r\n\r\n----------\r\nAlso contained in: ..."}}`,
	})
	s := testService("", srv.URL)
	d, err := s.Details(context.Background(), "openlibrary", "/works/OL893414W")
	if err != nil {
		t.Fatal(err)
	}
	want := Result{Source: "openlibrary", ExternalID: "/works/OL893414W", Title: "Dune", Creator: "Frank Herbert",
		Synopsis: "Set on the desert planet Arrakis.", Genres: "Fiction, Science fiction, Ecology", CoverURL: "https://covers.test/b/id/11481354-L.jpg"}
	if d.Title != want.Title || d.Creator != want.Creator || d.Synopsis != want.Synopsis || d.Genres != want.Genres || d.CoverURL != want.CoverURL || *d.Length != 608 || *d.ReleaseYear != 1965 {
		t.Fatalf("got %+v\nwant %+v", d, want)
	}
	if _, err := s.Details(context.Background(), "openlibrary", "/works/../../etc"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected bad key to be rejected, got %v", err)
	}
}

func TestAniListMangaUsesAuthorsAndChapters(t *testing.T) {
	var gotVars map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		io.WriteString(w, `{"data":{"Page":{"media":[{"id":30013,"type":"MANGA","title":{"romaji":"One Piece","english":""},
			"startDate":{"year":1997},"coverImage":{"extraLarge":"https://a.test/op.jpg"},"description":"Pirates.<br><br>\n(Source: Viz)",
			"genres":["Action","Adventure"],"chapters":0,"volumes":107,
			"staff":{"edges":[{"role":"Story & Art"},{"role":"Assistant"}],"nodes":[{"name":{"full":"Eiichiro Oda"}},{"name":{"full":"Someone Else"}}]}}]}}}`)
	}))
	defer srv.Close()
	s := testService("", srv.URL)
	s.AniListURL = srv.URL

	results, err := s.Search(context.Background(), models.MediaTypeComic, "one piece")
	if err != nil {
		t.Fatal(err)
	}
	if gotVars["type"] != "MANGA" || gotVars["search"] != "one piece" {
		t.Fatalf("unexpected variables %v", gotVars)
	}
	r := results[0]
	if r.Title != "One Piece" || r.Creator != "Eiichiro Oda" || *r.Length != 107 || r.Synopsis != "Pirates.\n\n(Source: Viz)" || r.ExternalID != "30013" {
		t.Fatalf("unexpected result %+v", r)
	}
}

func TestITunesPodcastHasNoReleaseYear(t *testing.T) {
	srv, _ := fakeAPI(t, map[string]string{
		"/search": `{"results":[{"kind":"podcast","collectionId":173001861,"collectionName":"Hardcore History","artistName":"Dan Carlin",
			"artworkUrl600":"https://m.test/600x600bb.jpg","primaryGenreName":"History","trackCount":13,"releaseDate":"2026-07-31T18:02:00Z"}]}`,
	})
	s := testService("", srv.URL)
	results, err := s.Search(context.Background(), models.MediaTypePodcast, "hardcore history")
	if err != nil {
		t.Fatal(err)
	}
	r := results[0]
	if r.ReleaseYear != nil || r.Creator != "Dan Carlin" || *r.Length != 13 || r.ExternalID != "173001861" {
		t.Fatalf("unexpected result %+v", r)
	}
}

func TestITunesAlbumUpscalesArtwork(t *testing.T) {
	r := iTunesResult(iTunesItem{CollectionID: 1, CollectionName: "In Rainbows", ReleaseDate: "2007-10-10T07:00:00Z",
		ArtworkURL100: "https://m.test/x/100x100bb.jpg"})
	if r.CoverURL != "https://m.test/x/600x600bb.jpg" || *r.ReleaseYear != 2007 {
		t.Fatalf("unexpected %+v", r)
	}
}

func TestJoinLimitDedupesCaseAndHyphens(t *testing.T) {
	if got := joinLimit([]string{"Fiction", "Science fiction", "Science-Fiction", " fiction ", "Ecology"}, 3); got != "Fiction, Science fiction, Ecology" {
		t.Fatalf("got %q", got)
	}
}

func TestUnsupportedAndUpstreamErrors(t *testing.T) {
	s := testService("", "http://unused.invalid")
	if _, err := s.Search(context.Background(), models.MediaTypeGame, "zelda"); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
	if _, err := s.Details(context.Background(), "nope", "1"); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("expected ErrUnsupported, got %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	s = testService("", srv.URL)
	if _, err := s.Search(context.Background(), models.MediaTypeBook, "x"); !errors.Is(err, ErrUpstream) || !strings.Contains(Detail(err), "500") {
		t.Fatalf("expected ErrUpstream with status, got %v", err)
	}
	if res, err := s.Search(context.Background(), models.MediaTypeBook, "   "); err != nil || len(res) != 0 {
		t.Fatalf("blank query should return no results, got %v %v", res, err)
	}
}

func TestTransportErrorHasNoAPIKey(t *testing.T) {
	srv, _ := fakeAPI(t, nil)
	base := srv.URL
	srv.Close() // connection refused
	svc := testService("secretkey123", base)
	_, err := svc.Search(context.Background(), models.MediaTypeMovie, "heat")
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("want ErrUpstream, got %v", err)
	}
	for _, s := range []string{err.Error(), Detail(err)} {
		if strings.Contains(s, "secretkey123") || strings.Contains(s, "api_key") || strings.Contains(s, "query=") {
			t.Errorf("leaks request URL query: %s", s)
		}
	}
}

func TestUpstreamBodyNotEchoed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "internal secret upstream body")
	}))
	defer srv.Close()
	_, err := testService("k", srv.URL).Search(context.Background(), models.MediaTypeBook, "dune")
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("want ErrUpstream, got %v", err)
	}
	if strings.Contains(err.Error(), "secret") || strings.Contains(Detail(err), "secret") {
		t.Errorf("body echoed: %v / %s", err, Detail(err))
	}
}

func TestRedact(t *testing.T) {
	in := `Get "https://x/3/search?api_key=abc&query=q&token=t2": boom`
	out := redact(in)
	if strings.Contains(out, "abc") || strings.Contains(out, "t2") || !strings.Contains(out, "query=q") {
		t.Errorf("bad redaction: %s", out)
	}
}

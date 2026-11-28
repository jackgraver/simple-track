package lookup

import "testing"

func TestCoverAllowed(t *testing.T) {
	s := NewService("")
	for url, want := range map[string]bool{
		"": true,
		"https://image.tmdb.org/t/p/w500/abc.jpg":       true,
		"https://covers.openlibrary.org/b/id/1-L.jpg":   true,
		"https://s4.anilist.co/file/x.png":              true,
		"https://is1-ssl.mzstatic.com/image/a/600x.jpg": true,
		"http://image.tmdb.org/t/p/w500/abc.jpg":        false, // not https
		"https://evil.example/pixel.gif":                false,
		"https://image.tmdb.org.evil.example/x.jpg":     false,
		"https://evilanilist.co/x.png":                  false,
		"https://user@image.tmdb.org/x.jpg":             false,
		"javascript:alert(1)":                           false,
		"data:image/png;base64,AAAA":                    false,
		"/media/covers/1.jpg":                           false,
	} {
		if got := s.CoverAllowed(url); got != want {
			t.Errorf("CoverAllowed(%q) = %v, want %v", url, got, want)
		}
	}
	s.TMDBImageBase = "http://127.0.0.1:9/tmdb/img"
	if !s.CoverAllowed("http://127.0.0.1:9/tmdb/img/w185/x.jpg") || s.CoverAllowed("http://127.0.0.1:9/other/x.jpg") {
		t.Error("configured image base handling wrong")
	}
}

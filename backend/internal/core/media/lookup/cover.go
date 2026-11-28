package lookup

import (
	"net/url"
	"strings"
)

// Cover hosts the lookup sources serve images from (BACKLOG #15). A catalog cover is shown to
// every user's browser, so its host learns who browses media; only these hosts are accepted.
var (
	coverHosts        = []string{"image.tmdb.org", "covers.openlibrary.org"}
	coverHostSuffixes = []string{".anilist.co", ".mzstatic.com"} // AniList CDN, iTunes artwork
)

// CoverAllowed reports whether rawURL may be stored as a shared-catalog cover: empty, or an https
// URL on one of the lookup sources' image hosts. A URL under one of this service's configured image
// bases also passes, which is how the e2e suite's fake upstream (plain http, overridden bases)
// works; in production those bases are the https defaults.
func (s *Service) CoverAllowed(rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return true
	}
	if len(rawURL) > 2048 {
		return false
	}
	for _, base := range []string{s.TMDBImageBase, s.OpenLibCovers} {
		if base != "" && strings.HasPrefix(rawURL, strings.TrimRight(base, "/")+"/") {
			return true
		}
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Hostname() == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, h := range coverHosts {
		if host == h {
			return true
		}
	}
	for _, suf := range coverHostSuffixes {
		if strings.HasSuffix(host, suf) {
			return true
		}
	}
	return false
}

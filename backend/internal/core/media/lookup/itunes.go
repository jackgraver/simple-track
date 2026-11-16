package lookup

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type iTunesItem struct {
	WrapperType      string `json:"wrapperType"`
	Kind             string `json:"kind"`
	CollectionID     int    `json:"collectionId"`
	CollectionName   string `json:"collectionName"`
	ArtistName       string `json:"artistName"`
	ArtworkURL100    string `json:"artworkUrl100"`
	ArtworkURL600    string `json:"artworkUrl600"`
	PrimaryGenreName string `json:"primaryGenreName"`
	TrackCount       int    `json:"trackCount"`
	ReleaseDate      string `json:"releaseDate"`
}

func (s *Service) iTunesSearch(ctx context.Context, entity, query string) ([]Result, error) {
	var resp struct {
		Results []iTunesItem `json:"results"`
	}
	params := url.Values{"term": {query}, "entity": {entity}, "limit": {strconv.Itoa(maxResults)}}
	if err := s.getJSON(ctx, withQuery(s.ITunesBase+"/search", params), nil, &resp); err != nil {
		return nil, err
	}
	results := make([]Result, 0, len(resp.Results))
	for _, it := range resp.Results {
		results = append(results, iTunesResult(it))
	}
	return results, nil
}

func (s *Service) iTunesDetails(ctx context.Context, externalID string) (*Result, error) {
	if _, err := strconv.Atoi(externalID); err != nil {
		return nil, fmt.Errorf("%w: bad itunes id %q", ErrNotFound, externalID)
	}
	var resp struct {
		Results []iTunesItem `json:"results"`
	}
	if err := s.getJSON(ctx, withQuery(s.ITunesBase+"/lookup", url.Values{"id": {externalID}}), nil, &resp); err != nil {
		return nil, err
	}
	for _, it := range resp.Results {
		if strconv.Itoa(it.CollectionID) == externalID {
			r := iTunesResult(it)
			return &r, nil
		}
	}
	return nil, ErrNotFound
}

func iTunesResult(it iTunesItem) Result {
	cover := it.ArtworkURL600
	if cover == "" && it.ArtworkURL100 != "" {
		cover = strings.Replace(it.ArtworkURL100, "100x100bb", "600x600bb", 1)
	}
	r := Result{
		Source:     "itunes",
		ExternalID: strconv.Itoa(it.CollectionID),
		Title:      it.CollectionName,
		Creator:    it.ArtistName,
		Genres:     it.PrimaryGenreName,
		CoverURL:   cover,
		Length:     intPtr(it.TrackCount),
	}
	// A podcast's releaseDate is its latest episode, not when it started.
	if it.Kind != "podcast" {
		r.ReleaseYear = yearFromDate(it.ReleaseDate)
	}
	return r
}

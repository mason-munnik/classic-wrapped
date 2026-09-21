// Package musicbrainz matches classified composer/performer credits against
// the MusicBrainz API to resolve canonical work/recording IDs.

package musicbrainz

import (
	"context"
	"net/http"
	"time"
)

// These structs mirror the JSON returned by /ws/2/recording?query=...
// Only the fields we use are modelled; the decoder ignores the rest.

type RecordingSearchResponse struct {
	Recordings []Recording `json:"recordings"`
}

type Recording struct {
	ID           string         `json:"id"`
	Score        int            `json:"score"`
	Title        string         `json:"title"`
	ArtistCredit []ArtistCredit `json:"artist-credit"`
}

type ArtistCredit struct {
	Name   string `json:"name"`
	Artist Artist `json:"artist"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Client talks to the MusicBrainz web service. BaseURL is a field rather
// than a constant so tests can point it at a local httptest server.
type Client struct {
	BaseURL    string
	UserAgent  string
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		BaseURL:    "https://musicbrainz.org/ws/2",
		UserAgent:  "classic-wrapped/0.1 ( masonmunnik@gmail.com )",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Search(context context.Context, query string) (*RecordingSearchResponse, error) {

}

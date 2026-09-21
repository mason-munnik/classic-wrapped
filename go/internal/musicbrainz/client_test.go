package musicbrainz

import (
	"context"
	"testing"
)

// TestSearchLive hits the real MusicBrainz API. Skipped under -short so CI
// stays offline; run by hand with:
//
//	go test ./internal/musicbrainz/ -run TestSearchLive -v
func TestSearchLive(t *testing.T) {
	if testing.Short() {
		t.Skip("hits the network")
	}

	c := NewClient()
	query := `recording:"Adagietto" AND artist:"London Philharmonic Orchestra"`

	resp, err := c.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Search(%q) failed: %v", query, err)
	}
	if len(resp.Recordings) == 0 {
		t.Fatalf("Search(%q) returned no recordings", query)
	}

	for _, rec := range resp.Recordings {
		names := make([]string, 0, len(rec.ArtistCredit))
		for _, credit := range rec.ArtistCredit {
			names = append(names, credit.Name)
		}
		t.Logf("score=%3d id=%s title=%q artists=%v", rec.Score, rec.ID, rec.Title, names)
	}
}

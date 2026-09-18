package classical

import (
	"testing"
	"time"

	"github.com/mason-munnik/classic-wrapped/go/internal/ingest"
)

func TestSplitArtistTokensFeat(t *testing.T) {
	artist := "Yo-Yo Ma feat. Kathryn Stott"
	got := SplitArtistTokens(artist)
	want := []string{"Yo-Yo Ma", "Kathryn Stott"}
	if len(got) != len(want) {
		t.Fatalf("SplitArtistTokens(%q) returned %v, want %v", artist, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SplitArtistTokens(%q) returned %v, want %v", artist, got, want)
		}
	}
}

func TestClassifyMahlerAlbumTitle(t *testing.T) {
	play := ingest.Play{
		TS:       time.Now(),
		MSPlayed: 120000,
		Track:    "Symphony No. 5 in C-Sharp Minor: IV. Adagietto. Sehr langsam",
		Artist:   "London Philharmonic Orchestra",
		Album:    "Mahler: Symphony No. 5",
		TrackURI: "spotify:track:test",
	}

	result := Classify(play)
	if !result.IsClassical {
		t.Fatalf("Classify(%+v) = %v; want classical", play, result)
	}
	if result.Composer.CanonicalName != "Gustav Mahler" {
		t.Fatalf("Classify(%+v) composer = %q; want %q", play, result.Composer.CanonicalName, "Gustav Mahler")
	}
	if result.MatchSource != "album_title" {
		t.Fatalf("Classify(%+v) match_source = %q; want %q", play, result.MatchSource, "album_title")
	}
}

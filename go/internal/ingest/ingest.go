// Package ingest parses a Spotify Extended Streaming History export into
// Play records.
//
// Ports python/app/ingest.py:
//   - Play struct       (ts, msPlayed, track, artist, album, trackURI)
//   - LoadJSONFile(path string) ([]map[string]any, error)
//   - ParseTimestamp(raw string) (time.Time, error)
//   - ParseEntry(entry map[string]any) (Play, error)
//
// Test fixture at testdata/spotify_sample.json (copied from
// python/tests/fixtures/spotify_sample.json).
package ingest

import (
	"encoding/json"
	"os"
	"time"
)

type Play struct {
	TS       time.Time
	MSPlayed int
	Track    string
	Artist   string
	Album    string
	TrackURI string
}

func LoadJSONFile(filePath string) ([]map[string]any, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	var records []map[string]any
	if err := json.NewDecoder(file).Decode(&records); err != nil {
		return nil, err
	}
	return records, nil
}

func ParseEntry(entry map[string]any) Play {
	tsRaw, _ := entry["ts"].(string)
	ts, _ := time.Parse(time.RFC3339, tsRaw)
	msPlayed, _ := entry["ms_played"].(float64)
	track, _ := entry["master_metadata_track_name"].(string)
	artist, _ := entry["master_metadata_album_artist_name"].(string)
	album, _ := entry["master_metadata_album_album_name"].(string)
	trackURI, _ := entry["spotify_track_uri"].(string)

	return Play{
		TS:       ts,
		MSPlayed: int(msPlayed),
		Track:    track,
		Artist:   artist,
		Album:    album,
		TrackURI: trackURI,
	}
}

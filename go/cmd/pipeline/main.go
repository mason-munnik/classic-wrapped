// Command pipeline runs the classic-wrapped ingest → classify → match →
// aggregate pipeline.
//
// Stages live in internal/ingest, internal/classical, internal/musicbrainz,
// and internal/aggregate — mirrors python/app/ in the sibling python/
// directory, which stays in place as a reference/backup during the port.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mason-munnik/classic-wrapped/go/internal/classical"
	"github.com/mason-munnik/classic-wrapped/go/internal/ingest"
)

func main() {
	path := resolveInputPath()
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	fmt.Printf("Classic-wrapped pipeline\n")
	fmt.Printf("Input: %s\n\n", path)

	records, err := ingest.LoadJSONFile(path)
	if err != nil {
		fmt.Printf("load error: %v\n", err)
		os.Exit(1)
	}
	if records == nil {
		fmt.Println("No records loaded.")
		os.Exit(1)
	}

	fmt.Printf("Loaded %d raw Spotify rows.\n\n", len(records))

	showNormalizationDemo()
	showAliasIndexDemo()
	showClassificationSummary(records)
}

func resolveInputPath() string {
	candidates := []string{
		filepath.Join("go", "internal", "ingest", "testdata", "spotify_sample.json"),
		filepath.Join("internal", "ingest", "testdata", "spotify_sample.json"),
		filepath.Join("..", "go", "internal", "ingest", "testdata", "spotify_sample.json"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return filepath.Join("go", "internal", "ingest", "testdata", "spotify_sample.json")
}

func showNormalizationDemo() {
	fmt.Println("Normalization demo:")
	fmt.Printf("  %q -> %q\n", "Hello,         world!", classical.NormalizeText("Hello,         world!"))
	fmt.Printf("  %q -> %q\n\n", "Café au lait", classical.NormalizeText("Café au lait"))
}

func showAliasIndexDemo() {
	composers := []classical.Composer{
		{CanonicalName: "Richard Strauss", Aliases: []string{"Strauss"}},
		{CanonicalName: "Johann Strauss II", Aliases: []string{"Johann Strauss"}},
		{CanonicalName: "Wolfgang Amadeus Mozart", Aliases: []string{"Mozart", "W. A. Mozart"}},
		{CanonicalName: "Gustav Mahler", Aliases: []string{"Mahler"}},
	}

	pairs := classical.BuildAliasIndex(composers)
	fmt.Println("Alias index preview:")
	for i := range pairs {
		if i == 8 {
			break
		}
		fmt.Printf("  %v\n", pairs[i])
	}
	fmt.Println()

	artists := []string{
		"Herbert von Karajan; Berlin Philharmonic",
		"Glenn Gould",
		"Yo-Yo Ma feat. Kathryn Stott",
		"A and B with C vs. D",
		"A, B / C & D",
		"",
	}

	fmt.Println("Artist tokenization preview:")
	for _, artist := range artists {
		fmt.Printf("  %q => %#v\n", artist, classical.SplitArtistTokens(artist))
	}
	fmt.Println()
}

func showClassificationSummary(records []map[string]any) {
	type classificationSummary struct {
		composer string
		count    int
	}

	stats := map[string]int{"classical": 0, "nonClassical": 0, "total": 0}
	composerCounts := map[string]int{}
	fmt.Println("Classification results:")
	fmt.Println("  TRACK                                        ARTIST                          COMPOSER                 SOURCE")
	fmt.Println("  ------------------------------------------ ------------------------------ ------------------------ --------")

	for _, raw := range records {
		play := ingest.ParseEntry(raw)
		if play.Track == "" && play.Artist == "" && play.Album == "" && play.TrackURI == "" {
			continue
		}

		stats["total"]++
		result := classical.Classify(play)
		composerName := "—"
		if result.Composer.CanonicalName != "" {
			composerName = result.Composer.CanonicalName
			composerCounts[composerName]++
		}
		if result.IsClassical {
			stats["classical"]++
		} else {
			stats["nonClassical"]++
		}

		trackName := strings.TrimSpace(play.Track)
		if trackName == "" {
			trackName = "(n/a)"
		}
		artistName := strings.TrimSpace(play.Artist)
		if artistName == "" {
			artistName = "(n/a)"
		}
		fmt.Printf("  %-42s %-30s %-24s %-12s\n", trunc(trackName, 42), trunc(artistName, 30), trunc(composerName, 24), result.MatchSource)
	}

	fmt.Println()
	fmt.Println("Pipeline summary:")
	fmt.Printf("  records scanned: %d\n", stats["total"])
	fmt.Printf("  classical matches: %d\n", stats["classical"])
	fmt.Printf("  non-classical: %d\n", stats["nonClassical"])
	if len(composerCounts) > 0 {
		fmt.Println("  detected composers:")
		ordered := make([]classificationSummary, 0, len(composerCounts))
		for name, count := range composerCounts {
			ordered = append(ordered, classificationSummary{composer: name, count: count})
		}
		sort.Slice(ordered, func(i, j int) bool {
			if ordered[i].count == ordered[j].count {
				return ordered[i].composer < ordered[j].composer
			}
			return ordered[i].count > ordered[j].count
		})
		for _, item := range ordered {
			fmt.Printf("    - %s: %d plays\n", item.composer, item.count)
		}
	}
}

func trunc(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return s[:width-1] + "…"
}

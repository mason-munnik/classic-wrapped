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

	"github.com/mason-munnik/classic-wrapped/go/internal/classical"
	"github.com/mason-munnik/classic-wrapped/go/internal/ingest"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pipeline <path-to-spotify-export.json>")
		os.Exit(1)
	}
	records, err := ingest.LoadJSONFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, r := range records {
		play := ingest.ParseEntry(r)
		fmt.Printf("%v\n", play)
	}

	fmt.Println(classical.NormalizeText("Hello,         world!"))

	// Demo: BuildAliasIndex on a couple of composers with a deliberately
	// ambiguous surname ("Strauss") to see the sort + normalize in action.
	composers := []classical.Composer{
		{CanonicalName: "Richard Strauss", Aliases: []string{"Strauss"}},
		{CanonicalName: "Johann Strauss II", Aliases: []string{"Johann Strauss"}},
		{CanonicalName: "Wolfgang Amadeus Mozart", Aliases: []string{"Mozart", "W. A. Mozart"}},
	}
	pairs := classical.BuildAliasIndex(composers)
	for _, p := range pairs {
		fmt.Printf("%+v\n", p)
	}

	// Demo: SplitArtistTokens on a few artist strings covering the
	// different separators (semicolon, comma, slash, &, feat./ft./and/with/vs.).
	artistSamples := []string{
		"Herbert von Karajan; Berlin Philharmonic",
		"Glenn Gould",
		"Yo-Yo Ma feat. Kathryn Stott",
		"A and B with C vs. D",
		"A, B / C & D",
		"",
	}
	for _, artist := range artistSamples {
		fmt.Printf("%q => %#v\n", artist, classical.SplitArtistTokens(artist))
	}
}

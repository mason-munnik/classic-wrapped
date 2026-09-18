// TODO:
//   - Classify(p ingest.Play) ClassicalMatch

package classical

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/mason-munnik/classic-wrapped/go/internal/ingest"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var splitPattern = regexp.MustCompile(
	`(?i)\s*(?:[;,/&]|\bfeat\.?|\bft\.?|\band\b|\bwith\b|\bvs\.?)\s*`,
)

type ClassicalMatch struct {
	Play        ingest.Play
	IsClassical bool
	Composer    Composer
	Performer   string
	MatchSource string
}

type Composer struct {
	CanonicalName string   `json:"canonical_name"`
	Aliases       []string `json:"aliases"`
}

type aliasPair struct {
	alias         string
	canonicalName string
}

var stripAccents = transform.Chain(
	norm.NFKD,
	runes.Remove(runes.In(unicode.Mn)),
)

func NormalizeText(s string) string {
	if s == "" {
		return ""
	}
	stripped, _, err := transform.String(stripAccents, s)
	if err != nil {
		return s
	}
	result := strings.Join(strings.Fields(stripped), " ")
	return strings.ToLower(result)
}

func BuildAliasIndex(composers []Composer) []aliasPair {
	var pairs []aliasPair
	for _, composer := range composers {
		pairs = append(pairs, aliasPair{alias: composer.CanonicalName, canonicalName: composer.CanonicalName})
		for _, alias := range composer.Aliases {
			pairs = append(pairs, aliasPair{alias: alias, canonicalName: composer.CanonicalName})
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return len(pairs[i].alias) > len(pairs[j].alias)
	})

	for i := range pairs {
		pairs[i].alias = NormalizeText(pairs[i].alias)
	}

	return pairs
}

func SplitArtistTokens(artist string) []string {
	if artist == "" {
		return nil
	}
	tokens := splitPattern.Split(artist, -1)
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func MatchComposerInText(text string, composers []Composer) (canonical string, ok bool) {
	normalized := NormalizeText(text)
	if normalized == "" {
		return "", false
	}
	composerAliasIndex := BuildAliasIndex(composers)

	for _, pair := range composerAliasIndex {
		pattern := `\b` + regexp.QuoteMeta(pair.alias) + `\b`
		matched, err := regexp.MatchString(pattern, normalized)
		if err != nil {
			continue
		}
		if matched {
			return pair.canonicalName, true
		}
	}

	return "", false
}

func AlbumPrefix(album string) string {
	if album == "" || !strings.Contains(album, ":") {
		return ""
	}
	prefix := strings.Split(album, ":")[0]
	return strings.TrimSpace(prefix)
}

func noMatch(p ingest.Play) ClassicalMatch {
	return ClassicalMatch{
		Play:        p,
		IsClassical: false,
		Composer:    Composer{},
		Performer:   "",
		MatchSource: "none",
	}
}

func loadComposers() []Composer {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil
	}

	candidates := []string{
		filepath.Join(filepath.Dir(file), "data", "composers.json"),
		filepath.Join(filepath.Dir(file), "..", "..", "..", "python", "app", "data", "composers.json"),
		filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "python", "app", "data", "composers.json"),
		filepath.Join(".", "python", "app", "data", "composers.json"),
		filepath.Join("..", "python", "app", "data", "composers.json"),
	}

	for _, path := range candidates {
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var composers []Composer
		if err := json.Unmarshal(raw, &composers); err != nil {
			continue
		}
		return composers
	}

	return nil
}

func Classify(play ingest.Play) ClassicalMatch {
	if play.Artist == "" && play.Album == "" {
		return noMatch(play)
	}

	composers := loadComposers()
	tokens := SplitArtistTokens(play.Artist)
	for index, token := range tokens {
		composerName, ok := MatchComposerInText(token, composers)
		if ok {
			remaining := make([]string, 0, len(tokens)-1)
			for i, part := range tokens {
				if i != index {
					remaining = append(remaining, part)
				}
			}
			return ClassicalMatch{
				Play:        play,
				IsClassical: true,
				Composer:    Composer{CanonicalName: composerName},
				Performer:   strings.Join(remaining, "; "),
				MatchSource: "artist",
			}
		}
	}

	composerName, ok := MatchComposerInText(AlbumPrefix(play.Album), composers)
	if ok {
		return ClassicalMatch{
			Play:        play,
			IsClassical: true,
			Composer:    Composer{CanonicalName: composerName},
			Performer:   strings.TrimSpace(play.Artist),
			MatchSource: "album_title",
		}
	}

	return noMatch(play)
}

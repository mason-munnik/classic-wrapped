"""Classical-music detection over ``Play`` records parsed from a Spotify export.

Spotify's Extended Streaming History export has no role-tagged credits: there
is a single flat artist string and an album title, never separate
composer/performer/conductor/orchestra fields. Depending on how a label
tagged a classical release, the artist field might hold the performer (with
the composer only appearing as a prefix on the album title, e.g.
``"Mahler: Symphony No. 5"``), the composer alone, a combined
"Composer; Performer" string, or just be an ordinary non-classical artist
name. ``classify`` decides whether a play is classical and, if so, splits out
the composer from everything else.

v1 limitation: since the export carries no role tags, anything that isn't
the matched composer collapses into a single ``performer`` string —
conductor, orchestra, and soloist are not distinguished from one another.

Surname collisions among composers are resolved by a fixed default-alias
rule baked into ``composers.json``, not by true disambiguation: only the
most likely composer for an ambiguous bare surname gets that surname as an
alias (e.g. "Bach" resolves to J.S. Bach, not C.P.E. or J.C. Bach; "Strauss"
resolves to Richard Strauss, not Johann Strauss II, which instead keys off
"Johann Strauss"). Composers whose surname doubles as a common English word
or name (Adams, Williams, Price, Still, Bliss, Arnold, Cage, Ireland,
Bridge, Wolf, Harris, and similar) get no bare-surname alias at all, so
matching for them only fires on their full canonical name.
"""

import json
import re
import unicodedata
from dataclasses import dataclass
from pathlib import Path

from app.ingest import Play

COMPOSER_DATA_PATH = Path(__file__).parent / "data" / "composers.json"

_SPLIT_PATTERN = re.compile(
    r"\s*(?:[;,/&]|\bfeat\.?\b|\bft\.?\b|\band\b|\bwith\b|\bvs\.?\b)\s*",
    re.IGNORECASE,
)


def load_composers():
    with open(COMPOSER_DATA_PATH, "r", encoding="utf-8") as file:
        return json.load(file)


def normalize_text(text):
    if text is None:
        return ""
    decomposed = unicodedata.normalize("NFKD", text)
    without_accents = "".join(ch for ch in decomposed if not unicodedata.combining(ch))
    return re.sub(r"\s+", " ", without_accents).strip().lower()


def build_alias_index(composers):
    pairs = []
    for composer in composers:
        canonical_name = composer["canonical_name"]
        pairs.append((canonical_name, canonical_name))
        for alias in composer["aliases"]:
            pairs.append((alias, canonical_name))
    pairs.sort(key=lambda pair: len(pair[0]), reverse=True)
    return [(normalize_text(alias), canonical_name) for alias, canonical_name in pairs]


COMPOSERS = load_composers()
COMPOSER_ALIAS_INDEX = build_alias_index(COMPOSERS)


def split_artist_tokens(artist):
    if not artist:
        return []
    tokens = _SPLIT_PATTERN.split(artist)
    return [token.strip() for token in tokens if token.strip()]


def match_composer_in_text(text):
    normalized = normalize_text(text)
    if not normalized:
        return None
    for alias, canonical_name in COMPOSER_ALIAS_INDEX:
        if re.search(rf"\b{re.escape(alias)}\b", normalized):
            return canonical_name
    return None


def album_prefix(album):
    if not album or ":" not in album:
        return None
    return album.split(":", 1)[0].strip()


@dataclass(frozen=True)
class ClassicalMatch:
    play: Play
    is_classical: bool
    composer: str | None
    performer: str | None
    match_source: str  # "artist" | "album_title" | "none"


def _no_match(play):
    return ClassicalMatch(
        play=play,
        is_classical=False,
        composer=None,
        performer=None,
        match_source="none",
    )


def classify(play: Play) -> ClassicalMatch:
    if play.artist is None and play.album is None:
        return _no_match(play)

    tokens = split_artist_tokens(play.artist)
    for index, token in enumerate(tokens):
        composer = match_composer_in_text(token)
        if composer is not None:
            remaining = [t for i, t in enumerate(tokens) if i != index]
            performer = "; ".join(remaining) if remaining else None
            return ClassicalMatch(
                play=play,
                is_classical=True,
                composer=composer,
                performer=performer,
                match_source="artist",
            )

    composer = match_composer_in_text(album_prefix(play.album))
    if composer is not None:
        performer = play.artist.strip() if play.artist else None
        return ClassicalMatch(
            play=play,
            is_classical=True,
            composer=composer,
            performer=performer,
            match_source="album_title",
        )

    return _no_match(play)

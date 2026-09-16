from pathlib import Path

from app.classical_detector import classify
from app.ingest import Play, load_json_file, parse_entry

FIXTURE_PATH = Path(__file__).parent / "fixtures" / "spotify_sample.json"


def _play(artist=None, album=None, track=None):
    return Play(ts=None, ms_played=0, track=track, artist=artist, album=album, track_uri=None)


def test_classify_composer_given_directly_as_artist_is_classical():
    play = _play(artist="Gustav Mahler", album="Symphony No. 5")
    result = classify(play)
    assert result.is_classical is True
    assert result.composer == "Gustav Mahler"
    assert result.performer is None
    assert result.match_source == "artist"


def test_classify_real_fixture_entry_matches_composer_via_album_title():
    records = load_json_file(str(FIXTURE_PATH))
    play = parse_entry(records[3])
    result = classify(play)
    assert result.is_classical is True
    assert result.composer == "Gustav Mahler"
    assert result.performer == "London Philharmonic Orchestra"
    assert result.match_source == "album_title"


def test_classify_combined_composer_and_performer_string_splits_on_semicolon():
    play = _play(artist="Gustav Mahler; London Philharmonic Orchestra")
    result = classify(play)
    assert result.is_classical is True
    assert result.composer == "Gustav Mahler"
    assert result.performer == "London Philharmonic Orchestra"
    assert result.match_source == "artist"


def test_classify_ampersand_delimiter_splits_correctly():
    play = _play(artist="Mozart & Vienna Philharmonic")
    result = classify(play)
    assert result.is_classical is True
    assert result.composer == "Wolfgang Amadeus Mozart"
    assert result.performer == "Vienna Philharmonic"
    assert result.match_source == "artist"


def test_classify_non_classical_pop_artist_is_not_classical():
    play = _play(artist="Taylor Swift", album="1989")
    result = classify(play)
    assert result.is_classical is False
    assert result.composer is None
    assert result.performer is None
    assert result.match_source == "none"


def test_classify_podcast_shaped_play_is_not_classical():
    play = _play(artist=None, album=None, track=None)
    result = classify(play)
    assert result.is_classical is False
    assert result.composer is None
    assert result.performer is None
    assert result.match_source == "none"


def test_classify_bachman_turner_overdrive_is_not_misdetected_as_bach():
    play = _play(artist="Bachman-Turner Overdrive", album="Not Fragile")
    result = classify(play)
    assert result.is_classical is False
    assert result.composer is None


def test_classify_artist_match_takes_priority_over_album_title_colon():
    play = _play(artist="Gustav Mahler", album="Mahler: Symphony No. 5")
    result = classify(play)
    assert result.is_classical is True
    assert result.match_source == "artist"


def test_classify_non_classical_artist_with_colon_in_album_title_is_not_classical():
    play = _play(artist="Taylor Swift", album="Greatest Hits: Deluxe Edition")
    result = classify(play)
    assert result.is_classical is False
    assert result.composer is None
    assert result.performer is None
    assert result.match_source == "none"

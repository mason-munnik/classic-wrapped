from datetime import datetime, timezone
from pathlib import Path

from app.ingest import Play, load_json_file, parse_entry, parse_timestamp

FIXTURE_PATH = Path(__file__).parent / "fixtures" / "spotify_sample.json"


def test_load_json_file_missing_file_returns_none(capsys, tmp_path):
    result = load_json_file(str(tmp_path / "does_not_exist.json"))
    assert result is None
    assert "does not exist" in capsys.readouterr().out


def test_load_json_file_invalid_json_returns_none(capsys, tmp_path):
    bad_file = tmp_path / "bad.json"
    bad_file.write_text("{not valid json")
    result = load_json_file(str(bad_file))
    assert result is None
    assert "is not valid JSON" in capsys.readouterr().out


def test_load_json_file_valid_file_loads_correctly():
    result = load_json_file(str(FIXTURE_PATH))
    assert isinstance(result, list)
    assert len(result) == 4
    assert result[0]["master_metadata_track_name"] == "Test Track One"
    assert result[2]["master_metadata_track_name"] is None
    assert result[3]["master_metadata_album_artist_name"] == "London Philharmonic Orchestra"


def test_load_json_file_generic_exception_returns_none(capsys, tmp_path, monkeypatch):
    exists_file = tmp_path / "exists.json"
    exists_file.write_text("{}")

    def raise_permission_error(*args, **kwargs):
        raise PermissionError("permission denied")

    monkeypatch.setattr("builtins.open", raise_permission_error)

    result = load_json_file(str(exists_file))
    assert result is None
    assert "An unexpected error occurred" in capsys.readouterr().out


def test_parse_timestamp_valid_iso8601_no_timezone():
    assert parse_timestamp("2024-01-15T10:30:00") == datetime(2024, 1, 15, 10, 30, 0)


def test_parse_timestamp_valid_iso8601_with_z_suffix():
    assert parse_timestamp("2024-01-15T10:30:00Z") == datetime(
        2024, 1, 15, 10, 30, 0, tzinfo=timezone.utc
    )


def test_parse_timestamp_none_returns_none():
    assert parse_timestamp(None) is None


def test_parse_entry_full_music_entry():
    entry = {
        "ts": "2024-01-15T10:30:00Z",
        "platform": "android",
        "ms_played": 215000,
        "conn_country": "US",
        "ip_addr": "0.0.0.0",
        "master_metadata_track_name": "Test Track One",
        "master_metadata_album_artist_name": "Test Artist One",
        "master_metadata_album_album_name": "Test Album One",
        "spotify_track_uri": "spotify:track:0000000000000000000001",
        "episode_name": None,
        "episode_show_name": None,
        "spotify_episode_uri": None,
        "reason_start": "trackdone",
        "reason_end": "trackdone",
        "shuffle": False,
        "skipped": None,
        "offline": False,
        "offline_timestamp": 0,
        "incognito_mode": False,
    }
    play = parse_entry(entry)
    assert play.ts == datetime(2024, 1, 15, 10, 30, 0, tzinfo=timezone.utc)
    assert play.ms_played == 215000
    assert play.track == "Test Track One"
    assert play.artist == "Test Artist One"
    assert play.album == "Test Album One"
    assert play.track_uri == "spotify:track:0000000000000000000001"


def test_parse_entry_null_timestamp():
    entry = {
        "ts": None,
        "platform": "android",
        "ms_played": 215000,
        "master_metadata_track_name": "Test Track One",
        "master_metadata_album_artist_name": "Test Artist One",
        "master_metadata_album_album_name": "Test Album One",
        "spotify_track_uri": "spotify:track:0000000000000000000001",
    }
    play = parse_entry(entry)
    assert play.ts is None
    assert play.ms_played == 215000
    assert play.track == "Test Track One"
    assert play.artist == "Test Artist One"
    assert play.album == "Test Album One"
    assert play.track_uri == "spotify:track:0000000000000000000001"


def test_parse_entry_podcast_shaped_entry_not_filtered():
    entry = {
        "ts": "2024-01-15T11:40:10Z",
        "platform": "ios",
        "ms_played": 63000,
        "conn_country": "US",
        "ip_addr": "0.0.0.0",
        "master_metadata_track_name": None,
        "master_metadata_album_artist_name": None,
        "master_metadata_album_album_name": None,
        "spotify_track_uri": None,
        "episode_name": "Test Episode One",
        "episode_show_name": "Test Podcast Show",
        "spotify_episode_uri": "spotify:episode:0000000000000000000099",
        "reason_start": "clickrow",
        "reason_end": "endplay",
        "shuffle": False,
        "skipped": None,
        "offline": False,
        "offline_timestamp": 0,
        "incognito_mode": False,
    }
    result = parse_entry(entry)
    assert isinstance(result, Play)
    assert result.track is None
    assert result.artist is None
    assert result.album is None
    assert result.track_uri is None
    assert isinstance(result.ts, datetime)
    assert result.ms_played == 63000


def test_parse_entries_from_fixture_file_all_parse_without_filtering():
    records = load_json_file(str(FIXTURE_PATH))
    plays = [parse_entry(r) for r in records]
    assert len(plays) == 4
    assert all(isinstance(p, Play) for p in plays)
    assert plays[2].track is None and plays[2].artist is None
    assert plays[3].artist == "London Philharmonic Orchestra"
    assert plays[3].album == "Mahler: Symphony No. 5"

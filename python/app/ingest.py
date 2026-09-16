import json
import os
import sys
from dataclasses import dataclass
from datetime import datetime

@dataclass(frozen=True)
class Play:
    ts: datetime
    ms_played: int
    track: str
    artist: str
    album: str
    track_uri: str

def load_json_file(file_path):
    """
    loads JSON file into python list
    """
    if not os.path.exists(file_path):
        print(f"Error: the file '{file_path}' does not exist")
        return None
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            data = json.load(file)
            return data

    except json.JSONDecodeError:
        print(f"Error: '{file_path}' is not valid JSON")
        return None

    except Exception as e:  # pylint: disable=broad-exception-caught
        # Intentional catch-all: any unexpected I/O error here should be
        # reported to the user, not crash the CLI.
        print(f"An unexpected error occurred: {e}")
        return None

def parse_timestamp(raw):
    if raw is None:
        return None
    return datetime.fromisoformat(raw)

def parse_entry(entry):
    return Play(
        ts= parse_timestamp(entry.get("ts")),
        ms_played= entry.get("ms_played"),
        track= entry.get("master_metadata_track_name"),
        artist= entry.get("master_metadata_album_artist_name"),
        album= entry.get("master_metadata_album_album_name"),
        track_uri= entry.get("spotify_track_uri")
    )

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python ingest.py <path-to-json>")
        sys.exit(1)

    records = load_json_file(sys.argv[1])
    if records is None:
        sys.exit(1)

    for r in records:
        print(parse_entry(r))

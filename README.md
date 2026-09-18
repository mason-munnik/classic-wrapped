# classic-wrapped
A service inspired by my love of classical music, and my anger with Spotify wrapped for having such poor listening data geared towards classical music fans.

Written by hand, with guidance with CI/CD and testing from Claude Code

## Running the Go demo

From the Go module directory:

```bash
cd go
go run ./cmd/pipeline
```

This runs the main pipeline entrypoint and prints a readable end-to-end demo of the classical detection flow, including:

- loading the Spotify export JSON
- normalizing text
- building the composer alias index
- splitting artist strings like "feat." and ";"
- classifying tracks as classical or non-classical
- summarizing detected composers

You can also point it at a different JSON export file:

```bash
go run ./cmd/pipeline /path/to/spotify-export.json
```
# RockList

RockList is a Go CLI tool that automatically creates playlists for your Rockbox media player based on artists' top tracks from Last.fm and Spotify.

## Features

- Read Rockbox database to identify artists in your collection
- Fetch artists' top tracks from Last.fm or Spotify
- Generate playlists compatible with Rockbox
- Dockerized for easy deployment
- Configurable via environment variables or command-line flags

## Requirements

- Go 1.20+ (for development)
- Docker (for containerized deployment)
- Last.fm API key
- Spotify API credentials

## Installation

### Using Docker (Recommended)

```bash
# Build the Docker image
docker build -t rocklist .

# Run with environment variables
docker run --rm -v /path/to/rockbox:/rockbox \
  -e LASTFM_API_KEY="your_api_key" \
  -e LASTFM_API_SECRET="your_api_secret" \
  -e SPOTIFY_CLIENT_ID="your_client_id" \
  -e SPOTIFY_CLIENT_SECRET="your_client_secret" \
  rocklist
```

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/rocklist.git
cd rocklist

# Build the binary
go build -o rocklist cmd/main.go

# Run the application
./rocklist
```

## Configuration

You can configure RockList using environment variables or command-line flags:

```bash
# Using environment variables
LASTFM_API_KEY="your_key" \
LASTFM_API_SECRET="your_secret" \
SPOTIFY_CLIENT_ID="your_id" \
SPOTIFY_CLIENT_SECRET="your_secret" \
ROCKBOX_PATH="/path/to/rockbox" \
PLAYLIST_PATH="/path/to/playlists" \
API_SOURCE="lastfm" \
./rocklist

# Or using command-line flags
./rocklist \
  --api-source lastfm \
  --rockbox-path /path/to/rockbox \
  --playlist-path /path/to/playlists \
  --artist "Artist Name" \
  --tracks 10
```

## Configuration File

You can also use a `.env` file:

```
LASTFM_API_KEY=your_key
LASTFM_API_SECRET=your_secret
SPOTIFY_CLIENT_ID=your_id
SPOTIFY_CLIENT_SECRET=your_secret
ROCKBOX_PATH=/path/to/rockbox
PLAYLIST_PATH=/path/to/playlists
API_SOURCE=lastfm
```

## Project Structure

- `cmd/` - Application entry point
- `internal/` - Private application code
  - `api/` - Last.fm and Spotify API clients
  - `config/` - Application configuration
  - `database/` - Rockbox database parsing
  - `playlist/` - Playlist generation logic
- `pkg/` - Public libraries that could be used by external applications

## File Descriptions

- `cmd/main.go` - Entry point for the CLI application
- `internal/config/config.go` - Configuration handling using environment variables and CLI flags
- `internal/database/rockbox.go` - Rockbox database parser
- `internal/api/lastfm.go` - Last.fm API client
- `internal/api/spotify.go` - Spotify API client
- `internal/playlist/generator.go` - Playlist creation logic
- `pkg/util/fileutil.go` - File handling utilities
- `Dockerfile` - Container definition for the application
- `.env.example` - Example environment configuration
- `go.mod` - Go module definition

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 

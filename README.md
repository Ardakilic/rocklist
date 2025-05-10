# RockList

RockList is a Go CLI tool that automatically creates playlists for your Rockbox media player based on artists' top tracks from Last.fm and Spotify.

## Features

- Read Rockbox database to identify artists in your collection
- Fetch artists' top tracks from Last.fm or Spotify
- Generate playlists compatible with Rockbox
- Dockerized for easy deployment
- Configurable via environment variables or command-line flags

## Requirements

- Go 1.24.3 (for development)
- Docker (for containerized deployment)
- Either one of these, to get top songs of an Artist:
  - Last.fm API credentials
  - Spotify API credentials

## How to obtain API credentials

### Last.fm API credentials

1. Create or sign in to your [Last.fm account](https://www.last.fm)
2. Visit the [Last.fm API page](https://www.last.fm/api/account/create)
3. Fill in the application details to create an API account
4. After submission, you'll receive your API key and secret
5. Store these credentials in your environment variables or `.env` file

### Spotify API credentials

1. Create or sign in to your [Spotify account](https://www.spotify.com)
2. Visit the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/)
3. Click "Create App" and fill in the required information
4. Set a redirect URI (you can use `http://localhost:8888/callback`)
5. Once created, you'll see your Client ID on the app page
6. Click "Show Client Secret" to reveal your Client Secret
7. Store both the Client ID and Secret in your environment variables or `.env` file

## Installation

### Using Pre-built Docker Image (Easiest)

```bash
# Pull the latest image from GitHub Container Registry
docker pull ghcr.io/ardakilic/rocklist:latest
# OR from Docker Hub
docker pull ardakilic/rocklist:latest

# Run with environment variables (using GitHub Container Registry)
docker run --rm -v /path/to/rockbox:/rockbox \
  -e LASTFM_API_KEY="your_api_key" \
  -e LASTFM_API_SECRET="your_api_secret" \
  -e SPOTIFY_CLIENT_ID="your_client_id" \
  -e SPOTIFY_CLIENT_SECRET="your_client_secret" \
  ghcr.io/ardakilic/rocklist:latest

# OR run with Docker Hub image
docker run --rm -v /path/to/rockbox:/rockbox \
  -e LASTFM_API_KEY="your_api_key" \
  -e LASTFM_API_SECRET="your_api_secret" \
  -e SPOTIFY_CLIENT_ID="your_client_id" \
  -e SPOTIFY_CLIENT_SECRET="your_client_secret" \
  ardakilic/rocklist:latest
```

### Building Docker Image Locally

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
git clone https://github.com/ardakilic/rocklist.git
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

- `.github/` - GitHub specific configurations
  - `workflows/` - CI/CD workflow definitions
    - `docker-publish.yml` - Workflow for building and publishing Docker images to GHCR and Docker Hub
    - `ci.yaml` - Workflow for continuous integration testing
- `cmd/` - Application entry points
  - `main.go` - Entry point for the CLI application
- `internal/` - Private application code
  - `api/` - API clients and models
    - `lastfm.go` - Last.fm API client implementation
    - `spotify.go` - Spotify API client implementation
    - `api_test.go` - Tests for API clients
    - `models/` - API data models
      - `models.go` - Shared data structures for API responses
  - `config/` - Application configuration
    - `config.go` - Configuration handling using environment variables and CLI flags
    - `config_test.go` - Tests for configuration
  - `database/` - Database handling
    - `rockbox.go` - Rockbox database parser implementation
    - `rockbox_test.go` - Tests for Rockbox database parser
  - `playlist/` - Playlist generation
    - `generator.go` - Playlist creation logic
    - `generator_test.go` - Tests for playlist generation
- `pkg/` - Public libraries
  - `util/` - Utility functions
    - `fileutil.go` - File handling utilities
    - `fileutil_test.go` - Tests for file utilities
- `Dockerfile` - Container definition for the application
- `.editorconfig` - Editor configuration for consistent code formatting
- `.env.example` - Example environment configuration
- `.gitignore` - Specifies intentionally untracked files to ignore
- `go.mod` - Go module definition and dependencies
- `go.sum` - Checksums of the expected content of Go module dependencies
- `LICENSE` - MIT License file

## License

Copyright (c) 2025 Arda Kılıçdağı

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 

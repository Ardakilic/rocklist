# Rocklist

[![CI](https://github.com/Ardakilic/rocklist/actions/workflows/ci.yml/badge.svg)](https://github.com/Ardakilic/rocklist/actions/workflows/ci.yml)
[![Release](https://github.com/Ardakilic/rocklist/actions/workflows/release.yml/badge.svg)](https://github.com/Ardakilic/rocklist/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Rocklist** is a powerful playlist generator for [Rockbox](https://www.rockbox.org/) firmware devices. It parses your Rockbox database and creates smart playlists using external music services like Last.fm, Spotify, and MusicBrainz.

![Rocklist Screenshot](docs/screenshot.png)

## 🎵 Features

- **Parse Rockbox Database** - Reads TagCache files from your Rockbox device
- **Multiple Data Sources** - Use Last.fm, Spotify, or MusicBrainz for playlist generation
- **Smart Playlists** - Generate playlists based on:
  - **Top Songs** - Most popular songs by an artist
  - **Mixed Songs** - A blend of top tracks and similar songs
  - **Similar Artists** - Discover songs from artists similar to your favorites
  - **Tag/Genre Radio** - Create genre-based playlists (e.g., "Death Metal Radio")
- **Offline Ready** - All matched songs come from your local library
- **GUI & CLI** - Use the beautiful desktop app or automate with command-line
- **Cross-Platform** - Works on Windows, macOS, and Linux

## 📥 Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/Ardakilic/rocklist/releases).

| Platform | Download |
|----------|----------|
| Windows | `rocklist-windows-amd64.zip` |
| macOS (Universal) | `rocklist-macos-universal.dmg` |
| Linux | `rocklist-linux-amd64.tar.gz` |

### macOS Installation

1. Download `rocklist-macos-universal.dmg`
2. Open the DMG and drag Rocklist to your Applications folder
3. On first launch, right-click and select "Open" to bypass Gatekeeper (the app is code-signed and notarized)

### Build from Source

#### Prerequisites

- Go 1.23 or later
- Node.js 18 or later
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

**Linux dependencies:**
```bash
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev
```

**Build:**
```bash
# Clone the repository
git clone https://github.com/Ardakilic/rocklist.git
cd rocklist

# Install dependencies
make install

# Build for your platform
make build

# Or build for all platforms
make build-all
```

### Using Docker (Recommended - No Tools Required)

All commands run via Docker - **no local Go or Node.js installation required**:

```bash
# Setup cache directories and install all dependencies
make setup
make install

# Run tests
make test

# Build for Linux
make build
```

## 🚀 Quick Start

### GUI Mode

1. Launch Rocklist
2. Go to **Settings** tab and configure your API credentials
3. Go to **Fetch** tab, enter your Rockbox device path, and click **Parse Now!**
4. Go to **Generate** tab, select a data source and playlist type, then generate!

### CLI Mode

```bash
# Parse Rockbox database
rocklist parse --rockbox-path /Volumes/IPOD

# Generate a playlist
rocklist generate \
  --source lastfm \
  --type top_songs \
  --artist "Metallica" \
  --lastfm-api-key YOUR_API_KEY

# Generate a tag-based playlist
rocklist generate \
  --source spotify \
  --type tag \
  --tag "death metal" \
  --spotify-client-id YOUR_CLIENT_ID \
  --spotify-client-secret YOUR_CLIENT_SECRET
```

## ⚙️ Configuration

### API Credentials

Rocklist needs API credentials to fetch music data. You can configure them in the Settings tab or via command line/config file.

#### Last.fm
1. Go to [Last.fm API](https://www.last.fm/api/account/create)
2. Create an application and get your API Key and Secret

#### Spotify
1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create an application and get your Client ID and Client Secret

#### MusicBrainz
MusicBrainz only requires a descriptive User Agent string (e.g., `Rocklist/1.0.0 (https://github.com/Ardakilic/rocklist)`)

### Config File

You can also use a config file at `~/.rocklist/config.yaml`:

```yaml
rockbox_path: /Volumes/IPOD
lastfm_api_key: your_api_key
lastfm_api_secret: your_api_secret
spotify_client_id: your_client_id
spotify_client_secret: your_client_secret
musicbrainz_user_agent: "Rocklist/1.0.0 (contact@example.com)"
```

## 🔧 Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Setup

```bash
# Clone the repository
git clone https://github.com/Ardakilic/rocklist.git
cd rocklist

# Install dependencies
make install

# Run in development mode
make dev
```

### Available Make Commands

All commands run via Docker - no local Go/Node.js required:

```bash
make help              # Show all available commands
make setup             # Create cache directories
make install           # Install all dependencies (Go + npm)
make dev               # Run in development mode
make build             # Build for Linux
make build-windows     # Build for Windows
make build-darwin      # Build for macOS
make build-all         # Build for all platforms
make test              # Run tests
make test-coverage     # Run tests with coverage report
make lint              # Run linters
make clean             # Clean build artifacts
make clean-all         # Clean everything including caches
make shell             # Open bash shell in Docker
make docker-build      # Build Docker image
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage (requires 90%+)
make test-coverage
```

## 📁 Project Structure

```
rocklist/
├── cmd/                    # CLI commands
├── frontend/               # React frontend (Vite + TypeScript)
│   ├── src/
│   │   ├── components/     # React components
│   │   └── lib/            # Utilities
│   └── package.json
├── internal/
│   ├── api/                # External API clients
│   ├── database/           # SQLite database
│   ├── models/             # Domain models
│   ├── repository/         # Data access layer
│   ├── rockbox/            # Rockbox parser
│   └── service/            # Business logic
├── .github/workflows/      # GitHub Actions
├── Dockerfile              # Docker build
├── docker-compose.yml      # Docker Compose config
├── Makefile                # Build commands
├── wails.json              # Wails configuration
└── go.mod                  # Go modules
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👤 Author

**Arda Kılıçdağı**

- GitHub: [@Ardakilic](https://github.com/Ardakilic)
- Website: [https://arda.pw](https://arda.pw)

## 🙏 Acknowledgments

- [Rockbox](https://www.rockbox.org/) - The amazing open-source firmware
- [Wails](https://wails.io/) - The Go framework for desktop apps
- [Last.fm](https://www.last.fm/api) - Music data API
- [Spotify](https://developer.spotify.com/) - Music data API
- [MusicBrainz](https://musicbrainz.org/) - Open music encyclopedia

---

Made with ❤️ for Rockbox users
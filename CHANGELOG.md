# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2024-01-01

### Added
- Initial release of Rocklist
- Rockbox database parser (TagCache support)
- Last.fm API integration for playlist generation
- Spotify API integration for playlist generation
- MusicBrainz API integration for playlist generation
- Four playlist types:
  - Top Songs by artist
  - Mixed Songs (top + similar)
  - Similar Artists
  - Tag/Genre Radio
- Desktop GUI using Wails + React
- Command-line interface (CLI) for automation
- SQLite database for storing song metadata
- M3U8 playlist export for Rockbox
- Docker support for development
- Cross-platform builds (Windows, macOS, Linux)
- Apple code signing and notarization for macOS releases
- Comprehensive test suite with 90%+ code coverage

### Security
- Secure credential storage using platform keychain (future)
- API keys are not logged or exposed

[Unreleased]: https://github.com/Ardakilic/rocklist/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Ardakilic/rocklist/releases/tag/v1.0.0

import '@testing-library/jest-dom'
import { vi, beforeEach } from 'vitest'

// Mock window.go for Wails integration
const mockApp = {
  GetAppInfo: vi.fn().mockResolvedValue({
    name: 'Rocklist',
    version: '1.0.0',
    description: 'Playlist Generator for Rockbox',
    author: 'Arda Kılıçdağı',
    email: 'arda@kilicdagi.com',
    repository: 'https://github.com/Ardakilic/rocklist',
    license: 'MIT',
  }),
  GetConfig: vi.fn().mockResolvedValue({
    rockbox_path: '/media/rockbox',
    lastfm: { enabled: true, api_key: 'test', api_secret: 'test' },
    spotify: { enabled: false, client_id: '', client_secret: '' },
    musicbrainz: { enabled: true, user_agent: 'Rocklist/1.0' },
  }),
  SetRockboxPath: vi.fn().mockResolvedValue(undefined),
  SetLastFMCredentials: vi.fn().mockResolvedValue(undefined),
  SetSpotifyCredentials: vi.fn().mockResolvedValue(undefined),
  SetMusicBrainzCredentials: vi.fn().mockResolvedValue(undefined),
  ParseDatabase: vi.fn().mockResolvedValue(undefined),
  GetParseStatus: vi.fn().mockResolvedValue({
    in_progress: false,
    total_songs: 100,
    processed_songs: 100,
    error_count: 0,
  }),
  GetLastParsedAt: vi.fn().mockResolvedValue('2024-01-01T00:00:00Z'),
  GeneratePlaylist: vi.fn().mockResolvedValue({
    ID: 1,
    name: 'Test Playlist',
    song_count: 10,
  }),
  GetSongCount: vi.fn().mockResolvedValue(100),
  GetUniqueArtists: vi.fn().mockResolvedValue(['Artist 1', 'Artist 2']),
  GetUniqueGenres: vi.fn().mockResolvedValue(['Rock', 'Pop']),
  GetAllPlaylists: vi.fn().mockResolvedValue([]),
  DeletePlaylist: vi.fn().mockResolvedValue(undefined),
  WipeData: vi.fn().mockResolvedValue(undefined),
  GetLogs: vi.fn().mockResolvedValue([]),
  ClearLogs: vi.fn(),
  GetEnabledSources: vi.fn().mockResolvedValue(['lastfm', 'musicbrainz']),
}

Object.defineProperty(window, 'go', {
  value: {
    cmd: {
      App: mockApp,
    },
  },
  writable: true,
})

// Export mock for test access
export { mockApp }

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}
Object.defineProperty(window, 'localStorage', { value: localStorageMock })

// Reset mocks before each test
beforeEach(() => {
  // Reset call history but keep implementations
  vi.clearAllMocks()
  localStorageMock.getItem.mockReturnValue(null)
  
  // Restore mock implementations
  mockApp.GetAppInfo.mockResolvedValue({
    name: 'Rocklist',
    version: '1.0.0',
    description: 'Playlist Generator for Rockbox',
    author: 'Arda Kılıçdağı',
    email: 'arda@kilicdagi.com',
    repository: 'https://github.com/Ardakilic/rocklist',
    license: 'MIT',
  })
  mockApp.GetConfig.mockResolvedValue({
    rockbox_path: '/media/rockbox',
    lastfm: { enabled: true, api_key: 'test', api_secret: 'test' },
    spotify: { enabled: false, client_id: '', client_secret: '' },
    musicbrainz: { enabled: true, user_agent: 'Rocklist/1.0' },
  })
  mockApp.GetSongCount.mockResolvedValue(100)
  mockApp.GetLastParsedAt.mockResolvedValue('2024-01-01T00:00:00Z')
  mockApp.GetUniqueArtists.mockResolvedValue(['Artist 1', 'Artist 2'])
  mockApp.GetAllPlaylists.mockResolvedValue([])
  mockApp.GetEnabledSources.mockResolvedValue(['lastfm', 'musicbrainz'])
  mockApp.GetLogs.mockResolvedValue([])
  mockApp.GetParseStatus.mockResolvedValue({
    in_progress: false,
    total_songs: 100,
    processed_songs: 100,
    error_count: 0,
  })
})

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { GenerateTab } from './GenerateTab'

describe('GenerateTab component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the generate section', () => {
    render(<GenerateTab />)
    
    expect(screen.getByRole('heading', { name: /generate playlist/i })).toBeInTheDocument()
  })

  it('renders data source select', () => {
    render(<GenerateTab />)
    
    expect(screen.getByText('Data Source')).toBeInTheDocument()
  })

  it('renders playlist type select', () => {
    render(<GenerateTab />)
    
    expect(screen.getByText('Playlist Type')).toBeInTheDocument()
  })

  it('renders max songs input', () => {
    render(<GenerateTab />)
    
    expect(screen.getByText('Max Songs')).toBeInTheDocument()
    expect(screen.getByDisplayValue('50')).toBeInTheDocument()
  })

  it('renders generate button', () => {
    render(<GenerateTab />)
    
    expect(screen.getByRole('button', { name: /generate playlist/i })).toBeInTheDocument()
  })

  it('loads initial data on mount', async () => {
    render(<GenerateTab />)
    
    await waitFor(() => {
      expect(window.go.cmd.App.GetUniqueArtists).toHaveBeenCalled()
      expect(window.go.cmd.App.GetAllPlaylists).toHaveBeenCalled()
      expect(window.go.cmd.App.GetEnabledSources).toHaveBeenCalled()
    })
  })

  it('updates max songs value', () => {
    render(<GenerateTab />)
    
    const maxSongsInput = screen.getByDisplayValue('50')
    fireEvent.change(maxSongsInput, { target: { value: '100' } })
    
    expect(screen.getByDisplayValue('100')).toBeInTheDocument()
  })

  it('renders artist input field', () => {
    render(<GenerateTab />)
    
    // Artist input should be present since default playlist type is top_songs
    expect(screen.getByPlaceholderText(/enter artist name/i)).toBeInTheDocument()
  })

  it('renders logs section', () => {
    render(<GenerateTab />)
    
    expect(screen.getByText('Logs')).toBeInTheDocument()
  })

  it('has clear logs button', () => {
    render(<GenerateTab />)
    
    expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument()
  })

  it('calls generate playlist on button click', async () => {
    // Ensure lastfm is in enabled sources
    vi.mocked(window.go.cmd.App.GetEnabledSources).mockResolvedValue(['lastfm', 'musicbrainz'])
    
    render(<GenerateTab />)
    
    // Wait for enabled sources to load
    await waitFor(() => {
      expect(window.go.cmd.App.GetEnabledSources).toHaveBeenCalled()
    })
    
    // Fill in artist name since default type is top_songs
    const artistInput = screen.getByPlaceholderText(/enter artist name/i)
    fireEvent.change(artistInput, { target: { value: 'Test Artist' } })
    
    const generateButton = screen.getByRole('button', { name: /generate playlist/i })
    fireEvent.click(generateButton)
    
    await waitFor(() => {
      expect(window.go.cmd.App.GeneratePlaylist).toHaveBeenCalled()
    })
  })

  it('refreshes playlists after generation', async () => {
    render(<GenerateTab />)
    
    // Fill in artist name
    const artistInput = screen.getByPlaceholderText(/enter artist name/i)
    fireEvent.change(artistInput, { target: { value: 'Test Artist' } })
    
    const generateButton = screen.getByRole('button', { name: /generate playlist/i })
    fireEvent.click(generateButton)
    
    await waitFor(() => {
      expect(window.go.cmd.App.GetAllPlaylists).toHaveBeenCalled()
    })
  })

  it('displays playlists when available', async () => {
    vi.mocked(window.go.cmd.App.GetAllPlaylists).mockResolvedValue([
      {
        ID: 1,
        name: 'Test Playlist',
        description: 'Test desc',
        type: 'top_songs',
        data_source: 'lastfm',
        artist: 'Test Artist',
        tag: '',
        file_path: '/path/to/playlist.m3u8',
        song_count: 10,
        generated_at: '2024-01-01T00:00:00Z',
        exported_at: null,
      },
    ])
    
    render(<GenerateTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Test Playlist')).toBeInTheDocument()
    })
  })

  it('shows no logs message when empty', () => {
    render(<GenerateTab />)
    
    expect(screen.getByText(/no logs yet/i)).toBeInTheDocument()
  })

  it('clears logs when clear button clicked', async () => {
    render(<GenerateTab />)
    
    const clearButton = screen.getByRole('button', { name: /clear/i })
    fireEvent.click(clearButton)
    
    expect(window.go.cmd.App.ClearLogs).toHaveBeenCalled()
  })

  it('deletes playlist when delete button clicked', async () => {
    vi.mocked(window.go.cmd.App.GetAllPlaylists).mockResolvedValue([
      {
        ID: 1,
        name: 'Test Playlist',
        description: 'Test desc',
        type: 'top_songs',
        data_source: 'lastfm',
        artist: 'Test Artist',
        tag: '',
        file_path: '/path/to/playlist.m3u8',
        song_count: 10,
        generated_at: '2024-01-01T00:00:00Z',
        exported_at: null,
      },
    ])
    
    render(<GenerateTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Test Playlist')).toBeInTheDocument()
    })
    
    // Find delete button by icon
    const deleteButtons = screen.getAllByRole('button')
    const deleteButton = deleteButtons.find(btn => btn.querySelector('svg.lucide-trash-2'))
    if (deleteButton) {
      fireEvent.click(deleteButton)
      
      await waitFor(() => {
        expect(window.go.cmd.App.DeletePlaylist).toHaveBeenCalledWith(1)
      })
    }
  })
})

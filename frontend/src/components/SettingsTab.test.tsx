import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { SettingsTab } from './SettingsTab'

describe('SettingsTab component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state initially', () => {
    vi.mocked(window.go.cmd.App.GetConfig).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    )
    
    render(<SettingsTab />)
    
    expect(screen.getByText(/loading settings/i)).toBeInTheDocument()
  })

  it('loads and displays config', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(window.go.cmd.App.GetConfig).toHaveBeenCalled()
    })
    
    await waitFor(() => {
      expect(screen.getByText('Last.fm')).toBeInTheDocument()
    })
  })

  it('renders Last.fm settings section', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Last.fm')).toBeInTheDocument()
      expect(screen.getByPlaceholderText(/last.fm api key/i)).toBeInTheDocument()
    })
  })

  it('renders Spotify settings section', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Spotify')).toBeInTheDocument()
    })
  })

  it('renders MusicBrainz settings section', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('MusicBrainz')).toBeInTheDocument()
    })
  })

  it('renders save button', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /save settings/i })).toBeInTheDocument()
    })
  })

  it('renders danger zone section', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Danger Zone')).toBeInTheDocument()
    })
  })

  it('calls save functions when Save is clicked', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /save settings/i })).toBeInTheDocument()
    })
    
    const saveButton = screen.getByRole('button', { name: /save settings/i })
    fireEvent.click(saveButton)
    
    await waitFor(() => {
      expect(window.go.cmd.App.SetLastFMCredentials).toHaveBeenCalled()
      expect(window.go.cmd.App.SetSpotifyCredentials).toHaveBeenCalled()
      expect(window.go.cmd.App.SetMusicBrainzCredentials).toHaveBeenCalled()
    })
  })

  it('renders enabled checkboxes', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Last.fm')).toBeInTheDocument()
    })
    
    // All service checkboxes should be rendered
    const checkboxes = screen.getAllByRole('checkbox')
    expect(checkboxes.length).toBeGreaterThanOrEqual(3)
  })

  it('shows wipe data confirmation dialog', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Danger Zone')).toBeInTheDocument()
    })
    
    const wipeButton = screen.getByRole('button', { name: /wipe pre-fetched data/i })
    fireEvent.click(wipeButton)
    
    await waitFor(() => {
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
    })
  })

  it('can cancel wipe data', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Danger Zone')).toBeInTheDocument()
    })
    
    const wipeButton = screen.getByRole('button', { name: /wipe pre-fetched data/i })
    fireEvent.click(wipeButton)
    
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
    })
    
    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    fireEvent.click(cancelButton)
    
    await waitFor(() => {
      expect(screen.queryByText(/are you sure/i)).not.toBeInTheDocument()
    })
  })

  it('calls WipeData when confirmed', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Danger Zone')).toBeInTheDocument()
    })
    
    const wipeButton = screen.getByRole('button', { name: /wipe pre-fetched data/i })
    fireEvent.click(wipeButton)
    
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /yes, wipe all data/i })).toBeInTheDocument()
    })
    
    const confirmButton = screen.getByRole('button', { name: /yes, wipe all data/i })
    fireEvent.click(confirmButton)
    
    await waitFor(() => {
      expect(window.go.cmd.App.WipeData).toHaveBeenCalled()
    })
  })

  it('updates API key input', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/last.fm api key/i)).toBeInTheDocument()
    })
    
    const apiKeyInput = screen.getByPlaceholderText(/last.fm api key/i)
    fireEvent.change(apiKeyInput, { target: { value: 'new-api-key' } })
    
    expect(apiKeyInput).toHaveValue('new-api-key')
  })

  it('displays service sections', async () => {
    render(<SettingsTab />)
    
    await waitFor(() => {
      expect(screen.getByText('Last.fm')).toBeInTheDocument()
      expect(screen.getByText('Spotify')).toBeInTheDocument()
      expect(screen.getByText('MusicBrainz')).toBeInTheDocument()
    })
  })
})

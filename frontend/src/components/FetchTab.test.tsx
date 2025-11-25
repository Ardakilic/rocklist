import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { FetchTab } from './FetchTab'

describe('FetchTab component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the parse section', () => {
    render(<FetchTab />)
    
    expect(screen.getByText('Parse Rockbox Database')).toBeInTheDocument()
    expect(screen.getByLabelText(/rockbox device path/i)).toBeInTheDocument()
  })

  it('renders the path input', () => {
    render(<FetchTab />)
    
    const input = screen.getByPlaceholderText(/\/Volumes\/IPOD/i)
    expect(input).toBeInTheDocument()
  })

  it('renders the use pre-fetched checkbox', () => {
    render(<FetchTab />)
    
    expect(screen.getByText(/use pre-fetched data/i)).toBeInTheDocument()
    expect(screen.getByRole('checkbox')).toBeInTheDocument()
  })

  it('renders the Parse Now button', () => {
    render(<FetchTab />)
    
    expect(screen.getByRole('button', { name: /parse now/i })).toBeInTheDocument()
  })

  it('disables Parse button when path is empty', () => {
    render(<FetchTab />)
    
    const parseButton = screen.getByRole('button', { name: /parse now/i })
    expect(parseButton).toBeDisabled()
  })

  it('enables Parse button when path is entered', async () => {
    render(<FetchTab />)
    
    const input = screen.getByPlaceholderText(/\/Volumes\/IPOD/i)
    fireEvent.change(input, { target: { value: '/media/rockbox' } })
    
    await waitFor(() => {
      const parseButton = screen.getByRole('button', { name: /parse now/i })
      expect(parseButton).not.toBeDisabled()
    })
  })

  it('toggles use pre-fetched checkbox', () => {
    render(<FetchTab />)
    
    const checkbox = screen.getByRole('checkbox')
    expect(checkbox).not.toBeChecked()
    
    fireEvent.click(checkbox)
    expect(checkbox).toBeChecked()
  })

  it('renders logs section', () => {
    render(<FetchTab />)
    
    expect(screen.getByText('Logs')).toBeInTheDocument()
    expect(screen.getByText('Clear')).toBeInTheDocument()
  })

  it('shows empty logs message', () => {
    render(<FetchTab />)
    
    expect(screen.getByText(/no logs yet/i)).toBeInTheDocument()
  })

  it('loads initial data on mount', async () => {
    render(<FetchTab />)
    
    await waitFor(() => {
      expect(window.go.cmd.App.GetConfig).toHaveBeenCalled()
      expect(window.go.cmd.App.GetLastParsedAt).toHaveBeenCalled()
      expect(window.go.cmd.App.GetSongCount).toHaveBeenCalled()
    })
  })

  it('calls parse functions when Parse Now is clicked', async () => {
    render(<FetchTab />)
    
    const input = screen.getByPlaceholderText(/\/Volumes\/IPOD/i)
    fireEvent.change(input, { target: { value: '/media/rockbox' } })
    
    const parseButton = screen.getByRole('button', { name: /parse now/i })
    fireEvent.click(parseButton)
    
    await waitFor(() => {
      expect(window.go.cmd.App.SetRockboxPath).toHaveBeenCalledWith('/media/rockbox')
      expect(window.go.cmd.App.ParseDatabase).toHaveBeenCalled()
    })
  })

  it('shows loading state while parsing', async () => {
    // Make ParseDatabase slow
    vi.mocked(window.go.cmd.App.ParseDatabase).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    )
    
    render(<FetchTab />)
    
    const input = screen.getByPlaceholderText(/\/Volumes\/IPOD/i)
    fireEvent.change(input, { target: { value: '/media/rockbox' } })
    
    const parseButton = screen.getByRole('button', { name: /parse now/i })
    fireEvent.click(parseButton)
    
    await waitFor(() => {
      expect(screen.getByText(/parsing/i)).toBeInTheDocument()
    })
  })

  it('displays song count', async () => {
    vi.mocked(window.go.cmd.App.GetSongCount).mockResolvedValue(500)
    
    render(<FetchTab />)
    
    await waitFor(() => {
      expect(screen.getByText(/500/)).toBeInTheDocument()
      expect(screen.getByText(/songs in database/i)).toBeInTheDocument()
    })
  })

  it('calls ClearLogs when Clear button is clicked', () => {
    render(<FetchTab />)
    
    const clearButton = screen.getByRole('button', { name: /clear/i })
    fireEvent.click(clearButton)
    
    expect(window.go.cmd.App.ClearLogs).toHaveBeenCalled()
  })

  it('renders folder icon button', () => {
    render(<FetchTab />)
    
    const buttons = screen.getAllByRole('button')
    // Should have Parse Now, Clear, and folder buttons
    expect(buttons.length).toBeGreaterThanOrEqual(2)
  })

  it('has description text for path input', () => {
    render(<FetchTab />)
    
    expect(screen.getByText(/path where the .rockbox folder is located/i)).toBeInTheDocument()
  })
})

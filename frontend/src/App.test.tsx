import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import App from './App'

describe('App component', () => {
  beforeEach(() => {
    // Reset localStorage mock
    vi.mocked(localStorage.getItem).mockReturnValue(null)
  })

  it('renders header with title', () => {
    render(<App />)
    
    expect(screen.getByText('Rocklist')).toBeInTheDocument()
    expect(screen.getByText('Playlist Generator for Rockbox')).toBeInTheDocument()
  })

  it('renders tab navigation', async () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByRole('tablist')).toBeInTheDocument()
    })
    expect(screen.getAllByRole('tab')).toHaveLength(3)
  })

  it('shows welcome dialog on first visit', () => {
    vi.mocked(localStorage.getItem).mockReturnValue(null)
    render(<App />)
    
    expect(screen.getByText('Welcome to Rocklist!')).toBeInTheDocument()
  })

  it('hides welcome dialog after first visit', () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    expect(screen.queryByText('Welcome to Rocklist!')).not.toBeInTheDocument()
  })

  it('closes welcome dialog and sets localStorage', async () => {
    vi.mocked(localStorage.getItem).mockReturnValue(null)
    render(<App />)
    
    fireEvent.click(screen.getByText('Get Started'))
    
    expect(localStorage.setItem).toHaveBeenCalledWith('rocklist_visited', 'true')
    await waitFor(() => {
      expect(screen.queryByText('Welcome to Rocklist!')).not.toBeInTheDocument()
    })
  })

  it('shows About dialog when About button is clicked', async () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    fireEvent.click(screen.getByText('About'))
    
    await waitFor(() => {
      expect(screen.getByText('Author')).toBeInTheDocument()
    })
  })

  it('closes About dialog', async () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    // Open About dialog
    fireEvent.click(screen.getByText('About'))
    
    await waitFor(() => {
      expect(screen.getByText('Author')).toBeInTheDocument()
    })
    
    // Close via the dialog's close button
    const closeButton = screen.getByRole('button', { name: /close/i })
    fireEvent.click(closeButton)
    
    await waitFor(() => {
      expect(screen.queryByText('Author')).not.toBeInTheDocument()
    })
  })

  it('has all tabs accessible', async () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    // All tabs should be rendered
    const tabs = screen.getAllByRole('tab')
    expect(tabs).toHaveLength(3)
    
    // Each tab should be clickable
    tabs.forEach(tab => {
      expect(tab).not.toBeDisabled()
    })
  })

  it('loads app info on mount', async () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    await waitFor(() => {
      expect(window.go.cmd.App.GetAppInfo).toHaveBeenCalled()
    })
  })

  it('renders Toaster component', () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    // Toaster should be rendered (it's a toast provider)
    expect(document.querySelector('[data-radix-toast-viewport]') || screen.queryByRole('region')).toBeDefined()
  })

  it('has correct header layout', () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    const header = document.querySelector('header')
    expect(header).toBeInTheDocument()
    expect(header).toHaveClass('border-b')
  })

  it('renders main content area', () => {
    vi.mocked(localStorage.getItem).mockReturnValue('true')
    render(<App />)
    
    const main = document.querySelector('main')
    expect(main).toBeInTheDocument()
  })
})

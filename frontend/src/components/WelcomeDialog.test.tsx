import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { WelcomeDialog } from './WelcomeDialog'

describe('WelcomeDialog component', () => {
  it('renders when open', () => {
    render(<WelcomeDialog open={true} onClose={vi.fn()} />)
    
    expect(screen.getByText('Welcome to Rocklist!')).toBeInTheDocument()
    expect(screen.getByText(/Create amazing playlists/)).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    render(<WelcomeDialog open={false} onClose={vi.fn()} />)
    
    expect(screen.queryByText('Welcome to Rocklist!')).not.toBeInTheDocument()
  })

  it('displays feature descriptions', () => {
    render(<WelcomeDialog open={true} onClose={vi.fn()} />)
    
    expect(screen.getByText('Parse Your Library')).toBeInTheDocument()
    expect(screen.getByText('Generate Smart Playlists')).toBeInTheDocument()
    expect(screen.getByText('Enjoy Offline')).toBeInTheDocument()
  })

  it('calls onClose when Get Started is clicked', () => {
    const onClose = vi.fn()
    render(<WelcomeDialog open={true} onClose={onClose} />)
    
    fireEvent.click(screen.getByText('Get Started'))
    expect(onClose).toHaveBeenCalled()
  })

  it('displays correct description texts', () => {
    render(<WelcomeDialog open={true} onClose={vi.fn()} />)
    
    expect(screen.getByText(/Connect your Rockbox device/)).toBeInTheDocument()
    expect(screen.getByText(/Create playlists based on top songs/)).toBeInTheDocument()
    expect(screen.getByText(/All matched songs are from your local library/)).toBeInTheDocument()
  })

  it('has a close button', () => {
    render(<WelcomeDialog open={true} onClose={vi.fn()} />)
    
    expect(screen.getByRole('button', { name: /get started/i })).toBeInTheDocument()
  })
})

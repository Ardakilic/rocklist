import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { AboutDialog } from './AboutDialog'

describe('AboutDialog component', () => {
  const defaultAppInfo = {
    name: 'Rocklist',
    version: '1.0.0',
    description: 'A tool for creating playlists for Rockbox firmware devices',
    author: 'Arda Kılıçdağı',
    email: 'arda@kilicdagi.com',
    repository: 'https://github.com/Ardakilic/rocklist',
    license: 'MIT',
  }

  it('renders when open', () => {
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={defaultAppInfo} />)
    
    expect(screen.getByText('Rocklist')).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    render(<AboutDialog open={false} onClose={vi.fn()} appInfo={defaultAppInfo} />)
    
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('displays app name and version', () => {
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={defaultAppInfo} />)
    
    expect(screen.getByText('Rocklist')).toBeInTheDocument()
    expect(screen.getByText('Version 1.0.0')).toBeInTheDocument()
  })

  it('displays description', () => {
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={defaultAppInfo} />)
    
    expect(screen.getByText(defaultAppInfo.description)).toBeInTheDocument()
  })

  it('displays author information', () => {
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={defaultAppInfo} />)
    
    expect(screen.getByText('Author')).toBeInTheDocument()
    expect(screen.getByText('Arda Kılıçdağı')).toBeInTheDocument()
  })

  it('displays email link', () => {
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={defaultAppInfo} />)
    
    const emailLink = screen.getByRole('link', { name: 'arda@kilicdagi.com' })
    expect(emailLink).toBeInTheDocument()
    expect(emailLink).toHaveAttribute('href', 'mailto:arda@kilicdagi.com')
  })

  it('displays repository link', () => {
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={defaultAppInfo} />)
    
    const repoLink = screen.getByRole('link', { name: defaultAppInfo.repository })
    expect(repoLink).toBeInTheDocument()
    expect(repoLink).toHaveAttribute('href', defaultAppInfo.repository)
    expect(repoLink).toHaveAttribute('target', '_blank')
    expect(repoLink).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('displays license', () => {
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={defaultAppInfo} />)
    
    expect(screen.getByText(/Licensed under MIT/)).toBeInTheDocument()
  })

  it('uses default values when appInfo is empty', () => {
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={{}} />)
    
    expect(screen.getByText('Rocklist')).toBeInTheDocument()
    expect(screen.getByText('Version 1.0.0')).toBeInTheDocument()
    expect(screen.getByText('Arda Kılıçdağı')).toBeInTheDocument()
    expect(screen.getByText(/Licensed under MIT/)).toBeInTheDocument()
  })

  it('uses custom appInfo values', () => {
    const customInfo = {
      name: 'Custom App',
      version: '2.0.0',
      description: 'Custom description',
      author: 'Custom Author',
      email: 'custom@example.com',
      repository: 'https://github.com/custom/repo',
      license: 'Apache 2.0',
    }
    
    render(<AboutDialog open={true} onClose={vi.fn()} appInfo={customInfo} />)
    
    expect(screen.getByText('Custom App')).toBeInTheDocument()
    expect(screen.getByText('Version 2.0.0')).toBeInTheDocument()
    expect(screen.getByText('Custom Author')).toBeInTheDocument()
    expect(screen.getByText(/Licensed under Apache 2.0/)).toBeInTheDocument()
  })
})

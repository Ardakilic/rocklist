import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from './dialog'

describe('Dialog component', () => {
  it('renders when open', () => {
    render(
      <Dialog open>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Test Dialog</DialogTitle>
            <DialogDescription>Test description</DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    )
    
    expect(screen.getByText('Test Dialog')).toBeInTheDocument()
    expect(screen.getByText('Test description')).toBeInTheDocument()
  })

  it('does not render when closed', () => {
    render(
      <Dialog open={false}>
        <DialogContent>
          <DialogTitle>Hidden Dialog</DialogTitle>
        </DialogContent>
      </Dialog>
    )
    
    expect(screen.queryByText('Hidden Dialog')).not.toBeInTheDocument()
  })

  it('calls onOpenChange when closed', () => {
    const onOpenChange = vi.fn()
    render(
      <Dialog open onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogTitle>Closeable Dialog</DialogTitle>
          <DialogClose data-testid="close-btn">Close Me</DialogClose>
        </DialogContent>
      </Dialog>
    )
    
    // Click the close button
    fireEvent.click(screen.getByTestId('close-btn'))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it('renders DialogFooter', () => {
    render(
      <Dialog open>
        <DialogContent>
          <DialogTitle>Dialog with Footer</DialogTitle>
          <DialogFooter>
            <button>Cancel</button>
            <button>Confirm</button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )
    
    expect(screen.getByText('Cancel')).toBeInTheDocument()
    expect(screen.getByText('Confirm')).toBeInTheDocument()
  })

  it('applies custom className to DialogContent', () => {
    render(
      <Dialog open>
        <DialogContent className="custom-content">
          <DialogTitle>Custom Dialog</DialogTitle>
        </DialogContent>
      </Dialog>
    )
    
    const content = screen.getByRole('dialog')
    expect(content).toHaveClass('custom-content')
  })

  it('applies custom className to DialogHeader', () => {
    render(
      <Dialog open>
        <DialogContent>
          <DialogHeader className="custom-header">
            <DialogTitle>Header Test</DialogTitle>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    )
    
    expect(screen.getByText('Header Test').parentElement).toHaveClass('custom-header')
  })

  it('applies custom className to DialogTitle', () => {
    render(
      <Dialog open>
        <DialogContent>
          <DialogTitle className="custom-title">Title Test</DialogTitle>
        </DialogContent>
      </Dialog>
    )
    
    expect(screen.getByText('Title Test')).toHaveClass('custom-title')
  })

  it('applies custom className to DialogDescription', () => {
    render(
      <Dialog open>
        <DialogContent>
          <DialogTitle>Title</DialogTitle>
          <DialogDescription className="custom-desc">Description</DialogDescription>
        </DialogContent>
      </Dialog>
    )
    
    expect(screen.getByText('Description')).toHaveClass('custom-desc')
  })

  it('applies custom className to DialogFooter', () => {
    render(
      <Dialog open>
        <DialogContent>
          <DialogTitle>Title</DialogTitle>
          <DialogFooter className="custom-footer">
            <button>Button</button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )
    
    expect(screen.getByText('Button').parentElement).toHaveClass('custom-footer')
  })
})

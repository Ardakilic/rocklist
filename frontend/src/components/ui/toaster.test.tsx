import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import {
  Toast,
  ToastProvider,
  ToastViewport,
  ToastTitle,
  ToastDescription,
  ToastClose,
  Toaster,
} from './toaster'

describe('Toast components', () => {
  it('renders ToastProvider', () => {
    render(
      <ToastProvider>
        <div data-testid="child">Child content</div>
        <ToastViewport />
      </ToastProvider>
    )
    
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })

  it('renders ToastViewport', () => {
    render(
      <ToastProvider>
        <ToastViewport data-testid="viewport" />
      </ToastProvider>
    )
    
    expect(screen.getByTestId('viewport')).toBeInTheDocument()
  })

  it('renders Toast with title', () => {
    render(
      <ToastProvider>
        <Toast>
          <ToastTitle>Test Title</ToastTitle>
        </Toast>
        <ToastViewport />
      </ToastProvider>
    )
    
    expect(screen.getByText('Test Title')).toBeInTheDocument()
  })

  it('renders Toast with description', () => {
    render(
      <ToastProvider>
        <Toast>
          <ToastDescription>Test Description</ToastDescription>
        </Toast>
        <ToastViewport />
      </ToastProvider>
    )
    
    expect(screen.getByText('Test Description')).toBeInTheDocument()
  })

  it('renders Toast with close button', () => {
    render(
      <ToastProvider>
        <Toast>
          <ToastTitle>Title</ToastTitle>
          <ToastClose />
        </Toast>
        <ToastViewport />
      </ToastProvider>
    )
    
    expect(screen.getByRole('button')).toBeInTheDocument()
  })

  it('applies custom className to Toast', () => {
    render(
      <ToastProvider>
        <Toast className="custom-toast" data-testid="toast">
          <ToastTitle>Title</ToastTitle>
        </Toast>
        <ToastViewport />
      </ToastProvider>
    )
    
    expect(screen.getByTestId('toast')).toHaveClass('custom-toast')
  })

  it('applies custom className to ToastTitle', () => {
    render(
      <ToastProvider>
        <Toast>
          <ToastTitle className="custom-title">Title</ToastTitle>
        </Toast>
        <ToastViewport />
      </ToastProvider>
    )
    
    expect(screen.getByText('Title')).toHaveClass('custom-title')
  })

  it('applies custom className to ToastDescription', () => {
    render(
      <ToastProvider>
        <Toast>
          <ToastDescription className="custom-desc">Description</ToastDescription>
        </Toast>
        <ToastViewport />
      </ToastProvider>
    )
    
    expect(screen.getByText('Description')).toHaveClass('custom-desc')
  })

  it('renders Toaster component', () => {
    render(<Toaster />)
    
    // Toaster should render without errors
    expect(document.body).toBeInTheDocument()
  })
})

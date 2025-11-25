import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Checkbox } from './checkbox'

describe('Checkbox component', () => {
  it('renders correctly', () => {
    render(<Checkbox aria-label="Test checkbox" />)
    const checkbox = screen.getByRole('checkbox')
    expect(checkbox).toBeInTheDocument()
  })

  it('can be checked', () => {
    const onCheckedChange = vi.fn()
    render(<Checkbox onCheckedChange={onCheckedChange} aria-label="Test" />)
    
    const checkbox = screen.getByRole('checkbox')
    fireEvent.click(checkbox)
    
    expect(onCheckedChange).toHaveBeenCalledWith(true)
  })

  it('can be unchecked', () => {
    const onCheckedChange = vi.fn()
    render(<Checkbox defaultChecked onCheckedChange={onCheckedChange} aria-label="Test" />)
    
    const checkbox = screen.getByRole('checkbox')
    fireEvent.click(checkbox)
    
    expect(onCheckedChange).toHaveBeenCalledWith(false)
  })

  it('can be disabled', () => {
    render(<Checkbox disabled aria-label="Test" />)
    const checkbox = screen.getByRole('checkbox')
    expect(checkbox).toBeDisabled()
  })

  it('applies custom className', () => {
    render(<Checkbox className="custom-class" aria-label="Test" />)
    const checkbox = screen.getByRole('checkbox')
    expect(checkbox).toHaveClass('custom-class')
  })

  it('can be controlled', () => {
    const { rerender } = render(<Checkbox checked={false} aria-label="Test" />)
    expect(screen.getByRole('checkbox')).not.toBeChecked()

    rerender(<Checkbox checked={true} aria-label="Test" />)
    expect(screen.getByRole('checkbox')).toBeChecked()
  })

  it('forwards ref correctly', () => {
    const ref = vi.fn()
    render(<Checkbox ref={ref} aria-label="Test" />)
    expect(ref).toHaveBeenCalled()
  })
})

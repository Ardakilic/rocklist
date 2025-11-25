import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Label } from './label'

describe('Label component', () => {
  it('renders with text content', () => {
    render(<Label>Test Label</Label>)
    expect(screen.getByText('Test Label')).toBeInTheDocument()
  })

  it('renders with htmlFor attribute', () => {
    render(<Label htmlFor="input-id">Label for input</Label>)
    const label = screen.getByText('Label for input')
    expect(label).toHaveAttribute('for', 'input-id')
  })

  it('applies custom className', () => {
    render(<Label className="custom-class">Custom Label</Label>)
    expect(screen.getByText('Custom Label')).toHaveClass('custom-class')
  })

  it('forwards ref correctly', () => {
    const ref = vi.fn()
    render(<Label ref={ref}>Ref Label</Label>)
    expect(ref).toHaveBeenCalled()
  })

  it('has correct default styling', () => {
    render(<Label>Styled Label</Label>)
    const label = screen.getByText('Styled Label')
    expect(label).toHaveClass('text-sm')
    expect(label).toHaveClass('font-medium')
  })

  it('can render children elements', () => {
    render(
      <Label>
        <span data-testid="child">Child Element</span>
      </Label>
    )
    expect(screen.getByTestId('child')).toBeInTheDocument()
  })
})

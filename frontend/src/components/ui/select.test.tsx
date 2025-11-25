import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
  SelectGroup,
} from './select'

describe('Select component', () => {
  const renderSelect = (props = {}) => {
    return render(
      <Select {...props}>
        <SelectTrigger>
          <SelectValue placeholder="Select an option" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="option1">Option 1</SelectItem>
          <SelectItem value="option2">Option 2</SelectItem>
          <SelectItem value="option3">Option 3</SelectItem>
        </SelectContent>
      </Select>
    )
  }

  it('renders trigger with placeholder', () => {
    renderSelect()
    expect(screen.getByText('Select an option')).toBeInTheDocument()
  })

  it('renders trigger button', () => {
    renderSelect()
    expect(screen.getByRole('combobox')).toBeInTheDocument()
  })

  it('opens on click', () => {
    renderSelect()
    fireEvent.click(screen.getByRole('combobox'))
    expect(screen.getByRole('listbox')).toBeInTheDocument()
  })

  it('shows options when opened', () => {
    renderSelect()
    fireEvent.click(screen.getByRole('combobox'))
    
    expect(screen.getByText('Option 1')).toBeInTheDocument()
    expect(screen.getByText('Option 2')).toBeInTheDocument()
    expect(screen.getByText('Option 3')).toBeInTheDocument()
  })

  it('calls onValueChange when option selected', () => {
    const onValueChange = vi.fn()
    renderSelect({ onValueChange })
    
    fireEvent.click(screen.getByRole('combobox'))
    fireEvent.click(screen.getByText('Option 2'))
    
    expect(onValueChange).toHaveBeenCalledWith('option2')
  })

  it('displays selected value', () => {
    render(
      <Select value="option1">
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="option1">Option 1</SelectItem>
          <SelectItem value="option2">Option 2</SelectItem>
        </SelectContent>
      </Select>
    )
    
    expect(screen.getByText('Option 1')).toBeInTheDocument()
  })

  it('applies custom className to trigger', () => {
    render(
      <Select>
        <SelectTrigger className="custom-trigger">
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="option1">Option 1</SelectItem>
        </SelectContent>
      </Select>
    )
    
    expect(screen.getByRole('combobox')).toHaveClass('custom-trigger')
  })

  it('can be disabled', () => {
    render(
      <Select disabled>
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="option1">Option 1</SelectItem>
        </SelectContent>
      </Select>
    )
    
    expect(screen.getByRole('combobox')).toBeDisabled()
  })

  it('renders with SelectGroup', () => {
    render(
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="Select" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectItem value="apple">Apple</SelectItem>
            <SelectItem value="banana">Banana</SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    )
    
    fireEvent.click(screen.getByRole('combobox'))
    expect(screen.getByText('Apple')).toBeInTheDocument()
    expect(screen.getByText('Banana')).toBeInTheDocument()
  })
})

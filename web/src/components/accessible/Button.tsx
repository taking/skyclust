import React from 'react'
import { Button as MantineButton, ButtonProps as MantineButtonProps } from '@mantine/core'

interface AccessibleButtonProps extends MantineButtonProps {
  loading?: boolean
  disabled?: boolean
  'aria-label'?: string
  'aria-describedby'?: string
  'aria-expanded'?: boolean
  'aria-controls'?: string
  'aria-pressed'?: boolean
  'aria-current'?: boolean
  role?: string
  tabIndex?: number
  onFocus?: (event: React.FocusEvent<HTMLButtonElement>) => void
  onBlur?: (event: React.FocusEvent<HTMLButtonElement>) => void
  onKeyDown?: (event: React.KeyboardEvent<HTMLButtonElement>) => void
}

export const AccessibleButton: React.FC<AccessibleButtonProps> = ({
  children,
  loading = false,
  disabled = false,
  'aria-label': ariaLabel,
  'aria-describedby': ariaDescribedBy,
  'aria-expanded': ariaExpanded,
  'aria-controls': ariaControls,
  'aria-pressed': ariaPressed,
  'aria-current': ariaCurrent,
  role,
  tabIndex,
  onFocus,
  onBlur,
  onKeyDown,
  ...props
}) => {
  const handleKeyDown = (event: React.KeyboardEvent<HTMLButtonElement>) => {
    // Handle Enter and Space key presses for accessibility
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      if (!disabled && !loading && props.onClick) {
        props.onClick(event as any)
      }
    }
    
    onKeyDown?.(event)
  }

  return (
    <MantineButton
      {...props}
      loading={loading}
      disabled={disabled || loading}
      aria-label={ariaLabel}
      aria-describedby={ariaDescribedBy}
      aria-expanded={ariaExpanded}
      aria-controls={ariaControls}
      aria-pressed={ariaPressed}
      aria-current={ariaCurrent}
      role={role}
      tabIndex={tabIndex}
      onFocus={onFocus}
      onBlur={onBlur}
      onKeyDown={handleKeyDown}
      // Ensure proper focus management
      style={{
        ...props.style,
        outline: 'none', // We'll handle focus styles with CSS
      }}
    >
      {children}
    </MantineButton>
  )
}

// Focus styles for better accessibility
export const buttonFocusStyles = `
  .mantine-Button:focus-visible {
    outline: 2px solid #228be6;
    outline-offset: 2px;
  }
  
  .mantine-Button:focus:not(:focus-visible) {
    outline: none;
  }
`

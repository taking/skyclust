import React, { useEffect, useRef } from 'react'
import { Modal as MantineModal, ModalProps as MantineModalProps } from '@mantine/core'

interface AccessibleModalProps extends MantineModalProps {
  'aria-labelledby'?: string
  'aria-describedby'?: string
  'aria-modal'?: boolean
  role?: string
  onClose: () => void
}

export const AccessibleModal: React.FC<AccessibleModalProps> = ({
  children,
  opened,
  onClose,
  title,
  'aria-labelledby': ariaLabelledBy,
  'aria-describedby': ariaDescribedBy,
  'aria-modal': ariaModal = true,
  role = 'dialog',
  ...props
}) => {
  const modalRef = useRef<HTMLDivElement>(null)
  const previousActiveElement = useRef<HTMLElement | null>(null)

  // Focus management
  useEffect(() => {
    if (opened) {
      // Store the previously focused element
      previousActiveElement.current = document.activeElement as HTMLElement
      
      // Focus the modal when it opens
      if (modalRef.current) {
        modalRef.current.focus()
      }
    } else {
      // Restore focus to the previously focused element when modal closes
      if (previousActiveElement.current) {
        previousActiveElement.current.focus()
      }
    }
  }, [opened])

  // Handle escape key
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && opened) {
        onClose()
      }
    }

    if (opened) {
      document.addEventListener('keydown', handleKeyDown)
      return () => document.removeEventListener('keydown', handleKeyDown)
    }
  }, [opened, onClose])

  // Trap focus within modal
  useEffect(() => {
    if (!opened) return

    const handleTabKey = (event: KeyboardEvent) => {
      if (event.key !== 'Tab') return

      const modal = modalRef.current
      if (!modal) return

      const focusableElements = modal.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      )
      const firstElement = focusableElements[0] as HTMLElement
      const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement

      if (event.shiftKey) {
        // Shift + Tab
        if (document.activeElement === firstElement) {
          event.preventDefault()
          lastElement?.focus()
        }
      } else {
        // Tab
        if (document.activeElement === lastElement) {
          event.preventDefault()
          firstElement?.focus()
        }
      }
    }

    document.addEventListener('keydown', handleTabKey)
    return () => document.removeEventListener('keydown', handleTabKey)
  }, [opened])

  return (
    <MantineModal
      {...props}
      opened={opened}
      onClose={onClose}
      title={title}
      aria-labelledby={ariaLabelledBy}
      aria-describedby={ariaDescribedBy}
      aria-modal={ariaModal}
      role={role}
      ref={modalRef}
      tabIndex={-1}
      // Ensure proper focus management
      style={{
        ...props.style,
        outline: 'none',
      }}
    >
      {children}
    </MantineModal>
  )
}

// Focus trap hook for complex modals
export const useFocusTrap = (isActive: boolean) => {
  const containerRef = useRef<HTMLElement>(null)

  useEffect(() => {
    if (!isActive || !containerRef.current) return

    const focusableElements = containerRef.current.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    )

    const firstElement = focusableElements[0] as HTMLElement
    const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Tab') return

      if (event.shiftKey) {
        if (document.activeElement === firstElement) {
          event.preventDefault()
          lastElement?.focus()
        }
      } else {
        if (document.activeElement === lastElement) {
          event.preventDefault()
          firstElement?.focus()
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isActive])

  return containerRef
}

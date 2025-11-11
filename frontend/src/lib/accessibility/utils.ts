/**
 * Accessibility utility functions
 */

/**
 * Generate unique ID for accessibility attributes
 */
export function generateId(prefix: string = 'id'): string {
  return `${prefix}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Get ARIA label for status badges
 */
export function getStatusAriaLabel(status: string): string {
  const statusMap: Record<string, string> = {
    'running': 'Running',
    'stopped': 'Stopped',
    'pending': 'Pending',
    'error': 'Error',
    'active': 'Active',
    'inactive': 'Inactive',
    'available': 'Available',
    'unavailable': 'Unavailable',
  };
  
  return statusMap[status.toLowerCase()] || status;
}

/**
 * Get ARIA label for action buttons
 */
export function getActionAriaLabel(action: string, itemName: string): string {
  const actionMap: Record<string, string> = {
    'start': `Start ${itemName}`,
    'stop': `Stop ${itemName}`,
    'restart': `Restart ${itemName}`,
    'delete': `Delete ${itemName}`,
    'edit': `Edit ${itemName}`,
    'view': `View ${itemName}`,
    'show': `Show ${itemName}`,
    'hide': `Hide ${itemName}`,
  };
  
  return actionMap[action.toLowerCase()] || `${action} ${itemName}`;
}

/**
 * Get ARIA description for form fields
 */
export function getFieldDescription(fieldName: string, required: boolean = false): string {
  const baseDescription = `Enter ${fieldName.toLowerCase()}`;
  return required ? `${baseDescription} (required)` : baseDescription;
}

/**
 * Get ARIA live region announcement
 */
export function getLiveRegionMessage(action: string, itemName: string, success: boolean = true): string {
  const status = success ? 'successfully' : 'failed';
  return `${itemName} ${action} ${status}`;
}

/**
 * Focus management utilities
 */
export function focusElement(elementId: string): void {
  const element = document.getElementById(elementId);
  if (element) {
    element.focus();
  }
}

export function trapFocus(container: HTMLElement): void {
  const focusableElements = container.querySelectorAll(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
  );
  
  const firstElement = focusableElements[0] as HTMLElement;
  const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;
  
  container.addEventListener('keydown', (e) => {
    if (e.key === 'Tab') {
      if (e.shiftKey) {
        if (document.activeElement === firstElement) {
          lastElement.focus();
          e.preventDefault();
        }
      } else {
        if (document.activeElement === lastElement) {
          firstElement.focus();
          e.preventDefault();
        }
      }
    }
  });
}

/**
 * Screen reader only text utility
 */
export function srOnly(text: string): string {
  return text;
}

/**
 * Get keyboard navigation instructions
 */
export function getKeyboardInstructions(): string {
  return 'Use Tab to navigate, Enter to activate, Escape to close';
}


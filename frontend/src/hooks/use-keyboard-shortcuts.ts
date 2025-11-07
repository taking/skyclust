import { useEffect } from 'react';

export interface KeyboardShortcut {
  key: string;
  ctrlKey?: boolean;
  metaKey?: boolean;
  shiftKey?: boolean;
  altKey?: boolean;
  handler: () => void;
  description: string;
}

interface UseKeyboardShortcutsOptions {
  enabled?: boolean;
  preventDefault?: boolean;
}

/**
 * Hook for managing keyboard shortcuts globally
 */
export function useKeyboardShortcuts(
  shortcuts: KeyboardShortcut[],
  options: UseKeyboardShortcutsOptions = {}
) {
  const { enabled = true, preventDefault = true } = options;

  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      const matchingShortcut = shortcuts.find((shortcut) => {
        const keyMatches = 
          shortcut.key.toLowerCase() === event.key.toLowerCase() ||
          shortcut.key === event.key;

        const ctrlMatches = shortcut.ctrlKey ? (event.ctrlKey || event.metaKey) : !(event.ctrlKey || event.metaKey);
        const metaMatches = shortcut.metaKey ? (event.metaKey || event.ctrlKey) : !(event.metaKey || event.ctrlKey);
        const shiftMatches = shortcut.shiftKey === undefined || shortcut.shiftKey === event.shiftKey;
        const altMatches = shortcut.altKey === undefined || shortcut.altKey === event.altKey;

        return keyMatches && ctrlMatches && metaMatches && shiftMatches && altMatches;
      });

      if (matchingShortcut) {
        if (preventDefault) {
          event.preventDefault();
        }
        
        // Only trigger if not typing in an input/textarea/select
        const target = event.target as HTMLElement;
        const isInputElement = 
          target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.tagName === 'SELECT' ||
          target.isContentEditable;

        if (!isInputElement) {
          matchingShortcut.handler();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [shortcuts, enabled, preventDefault]);
}

/**
 * Common keyboard shortcuts for the application
 */
export const commonShortcuts = {
  newResource: (handler: () => void): KeyboardShortcut => ({
    key: 'n',
    ctrlKey: true,
    handler,
    description: 'Create new resource',
  }),
  
  search: (handler: () => void): KeyboardShortcut => ({
    key: 'k',
    ctrlKey: true,
    handler,
    description: 'Open search',
  }),
  
  delete: (handler: () => void): KeyboardShortcut => ({
    key: 'Delete',
    handler,
    description: 'Delete selected item(s)',
  }),
  
  escape: (handler: () => void): KeyboardShortcut => ({
    key: 'Escape',
    handler,
    description: 'Close dialog/cancel',
  }),
  
  save: (handler: () => void): KeyboardShortcut => ({
    key: 's',
    ctrlKey: true,
    handler,
    description: 'Save changes',
  }),
};


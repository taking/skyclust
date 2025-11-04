'use client';

import { useRouter, usePathname } from 'next/navigation';
import { useKeyboardShortcuts, KeyboardShortcut, commonShortcuts } from '@/hooks/use-keyboard-shortcuts';
import { EVENTS, KEYBOARD_SHORTCUTS } from '@/lib/constants';
import { useState } from 'react';

export function GlobalKeyboardShortcuts() {
  const router = useRouter();
  const pathname = usePathname();
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  // Helper to open help dialog
  const openHelpDialog = () => {
    // Dispatch custom event to open help dialog
    const event = new CustomEvent(EVENTS.UI.SHOW_KEYBOARD_SHORTCUTS);
    window.dispatchEvent(event);
  };

  // Helper to toggle sidebar (mobile menu)
  const toggleSidebar = () => {
    // For mobile: toggle Sheet (mobile nav)
    const mobileMenuButton = document.querySelector('button[aria-label*="menu"], button[aria-label*="Menu"]') as HTMLElement;
    if (mobileMenuButton) {
      mobileMenuButton.click();
      return;
    }
    
    // For desktop: dispatch custom event (currently desktop sidebar is always visible)
    // This can be extended later to toggle desktop sidebar if needed
    const event = new CustomEvent(EVENTS.UI.TOGGLE_SIDEBAR);
    window.dispatchEvent(event);
  };

  // Helper to open create dialog based on current page
  const openCreateDialog = () => {
    if (pathname.startsWith('/compute/vms')) {
      const event = new CustomEvent(EVENTS.CREATE_DIALOG.VM);
      window.dispatchEvent(event);
    } else if (pathname.startsWith('/kubernetes/clusters')) {
      const event = new CustomEvent(EVENTS.CREATE_DIALOG.CLUSTER);
      window.dispatchEvent(event);
    } else if (pathname.startsWith('/networks/vpcs')) {
      const event = new CustomEvent(EVENTS.CREATE_DIALOG.VPC);
      window.dispatchEvent(event);
    } else if (pathname.startsWith('/networks/subnets')) {
      const event = new CustomEvent(EVENTS.CREATE_DIALOG.SUBNET);
      window.dispatchEvent(event);
    } else if (pathname.startsWith('/networks/security-groups')) {
      const event = new CustomEvent(EVENTS.CREATE_DIALOG.SECURITY_GROUP);
      window.dispatchEvent(event);
    } else if (pathname.startsWith('/credentials')) {
      const event = new CustomEvent(EVENTS.CREATE_DIALOG.CREDENTIAL);
      window.dispatchEvent(event);
    }
  };

  // Build keyboard shortcuts
  const shortcuts: KeyboardShortcut[] = [
    // Navigation (Shift combinations)
    {
      key: KEYBOARD_SHORTCUTS.ACTIONS.MENU_LIST.key,
      shiftKey: KEYBOARD_SHORTCUTS.ACTIONS.MENU_LIST.shiftKey,
      handler: toggleSidebar,
      description: 'Menu List',
    },
    {
      key: KEYBOARD_SHORTCUTS.ACTIONS.HELP.key,
      shiftKey: KEYBOARD_SHORTCUTS.ACTIONS.HELP.shiftKey,
      handler: openHelpDialog,
      description: 'Show keyboard shortcuts',
    },
    
    // Action (Shift combinations)
    {
      key: KEYBOARD_SHORTCUTS.ACTIONS.CREATE_NEW.key,
      shiftKey: KEYBOARD_SHORTCUTS.ACTIONS.CREATE_NEW.shiftKey,
      handler: openCreateDialog,
      description: 'Create New Resource',
    },
    
    // General Navigation (single keys)
    {
      key: KEYBOARD_SHORTCUTS.NAVIGATION.DASHBOARD,
      handler: () => router.push('/dashboard'),
      description: 'Go to Dashboard',
    },
    {
      key: KEYBOARD_SHORTCUTS.NAVIGATION.COMPUTE,
      handler: () => router.push('/compute/vms'),
      description: 'Go to Compute',
    },
    {
      key: KEYBOARD_SHORTCUTS.NAVIGATION.KUBERNETES,
      handler: () => router.push('/kubernetes/clusters'),
      description: 'Go to Kubernetes',
    },
    {
      key: KEYBOARD_SHORTCUTS.NAVIGATION.NETWORKS,
      handler: () => router.push('/networks/vpcs'),
      description: 'Go to Networks',
    },
    {
      key: KEYBOARD_SHORTCUTS.NAVIGATION.CREDENTIALS,
      handler: () => router.push('/credentials'),
      description: 'Go to Credentials',
    },
    
    // Global search (Ctrl/Cmd + K) - Keep existing
    commonShortcuts.search(() => {
      setIsSearchOpen(true);
      // Focus search input if it exists
      const searchInput = document.querySelector('input[type="search"], input[placeholder*="Search"]') as HTMLInputElement;
      if (searchInput) {
        setTimeout(() => searchInput.focus(), 100);
      }
    }),
  ];

  useKeyboardShortcuts(shortcuts);

  // Store shortcuts globally for help dialog
  if (typeof window !== 'undefined') {
    // Window 타입 확장
    interface WindowWithShortcuts extends Window {
      __keyboardShortcuts?: KeyboardShortcut[];
    }
    (window as WindowWithShortcuts).__keyboardShortcuts = shortcuts;
  }

  return null;
}


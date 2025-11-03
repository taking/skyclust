'use client';

import { useRouter, usePathname } from 'next/navigation';
import { useKeyboardShortcuts, KeyboardShortcut, commonShortcuts } from '@/hooks/use-keyboard-shortcuts';
import { useState } from 'react';
import { KeyboardShortcutsHelp } from './keyboard-shortcuts-help';

export function GlobalKeyboardShortcuts() {
  const router = useRouter();
  const pathname = usePathname();
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  // Build page-specific shortcuts
  const shortcuts: KeyboardShortcut[] = [
    // Global shortcuts
    {
      key: '?',
      shiftKey: true,
      handler: () => {
        // Will be handled by KeyboardShortcutsHelp dialog
      },
      description: 'Show keyboard shortcuts',
    },
    
    // Navigation shortcuts
    {
      key: 'g',
      handler: () => router.push('/dashboard'),
      description: 'Go to Dashboard',
    },
    {
      key: 'g',
      ctrlKey: true,
      handler: () => router.push('/dashboard'),
      description: 'Go to Dashboard',
    },
    
    // Page-specific shortcuts
    ...(pathname === '/vms' ? [
      commonShortcuts.newResource(() => {
        // Trigger VM creation dialog
        const event = new CustomEvent('open-create-vm-dialog');
        window.dispatchEvent(event);
      }),
      {
        key: 'g',
        handler: () => router.push('/vms'),
        description: 'Go to VMs',
      },
    ] : []),
    
    ...(pathname?.startsWith('/kubernetes') ? [
      commonShortcuts.newResource(() => {
        const event = new CustomEvent('open-create-cluster-dialog');
        window.dispatchEvent(event);
      }),
      {
        key: 'g',
        handler: () => router.push('/kubernetes'),
        description: 'Go to Kubernetes',
      },
    ] : []),
    
    ...(pathname?.startsWith('/networks') ? [
      commonShortcuts.newResource(() => {
        const event = new CustomEvent('open-create-network-dialog');
        window.dispatchEvent(event);
      }),
      {
        key: 'g',
        handler: () => router.push('/networks'),
        description: 'Go to Networks',
      },
    ] : []),
    
    // Global search (Ctrl/Cmd + K)
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


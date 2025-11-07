'use client';

import { useRouter, usePathname } from 'next/navigation';
import { useKeyboardShortcuts, KeyboardShortcut, commonShortcuts } from '@/hooks/use-keyboard-shortcuts';
import { EVENTS, KEYBOARD_SHORTCUTS } from '@/lib/constants';
import { useState } from 'react';

export function GlobalKeyboardShortcuts() {
  const router = useRouter();
  const pathname = usePathname();
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  // 도움말 다이얼로그 열기 헬퍼
  const openHelpDialog = () => {
    // 도움말 다이얼로그를 열기 위한 커스텀 이벤트 발생
    const event = new CustomEvent(EVENTS.UI.SHOW_KEYBOARD_SHORTCUTS);
    window.dispatchEvent(event);
  };

  // 사이드바 토글 헬퍼 (모바일 메뉴)
  const toggleSidebar = () => {
    // 모바일: Sheet 토글 (모바일 네비게이션)
    const mobileMenuButton = document.querySelector('button[aria-label*="menu"], button[aria-label*="Menu"]') as HTMLElement;
    if (mobileMenuButton) {
      mobileMenuButton.click();
      return;
    }
    
    // 데스크톱: 커스텀 이벤트 발생 (현재 데스크톱 사이드바는 항상 표시됨)
    // 필요시 나중에 데스크톱 사이드바 토글 기능 확장 가능
    const event = new CustomEvent(EVENTS.UI.TOGGLE_SIDEBAR);
    window.dispatchEvent(event);
  };

  // 현재 페이지에 따라 생성 다이얼로그 열기 헬퍼
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

  // 키보드 단축키 구성
  const shortcuts: KeyboardShortcut[] = [
    // 네비게이션 (Shift 조합)
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
    // 단일 ? 키로 도움말 열기 (입력 필드가 아닐 때)
    // 참고: useKeyboardShortcuts 훅에서 이미 입력 필드를 확인함
    {
      key: '?',
      handler: openHelpDialog,
      description: 'Show keyboard shortcuts (press ?)',
    },
    
    // 액션 (Shift 조합)
    {
      key: KEYBOARD_SHORTCUTS.ACTIONS.CREATE_NEW.key,
      shiftKey: KEYBOARD_SHORTCUTS.ACTIONS.CREATE_NEW.shiftKey,
      handler: openCreateDialog,
      description: 'Create New Resource',
    },
    
    // 일반 네비게이션 (단일 키)
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
    
    // 전역 검색 (Ctrl/Cmd + K) - 기존 유지
    commonShortcuts.search(() => {
      setIsSearchOpen(true);
      // 검색 입력 필드가 있으면 포커스
      const searchInput = document.querySelector('input[type="search"], input[placeholder*="Search"]') as HTMLInputElement;
      if (searchInput) {
        setTimeout(() => searchInput.focus(), 100);
      }
    }),
  ];

  useKeyboardShortcuts(shortcuts);

  // 도움말 다이얼로그를 위해 단축키를 전역으로 저장
  if (typeof window !== 'undefined') {
    // Window 타입 확장
    interface WindowWithShortcuts extends Window {
      __keyboardShortcuts?: KeyboardShortcut[];
    }
    (window as WindowWithShortcuts).__keyboardShortcuts = shortcuts;
  }

  return null;
}


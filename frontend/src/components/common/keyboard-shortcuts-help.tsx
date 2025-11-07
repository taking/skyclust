'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Keyboard } from 'lucide-react';
import { KeyboardShortcut } from '@/hooks/use-keyboard-shortcuts';
import { EVENTS } from '@/lib/constants';
import { useTranslation } from '@/hooks/use-translation';

interface KeyboardShortcutsHelpProps {
  shortcuts: KeyboardShortcut[];
  trigger?: React.ReactNode;
}

export function KeyboardShortcutsHelp({ shortcuts: propShortcuts, trigger }: KeyboardShortcutsHelpProps) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = React.useState(false);
  const [shortcuts, setShortcuts] = React.useState<KeyboardShortcut[]>([]);

  /**
   * window.__keyboardShortcuts에서 shortcuts를 가져오는 함수
   */
  const loadShortcuts = React.useCallback(() => {
    // 1. propShortcuts가 제공되고 비어있지 않으면 사용
    if (propShortcuts && propShortcuts.length > 0) {
      setShortcuts(propShortcuts);
      return;
    }
    
    // 2. window.__keyboardShortcuts에서 가져오기
    if (typeof window !== 'undefined') {
      const windowShortcuts = (window as Window & { __keyboardShortcuts?: KeyboardShortcut[] }).__keyboardShortcuts;
      if (windowShortcuts && windowShortcuts.length > 0) {
        setShortcuts(windowShortcuts);
        return;
      }
    }
    
    // 3. 둘 다 없으면 빈 배열로 설정
    setShortcuts([]);
  }, [propShortcuts]);

  // 컴포넌트 마운트 시 및 propShortcuts 변경 시 shortcuts 로드
  React.useEffect(() => {
    loadShortcuts();
  }, [loadShortcuts]);

  // 다이얼로그가 열릴 때마다 shortcuts 다시 로드 (window.__keyboardShortcuts가 나중에 설정될 수 있음)
  React.useEffect(() => {
    if (isOpen) {
      loadShortcuts();
    }
  }, [isOpen, loadShortcuts]);

  // Shift + ? 키보드 단축키 리스닝
  React.useEffect(() => {
    const handleShowShortcuts = () => {
      setIsOpen(true);
    };
    window.addEventListener(EVENTS.UI.SHOW_KEYBOARD_SHORTCUTS, handleShowShortcuts);
    return () => {
      window.removeEventListener(EVENTS.UI.SHOW_KEYBOARD_SHORTCUTS, handleShowShortcuts);
    };
  }, []);

  const formatKeyCombo = (shortcut: KeyboardShortcut): string => {
    const parts: string[] = [];
    
    if (shortcut.ctrlKey || shortcut.metaKey) {
      parts.push(navigator.platform.includes('Mac') ? '⌘' : 'Ctrl');
    }
    if (shortcut.shiftKey) {
      parts.push('Shift');
    }
    if (shortcut.altKey) {
      parts.push('Alt');
    }
    
    // 키 표시 형식화
    let keyDisplay = shortcut.key;
    if (keyDisplay === ' ') {
      keyDisplay = 'Space';
    } else if (keyDisplay === '?') {
      keyDisplay = '?';
    } else if (keyDisplay.length === 1 && keyDisplay.match(/[a-z]/i)) {
      // 단일 문자 키 - 표시를 위해 대문자로 변환
      keyDisplay = keyDisplay.toUpperCase();
    }
    
    parts.push(keyDisplay);
    
    return parts.join(' + ');
  };

  const groupedShortcuts = React.useMemo(() => {
    const groups: Record<string, KeyboardShortcut[]> = {
      navigation: [],
      action: [],
      general: [],
    };

    shortcuts.forEach((shortcut) => {
      const desc = shortcut.description.toLowerCase();
      
      // 설명과 키 조합에 따라 분류
      if (desc.includes('menu list') || desc.includes('menu')) {
        groups.navigation.push(shortcut);
      } else if (desc.includes('go to') || desc.includes('dashboard') || desc.includes('compute') || 
                 desc.includes('kubernetes') || desc.includes('networks') || desc.includes('credentials')) {
        groups.general.push(shortcut);
      } else if (desc.includes('create') || desc.includes('new resource')) {
        groups.action.push(shortcut);
      } else if (desc.includes('show keyboard shortcuts') || desc.includes('keyboard shortcuts')) {
        groups.general.push(shortcut);
      } else if (desc.includes('search') || desc.includes('open')) {
        groups.navigation.push(shortcut);
      } else if (desc.includes('delete') || desc.includes('save')) {
        groups.action.push(shortcut);
      } else {
        groups.general.push(shortcut);
      }
    });

    return Object.entries(groups).filter(([_, shortcuts]) => shortcuts.length > 0);
  }, [shortcuts]);
  
  const categoryLabels: Record<string, string> = {
    navigation: t('shortcuts.categories.navigation'),
    action: t('shortcuts.categories.action'),
    general: t('shortcuts.categories.general'),
  };

  return (
    <>
      {trigger ? (
        <div onClick={() => setIsOpen(true)} style={{ display: 'inline-block' }}>
          {trigger}
        </div>
      ) : (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setIsOpen(true)}
          className="text-xs"
        >
          <Keyboard className="mr-2 h-4 w-4" />
          {t('shortcuts.shortcutsLabel')}
        </Button>
      )}
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{t('shortcuts.title')}</DialogTitle>
            <DialogDescription>
              {t('shortcuts.description')}
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-6 mt-4">
            {shortcuts.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <p className="text-sm">키보드 단축키를 불러오는 중...</p>
              </div>
            ) : groupedShortcuts.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <p className="text-sm">표시할 키보드 단축키가 없습니다.</p>
              </div>
            ) : (
              groupedShortcuts.map(([groupName, groupShortcuts]) => (
                <div key={groupName}>
                  <h3 className="text-sm font-semibold text-gray-900 mb-3">{categoryLabels[groupName] || groupName}</h3>
                  <div className="space-y-2">
                    {groupShortcuts.map((shortcut, index) => (
                      <div
                        key={index}
                        className="flex items-center justify-between py-2 px-3 rounded-md hover:bg-gray-50"
                      >
                        <span className="text-sm text-gray-700">{shortcut.description}</span>
                        <kbd className="px-2 py-1 text-xs font-semibold text-gray-800 bg-gray-100 border border-gray-300 rounded-md shadow-sm">
                          {formatKeyCombo(shortcut)}
                        </kbd>
                      </div>
                    ))}
                  </div>
                </div>
              ))
            )}
          </div>
          
          <div className="mt-6 pt-4 border-t">
            <p className="text-xs text-gray-500">
              {t('shortcuts.tip')}
            </p>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}


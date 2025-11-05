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

export function KeyboardShortcutsHelp({ shortcuts, trigger }: KeyboardShortcutsHelpProps) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = React.useState(false);

  // Listen for Shift + ? keyboard shortcut
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
    
    // Format key display
    let keyDisplay = shortcut.key;
    if (keyDisplay === ' ') {
      keyDisplay = 'Space';
    } else if (keyDisplay === '?') {
      keyDisplay = '?';
    } else if (keyDisplay.length === 1 && keyDisplay.match(/[a-z]/i)) {
      // Single letter keys - capitalize for display
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
      
      // Categorize based on description and key combination
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
  }, [shortcuts, t]);
  
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
            {groupedShortcuts.map(([groupName, groupShortcuts]) => (
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
            ))}
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


'use client';

import * as React from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Keyboard } from 'lucide-react';
import { KeyboardShortcut } from '@/hooks/use-keyboard-shortcuts';

interface KeyboardShortcutsHelpProps {
  shortcuts: KeyboardShortcut[];
  trigger?: React.ReactNode;
}

export function KeyboardShortcutsHelp({ shortcuts, trigger }: KeyboardShortcutsHelpProps) {
  const [isOpen, setIsOpen] = React.useState(false);

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
    parts.push(shortcut.key === ' ' ? 'Space' : shortcut.key);
    
    return parts.join(' + ');
  };

  const groupedShortcuts = React.useMemo(() => {
    const groups: Record<string, KeyboardShortcut[]> = {
      'Navigation': [],
      'Actions': [],
      'General': [],
    };

    shortcuts.forEach((shortcut) => {
      const desc = shortcut.description.toLowerCase();
      if (desc.includes('search') || desc.includes('open')) {
        groups['Navigation'].push(shortcut);
      } else if (desc.includes('delete') || desc.includes('save') || desc.includes('create')) {
        groups['Actions'].push(shortcut);
      } else {
        groups['General'].push(shortcut);
      }
    });

    return Object.entries(groups).filter(([_, shortcuts]) => shortcuts.length > 0);
  }, [shortcuts]);

  return (
    <>
      {trigger ? (
        <DialogTrigger asChild onClick={() => setIsOpen(true)}>
          {trigger}
        </DialogTrigger>
      ) : (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setIsOpen(true)}
          className="text-xs"
        >
          <Keyboard className="mr-2 h-4 w-4" />
          Shortcuts
        </Button>
      )}

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Keyboard Shortcuts</DialogTitle>
            <DialogDescription>
              Use these keyboard shortcuts to navigate and perform actions faster
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-6 mt-4">
            {groupedShortcuts.map(([groupName, groupShortcuts]) => (
              <div key={groupName}>
                <h3 className="text-sm font-semibold text-gray-900 mb-3">{groupName}</h3>
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
              Tip: Keyboard shortcuts are disabled when typing in input fields
            </p>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}


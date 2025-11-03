'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useAuthStore } from '@/store/auth';
import { useRouter } from 'next/navigation';
import { LogOut, User, Settings } from 'lucide-react';
import { MobileNav } from './mobile-nav';
import { ThemeToggle } from '@/components/theme/theme-toggle';
import { ScreenReaderOnly } from '@/components/accessibility/screen-reader-only';
import { getActionAriaLabel } from '@/lib/accessibility';
import { KeyboardShortcutsHelp } from '@/components/common/keyboard-shortcuts-help';
import { KeyboardShortcut } from '@/hooks/use-keyboard-shortcuts';

function HeaderComponent() {
  const { user, logout } = useAuthStore();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <header className="border-b bg-background" role="banner">
      <div className="flex h-16 items-center justify-between px-4 sm:px-6">
        <div className="flex items-center space-x-4">
          <MobileNav />
          <h1 className="text-lg sm:text-xl font-bold text-foreground">
            SkyClust
            <ScreenReaderOnly>Multi-Cloud Management Platform</ScreenReaderOnly>
          </h1>
        </div>

        <div className="flex items-center space-x-4">
          <KeyboardShortcutsHelp 
            shortcuts={(typeof window !== 'undefined' && (window as Window & { __keyboardShortcuts?: KeyboardShortcut[] }).__keyboardShortcuts) || []}
          />
          <ThemeToggle />
          {user ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button 
                  variant="ghost" 
                  className="relative h-8 w-8 rounded-full"
                  aria-label={`User menu for ${user.username}`}
                  aria-haspopup="menu"
                >
                  <Avatar className="h-8 w-8">
                    <AvatarImage src="" alt={`${user.username}'s avatar`} />
                    <AvatarFallback>
                      {user.username.charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56" align="end" forceMount role="menu">
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">{user.username}</p>
                    <p className="text-xs leading-none text-muted-foreground">
                      {user.email}
                    </p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={() => router.push('/profile')}
                  role="menuitem"
                  aria-label="Go to profile page"
                >
                  <User className="mr-2 h-4 w-4" aria-hidden="true" />
                  <span>Profile</span>
                </DropdownMenuItem>
                <DropdownMenuItem 
                  onClick={() => router.push('/settings')}
                  role="menuitem"
                  aria-label="Go to settings page"
                >
                  <Settings className="mr-2 h-4 w-4" aria-hidden="true" />
                  <span>Settings</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={handleLogout}
                  role="menuitem"
                  aria-label="Log out of account"
                >
                  <LogOut className="mr-2 h-4 w-4" aria-hidden="true" />
                  <span>Log out</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <div className="flex items-center space-x-2">
              <Button variant="ghost" onClick={() => router.push('/login')}>
                Login
              </Button>
              <Button onClick={() => router.push('/register')}>
                Sign Up
              </Button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}

export const Header = React.memo(HeaderComponent);

'use client';

import * as React from 'react';
import { usePathname, useSearchParams } from 'next/navigation';
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useAuthStore } from '@/store/auth';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useRouter } from 'next/navigation';
import { LogOut, User, Settings, Plus } from 'lucide-react';
import { useCredentials } from '@/hooks/use-credentials';
import { MobileNav } from './mobile-nav';
import { ThemeToggle } from '@/components/theme/theme-toggle';
import { ScreenReaderOnly } from '@/components/accessibility/screen-reader-only';
import { getActionAriaLabel } from '@/lib/accessibility';
import { Breadcrumb } from '@/components/common/breadcrumb';
import { getRegionsForProvider, supportsRegionSelection } from '@/lib/regions';
import type { CloudProvider } from '@/lib/types';
import { useTranslation } from '@/hooks/use-translation';
import { locales, localeNames, type Locale } from '@/i18n/config';

function HeaderComponent() {
  const { user, logout } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion } = useCredentialContextStore();
  const { t, locale, setLocale } = useTranslation();

  // Check if we should show credential/region selectors
  const shouldShowSelectors = React.useMemo(() => {
    return pathname.startsWith('/compute') || 
           pathname.startsWith('/kubernetes') || 
           pathname.startsWith('/networks');
  }, [pathname]);

  // Fetch credentials for current workspace using unified hook
  const { credentials, selectedCredential: selectedCredentialFromHook } = useCredentials({
    workspaceId: currentWorkspace?.id,
    selectedCredentialId: selectedCredentialId || undefined,
    enabled: !!currentWorkspace && shouldShowSelectors,
  });

  // Clear credential selection when workspace changes or selected credential is not in current workspace
  React.useEffect(() => {
    if (!shouldShowSelectors || !currentWorkspace) return;

    // Check if selected credential belongs to current workspace
    if (selectedCredentialId) {
      const credentialExists = credentials.some(c => c.id === selectedCredentialId);
      if (!credentialExists) {
        // Clear selection if credential doesn't exist in current workspace
        setSelectedCredential(null);
        setSelectedRegion(null);
        // Update URL
        const params = new URLSearchParams(searchParams.toString());
        params.delete('credentialId');
        params.delete('region');
        router.replace(`${pathname}?${params.toString()}`, { scroll: false });
      }
    }
  }, [currentWorkspace?.id, credentials, selectedCredentialId, shouldShowSelectors, pathname, router, searchParams, setSelectedCredential, setSelectedRegion]);

  // Get selected credential and provider (from hook)
  const selectedCredential = selectedCredentialFromHook;
  const selectedProvider = selectedCredential?.provider as CloudProvider | undefined;

  // Get regions for selected provider
  const regions = React.useMemo(() => getRegionsForProvider(selectedProvider), [selectedProvider]);
  const showRegionSelector = supportsRegionSelection(selectedProvider) && shouldShowSelectors;

  // Sync with URL query parameters
  React.useEffect(() => {
    if (!shouldShowSelectors) return;

    const urlCredentialId = searchParams.get('credentialId');
    const urlRegion = searchParams.get('region');

    // Sync from URL to store
    if (urlCredentialId && urlCredentialId !== selectedCredentialId) {
      setSelectedCredential(urlCredentialId);
    }
    if (urlRegion !== null && urlRegion !== selectedRegion) {
      setSelectedRegion(urlRegion || null);
    }
  }, [searchParams, shouldShowSelectors, selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion]);

  // Update URL when selection changes
  const handleCredentialChange = React.useCallback((credentialId: string) => {
    setSelectedCredential(credentialId);
    
    // Update URL
    const params = new URLSearchParams(searchParams.toString());
    if (credentialId) {
      params.set('credentialId', credentialId);
    } else {
      params.delete('credentialId');
    }
    // Clear region when credential changes (provider might change)
    if (credentialId) {
      const newCredential = credentials.find(c => c.id === credentialId);
      const newProvider = newCredential?.provider;
      if (!supportsRegionSelection(newProvider)) {
        params.delete('region');
        setSelectedRegion(null);
      }
    } else {
      params.delete('region');
      setSelectedRegion(null);
    }
    
    router.replace(`${pathname}?${params.toString()}`, { scroll: false });
  }, [credentials, pathname, router, searchParams, setSelectedCredential, setSelectedRegion]);

  const handleRegionChange = React.useCallback((region: string) => {
    const regionValue = region === 'all' ? '' : region;
    setSelectedRegion(regionValue || null);
    
    // Update URL
    const params = new URLSearchParams(searchParams.toString());
    if (regionValue) {
      params.set('region', regionValue);
    } else {
      params.delete('region');
    }
    
    router.replace(`${pathname}?${params.toString()}`, { scroll: false });
  }, [pathname, router, searchParams, setSelectedRegion]);

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <header className="border-b bg-background" role="banner">
      <div className="flex flex-col">
        {/* Main Header Row */}
        <div className="flex h-16 items-center justify-between px-4 sm:px-6">
          <div className="flex items-center space-x-4 flex-1 min-w-0">
            <MobileNav />
            <div className="flex-1 min-w-0">
              <Breadcrumb className="text-sm" />
            </div>
          </div>

          <div className="flex items-center space-x-4 flex-shrink-0">
            {/* Credential and Region Selectors */}
            {shouldShowSelectors && (
              <>
                <Select
                  value={selectedCredentialId || ''}
                  onValueChange={handleCredentialChange}
                >
                  <SelectTrigger className="w-[200px] h-8 text-xs">
                    <SelectValue placeholder={typeof t === 'function' ? t('credential.select') : 'Select Credential'} />
                  </SelectTrigger>
                  <SelectContent>
                    {credentials.map((credential) => (
                      <SelectItem key={credential.id} value={credential.id} className="text-xs">
                        {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
                      </SelectItem>
                    ))}
                    <div className="h-px bg-border my-1" />
                    <div className="px-2 py-1.5 border-t border-border">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start"
                        onClick={(e) => {
                          e.stopPropagation();
                          router.push('/credentials');
                        }}
                        aria-label="Create a new credential"
                      >
                        <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                        {typeof t === 'function' ? t('credential.create') : 'Create Credential'}
                      </Button>
                    </div>
                  </SelectContent>
                </Select>
                {showRegionSelector && (
                  <Select
                    value={selectedRegion || 'all'}
                    onValueChange={handleRegionChange}
                  >
                    <SelectTrigger className="w-[180px] h-8 text-xs">
                      <SelectValue placeholder={typeof t === 'function' ? t('region.select') : 'Select Region'} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">{typeof t === 'function' ? t('region.select') : 'All Regions'}</SelectItem>
                      {regions.map((region) => (
                        <SelectItem key={region.value} value={region.value} className="text-xs">
                          {region.value} - {region.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </>
            )}

            {/* Language Selector */}
            <Select
              value={locale}
              onValueChange={(value) => setLocale(value as Locale)}
            >
              <SelectTrigger className="w-[120px] h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {locales.map((loc) => (
                  <SelectItem key={loc} value={loc} className="text-xs">
                    {localeNames[loc]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

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
                  <span>{typeof t === 'function' ? t('user.profile') : 'Profile'}</span>
                </DropdownMenuItem>
                <DropdownMenuItem 
                  onClick={() => router.push('/settings')}
                  role="menuitem"
                  aria-label="Go to settings page"
                >
                  <Settings className="mr-2 h-4 w-4" aria-hidden="true" />
                  <span>{typeof t === 'function' ? t('user.settings') : 'Settings'}</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={handleLogout}
                  role="menuitem"
                  aria-label="Log out of account"
                >
                  <LogOut className="mr-2 h-4 w-4" aria-hidden="true" />
                  <span>{typeof t === 'function' ? t('user.logout') : 'Log out'}</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <div className="flex items-center space-x-2">
              <Button variant="ghost" onClick={() => router.push('/login')}>
                {typeof t === 'function' ? t('user.login') : 'Login'}
              </Button>
              <Button onClick={() => router.push('/register')}>
                {typeof t === 'function' ? t('user.signUp') : 'Sign Up'}
              </Button>
            </div>
          )}
          </div>
        </div>
      </div>
    </header>
  );
}

export const Header = React.memo(HeaderComponent);

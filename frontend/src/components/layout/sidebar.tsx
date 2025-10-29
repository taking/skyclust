'use client';

import { Button } from '@/components/ui/button';
import { useWorkspaceStore } from '@/store/workspace';
import { useRouter, usePathname } from 'next/navigation';
import {
  Home,
  Server,
  Key,
  Settings,
  Users,
  Plus,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { ScreenReaderOnly } from '@/components/accessibility/screen-reader-only';
import { getActionAriaLabel } from '@/lib/accessibility';

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: Home },
  { name: 'VMs', href: '/vms', icon: Server },
  { name: 'Credentials', href: '/credentials', icon: Key },
  { name: 'Workspaces', href: '/workspaces', icon: Users },
  { name: 'Profile', href: '/profile', icon: Settings },
];

export function Sidebar() {
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const pathname = usePathname();

  return (
    <nav className="flex h-full w-64 flex-col bg-card border-r lg:block hidden" role="navigation" aria-label="Main navigation">
      <div className="flex flex-col flex-1 overflow-y-auto">
        <div className="flex flex-col flex-1 px-3 py-4">
          <div className="flex items-center flex-shrink-0 px-4">
            <h2 className="text-lg font-semibold text-card-foreground truncate">
              {currentWorkspace?.name || 'Select Workspace'}
              <ScreenReaderOnly>
                {currentWorkspace ? 'Current workspace' : 'No workspace selected'}
              </ScreenReaderOnly>
            </h2>
          </div>
          
          {!currentWorkspace && (
            <div className="mt-4">
              <Button
                onClick={() => router.push('/workspaces')}
                className="w-full justify-start"
                variant="outline"
                aria-label="Create a new workspace"
              >
                <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                Create Workspace
              </Button>
            </div>
          )}

          <div className="mt-8 flex-1 space-y-1" role="list">
            {navigation.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Button
                  key={item.name}
                  onClick={() => router.push(item.href)}
                  variant={isActive ? 'secondary' : 'ghost'}
                  className={cn(
                    'w-full justify-start',
                    isActive && 'bg-accent'
                  )}
                  role="listitem"
                  aria-current={isActive ? 'page' : undefined}
                  aria-label={`Navigate to ${item.name} page`}
                >
                  <item.icon className="mr-3 h-5 w-5" aria-hidden="true" />
                  {item.name}
                </Button>
              );
            })}
          </div>
        </div>
      </div>
    </nav>
  );
}

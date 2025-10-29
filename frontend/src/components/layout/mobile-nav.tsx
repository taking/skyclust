'use client';

import { useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { useWorkspaceStore } from '@/store/workspace';
import { Menu, Home, Server, Key, Users, Settings } from 'lucide-react';

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: Home },
  { name: 'VMs', href: '/vms', icon: Server },
  { name: 'Credentials', href: '/credentials', icon: Key },
  { name: 'Workspaces', href: '/workspaces', icon: Users },
  { name: 'Profile', href: '/profile', icon: Settings },
];

export function MobileNav() {
  const [open, setOpen] = useState(false);
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const pathname = usePathname();

  const handleNavigation = (href: string) => {
    router.push(href);
    setOpen(false);
  };

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="sm" className="lg:hidden">
          <Menu className="h-5 w-5" />
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-64">
        <div className="flex flex-col h-full">
          <div className="flex items-center space-x-2 mb-6">
            <h2 className="text-lg font-semibold text-gray-900">SkyClust</h2>
          </div>
          
          <div className="mb-4">
            <h3 className="text-sm font-medium text-gray-500 uppercase tracking-wider">
              Workspace
            </h3>
            <p className="text-sm text-gray-900 truncate mt-1">
              {currentWorkspace?.name || 'Select Workspace'}
            </p>
          </div>

          <nav className="flex-1 space-y-1">
            {navigation.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Button
                  key={item.name}
                  onClick={() => handleNavigation(item.href)}
                  variant={isActive ? 'secondary' : 'ghost'}
                  className="w-full justify-start"
                >
                  <item.icon className="mr-3 h-5 w-5" />
                  {item.name}
                </Button>
              );
            })}
          </nav>

          {!currentWorkspace && (
            <div className="mt-4">
              <Button
                onClick={() => handleNavigation('/workspaces')}
                className="w-full"
                variant="outline"
              >
                Create Workspace
              </Button>
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}

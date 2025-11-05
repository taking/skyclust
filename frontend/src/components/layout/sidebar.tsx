'use client';

import * as React from 'react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Form } from '@/components/ui/form';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { useRouter, usePathname } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Home,
  Server,
  Key,
  Plus,
  Container,
  Network,
  Settings,
  ChevronDown,
  Image,
  HardDrive,
  Layers,
  Shield,
  Building2,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { ScreenReaderOnly } from '@/components/accessibility/screen-reader-only';
import { workspaceService } from '@/features/workspaces';
import { useToast } from '@/hooks/use-toast';
import { ErrorHandler } from '@/lib/error-handler';
import { useFormWithValidation, EnhancedField } from '@/hooks/use-form-with-validation';
import { CreateWorkspaceForm } from '@/lib/types';
import * as z from 'zod';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import { useTranslation } from '@/hooks/use-translation';

const createWorkspaceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters'),
});

// Navigation structure with hierarchical menu
interface NavigationItem {
  name: string;
  href?: string;
  icon: React.ComponentType<{ className?: string }>;
  children?: NavigationItem[];
}

// Navigation will be generated dynamically with translations
function getNavigation(t: (key: string) => string): NavigationItem[] {
  return [
    { name: t('nav.dashboard'), href: '/dashboard', icon: Home },
    {
      name: t('nav.compute'),
      icon: Server,
      children: [
        { name: t('nav.vms'), href: '/compute/vms', icon: Server },
        { name: t('nav.images'), href: '/compute/images', icon: Image },
        { name: t('nav.snapshots'), href: '/compute/snapshots', icon: HardDrive },
      ],
    },
    {
      name: t('nav.kubernetes'),
      icon: Container,
      children: [
        { name: t('nav.clusters'), href: '/kubernetes/clusters', icon: Container },
        { name: t('nav.nodePools'), href: '/kubernetes/node-pools', icon: Layers },
        { name: t('nav.nodes'), href: '/kubernetes/nodes', icon: Building2 },
      ],
    },
    {
      name: t('nav.networks'),
      icon: Network,
      children: [
        { name: t('nav.vpcs'), href: '/networks/vpcs', icon: Network },
        { name: t('nav.subnets'), href: '/networks/subnets', icon: Layers },
        { name: t('nav.securityGroups'), href: '/networks/security-groups', icon: Shield },
      ],
    },
    { name: t('nav.credentials'), href: '/credentials', icon: Key },
  ];
}

function SidebarComponent() {
  const { currentWorkspace, setCurrentWorkspace, workspaces, setWorkspaces } = useWorkspaceStore();
  const router = useRouter();
  const pathname = usePathname();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const { t } = useTranslation();
  
  // navigation을 안전하게 생성 (t가 함수인지 확인)
  const navigation = React.useMemo(() => {
    if (typeof t === 'function') {
      return getNavigation(t);
    }
    // Fallback: 기본 영어 메뉴
    return [
      { name: 'Dashboard', href: '/dashboard', icon: Home },
      {
        name: 'Compute',
        icon: Server,
        children: [
          { name: 'VMs', href: '/compute/vms', icon: Server },
          { name: 'Images', href: '/compute/images', icon: Image },
          { name: 'Snapshots', href: '/compute/snapshots', icon: HardDrive },
        ],
      },
      {
        name: 'Kubernetes',
        icon: Container,
        children: [
          { name: 'Clusters', href: '/kubernetes/clusters', icon: Container },
          { name: 'Node Pools', href: '/kubernetes/node-pools', icon: Layers },
          { name: 'Nodes', href: '/kubernetes/nodes', icon: Building2 },
        ],
      },
      {
        name: 'Networks',
        icon: Network,
        children: [
          { name: 'VPCs', href: '/networks/vpcs', icon: Network },
          { name: 'Subnets', href: '/networks/subnets', icon: Layers },
          { name: 'Security Groups', href: '/networks/security-groups', icon: Shield },
        ],
      },
      { name: 'Credentials', href: '/credentials', icon: Key },
    ];
  }, [t]);

  // Fetch workspaces
  const { data: fetchedWorkspaces = [] } = useQuery({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    staleTime: CACHE_TIMES.STABLE,
    gcTime: GC_TIMES.LONG,
    retry: 3,
    retryDelay: 1000,
  });

  // Update store when workspaces are fetched
  React.useEffect(() => {
    if (fetchedWorkspaces.length > 0 && workspaces.length === 0) {
      setWorkspaces(fetchedWorkspaces);
    }
  }, [fetchedWorkspaces, workspaces.length, setWorkspaces]);

  // Sync currentWorkspace from URL parameter for Settings/Members pages
  React.useEffect(() => {
    if (fetchedWorkspaces.length === 0) return;
    
    // Check if we're on Settings or Members page
    if (pathname.startsWith('/workspaces/') && (pathname.includes('/settings') || pathname.includes('/members'))) {
      // Extract workspace ID from URL path
      const match = pathname.match(/\/workspaces\/([^/]+)\/(settings|members)/);
      if (match && match[1]) {
        const urlWorkspaceId = match[1];
        const workspaceFromUrl = fetchedWorkspaces.find(w => w.id === urlWorkspaceId);
        
        // If workspace exists and current workspace doesn't match, update it
        if (workspaceFromUrl && currentWorkspace?.id !== urlWorkspaceId) {
          setCurrentWorkspace(workspaceFromUrl);
        }
      }
    }
  }, [pathname, fetchedWorkspaces, currentWorkspace?.id, setCurrentWorkspace]);

  // Workspace creation form
  const {
    form,
    handleSubmit,
    isLoading: isFormLoading,
    error: formError,
    reset,
    getFieldError,
    getFieldValidationState,
  } = useFormWithValidation<CreateWorkspaceForm>({
    schema: createWorkspaceSchema,
    defaultValues: {
      name: '',
      description: '',
    },
    onSubmit: async (data) => {
      await workspaceService.createWorkspace(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.workspaces.all });
      setIsCreateDialogOpen(false);
      reset();
      // t 함수를 안전하게 사용
      const successMessage = typeof t === 'function' 
        ? t('messages.created', { resource: t('workspace.title') })
        : 'Workspace created successfully';
      success(successMessage);
    },
    onError: (error) => {
      ErrorHandler.logError(error, { operation: 'createWorkspace', source: 'sidebar' });
      // t 함수를 안전하게 사용
      const errorMessage = typeof t === 'function'
        ? t('messages.operationFailed')
        : 'Failed to create workspace';
      showError(errorMessage);
    },
    resetOnSuccess: true,
  });

  const handleWorkspaceChange = (workspaceId: string) => {
    if (workspaceId === 'all' || !workspaceId) return;
    
    const workspace = fetchedWorkspaces.find(w => w.id === workspaceId);
    if (!workspace) return;
    
    const previousWorkspaceId = currentWorkspace?.id;
    const isWorkspaceChanged = previousWorkspaceId !== workspace.id;
    
    setCurrentWorkspace(workspace);
    
    if (isWorkspaceChanged) {
      const { clearSelection } = useCredentialContextStore.getState();
      
      clearSelection();
      
      const params = new URLSearchParams(window.location.search);
      params.set('workspaceId', workspace.id);
      params.delete('credentialId');
      params.delete('region');
      
      queryClient.invalidateQueries({
        predicate: (query) => {
          const key = query.queryKey;
          if (!Array.isArray(key)) return false;
          
          const keyString = key.join('-').toLowerCase();
          
          const shouldInvalidate = 
            (previousWorkspaceId && key.includes(previousWorkspaceId)) ||
            keyString.includes('vms') ||
            keyString.includes('credentials') ||
            keyString.includes('kubernetes') ||
            keyString.includes('clusters') ||
            keyString.includes('node-pools') ||
            keyString.includes('node-groups') ||
            keyString.includes('nodes') ||
            keyString.includes('vpcs') ||
            keyString.includes('subnets') ||
            keyString.includes('security-groups');
          
          return shouldInvalidate;
        },
      });
      
      if (pathname.startsWith('/workspaces/') && (pathname.includes('/settings') || pathname.includes('/members'))) {
        const pageType = pathname.includes('/settings') ? 'settings' : 'members';
        router.push(`/workspaces/${workspaceId}/${pageType}`);
        return;
      }
      
      const currentPath = pathname;
      router.replace(`${currentPath}?${params.toString()}`, { scroll: false });
    } else {
      const currentPath = pathname;
      const params = new URLSearchParams(window.location.search);
      params.set('workspaceId', workspace.id);
      router.replace(`${currentPath}?${params.toString()}`, { scroll: false });
    }
  };

  const displayWorkspaces = fetchedWorkspaces.length > 0 ? fetchedWorkspaces : workspaces;

  // Determine which accordion items should be open based on current path
  const getDefaultOpenItems = () => {
    const openItems: string[] = [];
    if (pathname.startsWith('/compute')) openItems.push('compute');
    if (pathname.startsWith('/kubernetes')) openItems.push('kubernetes');
    if (pathname.startsWith('/networks')) openItems.push('networks');
    return openItems;
  };

  // Check if a navigation item is active
  const isItemActive = (item: NavigationItem): boolean => {
    if (item.href) {
      return pathname === item.href || pathname.startsWith(item.href + '/');
    }
    if (item.children) {
      return item.children.some(child => isItemActive(child));
    }
    return false;
  };

  // Check if a child item is active
  const isChildActive = (item: NavigationItem): boolean => {
    if (item.href) {
      return pathname === item.href || pathname.startsWith(item.href + '/');
    }
    return false;
  };

  return (
    <nav className="flex h-full w-64 flex-col bg-card border-r lg:block hidden" role="navigation" aria-label="Main navigation">
      <div className="flex flex-col flex-1 overflow-y-auto">
        <div className="flex flex-col flex-1 px-3 py-4">
          {/* Logo */}
          <div className="mb-6 pt-4">
            <Button
              variant="ghost"
              className="w-full justify-start px-2 h-auto hover:bg-transparent"
              onClick={() => router.push('/dashboard')}
              aria-label="Go to dashboard"
            >
              <h1 className="text-lg sm:text-xl font-bold text-foreground">
                SkyClust
                <ScreenReaderOnly>Multi-Cloud Management Platform</ScreenReaderOnly>
              </h1>
            </Button>
          </div>

          {/* Workspace Selector */}
          <div className="mb-4 space-y-2">
            <label className="text-sm font-medium text-muted-foreground">
              <ScreenReaderOnly>{t('workspace.select')}</ScreenReaderOnly>
              {t('workspace.title')}
            </label>
            <Select
              value={currentWorkspace?.id || 'all'}
              onValueChange={handleWorkspaceChange}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder={t('workspace.select')} />
              </SelectTrigger>
              <SelectContent>
                {displayWorkspaces.map((workspace) => (
                  <SelectItem key={workspace.id} value={workspace.id} className="truncate">
                    {workspace.name}
                  </SelectItem>
                ))}
                <div className="h-px bg-border my-1" />
                {currentWorkspace && (
                  <div className="px-2 py-1.5">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start"
                      onClick={(e) => {
                        e.stopPropagation();
                        router.push(`/workspaces/${currentWorkspace.id}/settings`);
                      }}
                      aria-label="Open workspace settings"
                    >
                      <Settings className="mr-2 h-4 w-4" aria-hidden="true" />
                      {t('workspace.settings')}
                    </Button>
                  </div>
                )}
                <div className="px-2 py-1.5 border-t border-border">
                  <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
                    <DialogTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="w-full justify-start"
                        onClick={(e) => {
                          e.stopPropagation();
                        }}
                        aria-label="Create a new workspace"
                      >
                        <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                        {t('workspace.createNew')}
                      </Button>
                        </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>{t('workspace.createNew')}</DialogTitle>
                        <DialogDescription>
                          {t('workspace.createNewDescription')}
                        </DialogDescription>
                      </DialogHeader>
                      <Form {...form}>
                        <form onSubmit={handleSubmit} className="space-y-4">
                          <EnhancedField
                            name="name"
                            label={t('workspace.name')}
                            type="text"
                            placeholder={t('workspace.name')}
                            required
                            getFieldError={getFieldError}
                            getFieldValidationState={getFieldValidationState}
                          />
                          <EnhancedField
                            name="description"
                            label={t('workspace.description')}
                            type="textarea"
                            placeholder={t('workspace.description')}
                            required
                            getFieldError={getFieldError}
                            getFieldValidationState={getFieldValidationState}
                          />
                          
                          {formError && (
                            <div className="text-sm text-red-600 text-center" role="alert">
                              {formError}
                            </div>
                          )}

                          <div className="flex justify-end space-x-2">
                            <Button
                              type="button"
                              variant="outline"
                              onClick={() => {
                                reset();
                                setIsCreateDialogOpen(false);
                              }}
                            >
                              {t('common.cancel')}
                            </Button>
                            <Button type="submit" disabled={isFormLoading}>
                              {isFormLoading ? t('common.loading') : t('workspace.create')}
                            </Button>
                          </div>
                        </form>
                      </Form>
                    </DialogContent>
                  </Dialog>
                </div>
              </SelectContent>
            </Select>
          </div>

          {/* Navigation Menu */}
          <div className="mt-4 flex-1" role="list">
            <Accordion type="multiple" defaultValue={getDefaultOpenItems()} className="w-full">
              {navigation.map((item) => {
                if (item.children) {
                  // Parent item with children (accordion)
                  const isActive = isItemActive(item);
                  return (
                    <AccordionItem key={item.name} value={item.name.toLowerCase()} className="border-none">
                      <AccordionTrigger
                        className={cn(
                          'py-2 px-3 hover:no-underline',
                          isActive && 'bg-accent'
                        )}
                      >
                        <div className="flex items-center w-full">
                          <item.icon className="mr-3 h-5 w-5 flex-shrink-0" aria-hidden="true" />
                          <span className="text-sm font-medium">{item.name}</span>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent className="pb-1 pt-0">
                        <div className="ml-4 space-y-1">
                          {item.children.map((child) => {
                            const isChildItemActive = isChildActive(child);
                            return (
                              <Button
                                key={child.name}
                                onClick={() => router.push(child.href || '#')}
                                variant={isChildItemActive ? 'secondary' : 'ghost'}
                                size="sm"
                                className={cn(
                                  'w-full justify-start text-sm',
                                  isChildItemActive && 'bg-accent font-medium'
                                )}
                                role="listitem"
                                aria-current={isChildItemActive ? 'page' : undefined}
                                aria-label={`Navigate to ${child.name} page`}
                              >
                                <child.icon className="mr-2 h-4 w-4" aria-hidden="true" />
                                {child.name}
                              </Button>
                            );
                          })}
                        </div>
                      </AccordionContent>
                    </AccordionItem>
                  );
                } else {
                  // Single item without children
                  const isActive = pathname === item.href;
                  return (
                    <Button
                      key={item.name}
                      onClick={() => item.href && router.push(item.href)}
                      variant={isActive ? 'secondary' : 'ghost'}
                      className={cn(
                        'w-full justify-start mb-1',
                        isActive && 'bg-accent font-medium'
                      )}
                      role="listitem"
                      aria-current={isActive ? 'page' : undefined}
                      aria-label={`Navigate to ${item.name} page`}
                    >
                      <item.icon className="mr-3 h-5 w-5" aria-hidden="true" />
                      {item.name}
                    </Button>
                  );
                }
              })}
            </Accordion>
          </div>
        </div>
      </div>
    </nav>
  );
}

export const Sidebar = React.memo(SidebarComponent);

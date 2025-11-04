/**
 * Breadcrumb Component
 * 경로 표시를 위한 브레드크럼 컴포넌트
 */

'use client';

import * as React from 'react';
import Link from 'next/link';
import { usePathname, useSearchParams } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { ChevronRight, Home } from 'lucide-react';
import { cn } from '@/lib/utils';

interface BreadcrumbItem {
  label: string;
  href: string;
}

interface BreadcrumbProps {
  items?: BreadcrumbItem[];
  className?: string;
}

export function Breadcrumb({ items, className }: BreadcrumbProps) {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();

  // Get workspace ID from URL or store
  const workspaceId = React.useMemo(() => {
    const urlWorkspaceId = searchParams.get('workspaceId');
    return urlWorkspaceId || currentWorkspace?.id || '';
  }, [searchParams, currentWorkspace?.id]);

  // Map 1 depth paths to default sub-paths
  const defaultPaths: Record<string, string> = {
    '/compute': '/compute/vms',
    '/kubernetes': '/kubernetes/clusters',
    '/networks': '/networks/vpcs',
  };

  // Auto-generate breadcrumb from pathname if items not provided
  const breadcrumbItems = React.useMemo(() => {
    if (items) return items;

    const paths = pathname.split('/').filter(Boolean);
    const result: BreadcrumbItem[] = [
      { label: 'Home', href: '/dashboard' },
    ];

    // Map path segments to labels
    const pathLabels: Record<string, string> = {
      'compute': 'Compute',
      'vms': 'VMs',
      'images': 'Images',
      'snapshots': 'Snapshots',
      'kubernetes': 'Kubernetes',
      'clusters': 'Clusters',
      'node-pools': 'Node Pools',
      'nodes': 'Nodes',
      'networks': 'Networks',
      'vpcs': 'VPCs',
      'subnets': 'Subnets',
      'security-groups': 'Security Groups',
      'credentials': 'Credentials',
      'workspaces': 'Workspaces',
      'settings': 'Settings',
      'members': 'Members',
      'dashboard': 'Dashboard',
      'profile': 'Profile',
      'cost-analysis': 'Cost Analysis',
      'exports': 'Exports',
      'notifications': 'Notifications',
    };

    let currentPath = '';
    const seenHrefs = new Set<string>(); // Track seen hrefs to prevent duplicates
    
    paths.forEach((segment, index) => {
      currentPath += `/${segment}`;
      const label = pathLabels[segment] || segment.charAt(0).toUpperCase() + segment.slice(1);
      
      // Don't add UUIDs or IDs as breadcrumb items
      if (segment.match(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i)) {
        return;
      }

      // Check if this is a 1 depth path that needs redirection
      let href = currentPath;
      if (defaultPaths[currentPath]) {
        href = defaultPaths[currentPath];
      }

      // Add workspaceId query parameter if available
      if (workspaceId && (defaultPaths[currentPath] || currentPath.startsWith('/compute') || currentPath.startsWith('/kubernetes') || currentPath.startsWith('/networks'))) {
        href = `${href}?workspaceId=${workspaceId}`;
      }

      // Skip if this href was already added (prevents duplicate breadcrumbs)
      if (seenHrefs.has(href)) {
        return;
      }

      seenHrefs.add(href);
      result.push({
        label,
        href,
      });
    });

    return result;
  }, [pathname, items, workspaceId]);

  if (breadcrumbItems.length <= 1) {
    return null;
  }

  return (
    <nav
      className={cn('flex items-center space-x-1 text-sm text-muted-foreground', className)}
      aria-label="Breadcrumb"
    >
      <ol className="flex items-center space-x-1">
        {breadcrumbItems.map((item, index) => {
          const isLast = index === breadcrumbItems.length - 1;
          
          return (
            <li key={`${item.href}-${index}`} className="flex items-center">
              {index === 0 ? (
                <Link
                  href={item.href}
                  className="flex items-center hover:text-foreground transition-colors"
                  aria-label="Home"
                >
                  <Home className="h-4 w-4" />
                </Link>
              ) : (
                <>
                  <ChevronRight className="h-4 w-4 mx-1 text-muted-foreground" />
                  {isLast ? (
                    <span className="font-medium text-foreground" aria-current="page">
                      {item.label}
                    </span>
                  ) : (
                    <Link
                      href={item.href}
                      className="hover:text-foreground transition-colors"
                    >
                      {item.label}
                    </Link>
                  )}
                </>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}


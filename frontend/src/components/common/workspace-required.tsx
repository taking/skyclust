'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { Layout } from '@/components/layout/layout';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface WorkspaceRequiredProps {
  children: React.ReactNode;
  allowAutoSelect?: boolean; // Allow waiting for auto-selection (for dashboard)
}

export function WorkspaceRequired({ children, allowAutoSelect = false }: WorkspaceRequiredProps) {
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const pathname = usePathname();
  const [isChecking, setIsChecking] = useState(allowAutoSelect);

  useEffect(() => {
    // If auto-select is allowed, wait a bit before redirecting
    if (allowAutoSelect && !currentWorkspace) {
      const timer = setTimeout(() => {
        setIsChecking(false);
        if (!currentWorkspace && pathname !== '/workspaces') {
          // Use replace to prevent back button issues
          router.replace('/workspaces');
        }
      }, 800); // Wait 800ms for workspace auto-selection to complete
      return () => clearTimeout(timer);
    } else if (!allowAutoSelect && !currentWorkspace) {
      // Immediate redirect for other pages
      if (pathname !== '/workspaces') {
        router.replace('/workspaces');
      }
    } else if (currentWorkspace) {
      setIsChecking(false);
    }
  }, [currentWorkspace, router, pathname, allowAutoSelect]);

  if (isChecking || !currentWorkspace) {
    return (
      <Layout>
        <div className="flex items-center justify-center min-h-screen">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>Workspace Required</CardTitle>
              <CardDescription>
                {isChecking ? 'Loading workspace...' : 'Please select a workspace to continue.'}
              </CardDescription>
            </CardHeader>
            {!isChecking && (
              <CardContent>
                <Button onClick={() => router.push('/workspaces')} className="w-full">
                  Go to Workspaces
                </Button>
              </CardContent>
            )}
          </Card>
        </div>
      </Layout>
    );
  }

  return <>{children}</>;
}


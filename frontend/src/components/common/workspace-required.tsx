'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { Layout } from '@/components/layout/layout';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';

interface WorkspaceRequiredProps {
  children: React.ReactNode;
  allowAutoSelect?: boolean; // Allow waiting for auto-selection (for dashboard)
}

export function WorkspaceRequired({ children, allowAutoSelect = false }: WorkspaceRequiredProps) {
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const pathname = usePathname();
  const { t } = useTranslation();
  const [isChecking, setIsChecking] = useState(allowAutoSelect);
  const [hasRedirected, setHasRedirected] = useState(false);

  useEffect(() => {
    // Prevent multiple redirects
    if (hasRedirected) return;

    // If auto-select is allowed, wait longer for workspace auto-selection
    if (allowAutoSelect && !currentWorkspace) {
      const timer = setTimeout(() => {
        // Double-check workspace is still not set before redirecting
        if (!currentWorkspace && pathname !== '/workspaces' && !hasRedirected) {
          setHasRedirected(true);
          router.replace('/workspaces');
        }
        setIsChecking(false);
      }, 1500); // Wait 1.5s for workspace auto-selection to complete
      return () => clearTimeout(timer);
    } else if (!allowAutoSelect && !currentWorkspace) {
      // Immediate redirect for other pages
      if (pathname !== '/workspaces' && !hasRedirected) {
        setHasRedirected(true);
        router.replace('/workspaces');
      }
    } else if (currentWorkspace) {
      setIsChecking(false);
      setHasRedirected(false); // Reset if workspace is set
    }
  }, [currentWorkspace, router, pathname, allowAutoSelect, hasRedirected]);

  if (isChecking || !currentWorkspace) {
    return (
      <Layout>
        <div className="flex items-center justify-center min-h-screen">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>{t('workspace.title')}</CardTitle>
              <CardDescription>
                {isChecking ? t('components.workspaceRequired.loading') : t('components.workspaceRequired.selectWorkspace')}
              </CardDescription>
            </CardHeader>
            {!isChecking && (
              <CardContent>
                <Button onClick={() => router.push('/workspaces')} className="w-full">
                  {t('workspace.goToWorkspaces')}
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


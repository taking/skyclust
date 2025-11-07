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
  allowAutoSelect?: boolean; // 자동 선택 대기 허용 (대시보드용)
}

export function WorkspaceRequired({ children, allowAutoSelect = false }: WorkspaceRequiredProps) {
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();
  const pathname = usePathname();
  const { t } = useTranslation();
  const [isChecking, setIsChecking] = useState(allowAutoSelect);
  const [hasRedirected, setHasRedirected] = useState(false);

  useEffect(() => {
    // 여러 번 리다이렉트 방지
    if (hasRedirected) return;

    // 자동 선택이 허용된 경우 워크스페이스 자동 선택을 더 오래 대기
    if (allowAutoSelect && !currentWorkspace) {
      const timer = setTimeout(() => {
        // 리다이렉트 전 워크스페이스가 여전히 설정되지 않았는지 재확인
        if (!currentWorkspace && pathname !== '/workspaces' && !hasRedirected) {
          setHasRedirected(true);
          router.replace('/workspaces');
        }
        setIsChecking(false);
      }, 1500); // 워크스페이스 자동 선택 완료를 위해 1.5초 대기
      return () => clearTimeout(timer);
    } else if (!allowAutoSelect && !currentWorkspace) {
      // 다른 페이지는 즉시 리다이렉트
      if (pathname !== '/workspaces' && !hasRedirected) {
        setHasRedirected(true);
        router.replace('/workspaces');
      }
    } else if (currentWorkspace) {
      setIsChecking(false);
      setHasRedirected(false); // 워크스페이스가 설정되면 리셋
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


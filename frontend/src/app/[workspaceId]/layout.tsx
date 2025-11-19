/**
 * Workspace Layout
 * 워크스페이스 컨텍스트를 제공하는 레이아웃
 * 
 * 모든 워크스페이스 관련 페이지는 이 레이아웃을 사용합니다.
 * - 리소스 페이지: /{workspaceId}/{credentialId}/...
 * - 관리 페이지: /{workspaceId}/...
 */

'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Layout } from '@/components/layout/layout';
import { useWorkspaceStore } from '@/store/workspace';
import { useQuery } from '@tanstack/react-query';
import { workspaceService } from '@/features/workspaces';
import { queryKeys } from '@/lib/query';
import { API } from '@/lib/constants';
import { Spinner } from '@/components/ui/spinner';
import { AppErrorBoundary } from '@/components/error-boundary';
import { log } from '@/lib/logging';

export default function WorkspaceLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const params = useParams();
  const router = useRouter();
  const workspaceId = params.workspaceId as string;
  const { currentWorkspace, setCurrentWorkspace, workspaces, setWorkspaces } = useWorkspaceStore();

  // 워크스페이스 목록 조회
  const { data: fetchedWorkspaces = [], isLoading: isLoadingWorkspaces } = useQuery({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    retry: API.REQUEST.MAX_RETRIES,
    retryDelay: API.REQUEST.RETRY_DELAY,
  });

  // 워크스페이스 목록 업데이트
  useEffect(() => {
    if (!isLoadingWorkspaces && fetchedWorkspaces.length > 0) {
      setWorkspaces(fetchedWorkspaces);
    }
  }, [fetchedWorkspaces, isLoadingWorkspaces, setWorkspaces]);

  // URL의 workspaceId와 현재 선택된 workspaceId 일치 확인
  useEffect(() => {
    if (!workspaceId || isLoadingWorkspaces) return;

    // URL의 workspaceId가 현재 선택된 workspace와 다르면 업데이트
    if (currentWorkspace?.id !== workspaceId) {
      const workspace = fetchedWorkspaces.find((w) => w.id === workspaceId);
      if (workspace) {
        setCurrentWorkspace(workspace);
      } else if (fetchedWorkspaces.length > 0) {
        // URL의 workspaceId가 유효하지 않으면 첫 번째 워크스페이스로 리다이렉트
        const firstWorkspace = fetchedWorkspaces[0];
        setCurrentWorkspace(firstWorkspace);
        // 현재 경로에서 workspaceId만 교체
        const currentPath = window.location.pathname;
        const newPath = currentPath.replace(`/${workspaceId}`, `/${firstWorkspace.id}`);
        router.replace(newPath);
      }
    }
  }, [workspaceId, currentWorkspace, fetchedWorkspaces, isLoadingWorkspaces, setCurrentWorkspace, router]);

  // 워크스페이스가 없으면 로딩 표시
  if (isLoadingWorkspaces || !workspaceId) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-screen">
          <div className="text-center">
            <Spinner size="lg" label="Loading workspace..." />
          </div>
        </div>
      </Layout>
    );
  }

  // 워크스페이스가 유효하지 않으면 에러 표시
  if (fetchedWorkspaces.length > 0 && !fetchedWorkspaces.find((w) => w.id === workspaceId)) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-screen">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Workspace Not Found</h2>
            <p className="text-gray-600 mb-4">The workspace you're looking for doesn't exist.</p>
            <button
              onClick={() => router.push(`/${fetchedWorkspaces[0].id}/workspaces`)}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Go to Workspaces
            </button>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <AppErrorBoundary
      resetKeys={[workspaceId]}
      onError={(error, errorInfo) => {
        // Workspace 레벨 에러 로깅
        log.error('Workspace layout error', error, {
          service: 'Layout',
          action: 'workspaceLayout',
          workspaceId,
          componentStack: errorInfo.componentStack,
        });
      }}
    >
      {children}
    </AppErrorBoundary>
  );
}


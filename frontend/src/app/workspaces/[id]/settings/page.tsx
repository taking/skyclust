/**
 * Legacy Route Redirect
 * 기존 라우트를 새로운 구조로 리다이렉트
 * /workspaces/{id}/settings -> /{workspaceId}/workspaces/{id}/settings
 */

'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { buildManagementPath } from '@/lib/routing/helpers';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';

export default function WorkspaceSettingsRedirectPage() {
  const params = useParams();
  const router = useRouter();
  const { currentWorkspace } = useWorkspaceStore();
  const workspaceId = params.id as string;

  useEffect(() => {
    // URL의 workspaceId와 현재 선택된 workspaceId 중 하나를 사용
    const targetWorkspaceId = currentWorkspace?.id || workspaceId;
    if (targetWorkspaceId) {
      const newPath = buildManagementPath(targetWorkspaceId, `workspaces/${workspaceId}/settings`);
      router.replace(newPath);
    } else {
      router.replace('/workspaces');
    }
  }, [currentWorkspace, workspaceId, router]);

  return (
    <Layout>
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <Spinner size="lg" label="Redirecting..." />
        </div>
      </div>
    </Layout>
  );
}

/**
 * Legacy Route Redirect
 * 기존 라우트를 새로운 구조로 리다이렉트
 * /workspaces/{id}/members -> /{workspaceId}/workspaces/{id}/members
 */

'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { buildManagementPath } from '@/lib/routing/helpers';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';

export default function WorkspaceMembersRedirectPage() {
  const params = useParams();
  const router = useRouter();
  const { currentWorkspace } = useWorkspaceStore();
  const workspaceId = params.id as string;

  useEffect(() => {
    const targetWorkspaceId = currentWorkspace?.id || workspaceId;
    if (targetWorkspaceId) {
      const newPath = buildManagementPath(targetWorkspaceId, `workspaces/${workspaceId}/members`);
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

/**
 * Workspace Credentials Page
 * 워크스페이스 Credentials 관리 페이지
 * 
 * 새로운 라우팅 구조: /w/{workspaceId}/workspaces/{id}/credentials
 * 
 * 이 페이지는 Workspace 상세 페이지의 탭으로 통합되었지만,
 * 직접 접근을 위해 별도 페이지로도 제공됩니다.
 */

'use client';

import * as React from 'react';
import { Suspense } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useResourceContext } from '@/hooks/use-resource-context';
import { buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { Layout } from '@/components/layout/layout';
import { Spinner } from '@/components/ui/spinner';

function WorkspaceCredentialsPageContent() {
  const params = useParams();
  const router = useRouter();
  const { workspaceId: pathWorkspaceId } = useResourceContext();
  const workspaceId = params.id as string;

  // 최종 workspaceId 결정: path parameter 우선
  const finalWorkspaceId = pathWorkspaceId || workspaceId;

  // Workspace 상세 페이지의 credentials 탭으로 리다이렉트
  React.useEffect(() => {
    if (finalWorkspaceId && workspaceId) {
      router.replace(`${buildWorkspaceResourcePath(finalWorkspaceId, workspaceId, 'overview')}?tab=credentials`);
    }
  }, [finalWorkspaceId, workspaceId, router]);

  return (
    <Layout>
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" label="Redirecting..." />
      </div>
    </Layout>
  );
}

export default function WorkspaceCredentialsPage() {
  return (
    <Suspense fallback={
      <Layout>
        <div className="flex items-center justify-center h-64">
          <Spinner size="lg" label="Loading..." />
        </div>
      </Layout>
    }>
      <WorkspaceCredentialsPageContent />
    </Suspense>
  );
}


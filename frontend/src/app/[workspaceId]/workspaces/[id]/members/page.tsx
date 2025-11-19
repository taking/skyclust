/**
 * Workspace Members Page Redirect
 * 워크스페이스 멤버 관리 페이지 리다이렉트
 * 
 * 기존 라우팅 구조: /{workspaceId}/workspaces/{id}/members
 * 새로운 라우팅 구조: /w/{workspaceId}/workspaces/{id}?tab=members
 * 
 * Workspace 상세 페이지의 Members 탭으로 리다이렉트
 */

'use client';

import * as React from 'react';
import { Suspense } from 'react';
import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useResourceContext } from '@/hooks/use-resource-context';
import { buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { Layout } from '@/components/layout/layout';
import { Spinner } from '@/components/ui/spinner';

function WorkspaceMembersPageContent() {
  const params = useParams();
  const router = useRouter();
  const { workspaceId: pathWorkspaceId } = useResourceContext();
  const workspaceId = params.id as string;

  // 최종 workspaceId 결정: path parameter 우선
  const finalWorkspaceId = pathWorkspaceId || workspaceId;

  // Workspace 상세 페이지의 members 탭으로 리다이렉트
  useEffect(() => {
    if (finalWorkspaceId && workspaceId) {
      router.replace(`${buildWorkspaceResourcePath(finalWorkspaceId, workspaceId, 'overview')}?tab=members`);
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

export default function WorkspaceMembersPage() {
  return (
    <Suspense fallback={
      <Layout>
        <div className="flex items-center justify-center h-64">
          <Spinner size="lg" label="Loading..." />
        </div>
      </Layout>
    }>
      <WorkspaceMembersPageContent />
    </Suspense>
  );
}


/**
 * Credentials Page Redirect
 * Credentials 관리 페이지 리다이렉트
 * 
 * 기존 라우팅 구조: /{workspaceId}/credentials
 * 새로운 라우팅 구조: /w/{workspaceId}/workspaces/{workspaceId}/credentials
 * 
 * Workspace 상세 페이지의 Credentials 탭으로 리다이렉트
 */

'use client';

import * as React from 'react';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useResourceContext } from '@/hooks/use-resource-context';
import { buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { Layout } from '@/components/layout/layout';
import { Spinner } from '@/components/ui/spinner';

export default function CredentialsPage() {
  const router = useRouter();
  const { workspaceId } = useResourceContext();

  useEffect(() => {
    if (workspaceId) {
      // Workspace 상세 페이지의 credentials 탭으로 리다이렉트
      router.replace(`${buildWorkspaceResourcePath(workspaceId, workspaceId, 'overview')}?tab=credentials`);
    } else {
      // Workspace가 없으면 workspaces 페이지로
      router.replace('/workspaces');
    }
  }, [workspaceId, router]);

  return (
    <Layout>
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" label="Redirecting..." />
      </div>
    </Layout>
  );
}


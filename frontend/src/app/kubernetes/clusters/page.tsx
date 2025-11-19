/**
 * Legacy Route Redirect
 * 기존 라우트를 새로운 구조로 리다이렉트
 * /kubernetes/clusters -> /{workspaceId}/{credentialId}/kubernetes/clusters
 */

'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { buildResourcePath } from '@/lib/routing/helpers';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';

export default function KubernetesClustersRedirectPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion } = useCredentialContextStore.getState();

  useEffect(() => {
    // Workspace와 Credential이 있으면 새 경로로 리다이렉트
    if (currentWorkspace?.id && selectedCredentialId) {
      const filters: Record<string, string | undefined> = {};
      if (selectedRegion) filters.region = selectedRegion;
      
      // Query parameter에서 추가 필터 가져오기
      const region = searchParams.get('region');
      if (region) filters.region = region;

      const newPath = buildResourcePath(
        currentWorkspace.id,
        selectedCredentialId,
        'kubernetes',
        '/clusters',
        filters
      );
      router.replace(newPath);
    } else if (currentWorkspace?.id) {
      // Workspace만 있으면 credentials 페이지로
      router.replace(`/${currentWorkspace.id}/credentials`);
    } else {
      // Workspace가 없으면 workspaces 페이지로
      router.replace('/workspaces');
    }
  }, [currentWorkspace, selectedCredentialId, router, searchParams, selectedRegion]);

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

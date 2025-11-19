/**
 * Legacy Route Redirect
 * 기존 라우트를 새로운 구조로 리다이렉트
 * /kubernetes/node-groups/{name} -> /{workspaceId}/{credentialId}/kubernetes/node-groups/{name}
 */

'use client';

import { useEffect } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { buildResourceDetailPath } from '@/lib/routing/helpers';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';

export default function NodeGroupDetailRedirectPage() {
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion } = useCredentialContextStore.getState();

  const nodeGroupName = params.name as string;

  useEffect(() => {
    if (currentWorkspace?.id && selectedCredentialId && nodeGroupName) {
      const filters: Record<string, string | undefined> = {};
      if (selectedRegion) filters.region = selectedRegion;
      
      const region = searchParams.get('region');
      if (region) filters.region = region;

      const newPath = buildResourceDetailPath(
        currentWorkspace.id,
        selectedCredentialId,
        'kubernetes',
        'node-groups',
        nodeGroupName,
        filters
      );
      router.replace(newPath);
    } else if (currentWorkspace?.id && selectedCredentialId) {
      const path = buildResourceDetailPath(
        currentWorkspace.id,
        selectedCredentialId,
        'kubernetes',
        'node-groups',
        nodeGroupName
      );
      router.replace(path);
    } else if (currentWorkspace?.id) {
      router.replace(`/${currentWorkspace.id}/credentials`);
    } else {
      router.replace('/workspaces');
    }
  }, [currentWorkspace, selectedCredentialId, nodeGroupName, router, searchParams, selectedRegion]);

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

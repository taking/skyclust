/**
 * Legacy Route Redirect
 * 기존 라우트를 새로운 구조로 리다이렉트
 * /dashboard -> /{workspaceId}/dashboard
 */

'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { buildManagementPath } from '@/lib/routing/helpers';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';

export default function DashboardRedirectPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion } = useCredentialContextStore.getState();

  useEffect(() => {
    if (currentWorkspace?.id) {
      const filters: Record<string, string | undefined> = {};
      if (selectedCredentialId) filters.credentialId = selectedCredentialId;
      if (selectedRegion) filters.region = selectedRegion;
      
      const credentialId = searchParams.get('credentialId');
      const region = searchParams.get('region');
      if (credentialId) filters.credentialId = credentialId;
      if (region) filters.region = region;

      const newPath = buildManagementPath(currentWorkspace.id, 'dashboard', filters);
      router.replace(newPath);
    } else {
      router.replace('/workspaces');
    }
  }, [currentWorkspace, router, searchParams, selectedCredentialId, selectedRegion]);

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

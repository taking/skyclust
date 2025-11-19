/**
 * Legacy Route Redirect
 * 기존 라우트를 새로운 구조로 리다이렉트
 * /networks/subnets/create -> /{workspaceId}/{credentialId}/networks/subnets/create
 */

'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { buildResourceCreatePath } from '@/lib/routing/helpers';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';

export default function CreateSubnetRedirectPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId, selectedRegion } = useCredentialContextStore.getState();

  useEffect(() => {
    if (currentWorkspace?.id && selectedCredentialId) {
      const filters: Record<string, string | undefined> = {};
      if (selectedRegion) filters.region = selectedRegion;
      
      const region = searchParams.get('region');
      const vpcId = searchParams.get('vpc_id');
      if (region) filters.region = region;
      if (vpcId) filters.vpc_id = vpcId;

      const newPath = buildResourceCreatePath(
        currentWorkspace.id,
        selectedCredentialId,
        'networks',
        'subnets',
        filters
      );
      router.replace(newPath);
    } else if (currentWorkspace?.id) {
      router.replace(`/${currentWorkspace.id}/credentials`);
    } else {
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

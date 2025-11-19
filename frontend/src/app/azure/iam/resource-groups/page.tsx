/**
 * Legacy Route Redirect
 * 기존 라우트를 새로운 구조로 리다이렉트
 * /azure/iam/resource-groups -> /{workspaceId}/{credentialId}/azure/iam/resource-groups
 */

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentialContextStore } from '@/store/credential-context';
import { buildResourcePath } from '@/lib/routing/helpers';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';

export default function ResourceGroupsRedirectPage() {
  const router = useRouter();
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedCredentialId } = useCredentialContextStore.getState();

  useEffect(() => {
    if (currentWorkspace?.id && selectedCredentialId) {
      const newPath = buildResourcePath(
        currentWorkspace.id,
        selectedCredentialId,
        'azure',
        '/iam/resource-groups'
      );
      router.replace(newPath);
    } else if (currentWorkspace?.id) {
      router.replace(`/${currentWorkspace.id}/credentials`);
    } else {
      router.replace('/workspaces');
    }
  }, [currentWorkspace, selectedCredentialId, router]);

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

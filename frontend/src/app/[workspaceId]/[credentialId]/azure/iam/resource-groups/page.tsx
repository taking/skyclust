/**
 * Azure Resource Groups Page
 * Azure Resource Groups 관리 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/azure/iam/resource-groups
 */

'use client';

import { Suspense } from 'react';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildResourceCreatePath } from '@/lib/routing/helpers';
import { ResourceGroupsPageContent } from '@/features/resource-groups';

function ResourceGroupsPage() {
  const { workspaceId, credentialId } = useRequiredResourceContext();

  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <ResourceGroupsPageContent
        workspaceId={workspaceId}
        credentialId={credentialId}
        onCreateClick={() => {
          if (workspaceId && credentialId) {
            const path = buildResourceCreatePath(
              workspaceId,
              credentialId,
              'azure',
              '/iam/resource-groups'
            );
            // router.push will be handled by the component
          }
        }}
      />
    </Suspense>
  );
}

export default ResourceGroupsPage;


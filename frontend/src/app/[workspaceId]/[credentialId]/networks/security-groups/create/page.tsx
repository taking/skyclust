/**
 * Create Security Group Page
 * Security Group 생성 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/networks/security-groups/create
 */

'use client';

import { Suspense } from 'react';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildResourcePath } from '@/lib/routing/helpers';
import { CreateSecurityGroupPageContent } from '@/features/networks';

function CreateSecurityGroupPage() {
  const { workspaceId, credentialId, region } = useRequiredResourceContext();

  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <CreateSecurityGroupPageContent
        workspaceId={workspaceId}
        credentialId={credentialId}
        region={region}
        onCancel={() => {
          if (workspaceId && credentialId) {
            const path = buildResourcePath(
              workspaceId,
              credentialId,
              'networks',
              '/security-groups',
              { region: region || undefined }
            );
            // router.push will be handled by the component
          }
        }}
      />
    </Suspense>
  );
}

export default CreateSecurityGroupPage;


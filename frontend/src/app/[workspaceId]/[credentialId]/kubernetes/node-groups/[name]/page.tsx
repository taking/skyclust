/**
 * Kubernetes Node Group Detail Page
 * Kubernetes Node Group 상세 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/kubernetes/node-groups/{name}
 */

'use client';

import { Suspense } from 'react';
import { useParams } from 'next/navigation';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildResourcePath } from '@/lib/routing/helpers';
import { NodeGroupDetailPageContent } from '@/features/kubernetes';

function NodeGroupDetailPage() {
  const params = useParams();
  const { workspaceId, credentialId, region } = useRequiredResourceContext();
  const nodeGroupName = params.name as string;

  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <NodeGroupDetailPageContent
        nodeGroupName={nodeGroupName}
        workspaceId={workspaceId}
        credentialId={credentialId}
        region={region}
        onBack={() => {
          if (workspaceId && credentialId) {
            const path = buildResourcePath(
              workspaceId,
              credentialId,
              'kubernetes',
              '/node-groups',
              { region: region || undefined }
            );
            // router.push will be handled by the component
          }
        }}
      />
    </Suspense>
  );
}

export default NodeGroupDetailPage;


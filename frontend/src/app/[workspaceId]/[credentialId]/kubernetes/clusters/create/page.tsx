/**
 * Create Kubernetes Cluster Page
 * Kubernetes 클러스터 생성 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/kubernetes/clusters/create
 */

'use client';

import { Suspense } from 'react';
import { useRequiredResourceContext } from '@/hooks/use-resource-context';
import { buildCredentialResourcePath } from '@/lib/routing/helpers';
import { CreateClusterPageContent } from '@/features/kubernetes';

function CreateClusterPage() {
  const { workspaceId, credentialId, region } = useRequiredResourceContext();

  if (!workspaceId || !credentialId) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <p className="text-gray-600">Workspace and Credential are required</p>
        </div>
      </div>
    );
  }

  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
          <p className="mt-2 text-gray-600">Loading...</p>
        </div>
      </div>
    }>
      <CreateClusterPageContent
        workspaceId={workspaceId}
        credentialId={credentialId}
        region={region}
        onCancel={() => {
          if (workspaceId && credentialId) {
            const path = buildCredentialResourcePath(
              workspaceId,
              credentialId,
              'k8s',
              '/clusters',
              { region: region || undefined }
            );
            // router.push will be handled by the component
          }
        }}
      />
    </Suspense>
  );
}

export default CreateClusterPage;


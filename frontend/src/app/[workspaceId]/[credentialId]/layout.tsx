/**
 * Credential Layout
 * Credential 컨텍스트를 제공하는 레이아웃
 * 
 * 리소스 페이지 전용 레이아웃입니다.
 * - 리소스 페이지: /{workspaceId}/{credentialId}/kubernetes/...
 * - 리소스 페이지: /{workspaceId}/{credentialId}/networks/...
 * - 리소스 페이지: /{workspaceId}/{credentialId}/compute/...
 * - 리소스 페이지: /{workspaceId}/{credentialId}/azure/...
 */

'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useCredentialContextStore } from '@/store/credential-context';
import { useQuery } from '@tanstack/react-query';
import { useCredentials } from '@/hooks/use-credentials';
import { queryKeys } from '@/lib/query';
import { API } from '@/lib/constants';
import { Spinner } from '@/components/ui/spinner';
import { Layout } from '@/components/layout/layout';
import { AppErrorBoundary } from '@/components/error-boundary';
import { log } from '@/lib/logging';

export default function CredentialLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const params = useParams();
  const router = useRouter();
  const workspaceId = params.workspaceId as string;
  const credentialId = params.credentialId as string;
  const { setSelectedCredential, setSelectedRegion } = useCredentialContextStore();

  // Credential 목록 조회
  const { credentials, isLoading: isLoadingCredentials } = useCredentials({
    workspaceId,
    selectedCredentialId: credentialId,
    enabled: !!workspaceId,
  });

  // URL의 credentialId와 현재 선택된 credentialId 일치 확인
  useEffect(() => {
    if (!credentialId || isLoadingCredentials) return;

    // URL의 credentialId가 유효한지 확인
    const credential = credentials.find((c) => c.id === credentialId);
    if (credential) {
      // Credential Context 업데이트
      if (credentialId !== useCredentialContextStore.getState().selectedCredentialId) {
        setSelectedCredential(credentialId);
      }
    } else if (credentials.length > 0) {
      // URL의 credentialId가 유효하지 않으면 첫 번째 credential로 리다이렉트
      const firstCredential = credentials[0];
      setSelectedCredential(firstCredential.id);
      // 현재 경로에서 credentialId만 교체
      const currentPath = window.location.pathname;
      const newPath = currentPath.replace(`/${credentialId}`, `/${firstCredential.id}`);
      router.replace(newPath);
    }
  }, [credentialId, credentials, isLoadingCredentials, setSelectedCredential, router]);

  // Credential이 없으면 로딩 표시
  if (isLoadingCredentials || !credentialId) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-screen">
          <div className="text-center">
            <Spinner size="lg" label="Loading credential..." />
          </div>
        </div>
      </Layout>
    );
  }

  // Credential이 유효하지 않으면 에러 표시
  if (credentials.length > 0 && !credentials.find((c) => c.id === credentialId)) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-screen">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Credential Not Found</h2>
            <p className="text-gray-600 mb-4">The credential you're looking for doesn't exist.</p>
            <button
              onClick={() => router.push(`/${workspaceId}/credentials`)}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Go to Credentials
            </button>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <AppErrorBoundary
      resetKeys={[workspaceId, credentialId]}
      onError={(error, errorInfo) => {
        // Credential 레벨 에러 로깅
        log.error('Credential layout error', error, {
          service: 'Layout',
          action: 'credentialLayout',
          workspaceId,
          credentialId,
          componentStack: errorInfo.componentStack,
        });
      }}
    >
      {children}
    </AppErrorBoundary>
  );
}


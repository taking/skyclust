/**
 * useCredentialContext Hook
 * Credential과 Region 선택 상태를 URL과 동기화하는 hook
 */

import { useEffect } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { useCredentialContextStore } from '@/store/credential-context';

/**
 * URL 쿼리 파라미터와 Credential/Region 상태를 동기화하는 hook
 */
export function useCredentialContext() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { 
    selectedCredentialId, 
    selectedRegion, 
    setSelectedCredential, 
    setSelectedRegion 
  } = useCredentialContextStore();

  // Check if we should sync for this path
  const shouldSync = pathname.startsWith('/compute') || 
                     pathname.startsWith('/kubernetes') || 
                     pathname.startsWith('/networks');

  // Sync from URL to store on mount and when URL changes (read-only)
  // URL 업데이트는 Header 컴포넌트에서만 처리합니다
  useEffect(() => {
    if (!shouldSync) return;

    const urlCredentialId = searchParams.get('credentialId');
    const urlRegion = searchParams.get('region');

    // Sync credential from URL to store only
    if (urlCredentialId && urlCredentialId !== selectedCredentialId) {
      setSelectedCredential(urlCredentialId);
    } else if (!urlCredentialId && selectedCredentialId) {
      // URL에 credentialId가 없으면 store도 초기화 (workspace 변경 시)
      setSelectedCredential(null);
    }

    // Sync region from URL to store only
    if (urlRegion !== null && urlRegion !== selectedRegion) {
      setSelectedRegion(urlRegion || null);
    } else if (urlRegion === null && selectedRegion) {
      // URL에 region이 없으면 store도 초기화 (workspace 변경 시)
      setSelectedRegion(null);
    }
  }, [searchParams, shouldSync, selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion]);

  return {
    selectedCredentialId,
    selectedRegion,
    setSelectedCredential,
    setSelectedRegion,
  };
}


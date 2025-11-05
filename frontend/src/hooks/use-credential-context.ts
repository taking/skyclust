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
  // Store → URL 동기화는 Header에서 처리하므로, 여기서는 URL → Store만 처리
  useEffect(() => {
    if (!shouldSync) return;

    const urlCredentialId = searchParams.get('credentialId');
    const urlRegion = searchParams.get('region');

    // Sync credential from URL to store only
    if (urlCredentialId && urlCredentialId !== selectedCredentialId) {
      setSelectedCredential(urlCredentialId);
    }
    // URL이 없을 때 store를 초기화하지 않음 (Header에서 Store → URL 동기화 처리)

    // Sync region from URL to store only
    if (urlRegion !== null && urlRegion !== selectedRegion) {
      setSelectedRegion(urlRegion || null);
    }
    // URL이 없을 때 store를 초기화하지 않음 (Header에서 Store → URL 동기화 처리)
  }, [searchParams, shouldSync, selectedCredentialId, selectedRegion, setSelectedCredential, setSelectedRegion]);

  return {
    selectedCredentialId,
    selectedRegion,
    setSelectedCredential,
    setSelectedRegion,
  };
}


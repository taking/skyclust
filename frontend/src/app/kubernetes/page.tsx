/**
 * Redirect from old /kubernetes to new /kubernetes/clusters
 */

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function KubernetesRedirectPage() {
  const router = useRouter();
  
  useEffect(() => {
    router.replace('/kubernetes/clusters');
  }, [router]);
  
  return null;
}

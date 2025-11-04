/**
 * Redirect from old /networks to new /networks/vpcs
 */

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function NetworksRedirectPage() {
  const router = useRouter();
  
  useEffect(() => {
    router.replace('/networks/vpcs');
  }, [router]);
  
  return null;
}

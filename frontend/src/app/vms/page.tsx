/**
 * Redirect from old /vms to new /compute/vms
 */

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function VMsRedirectPage() {
  const router = useRouter();
  
  useEffect(() => {
    router.replace('/compute/vms');
  }, [router]);
  
  return null;
}

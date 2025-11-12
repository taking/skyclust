/**
 * Kubernetes Cluster Detail Page Redirect
 * 기존 라우팅 구조에서 새로운 구조로 리다이렉트
 */

'use client';

import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';

export default function KubernetesClusterDetailRedirectPage() {
  const params = useParams();
  const router = useRouter();
  const clusterName = params.name as string;

  useEffect(() => {
    if (clusterName) {
      router.replace(`/kubernetes/clusters/${clusterName}`);
    }
  }, [clusterName, router]);

  return null;
}


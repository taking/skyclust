/**
 * Cluster Detail Header Component
 * 클러스터 상세 페이지 헤더 컴포넌트
 */

'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Download, ArrowUp } from 'lucide-react';

interface ClusterHeaderProps {
  clusterName: string;
  onUpgradeClick: () => void;
  onDownloadKubeconfigClick: () => void;
  isUpgradeDisabled: boolean;
  isDownloadDisabled: boolean;
  isDownloadPending: boolean;
}

export function ClusterHeader({
  clusterName,
  onUpgradeClick,
  onDownloadKubeconfigClick,
  isUpgradeDisabled,
  isDownloadDisabled,
  isDownloadPending,
}: ClusterHeaderProps) {
  const router = useRouter();

  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center space-x-4">
        <Button variant="ghost" onClick={() => router.push('/kubernetes')}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back
        </Button>
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{clusterName}</h1>
          <p className="text-gray-600">Kubernetes cluster details</p>
        </div>
      </div>
      <div className="flex items-center space-x-2">
        <Button
          variant="outline"
          onClick={onUpgradeClick}
          disabled={isUpgradeDisabled}
        >
          <ArrowUp className="mr-2 h-4 w-4" />
          Upgrade Cluster
        </Button>
        <Button
          variant="outline"
          onClick={onDownloadKubeconfigClick}
          disabled={isDownloadDisabled || isDownloadPending}
        >
          <Download className="mr-2 h-4 w-4" />
          {isDownloadPending ? 'Downloading...' : 'Download Kubeconfig'}
        </Button>
      </div>
    </div>
  );
}


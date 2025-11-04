/**
 * Cluster Upgrade Status Card Component
 * 클러스터 업그레이드 상태 카드 컴포넌트
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import { ArrowUp, AlertTriangle, CheckCircle } from 'lucide-react';

interface UpgradeStatus {
  status?: string;
  current_version?: string;
  target_version?: string;
  progress?: number;
  error?: string;
}

interface ClusterUpgradeStatusCardProps {
  upgradeStatus: UpgradeStatus | undefined;
  currentClusterVersion?: string;
}

export function ClusterUpgradeStatusCard({
  upgradeStatus,
  currentClusterVersion,
}: ClusterUpgradeStatusCardProps) {
  if (!upgradeStatus) {
    return null;
  }

  return (
    <Card className={
      upgradeStatus.status === 'FAILED' 
        ? 'border-red-500' 
        : upgradeStatus.status === 'COMPLETED' 
        ? 'border-green-500' 
        : ''
    }>
      <CardHeader>
        <CardTitle className="flex items-center">
          <ArrowUp className="mr-2 h-5 w-5" />
          Upgrade Status
          <Badge
            variant={
              upgradeStatus.status === 'COMPLETED'
                ? 'default'
                : upgradeStatus.status === 'IN_PROGRESS' || upgradeStatus.status === 'PENDING'
                ? 'secondary'
                : upgradeStatus.status === 'FAILED'
                ? 'destructive'
                : 'outline'
            }
            className="ml-2"
          >
            {upgradeStatus.status}
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label className="text-sm text-gray-500">Current Version</Label>
              <p className="text-sm font-medium">{upgradeStatus.current_version || currentClusterVersion || '-'}</p>
            </div>
            <div>
              <Label className="text-sm text-gray-500">Target Version</Label>
              <p className="text-sm font-medium">{upgradeStatus.target_version || '-'}</p>
            </div>
          </div>
          {upgradeStatus.progress !== undefined && (
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Progress</span>
                <span>{upgradeStatus.progress}%</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${upgradeStatus.progress}%` }}
                />
              </div>
            </div>
          )}
          {upgradeStatus.error && (
            <div className="flex items-start space-x-2 p-3 bg-red-50 rounded-md">
              <AlertTriangle className="h-5 w-5 text-red-600 mt-0.5" />
              <div className="flex-1">
                <p className="text-sm font-medium text-red-900">Upgrade Error</p>
                <p className="text-sm text-red-700">{upgradeStatus.error}</p>
              </div>
            </div>
          )}
          {upgradeStatus.status === 'COMPLETED' && (
            <div className="flex items-center space-x-2 p-3 bg-green-50 rounded-md">
              <CheckCircle className="h-5 w-5 text-green-600" />
              <p className="text-sm font-medium text-green-900">Upgrade completed successfully</p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}


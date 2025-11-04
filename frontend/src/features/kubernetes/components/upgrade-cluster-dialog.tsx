/**
 * Upgrade Cluster Dialog Component
 * 클러스터 업그레이드 다이얼로그 컴포넌트
 */

'use client';

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { AlertTriangle } from 'lucide-react';

interface UpgradeClusterDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clusterName: string;
  currentVersion: string | undefined;
  upgradeVersion: string;
  onUpgradeVersionChange: (version: string) => void;
  onUpgrade: (version: string) => void;
  isPending: boolean;
}

export function UpgradeClusterDialog({
  open,
  onOpenChange,
  clusterName,
  currentVersion,
  upgradeVersion,
  onUpgradeVersionChange,
  onUpgrade,
  isPending,
}: UpgradeClusterDialogProps) {
  const handleSubmit = () => {
    if (upgradeVersion) {
      onUpgrade(upgradeVersion);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Upgrade Cluster</DialogTitle>
          <DialogDescription>
            Upgrade {clusterName} to a new Kubernetes version
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="current-version">Current Version</Label>
            <Input
              id="current-version"
              value={currentVersion || ''}
              disabled
              className="bg-gray-50"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="target-version">Target Version *</Label>
            <Input
              id="target-version"
              value={upgradeVersion}
              onChange={(e) => onUpgradeVersionChange(e.target.value)}
              placeholder="e.g., 1.29.0"
              onKeyDown={(e) => {
                if (e.key === 'Enter' && upgradeVersion && !isPending) {
                  handleSubmit();
                }
              }}
            />
            <p className="text-xs text-gray-500">
              Enter the Kubernetes version to upgrade to (e.g., 1.29.0)
            </p>
          </div>
          <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-md">
            <div className="flex items-start space-x-2">
              <AlertTriangle className="h-5 w-5 text-yellow-600 mt-0.5" />
              <div className="flex-1">
                <p className="text-sm font-medium text-yellow-900">Important</p>
                <ul className="mt-1 text-sm text-yellow-800 list-disc list-inside space-y-1">
                  <li>Upgrading a cluster will cause temporary downtime</li>
                  <li>Ensure all node pools are compatible with the target version</li>
                  <li>Backup your workloads before upgrading</li>
                  <li>Upgrade process cannot be easily rolled back</li>
                </ul>
              </div>
            </div>
          </div>
          <div className="flex justify-end space-x-2">
            <Button
              variant="outline"
              onClick={() => {
                onOpenChange(false);
                onUpgradeVersionChange('');
              }}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={!upgradeVersion || isPending}
            >
              {isPending ? 'Upgrading...' : 'Upgrade Cluster'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}


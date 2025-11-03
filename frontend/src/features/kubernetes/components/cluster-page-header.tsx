/**
 * Cluster Page Header Component
 * Kubernetes 클러스터 페이지 헤더
 */

'use client';

import * as React from 'react';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogTrigger } from '@/components/ui/dialog';
import { Plus } from 'lucide-react';
import type { Credential, CloudProvider, CreateClusterForm } from '@/lib/types';

// Dynamic import for CreateClusterDialog
const CreateClusterDialog = dynamic(
  () => import('./create-cluster-dialog').then(mod => ({ default: mod.CreateClusterDialog })),
  { 
    ssr: false,
    loading: () => null, // Dialog is hidden by default, so no loading state needed
  }
);

interface ClusterPageHeaderProps {
  workspaceName?: string;
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  selectedProvider?: CloudProvider;
  onCreateCluster: (data: CreateClusterForm) => void;
  isCreatePending?: boolean;
  isCreateDialogOpen: boolean;
  onCreateDialogChange: (open: boolean) => void;
}

function ClusterPageHeaderComponent({
  workspaceName,
  credentials,
  selectedCredentialId,
  onCredentialChange,
  selectedProvider,
  onCreateCluster,
  isCreatePending = false,
  isCreateDialogOpen,
  onCreateDialogChange,
}: ClusterPageHeaderProps) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Kubernetes Clusters</h1>
        <p className="text-gray-600">
          Manage Kubernetes clusters{workspaceName ? ` for ${workspaceName}` : ''}
        </p>
      </div>
      <div className="flex items-center space-x-2">
        <Select
          value={selectedCredentialId}
          onValueChange={onCredentialChange}
        >
          <SelectTrigger className="w-[250px]">
            <SelectValue placeholder="Select Credential" />
          </SelectTrigger>
          <SelectContent>
            {credentials.map((credential) => (
              <SelectItem key={credential.id} value={credential.id}>
                {credential.name || `${credential.provider.toUpperCase()} (${credential.id.slice(0, 8)})`}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Dialog open={isCreateDialogOpen} onOpenChange={onCreateDialogChange}>
          <DialogTrigger asChild>
            <Button
              disabled={!selectedCredentialId || credentials.length === 0}
            >
              <Plus className="mr-2 h-4 w-4" />
              Create Cluster
            </Button>
          </DialogTrigger>
          <CreateClusterDialog
            open={isCreateDialogOpen}
            onOpenChange={onCreateDialogChange}
            onSubmit={onCreateCluster}
            credentials={credentials}
            selectedCredentialId={selectedCredentialId}
            onCredentialChange={onCredentialChange}
            selectedProvider={selectedProvider}
            isPending={isCreatePending}
          />
        </Dialog>
      </div>
    </div>
  );
}

export const ClusterPageHeader = React.memo(ClusterPageHeaderComponent);


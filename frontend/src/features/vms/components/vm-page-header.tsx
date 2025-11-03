/**
 * VM Page Header Component
 * Virtual Machine 페이지 헤더
 */

'use client';

import * as React from 'react';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogTrigger } from '@/components/ui/dialog';
import { Plus } from 'lucide-react';
import type { Credential, CreateVMForm } from '@/lib/types';

// Dynamic import for CreateVMDialog
const CreateVMDialog = dynamic(
  () => import('./create-vm-dialog').then(mod => ({ default: mod.CreateVMDialog })),
  { 
    ssr: false,
    loading: () => null, // Dialog is hidden by default, so no loading state needed
  }
);

interface VMPageHeaderProps {
  workspaceName?: string;
  credentials: Credential[];
  selectedCredentialId: string;
  onCredentialChange: (credentialId: string) => void;
  onCreateVM: (data: CreateVMForm) => void;
  isCreatePending?: boolean;
  isCreateDialogOpen: boolean;
  onCreateDialogChange: (open: boolean) => void;
}

function VMPageHeaderComponent({
  workspaceName,
  credentials,
  selectedCredentialId,
  onCredentialChange,
  onCreateVM,
  isCreatePending = false,
  isCreateDialogOpen,
  onCreateDialogChange,
}: VMPageHeaderProps) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Virtual Machines</h1>
        <p className="text-gray-600">
          Manage VMs{workspaceName ? ` in ${workspaceName} workspace` : ''}
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
            <Button disabled={!selectedCredentialId || credentials.length === 0}>
              <Plus className="mr-2 h-4 w-4" />
              Create VM
            </Button>
          </DialogTrigger>
          <CreateVMDialog
            open={isCreateDialogOpen}
            onOpenChange={onCreateDialogChange}
            onSubmit={onCreateVM}
            credentials={credentials}
            selectedCredentialId={selectedCredentialId}
            onCredentialChange={onCredentialChange}
            isPending={isCreatePending}
          />
        </Dialog>
      </div>
    </div>
  );
}

export const VMPageHeader = React.memo(VMPageHeaderComponent);


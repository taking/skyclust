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
import { useTranslation } from '@/hooks/use-translation';
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
  const { t } = useTranslation();
  
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">{t('vm.title')}</h1>
        <p className="text-gray-600">
          {workspaceName 
            ? t('vm.manageVMsWithWorkspace', { workspaceName }) 
            : t('vm.manageVMs')
          }
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {/* Credential selection is now handled in Header */}
        <Dialog open={isCreateDialogOpen} onOpenChange={onCreateDialogChange}>
          <DialogTrigger asChild>
            <Button disabled={!selectedCredentialId || credentials.length === 0}>
              <Plus className="mr-2 h-4 w-4" />
              {t('vm.create')}
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


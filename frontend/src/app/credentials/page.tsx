/**
 * Credentials Page (Refactored)
 * Credentials 관리 페이지 - 리팩토링된 버전
 */

'use client';

import { useState } from 'react';
import dynamic from 'next/dynamic';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { useWorkspaceStore } from '@/store/workspace';
import { useCredentials } from '@/hooks/use-credentials';
import { useCreateDialog } from '@/hooks/use-create-dialog';
import { EVENTS } from '@/lib/constants';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { ResourceEmptyState } from '@/components/common/resource-empty-state';
import { Key } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';
import {
  useCredentialActions,
  CredentialsPageHeader,
  CredentialList,
} from '@/features/credentials';
import type { CreateCredentialForm, Credential } from '@/lib/types';

// Dynamic imports for Dialog components
const CreateCredentialDialog = dynamic(
  () => import('@/features/credentials').then(mod => ({ default: mod.CreateCredentialDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

const EditCredentialDialog = dynamic(
  () => import('@/features/credentials').then(mod => ({ default: mod.EditCredentialDialog })),
  { 
    ssr: false,
    loading: () => null,
  }
);

export default function CredentialsPage() {
  const { t } = useTranslation();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useCreateDialog(EVENTS.CREATE_DIALOG.CREDENTIAL);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [editingCredential, setEditingCredential] = useState<Credential | null>(null);
  const [showCredentials, setShowCredentials] = useState<Record<string, boolean>>({});
  const [gcpInputMode, setGcpInputMode] = useState<'json' | 'file'>('json');
  const [deleteDialogState, setDeleteDialogState] = useState<{
    open: boolean;
    credentialId: string | null;
    credentialName?: string;
  }>({
    open: false,
    credentialId: null,
    credentialName: undefined,
  });
  const { currentWorkspace } = useWorkspaceStore();
  const router = useRouter();

  // Fetch credentials using unified hook
  const { credentials, isLoading } = useCredentials({
    workspaceId: currentWorkspace?.id,
  });

  const {
    createCredentialMutation,
    createCredentialFromFileMutation,
    updateCredentialMutation,
    deleteCredentialMutation,
  } = useCredentialActions({
    workspaceId: currentWorkspace?.id,
    onSuccess: () => {
      setIsCreateDialogOpen(false);
      setGcpInputMode('json');
    },
  });

  const handleCreateCredential = async (data: CreateCredentialForm) => {
    if (!currentWorkspace) return;
    
    // Handle GCP file upload
    if (data.provider === 'gcp' && gcpInputMode === 'file') {
      const file = (data.credentials as any)?._file as File; // eslint-disable-line @typescript-eslint/no-explicit-any
      if (!file) {
        alert(t('credential.selectGCPFile') || 'Please select a GCP service account JSON file');
        return;
      }
      
      createCredentialFromFileMutation.mutate({
        workspaceId: currentWorkspace.id,
        name: data.name || 'GCP Production',
        provider: 'gcp',
        file,
      });
      return;
    }
    
    // Transform credentials object to data field (remove _file if present)
    const credentials = { ...data.credentials };
    delete (credentials as any)._file; // eslint-disable-line @typescript-eslint/no-explicit-any
    
    const requestData = {
      workspace_id: currentWorkspace.id,
      name: data.name || `${data.provider.toUpperCase()} Credential`,
      provider: data.provider,
      data: credentials || {},
    };
    createCredentialMutation.mutate(requestData as any); // eslint-disable-line @typescript-eslint/no-explicit-any
  };

  const handleEditCredential = (credential: Credential) => {
    setEditingCredential(credential);
    setIsEditDialogOpen(true);
  };

  const handleUpdateCredential = (data: CreateCredentialForm) => {
    if (!editingCredential) return;
    updateCredentialMutation.mutate({
      id: editingCredential.id,
      data,
    });
  };

  const handleDeleteCredential = (credentialId: string) => {
    const credential = credentials.find(c => c.id === credentialId);
    setDeleteDialogState({ open: true, credentialId, credentialName: credential?.name });
  };

  const handleConfirmDelete = () => {
    if (deleteDialogState.credentialId) {
      deleteCredentialMutation.mutate(deleteDialogState.credentialId);
      setDeleteDialogState({ open: false, credentialId: null, credentialName: undefined });
    }
  };

  const toggleShowCredentials = (credentialId: string) => {
    setShowCredentials(prev => ({
      ...prev,
      [credentialId]: !prev[credentialId]
    }));
  };

  const handleCloseEdit = () => {
    setIsEditDialogOpen(false);
    setEditingCredential(null);
  };

  if (!currentWorkspace) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">
            {t('credential.noWorkspaceSelected')}
          </h2>
          <p className="text-gray-600 mb-6">
            {t('credential.selectWorkspaceMessage')}
          </p>
          <Button onClick={() => router.push('/workspaces')}>
            {t('credential.selectWorkspace')}
          </Button>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <WorkspaceRequired>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">{t('credential.loadingCredentials')}</p>
          </div>
        </div>
      </WorkspaceRequired>
    );
  }

  return (
    <WorkspaceRequired>
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <CredentialsPageHeader
            workspace={currentWorkspace}
            onCreateClick={() => setIsCreateDialogOpen(true)}
            isCreatePending={createCredentialMutation.isPending || createCredentialFromFileMutation.isPending}
          />

          {credentials.length === 0 ? (
            <ResourceEmptyState
              resourceName={t('credential.title')}
              icon={Key}
              description={t('credential.addCredentialsDescription')}
              onCreateClick={() => setIsCreateDialogOpen(true)}
            />
          ) : (
            <CredentialList
              credentials={credentials}
              showCredentials={showCredentials}
              onToggleShow={toggleShowCredentials}
              onEdit={handleEditCredential}
              onDelete={handleDeleteCredential}
              isDeleting={deleteCredentialMutation.isPending}
            />
          )}

          <CreateCredentialDialog
            open={isCreateDialogOpen}
            onOpenChange={setIsCreateDialogOpen}
            onSubmit={handleCreateCredential}
            isPending={createCredentialMutation.isPending || createCredentialFromFileMutation.isPending}
            gcpInputMode={gcpInputMode}
            onGcpInputModeChange={setGcpInputMode}
          />

          <EditCredentialDialog
            open={isEditDialogOpen}
            onOpenChange={setIsEditDialogOpen}
            credential={editingCredential}
            onSubmit={handleUpdateCredential}
            onClose={handleCloseEdit}
            isPending={updateCredentialMutation.isPending}
          />

          {/* Delete Credential Confirmation Dialog */}
          <DeleteConfirmationDialog
            open={deleteDialogState.open}
            onOpenChange={(open) => setDeleteDialogState({ ...deleteDialogState, open })}
            onConfirm={handleConfirmDelete}
            title={t('credential.deleteCredential')}
            description={t('credential.confirmDelete')}
            isLoading={deleteCredentialMutation.isPending}
            resourceName={deleteDialogState.credentialName}
            resourceNameLabel="자격 증명 이름"
          />
        </div>
      </div>
    </WorkspaceRequired>
  );
}


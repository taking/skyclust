/**
 * Workspace Credentials Tab
 * Workspace 상세 페이지의 Credentials 탭
 * 
 * 기존 CredentialsPage의 내용을 재사용
 */

'use client';

import { useState, useCallback } from 'react';
import dynamic from 'next/dynamic';
import { useCredentials } from '@/hooks/use-credentials';
import { useCreateDialog } from '@/hooks/use-create-dialog';
import { EVENTS } from '@/lib/constants';
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
import { Spinner } from '@/components/ui/spinner';

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

interface WorkspaceCredentialsTabProps {
  workspaceId: string;
}

export function WorkspaceCredentialsTab({ workspaceId }: WorkspaceCredentialsTabProps) {
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

  // Fetch credentials using unified hook
  const { credentials, isLoading } = useCredentials({
    workspaceId: workspaceId,
  });

  const {
    createCredentialMutation,
    createCredentialFromFileMutation,
    updateCredentialMutation,
    deleteCredentialMutation,
  } = useCredentialActions({
    workspaceId: workspaceId,
  });

  const handleCreateCredential = useCallback((data: CreateCredentialForm) => {
    createCredentialMutation.mutate(data, {
      onSuccess: () => {
        setIsCreateDialogOpen(false);
      },
    });
  }, [createCredentialMutation, setIsCreateDialogOpen]);

  const handleCreateCredentialFromFile = useCallback((file: File) => {
    createCredentialFromFileMutation.mutate(file, {
      onSuccess: () => {
        setIsCreateDialogOpen(false);
      },
    });
  }, [createCredentialFromFileMutation, setIsCreateDialogOpen]);

  const handleEditCredential = useCallback((credential: Credential) => {
    setEditingCredential(credential);
    setIsEditDialogOpen(true);
  }, [setEditingCredential, setIsEditDialogOpen]);

  const handleUpdateCredential = useCallback((data: CreateCredentialForm) => {
    if (!editingCredential) return;
    
    updateCredentialMutation.mutate(
      { id: editingCredential.id, data },
      {
        onSuccess: () => {
          setIsEditDialogOpen(false);
          setEditingCredential(null);
        },
      }
    );
  }, [editingCredential, updateCredentialMutation, setIsEditDialogOpen, setEditingCredential]);

  const handleCloseEdit = useCallback(() => {
    setIsEditDialogOpen(false);
    setEditingCredential(null);
  }, [setIsEditDialogOpen, setEditingCredential]);

  const handleDeleteCredential = useCallback((credentialId: string, credentialName?: string) => {
    setDeleteDialogState({
      open: true,
      credentialId,
      credentialName,
    });
  }, [setDeleteDialogState]);

  const handleConfirmDelete = useCallback(() => {
    if (!deleteDialogState.credentialId) return;
    
    deleteCredentialMutation.mutate(deleteDialogState.credentialId, {
      onSuccess: () => {
        setDeleteDialogState({ open: false, credentialId: null });
      },
    });
  }, [deleteDialogState.credentialId, deleteCredentialMutation, setDeleteDialogState]);

  const toggleShowCredentials = useCallback((credentialId: string) => {
    setShowCredentials((prev) => ({
      ...prev,
      [credentialId]: !prev[credentialId],
    }));
  }, []);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" label="Loading credentials..." />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <CredentialsPageHeader
        workspace={{ id: workspaceId, name: workspaceId }}
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

      <DeleteConfirmationDialog
        open={deleteDialogState.open}
        onOpenChange={(open) => {
          if (open && deleteDialogState.credentialId) {
            handleDeleteCredential(deleteDialogState.credentialId, deleteDialogState.credentialName);
          } else {
            setDeleteDialogState({ open: false, credentialId: null });
          }
        }}
        onConfirm={handleConfirmDelete}
        title={t('credential.deleteCredential')}
        description={deleteDialogState.credentialId ? t('credential.confirmDeleteCredential', { credentialId: deleteDialogState.credentialId }) : ''}
        isLoading={deleteCredentialMutation.isPending}
        resourceName={deleteDialogState.credentialName || deleteDialogState.credentialId || undefined}
        resourceNameLabel="Credential ID"
      />
    </div>
  );
}


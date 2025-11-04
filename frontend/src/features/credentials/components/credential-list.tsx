/**
 * Credential List Component
 * Credential 목록 그리드 컴포넌트
 */

'use client';

import { CredentialCard } from './credential-card';
import type { Credential } from '@/lib/types';

interface CredentialListProps {
  credentials: Credential[];
  showCredentials: Record<string, boolean>;
  onToggleShow: (credentialId: string) => void;
  onEdit: (credential: Credential) => void;
  onDelete: (credentialId: string) => void;
  isDeleting?: boolean;
}

export function CredentialList({
  credentials,
  showCredentials,
  onToggleShow,
  onEdit,
  onDelete,
  isDeleting = false,
}: CredentialListProps) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6">
      {credentials.map((credential) => (
        <CredentialCard
          key={credential.id}
          credential={credential}
          showCredentials={showCredentials[credential.id] || false}
          onToggleShow={() => onToggleShow(credential.id)}
          onEdit={() => onEdit(credential)}
          onDelete={() => onDelete(credential.id)}
          isDeleting={isDeleting}
        />
      ))}
    </div>
  );
}


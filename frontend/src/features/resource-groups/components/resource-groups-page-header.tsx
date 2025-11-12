/**
 * Resource Groups Page Header Component
 * Resource Groups 목록 페이지 헤더 컴포넌트
 */

'use client';

import * as React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { CredentialIndicator } from '@/components/common/credential-indicator';

export function ResourceGroupsPageHeader() {
  const { t } = useTranslation();
  const { selectedCredentialId } = useCredentialContext();

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-2xl font-bold">{t('nav.resourceGroups')}</h1>
        <p className="text-muted-foreground mt-1">
          Manage your Azure Resource Groups
        </p>
      </div>
      {selectedCredentialId && (
        <CredentialIndicator />
      )}
    </div>
  );
}


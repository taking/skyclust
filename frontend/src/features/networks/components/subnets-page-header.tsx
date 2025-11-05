/**
 * Subnets Page Header Component
 * Subnets 페이지 헤더 컴포넌트
 */

'use client';

import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';

export function SubnetsPageHeader() {
  const { currentWorkspace } = useWorkspaceStore();
  const { t } = useTranslation();

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">{t('network.subnets')}</h1>
        <p className="text-gray-600 mt-1">
          {currentWorkspace 
            ? t('network.manageSubnetsWithWorkspace', { workspaceName: currentWorkspace.name }) 
            : t('network.manageSubnets')
          }
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {/* Credential selection is now handled in Header */}
      </div>
    </div>
  );
}


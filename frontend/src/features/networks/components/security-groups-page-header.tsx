/**
 * Security Groups Page Header Component
 * Security Groups 페이지 헤더 컴포넌트
 */

'use client';

import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';

export function SecurityGroupsPageHeader() {
  const { currentWorkspace } = useWorkspaceStore();
  const { t } = useTranslation();

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">{t('network.securityGroups')}</h1>
        <p className="text-gray-600 mt-1">
          {currentWorkspace 
            ? t('network.manageSecurityGroupsWithWorkspace', { workspaceName: currentWorkspace.name }) 
            : t('network.manageSecurityGroups')
          }
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {/* Credential selection is now handled in Header */}
      </div>
    </div>
  );
}


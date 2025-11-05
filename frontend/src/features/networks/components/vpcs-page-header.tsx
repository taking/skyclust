/**
 * VPCs Page Header Component
 * VPCs 페이지 헤더 컴포넌트
 */

'use client';

import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';

export function VPCsPageHeader() {
  const { currentWorkspace } = useWorkspaceStore();
  const { t } = useTranslation();

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">{t('network.vpcs')}</h1>
        <p className="text-gray-600 mt-1">
          {currentWorkspace 
            ? t('network.manageVPCsWithWorkspace', { workspaceName: currentWorkspace.name }) 
            : t('network.manageVPCs')
          }
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {/* Credential selection is now handled in Header */}
      </div>
    </div>
  );
}

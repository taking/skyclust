/**
 * Security Groups Page Header Component
 * Security Groups 페이지 헤더 컴포넌트
 */

'use client';

import { useWorkspaceStore } from '@/store/workspace';

export function SecurityGroupsPageHeader() {
  const { currentWorkspace } = useWorkspaceStore();

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Security Groups</h1>
        <p className="text-gray-600 mt-1">
          Manage Security Groups{currentWorkspace ? ` for ${currentWorkspace.name}` : ''}
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {/* Credential selection is now handled in Header */}
      </div>
    </div>
  );
}


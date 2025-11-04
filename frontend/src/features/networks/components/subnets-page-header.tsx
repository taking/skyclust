/**
 * Subnets Page Header Component
 * Subnets 페이지 헤더 컴포넌트
 */

'use client';

import { useWorkspaceStore } from '@/store/workspace';

export function SubnetsPageHeader() {
  const { currentWorkspace } = useWorkspaceStore();

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Subnets</h1>
        <p className="text-gray-600 mt-1">
          Manage Subnets{currentWorkspace ? ` for ${currentWorkspace.name}` : ''}
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {/* Credential selection is now handled in Header */}
      </div>
    </div>
  );
}


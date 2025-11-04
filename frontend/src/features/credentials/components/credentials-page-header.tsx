/**
 * Credentials Page Header Component
 * Credentials 페이지 헤더 컴포넌트
 */

'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Home } from 'lucide-react';
import type { Workspace } from '@/lib/types';

interface CredentialsPageHeaderProps {
  workspace: Workspace | undefined;
  onCreateClick: () => void;
  isCreatePending: boolean;
}

export function CredentialsPageHeader({
  workspace,
  onCreateClick,
  isCreatePending,
}: CredentialsPageHeaderProps) {
  const router = useRouter();

  return (
    <div className="flex flex-col space-y-4 md:flex-row md:justify-between md:items-center md:space-y-0 mb-6 md:mb-8">
      <div>
        <h1 className="text-2xl md:text-3xl font-bold text-gray-900">Credentials</h1>
        <p className="text-sm md:text-base text-gray-600">
          {workspace ? `Manage cloud provider credentials for ${workspace.name}` : 'Manage cloud provider credentials'}
        </p>
      </div>
      <div className="flex items-center space-x-2">
        <Button variant="outline" onClick={() => router.push('/dashboard')}>
          <Home className="mr-2 h-4 w-4" />
          Home
        </Button>
        <Button onClick={onCreateClick} disabled={isCreatePending}>
          Add Credentials
        </Button>
      </div>
    </div>
  );
}


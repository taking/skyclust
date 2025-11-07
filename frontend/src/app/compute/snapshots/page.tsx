/**
 * VM Snapshots Page
 * VM 스냅샷 관리 페이지
 */

'use client';

import { Card, CardContent } from '@/components/ui/card';
import { HardDrive } from 'lucide-react';
import { useRequireAuth } from '@/hooks/use-auth';
import { useWorkspaceStore } from '@/store/workspace';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { Layout } from '@/components/layout/layout';
// import { useCredentialContext } from '@/hooks/use-credential-context'; // Not used yet
import { useTranslation } from '@/hooks/use-translation';

export default function SnapshotsPage() {
  const { currentWorkspace } = useWorkspaceStore();
  const { isLoading: authLoading } = useRequireAuth();
  const { t } = useTranslation();
  
  // Get credential context from global store (header에서 관리)
  // const { selectedCredentialId } = useCredentialContext(); // Not used yet

  if (authLoading) {
    return (
      <WorkspaceRequired>
        <Layout>
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
              <p className="mt-2 text-gray-600">{t('common.loading')}</p>
            </div>
          </div>
        </Layout>
      </WorkspaceRequired>
    );
  }

  return (
    <WorkspaceRequired>
      <Layout>
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">{t('vm.snapshotsTitle')}</h1>
              <p className="text-gray-600 mt-1">
                {currentWorkspace 
                  ? t('vm.snapshotsDescriptionWithWorkspace', { workspaceName: currentWorkspace.name })
                  : t('vm.snapshotsDescription')
                }
              </p>
            </div>
          </div>

          {/* Empty State - API not implemented yet */}
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <HardDrive className="h-12 w-12 text-gray-400 mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">{t('vm.snapshotsComingSoon')}</h3>
              <p className="text-sm text-gray-500 text-center">
                {t('vm.snapshotsComingSoonDescription')}
              </p>
            </CardContent>
          </Card>
        </div>
      </Layout>
    </WorkspaceRequired>
  );
}


/**
 * Workspace Settings Page
 * 워크스페이스 설정 페이지
 * 
 * 새로운 라우팅 구조: /w/{workspaceId}/workspaces/{id}/settings
 * 
 * Workspace 상세 페이지의 Settings 탭으로 통합되었지만,
 * breadcrumb을 유지하기 위해 별도 페이지로 제공됩니다.
 */

'use client';

import { Suspense } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useResourceContext } from '@/hooks/use-resource-context';
import { buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { Layout } from '@/components/layout/layout';
import { Spinner } from '@/components/ui/spinner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { WorkspaceOverviewTab, WorkspaceSettingsTab, WorkspaceMembersTab, WorkspaceCredentialsTab } from '@/features/workspaces';

function WorkspaceSettingsPageContent() {
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { workspaceId: pathWorkspaceId } = useResourceContext();
  const workspaceId = params.id as string;
  const tab = searchParams.get('tab') || 'settings';

  // 최종 workspaceId 결정: path parameter 우선
  const finalWorkspaceId = pathWorkspaceId || workspaceId;

  const handleTabChange = (value: string) => {
    if (value === 'overview') {
      router.push(buildWorkspaceResourcePath(finalWorkspaceId, finalWorkspaceId, 'overview'));
    } else {
      router.push(`${buildWorkspaceResourcePath(finalWorkspaceId, finalWorkspaceId, value as 'settings' | 'members' | 'credentials')}?tab=${value}`);
    }
  };

  return (
    <Layout>
      <div className="space-y-6">
        <Tabs value={tab} onValueChange={handleTabChange} className="space-y-4">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
            <TabsTrigger value="members">Members</TabsTrigger>
            <TabsTrigger value="credentials">Credentials</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-4">
            <Suspense fallback={
              <div className="flex items-center justify-center h-64">
                <Spinner size="lg" label="Loading overview..." />
              </div>
            }>
              <WorkspaceOverviewTab workspaceId={finalWorkspaceId} />
            </Suspense>
          </TabsContent>

          <TabsContent value="settings" className="space-y-4">
            <Suspense fallback={
              <div className="flex items-center justify-center h-64">
                <Spinner size="lg" label="Loading settings..." />
              </div>
            }>
              <WorkspaceSettingsTab workspaceId={finalWorkspaceId} />
            </Suspense>
          </TabsContent>

          <TabsContent value="members" className="space-y-4">
            <Suspense fallback={
              <div className="flex items-center justify-center h-64">
                <Spinner size="lg" label="Loading members..." />
              </div>
            }>
              <WorkspaceMembersTab workspaceId={finalWorkspaceId} />
            </Suspense>
          </TabsContent>

          <TabsContent value="credentials" className="space-y-4">
            <Suspense fallback={
              <div className="flex items-center justify-center h-64">
                <Spinner size="lg" label="Loading credentials..." />
              </div>
            }>
              <WorkspaceCredentialsTab workspaceId={finalWorkspaceId} />
            </Suspense>
          </TabsContent>
        </Tabs>
      </div>
    </Layout>
  );
}

export default function WorkspaceSettingsPage() {
  return (
    <Suspense fallback={
      <Layout>
        <div className="flex items-center justify-center h-64">
          <Spinner size="lg" label="Loading..." />
        </div>
      </Layout>
    }>
      <WorkspaceSettingsPageContent />
    </Suspense>
  );
}


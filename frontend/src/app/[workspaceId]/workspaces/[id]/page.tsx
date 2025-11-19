/**
 * Workspace Detail Page
 * 워크스페이스 상세 페이지 (탭 네비게이션 포함)
 * 
 * 새로운 라우팅 구조: /w/{workspaceId}/workspaces/{id}
 * 
 * 탭 구조:
 * - Overview: Workspace 정보, 통계, Quick Actions
 * - Settings: Workspace 설정
 * - Members: 멤버 관리
 * - Credentials: Credentials 관리
 */

'use client';

import { Suspense } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useResourceContext } from '@/hooks/use-resource-context';
import { buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Layout } from '@/components/layout/layout';
import { Spinner } from '@/components/ui/spinner';
import { WorkspaceOverviewTab, WorkspaceSettingsTab, WorkspaceMembersTab, WorkspaceCredentialsTab } from '@/features/workspaces';

function WorkspaceDetailPageContent() {
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { workspaceId: pathWorkspaceId } = useResourceContext();
  const workspaceId = params.id as string;
  const tab = searchParams.get('tab') || 'overview';

  // 최종 workspaceId 결정: path parameter 우선
  const finalWorkspaceId = pathWorkspaceId || workspaceId;

  const handleTabChange = (value: string) => {
    if (value === 'overview') {
      router.push(buildWorkspaceResourcePath(finalWorkspaceId, finalWorkspaceId, 'overview'));
    } else {
      // settings, members, credentials는 각각의 경로로 이동 (breadcrumb 유지)
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

export default function WorkspaceDetailPage() {
  return (
    <Suspense fallback={
      <Layout>
        <div className="flex items-center justify-center h-64">
          <Spinner size="lg" label="Loading workspace..." />
        </div>
      </Layout>
    }>
      <WorkspaceDetailPageContent />
    </Suspense>
  );
}


/**
 * Workspace Overview Tab
 * Workspace 상세 페이지의 Overview 탭
 * 
 * 통계, Quick Actions, Workspace 정보를 표시
 */

'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Key, Users, Settings, Calendar, ArrowRight } from 'lucide-react';
import { workspaceService } from '../services/workspace';
import { queryKeys } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { useRouter } from 'next/navigation';
import { buildWorkspaceResourcePath } from '@/lib/routing/helpers';
import { toLocaleDateString } from '@/lib/utils/date-format';
import { Spinner } from '@/components/ui/spinner';

interface WorkspaceOverviewTabProps {
  workspaceId: string;
}

export function WorkspaceOverviewTab({ workspaceId }: WorkspaceOverviewTabProps) {
  const { t, locale } = useTranslation();
  const router = useRouter();

  const { data: workspace, isLoading } = useQuery({
    queryKey: queryKeys.workspaces.detail(workspaceId),
    queryFn: () => workspaceService.getWorkspace(workspaceId),
    enabled: !!workspaceId,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" label="Loading workspace..." />
      </div>
    );
  }

  if (!workspace) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <p className="text-muted-foreground">{t('workspace.workspaceNotFound') || 'Workspace not found'}</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Workspace 정보 */}
      <Card>
        <CardHeader>
          <CardTitle>{t('workspace.workspaceInformation') || 'Workspace Information'}</CardTitle>
          <CardDescription>{t('workspace.workspaceInformationDescription') || 'Basic information about this workspace'}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <label className="text-sm font-medium text-muted-foreground">{t('workspace.name') || 'Name'}</label>
            <p className="text-lg font-semibold mt-1">{workspace.name}</p>
          </div>
          {workspace.description && (
            <div>
              <label className="text-sm font-medium text-muted-foreground">{t('workspace.description') || 'Description'}</label>
              <p className="text-sm mt-1">{workspace.description}</p>
            </div>
          )}
          <div>
            <label className="text-sm font-medium text-muted-foreground">{t('common.createdAt') || 'Created At'}</label>
            <p className="text-sm mt-1">
              {toLocaleDateString(workspace.created_at, locale as 'ko' | 'en')}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* 통계 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center">
              <Key className="mr-2 h-4 w-4" />
              {t('workspace.credentials') || 'Credentials'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <div className="text-3xl font-bold">{workspace.settings?.credential_count ?? 0}</div>
                <p className="text-sm text-muted-foreground mt-1">
                  {t('workspace.totalCredentials') || 'Total credentials'}
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  router.push(buildWorkspaceResourcePath(workspaceId, workspaceId, 'credentials'));
                }}
              >
                {t('common.view') || 'View'}
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center">
              <Users className="mr-2 h-4 w-4" />
              {t('workspace.members') || 'Members'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <div className="text-3xl font-bold">{workspace.settings?.member_count ?? 0}</div>
                <p className="text-sm text-muted-foreground mt-1">
                  {t('workspace.totalMembers') || 'Total members'}
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  router.push(buildWorkspaceResourcePath(workspaceId, workspaceId, 'members'));
                }}
              >
                {t('common.view') || 'View'}
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>{t('workspace.quickActions') || 'Quick Actions'}</CardTitle>
          <CardDescription>{t('workspace.quickActionsDescription') || 'Common workspace management tasks'}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Button
              variant="outline"
              className="h-auto py-4 flex flex-col items-start"
              onClick={() => {
                router.push(buildWorkspaceResourcePath(workspaceId, workspaceId, 'settings'));
              }}
            >
              <Settings className="mb-2 h-5 w-5" />
              <span className="font-semibold">{t('workspace.settings') || 'Settings'}</span>
              <span className="text-xs text-muted-foreground mt-1">
                {t('workspace.manageSettings') || 'Manage workspace settings'}
              </span>
            </Button>

            <Button
              variant="outline"
              className="h-auto py-4 flex flex-col items-start"
              onClick={() => {
                router.push(buildWorkspaceResourcePath(workspaceId, workspaceId, 'members'));
              }}
            >
              <Users className="mb-2 h-5 w-5" />
              <span className="font-semibold">{t('workspace.members') || 'Members'}</span>
              <span className="text-xs text-muted-foreground mt-1">
                {t('workspace.manageMembers') || 'Manage workspace members'}
              </span>
            </Button>

            <Button
              variant="outline"
              className="h-auto py-4 flex flex-col items-start"
              onClick={() => {
                router.push(buildWorkspaceResourcePath(workspaceId, workspaceId, 'credentials'));
              }}
            >
              <Key className="mb-2 h-5 w-5" />
              <span className="font-semibold">{t('workspace.credentials') || 'Credentials'}</span>
              <span className="text-xs text-muted-foreground mt-1">
                {t('workspace.manageCredentials') || 'Manage credentials'}
              </span>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}


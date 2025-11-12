'use client';

import { useState, useEffect, Suspense } from 'react';
import dynamicImport from 'next/dynamic';
import { Layout } from '@/components/layout/layout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useWorkspaceStore } from '@/store/workspace';
import { WidgetData, WidgetType, WidgetSize, WIDGET_CONFIGS } from '@/lib/widgets';
import { Server, Key, Users, Settings, RefreshCw, Container } from 'lucide-react';
import { workspaceService } from '@/features/workspaces';
import { useQuery } from '@tanstack/react-query';
import { WorkspaceRequired } from '@/components/common/workspace-required';
import { queryKeys } from '@/lib/query';
import { useTranslation } from '@/hooks/use-translation';
import { API } from '@/lib/constants';
import { useCredentialContext } from '@/hooks/use-credential-context';
import { useCredentialAutoSelect } from '@/hooks/use-credential-auto-select';
import { useDashboardSummary } from '@/features/dashboard/hooks/use-dashboard-summary';
import { useDashboardSSE } from '@/hooks/use-dashboard-sse';

import { Spinner } from '@/components/ui/loading-states';

// Dynamic imports for heavy components
const DraggableDashboard = dynamicImport(
  () => import('@/components/dashboard/draggable-dashboard').then(mod => ({ default: mod.DraggableDashboard })),
  { 
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" label="Loading dashboard..." />
      </div>
    ),
  }
);

const WidgetAddPanel = dynamicImport(
  () => import('@/components/dashboard/widget-add-panel').then(mod => ({ default: mod.WidgetAddPanel })),
  { 
    ssr: false,
    loading: () => (
      <div className="animate-pulse">
        <div className="h-10 w-32 bg-gray-200 rounded"></div>
      </div>
    ),
  }
);

const WidgetConfigDialog = dynamicImport(
  () => import('@/components/dashboard/widget-config-dialog').then(mod => ({ default: mod.WidgetConfigDialog })),
  { 
    ssr: false,
  }
);

const RealtimeNotifications = dynamicImport(
  () => import('@/components/monitoring/realtime-notifications').then(mod => ({ default: mod.RealtimeNotifications })),
  { 
    ssr: false,
    loading: () => (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center justify-center h-64">
            <Spinner size="lg" />
          </div>
        </CardContent>
      </Card>
    ),
  }
);

function DashboardContent() {
  const { currentWorkspace, setCurrentWorkspace, setWorkspaces } = useWorkspaceStore();
  const { t } = useTranslation();
  const { selectedCredentialId, selectedRegion } = useCredentialContext();
  const [widgets, setWidgets] = useState<WidgetData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [configuringWidget, setConfiguringWidget] = useState<WidgetData | null>(null);

  // Auto-select credential if not selected (like compute, kubernetes, networks pages)
  useCredentialAutoSelect({
    enabled: !!currentWorkspace,
    resourceType: 'compute', // Dashboard shows multiple resources, but use 'compute' as default
    updateUrl: true,
  });

  // 대시보드 요약 정보 조회
  // 멤버 통계는 credential 선택과 무관하므로 항상 조회
  // credentialId와 region은 VM, Cluster, Network 통계에만 사용됨
  const { data: dashboardSummary, isLoading: isLoadingSummary } = useDashboardSummary({
    workspaceId: currentWorkspace?.id || '',
    credentialId: selectedCredentialId || undefined, // VM, Cluster, Network 통계용
    region: selectedRegion || undefined, // VM, Cluster, Network 통계용
    enabled: !!currentWorkspace?.id, // 워크스페이스만 있으면 조회 (credential 선택 불필요)
  });

  // 대시보드 SSE 동적 구독 관리
  // 위젯 목록과 필터에 따라 필요한 이벤트만 구독
  // Dashboard summary는 항상 구독 (위젯이 없어도)
  useDashboardSSE({
    widgets,
    credentialId: selectedCredentialId || undefined,
    region: selectedRegion || undefined,
    includeSummary: true,
    enabled: !!currentWorkspace?.id, // 위젯이 없어도 summary는 구독
  });

  // Fetch workspaces
  const { data: fetchedWorkspaces = [], isLoading: isLoadingWorkspaces } = useQuery({
    queryKey: queryKeys.workspaces.list(),
    queryFn: () => workspaceService.getWorkspaces(),
    retry: API.REQUEST.MAX_RETRIES,
    retryDelay: API.REQUEST.RETRY_DELAY,
  });

  // Auto-select first workspace if available and none is selected
  useEffect(() => {
    if (!isLoadingWorkspaces && fetchedWorkspaces.length > 0) {
      setWorkspaces(fetchedWorkspaces);
      
      // If no workspace is selected, select the first one
      // Only set if we don't already have a current workspace to avoid unnecessary updates
      if (!currentWorkspace) {
        setCurrentWorkspace(fetchedWorkspaces[0]);
      }
    } else if (!isLoadingWorkspaces && fetchedWorkspaces.length === 0) {
      // No workspaces available, set empty array
      setWorkspaces([]);
    }
    // Remove currentWorkspace from dependencies to avoid infinite loops
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fetchedWorkspaces, isLoadingWorkspaces, setCurrentWorkspace, setWorkspaces]);

  // Load default widgets on mount
  useEffect(() => {
    const defaultWidgets: WidgetData[] = [
      {
        id: 'vm-status-1',
        type: 'vm-status',
        title: 'VM Status Overview',
        description: 'Overview of virtual machine statuses',
        size: 'medium',
        position: { x: 0, y: 0 },
        config: {},
        lastUpdated: new Date().toISOString(),
      },
      {
        id: 'resource-usage-1',
        type: 'resource-usage',
        title: 'Resource Usage',
        description: 'Current system resource utilization',
        size: 'medium',
        position: { x: 0, y: 0 },
        config: {},
        lastUpdated: new Date().toISOString(),
      },
      {
        id: 'cost-chart-1',
        type: 'cost-chart',
        title: 'Cost Analysis',
        description: 'Monthly cost trends and breakdown',
        size: 'large',
        position: { x: 0, y: 0 },
        config: {},
        lastUpdated: new Date().toISOString(),
      },
    ];

    // Load from localStorage or use defaults
    const savedWidgets = localStorage.getItem('dashboard-widgets');
    if (savedWidgets) {
      try {
        setWidgets(JSON.parse(savedWidgets));
      } catch {
        setWidgets(defaultWidgets);
      }
    } else {
      setWidgets(defaultWidgets);
    }
    
    setIsLoading(false);
  }, []);

  // Save widgets to localStorage
  useEffect(() => {
    if (widgets.length > 0) {
      localStorage.setItem('dashboard-widgets', JSON.stringify(widgets));
    }
  }, [widgets]);

  const handleWidgetsChange = (newWidgets: WidgetData[]) => {
    setWidgets(newWidgets);
  };

  const handleAddWidget = (type: WidgetType) => {
    const config = WIDGET_CONFIGS[type];
    const newWidget: WidgetData = {
      id: `${type}-${Date.now()}`,
      type,
      title: config.title,
      description: config.description,
      size: config.defaultSize,
      position: { x: 0, y: 0 },
      config: {},
      lastUpdated: new Date().toISOString(),
    };
    setWidgets([...widgets, newWidget]);
  };

  const handleWidgetRemove = (widgetId: string) => {
    setWidgets(widgets.filter(w => w.id !== widgetId));
  };

  const handleWidgetConfigure = (widgetId: string) => {
    const widget = widgets.find(w => w.id === widgetId);
    if (widget) {
      setConfiguringWidget(widget);
    }
  };

  const handleWidgetResize = (widgetId: string, size: WidgetSize) => {
    setWidgets(widgets.map(w => 
      w.id === widgetId ? { ...w, size } : w
    ));
  };

  const handleWidgetSave = (widgetId: string, updates: { size?: WidgetSize; title?: string; config?: Record<string, unknown> }) => {
    setWidgets(widgets.map(w => 
      w.id === widgetId 
        ? { ...w, ...updates, lastUpdated: new Date().toISOString() }
        : w
    ));
  };

  const handleResetDashboard = () => {
    if (confirm(t('dashboard.resetConfirm'))) {
      const defaultWidgets: WidgetData[] = [
        {
          id: 'vm-status-1',
          type: 'vm-status',
          title: 'VM Status Overview',
          description: 'Overview of virtual machine statuses',
          size: 'medium',
          position: { x: 0, y: 0 },
          config: {},
          lastUpdated: new Date().toISOString(),
        },
        {
          id: 'resource-usage-1',
          type: 'resource-usage',
          title: 'Resource Usage',
          description: 'Current system resource utilization',
          size: 'medium',
          position: { x: 0, y: 0 },
          config: {},
          lastUpdated: new Date().toISOString(),
        },
        {
          id: 'cost-chart-1',
          type: 'cost-chart',
          title: 'Cost Analysis',
          description: 'Monthly cost trends and breakdown',
          size: 'large',
          position: { x: 0, y: 0 },
          config: {},
          lastUpdated: new Date().toISOString(),
        },
      ];
      setWidgets(defaultWidgets);
    }
  };

  // WorkspaceRequired component handles the no-workspace case
  // So we can remove this check here

  if (isLoading) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">{t('dashboard.loading')}</p>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <WorkspaceRequired allowAutoSelect={true}>
      <Layout>
        <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{t('dashboard.title')}</h1>
            <p className="text-gray-600">
              {currentWorkspace?.name
                ? t('dashboard.welcomeWithWorkspace', { workspaceName: currentWorkspace.name || '' })
                : t('dashboard.welcome')
              }
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleResetDashboard}
            >
              <RefreshCw className="mr-2 h-4 w-4" />
              {t('dashboard.reset')}
            </Button>
            <WidgetAddPanel
              onAddWidget={handleAddWidget}
              existingWidgets={widgets.map(w => w.type)}
            />
          </div>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{t('dashboard.vms')}</CardTitle>
              <Server className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {isLoadingSummary ? (
                  <Spinner size="sm" />
                ) : (
                  dashboardSummary?.vms.total ?? 0
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                {t('dashboard.vmsDescription')}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{t('dashboard.clusters')}</CardTitle>
              <Container className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {isLoadingSummary ? (
                  <Spinner size="sm" />
                ) : (
                  dashboardSummary?.clusters.total ?? 0
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                {t('dashboard.clustersDescription')}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{t('dashboard.credentials')}</CardTitle>
              <Key className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {isLoadingSummary ? (
                  <Spinner size="sm" />
                ) : (
                  dashboardSummary?.credentials.total ?? 0
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                {t('dashboard.credentialsDescription')}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{t('dashboard.members')}</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {isLoadingSummary ? (
                  <Spinner size="sm" />
                ) : (
                  dashboardSummary?.members.total ?? 0
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                {t('dashboard.membersDescription')}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Draggable Dashboard */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold">{t('dashboard.widgets')}</h2>
              <p className="text-sm text-gray-600">
                {t('dashboard.widgetsDescription')}
              </p>
            </div>
            <Badge variant="outline">
              {widgets.length} {widgets.length !== 1 ? t('dashboard.widgets') : t('dashboard.widgets').replace(/s$/, '')}
            </Badge>
          </div>

          {widgets.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <div className="mx-auto h-12 w-12 text-gray-400">
                  <Settings className="h-12 w-12" />
                </div>
                <h3 className="mt-2 text-sm font-medium text-gray-900">{t('dashboard.noWidgets')}</h3>
                <p className="mt-1 text-sm text-gray-500">
                  {t('dashboard.noWidgetsDescription')}
                </p>
                <div className="mt-6">
                  <WidgetAddPanel
                    onAddWidget={handleAddWidget}
                    existingWidgets={[]}
                  />
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
              <div className="lg:col-span-3">
                <DraggableDashboard
                  widgets={widgets}
                  onWidgetsChange={handleWidgetsChange}
                  onWidgetRemove={handleWidgetRemove}
                  onWidgetConfigure={handleWidgetConfigure}
                  onWidgetResize={handleWidgetResize}
                />
                
                <WidgetConfigDialog
                  widget={configuringWidget}
                  open={!!configuringWidget}
                  onOpenChange={(open) => !open && setConfiguringWidget(null)}
                  onSave={handleWidgetSave}
                />
              </div>
              <div className="lg:col-span-1">
                <RealtimeNotifications className="sticky top-4" />
              </div>
            </div>
          )}
        </div>
      </div>
    </Layout>
    </WorkspaceRequired>
  );
}

export default function DashboardPage() {
  return (
    <Suspense fallback={
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-gray-600">Loading...</p>
          </div>
        </div>
      </Layout>
    }>
      <DashboardContent />
    </Suspense>
  );
}
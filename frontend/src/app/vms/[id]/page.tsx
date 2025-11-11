'use client';

import React from 'react';
import dynamic from 'next/dynamic';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Layout } from '@/components/layout/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { vmService, VMOverviewTab, VMMonitoringTab, VMNetworkingTab, VMStorageTab, VMDetailHeader, VMActionsCard } from '@/features/vms';
import { useToast } from '@/hooks/use-toast';
import { useRequireAuth } from '@/hooks/use-auth';
import { queryKeys } from '@/lib/query';
import { Spinner } from '@/components/ui/spinner';
import { DeleteConfirmationDialog } from '@/components/common/delete-confirmation-dialog';

// Dynamic imports for heavy components
const RealtimeVMMonitor = dynamic(
  () => import('@/components/monitoring/realtime-vm-monitor').then(mod => ({ default: mod.RealtimeVMMonitor })),
  { 
    ssr: false,
    loading: () => (
      <Card>
        <CardHeader>
          <CardTitle>Real-time Monitoring</CardTitle>
          <CardDescription>Loading monitoring data...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64">
            <Spinner size="lg" label="Loading real-time monitor..." />
          </div>
        </CardContent>
      </Card>
    ),
  }
);

// Mock data for charts
const cpuData = [
  { time: '00:00', cpu: 20 },
  { time: '01:00', cpu: 25 },
  { time: '02:00', cpu: 30 },
  { time: '03:00', cpu: 15 },
  { time: '04:00', cpu: 35 },
  { time: '05:00', cpu: 40 },
  { time: '06:00', cpu: 45 },
  { time: '07:00', cpu: 50 },
  { time: '08:00', cpu: 55 },
  { time: '09:00', cpu: 60 },
  { time: '10:00', cpu: 65 },
  { time: '11:00', cpu: 70 },
];

const memoryData = [
  { time: '00:00', memory: 30 },
  { time: '01:00', memory: 32 },
  { time: '02:00', memory: 35 },
  { time: '03:00', memory: 28 },
  { time: '04:00', memory: 38 },
  { time: '05:00', memory: 42 },
  { time: '06:00', memory: 45 },
  { time: '07:00', memory: 48 },
  { time: '08:00', memory: 52 },
  { time: '09:00', memory: 55 },
  { time: '10:00', memory: 58 },
  { time: '11:00', memory: 62 },
];

const networkData = [
  { time: '00:00', in: 100, out: 50 },
  { time: '01:00', in: 120, out: 60 },
  { time: '02:00', in: 80, out: 40 },
  { time: '03:00', in: 90, out: 45 },
  { time: '04:00', in: 110, out: 55 },
  { time: '05:00', in: 130, out: 65 },
  { time: '06:00', in: 150, out: 75 },
  { time: '07:00', in: 170, out: 85 },
  { time: '08:00', in: 190, out: 95 },
  { time: '09:00', in: 210, out: 105 },
  { time: '10:00', in: 230, out: 115 },
  { time: '11:00', in: 250, out: 125 },
];

export default function VMDetailPage() {
  const params = useParams();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { success, error: showError } = useToast();
  const { isLoading: authLoading } = useRequireAuth();
  const vmId = params.id as string;

  // Fetch VM details
  const { data: vm, isLoading, error } = useQuery({
    queryKey: queryKeys.vms.detail(vmId),
    queryFn: () => vmService.getVM(vmId),
    enabled: !!vmId,
  });

  // VM actions mutations
  const startVMMutation = useMutation({
    mutationFn: () => vmService.startVM(vmId),
    onSuccess: () => {
      success('VM started successfully');
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.detail(vmId) });
    },
    onError: (error) => {
      showError(`Failed to start VM: ${error.message}`);
    },
  });

  const stopVMMutation = useMutation({
    mutationFn: () => vmService.stopVM(vmId),
    onSuccess: () => {
      success('VM stopped successfully');
      queryClient.invalidateQueries({ queryKey: queryKeys.vms.detail(vmId) });
    },
    onError: (error) => {
      showError(`Failed to stop VM: ${error.message}`);
    },
  });

  const deleteVMMutation = useMutation({
    mutationFn: () => vmService.deleteVM(vmId),
    onSuccess: () => {
      success('VM deleted successfully');
      router.push('/vms');
    },
    onError: (error) => {
      showError(`Failed to delete VM: ${error.message}`);
    },
  });

  const handleStart = () => {
    startVMMutation.mutate();
  };

  const handleStop = () => {
    stopVMMutation.mutate();
  };

  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);

  const handleDelete = () => {
    setIsDeleteDialogOpen(true);
  };

  const handleConfirmDelete = () => {
    deleteVMMutation.mutate();
    setIsDeleteDialogOpen(false);
  };

  if (authLoading || isLoading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <Spinner size="lg" label="Loading VM details..." />
        </div>
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">Error loading VM</h3>
            <p className="mt-1 text-sm text-gray-500">
              {error instanceof Error ? error.message : 'Something went wrong'}
            </p>
            <Button onClick={() => router.push('/vms')} className="mt-4">
              Back to VMs
            </Button>
          </div>
        </div>
      </Layout>
    );
  }

  if (!vm) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <h3 className="text-lg font-medium text-gray-900">VM not found</h3>
            <p className="mt-1 text-sm text-gray-500">The requested VM could not be found.</p>
            <Button onClick={() => router.push('/vms')} className="mt-4">
              Back to VMs
            </Button>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
        {/* Header */}
        <VMDetailHeader
          vm={vm}
          onBack={() => router.push('/vms')}
          onDelete={handleDelete}
          isDeleting={deleteVMMutation.isPending}
        />

        {/* Action Buttons */}
        <VMActionsCard
          vm={vm}
          onStart={handleStart}
          onStop={handleStop}
          onRestart={() => {
            success('Restart functionality coming soon');
          }}
          isStarting={startVMMutation.isPending}
          isStopping={stopVMMutation.isPending}
        />

        {/* Main Content Tabs */}
        <Tabs defaultValue="overview" className="space-y-4">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="monitoring">Monitoring</TabsTrigger>
            <TabsTrigger value="realtime">Real-time</TabsTrigger>
            <TabsTrigger value="networking">Networking</TabsTrigger>
            <TabsTrigger value="storage">Storage</TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="space-y-4">
            <VMOverviewTab vm={vm} />
          </TabsContent>

          {/* Monitoring Tab */}
          <TabsContent value="monitoring" className="space-y-4">
            <VMMonitoringTab
              cpuData={cpuData}
              memoryData={memoryData}
              networkData={networkData}
            />
          </TabsContent>

          {/* Real-time Tab */}
          <TabsContent value="realtime" className="space-y-4">
            <RealtimeVMMonitor 
              vmId={vm.id} 
              vmName={vm.name}
              className="w-full"
            />
          </TabsContent>

          {/* Networking Tab */}
          <TabsContent value="networking" className="space-y-4">
            <VMNetworkingTab vm={vm} />
          </TabsContent>

          {/* Storage Tab */}
          <TabsContent value="storage" className="space-y-4">
            <VMStorageTab />
          </TabsContent>
        </Tabs>

        {/* Delete Confirmation Dialog */}
        <DeleteConfirmationDialog
          open={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
          onConfirm={handleConfirmDelete}
          title="VM 삭제 확인"
          description="이 VM을 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
          isLoading={deleteVMMutation.isPending}
          resourceName={vm?.name}
          resourceNameLabel="VM 이름"
        />
      </div>
    </Layout>
  );
}

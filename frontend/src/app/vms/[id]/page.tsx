'use client';

import React from 'react';
import dynamic from 'next/dynamic';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Layout } from '@/components/layout/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { vmService } from '@/features/vms';
import { useToast } from '@/hooks/use-toast';
import { useRequireAuth } from '@/hooks/use-auth';
import { 
  ArrowLeft, 
  Play, 
  Pause, 
  RotateCcw, 
  Trash2, 
  Monitor, 
  Network,
  Calendar,
  MapPin,
  Server
} from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts';
import { Spinner } from '@/components/ui/spinner';

// Dynamic import for RealtimeVMMonitor
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

  const handleDelete = () => {
    if (confirm('Are you sure you want to delete this VM? This action cannot be undone.')) {
      deleteVMMutation.mutate();
    }
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

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
        return 'bg-green-100 text-green-800';
      case 'stopped':
        return 'bg-red-100 text-red-800';
      case 'starting':
        return 'bg-yellow-100 text-yellow-800';
      case 'stopping':
        return 'bg-orange-100 text-orange-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <Layout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button
              variant="outline"
              size="sm"
              onClick={() => router.push('/vms')}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back
            </Button>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{vm.name}</h1>
              <p className="text-gray-600">VM Details and Management</p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <Badge className={getStatusColor(vm.status)}>
              {vm.status}
            </Badge>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDelete}
              disabled={deleteVMMutation.isPending}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </Button>
          </div>
        </div>

        {/* Action Buttons */}
        <Card>
          <CardHeader>
            <CardTitle>VM Actions</CardTitle>
            <CardDescription>Control your virtual machine</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex space-x-2">
              <Button
                onClick={handleStart}
                disabled={vm.status === 'running' || startVMMutation.isPending}
                className="flex-1"
              >
                <Play className="mr-2 h-4 w-4" />
                {startVMMutation.isPending ? 'Starting...' : 'Start'}
              </Button>
              <Button
                variant="outline"
                onClick={handleStop}
                disabled={vm.status === 'stopped' || stopVMMutation.isPending}
                className="flex-1"
              >
                <Pause className="mr-2 h-4 w-4" />
                {stopVMMutation.isPending ? 'Stopping...' : 'Stop'}
              </Button>
              <Button
                variant="outline"
                onClick={() => {
                  // Restart logic would go here
                  success('Restart functionality coming soon');
                }}
                disabled={vm.status === 'stopped'}
                className="flex-1"
              >
                <RotateCcw className="mr-2 h-4 w-4" />
                Restart
              </Button>
            </div>
          </CardContent>
        </Card>

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
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Instance Type</CardTitle>
                  <Server className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{vm.instance_type}</div>
                  <p className="text-xs text-muted-foreground">AWS EC2 Instance</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Public IP</CardTitle>
                  <Network className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{vm.public_ip || 'N/A'}</div>
                  <p className="text-xs text-muted-foreground">External access</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Private IP</CardTitle>
                  <Network className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{vm.private_ip || 'N/A'}</div>
                  <p className="text-xs text-muted-foreground">Internal network</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Region</CardTitle>
                  <MapPin className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{vm.region || 'N/A'}</div>
                  <p className="text-xs text-muted-foreground">Deployment region</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Created</CardTitle>
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {new Date(vm.created_at).toLocaleDateString()}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {new Date(vm.created_at).toLocaleTimeString()}
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Provider</CardTitle>
                  <Monitor className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{vm.provider}</div>
                  <p className="text-xs text-muted-foreground">Cloud provider</p>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* Monitoring Tab */}
          <TabsContent value="monitoring" className="space-y-4">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              <Card>
                <CardHeader>
                  <CardTitle>CPU Usage</CardTitle>
                  <CardDescription>CPU utilization over time</CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={cpuData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="time" />
                      <YAxis />
                      <Tooltip />
                      <Line type="monotone" dataKey="cpu" stroke="#8884d8" strokeWidth={2} />
                    </LineChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Memory Usage</CardTitle>
                  <CardDescription>Memory utilization over time</CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <AreaChart data={memoryData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="time" />
                      <YAxis />
                      <Tooltip />
                      <Area type="monotone" dataKey="memory" stroke="#82ca9d" fill="#82ca9d" />
                    </AreaChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>

              <Card className="lg:col-span-2">
                <CardHeader>
                  <CardTitle>Network Traffic</CardTitle>
                  <CardDescription>Network inbound and outbound traffic</CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <AreaChart data={networkData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="time" />
                      <YAxis />
                      <Tooltip />
                      <Area type="monotone" dataKey="in" stackId="1" stroke="#8884d8" fill="#8884d8" />
                      <Area type="monotone" dataKey="out" stackId="1" stroke="#82ca9d" fill="#82ca9d" />
                    </AreaChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>
            </div>
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
            <Card>
              <CardHeader>
                <CardTitle>Network Configuration</CardTitle>
                <CardDescription>Network settings and security groups</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="text-sm font-medium text-gray-700">Public IP Address</label>
                      <p className="text-lg font-mono">{vm.public_ip || 'Not assigned'}</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Private IP Address</label>
                      <p className="text-lg font-mono">{vm.private_ip || 'Not assigned'}</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Subnet ID</label>
                      <p className="text-lg font-mono">subnet-12345678</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">VPC ID</label>
                      <p className="text-lg font-mono">vpc-12345678</p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Storage Tab */}
          <TabsContent value="storage" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Storage Configuration</CardTitle>
                <CardDescription>Attached storage and volumes</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="text-sm font-medium text-gray-700">Root Volume</label>
                      <p className="text-lg font-mono">30 GB (gp3)</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Additional Storage</label>
                      <p className="text-lg font-mono">100 GB (gp3)</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Total Storage</label>
                      <p className="text-lg font-mono">130 GB</p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </Layout>
  );
}

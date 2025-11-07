'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area, BarChart, Bar } from 'recharts';
import { Activity } from 'lucide-react';

interface ClusterMetricsChartProps {
  clusterName: string;
  metrics?: {
    cpu?: Array<{ timestamp: string; value: number }>;
    memory?: Array<{ timestamp: string; value: number }>;
    pods?: Array<{ timestamp: string; value: number }>;
    nodes?: Array<{ timestamp: string; value: number }>;
  };
  isLoading?: boolean;
}

export function ClusterMetricsChart({ clusterName, metrics, isLoading }: ClusterMetricsChartProps) {
  // Mock data for demonstration
  const mockCpuData = metrics?.cpu || Array.from({ length: 24 }, (_, i) => ({
    timestamp: `${String(i).padStart(2, '0')}:00`,
    value: Math.floor(Math.random() * 30) + 40,
  }));

  const mockMemoryData = metrics?.memory || Array.from({ length: 24 }, (_, i) => ({
    timestamp: `${String(i).padStart(2, '0')}:00`,
    value: Math.floor(Math.random() * 20) + 50,
  }));

  const mockPodsData = metrics?.pods || Array.from({ length: 24 }, (_, i) => ({
    timestamp: `${String(i).padStart(2, '0')}:00`,
    value: Math.floor(Math.random() * 10) + 20,
  }));

  const mockNodesData = metrics?.nodes || Array.from({ length: 24 }, (_, i) => ({
    timestamp: `${String(i).padStart(2, '0')}:00`,
    value: 3 + Math.floor(Math.random() * 2),
  }));

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Cluster Metrics</CardTitle>
          <CardDescription>Loading metrics for {clusterName}...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Activity className="mr-2 h-5 w-5" />
          Cluster Metrics
        </CardTitle>
        <CardDescription>Performance metrics for {clusterName}</CardDescription>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="cpu" className="space-y-4">
          <TabsList>
            <TabsTrigger value="cpu">CPU</TabsTrigger>
            <TabsTrigger value="memory">Memory</TabsTrigger>
            <TabsTrigger value="pods">Pods</TabsTrigger>
            <TabsTrigger value="nodes">Nodes</TabsTrigger>
          </TabsList>

          <TabsContent value="cpu">
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={mockCpuData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis domain={[0, 100]} />
                <Tooltip />
                <Area type="monotone" dataKey="value" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.6} />
              </AreaChart>
            </ResponsiveContainer>
          </TabsContent>

          <TabsContent value="memory">
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={mockMemoryData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis domain={[0, 100]} />
                <Tooltip />
                <Area type="monotone" dataKey="value" stroke="#10b981" fill="#10b981" fillOpacity={0.6} />
              </AreaChart>
            </ResponsiveContainer>
          </TabsContent>

          <TabsContent value="pods">
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={mockPodsData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="value" fill="#8b5cf6" />
              </BarChart>
            </ResponsiveContainer>
          </TabsContent>

          <TabsContent value="nodes">
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={mockNodesData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis domain={[0, 10]} />
                <Tooltip />
                <Line type="monotone" dataKey="value" stroke="#f59e0b" strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}


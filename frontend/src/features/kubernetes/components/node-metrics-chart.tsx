'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Server, Cpu, MemoryStick, HardDrive } from 'lucide-react';
import { Node } from '@/lib/types';

interface NodeMetricsChartProps {
  node: Node;
  metrics?: {
    cpu?: Array<{ timestamp: string; value: number }>;
    memory?: Array<{ timestamp: string; value: number }>;
    disk?: Array<{ timestamp: string; value: number }>;
  };
  isLoading?: boolean;
}

export function NodeMetricsChart({ node, metrics, isLoading }: NodeMetricsChartProps) {
  const mockCpuData = metrics?.cpu || Array.from({ length: 12 }, (_, i) => ({
    timestamp: `${String(i * 2).padStart(2, '0')}:00`,
    value: Math.floor(Math.random() * 40) + 30,
  }));

  const mockMemoryData = metrics?.memory || Array.from({ length: 12 }, (_, i) => ({
    timestamp: `${String(i * 2).padStart(2, '0')}:00`,
    value: Math.floor(Math.random() * 30) + 40,
  }));

  const mockDiskData = metrics?.disk || Array.from({ length: 12 }, (_, i) => ({
    timestamp: `${String(i * 2).padStart(2, '0')}:00`,
    value: Math.floor(Math.random() * 20) + 60,
  }));

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Node Metrics</CardTitle>
          <CardDescription>Loading metrics...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-48 flex items-center justify-center">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-gray-900"></div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Server className="mr-2 h-5 w-5" />
          {node.name} Metrics
        </CardTitle>
        <CardDescription>Resource utilization over time</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <div className="flex items-center space-x-2 mb-2">
              <Cpu className="h-4 w-4 text-blue-500" />
              <span className="text-sm font-medium">CPU Usage</span>
            </div>
            <ResponsiveContainer width="100%" height={150}>
              <LineChart data={mockCpuData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" hide />
                <YAxis domain={[0, 100]} hide />
                <Tooltip />
                <Line type="monotone" dataKey="value" stroke="#3b82f6" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
            <div className="text-center mt-2">
              <span className="text-lg font-bold">{mockCpuData[mockCpuData.length - 1]?.value || 0}%</span>
            </div>
          </div>

          <div>
            <div className="flex items-center space-x-2 mb-2">
              <MemoryStick className="h-4 w-4 text-green-500" />
              <span className="text-sm font-medium">Memory Usage</span>
            </div>
            <ResponsiveContainer width="100%" height={150}>
              <LineChart data={mockMemoryData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" hide />
                <YAxis domain={[0, 100]} hide />
                <Tooltip />
                <Line type="monotone" dataKey="value" stroke="#10b981" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
            <div className="text-center mt-2">
              <span className="text-lg font-bold">{mockMemoryData[mockMemoryData.length - 1]?.value || 0}%</span>
            </div>
          </div>

          <div>
            <div className="flex items-center space-x-2 mb-2">
              <HardDrive className="h-4 w-4 text-orange-500" />
              <span className="text-sm font-medium">Disk Usage</span>
            </div>
            <ResponsiveContainer width="100%" height={150}>
              <LineChart data={mockDiskData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" hide />
                <YAxis domain={[0, 100]} hide />
                <Tooltip />
                <Line type="monotone" dataKey="value" stroke="#f59e0b" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
            <div className="text-center mt-2">
              <span className="text-lg font-bold">{mockDiskData[mockDiskData.length - 1]?.value || 0}%</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}


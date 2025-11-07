/**
 * VM Monitoring Tab Component
 * VM 상세 페이지의 Monitoring 탭 컴포넌트
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import dynamic from 'next/dynamic';

// Dynamic import for charts (recharts is heavy)
const CPUMetricsChart = dynamic(
  () => import('recharts').then(mod => ({
    default: ({ data }: { data: Array<{ time: string; cpu: number }> }) => (
      <mod.ResponsiveContainer width="100%" height={300}>
        <mod.LineChart data={data}>
          <mod.CartesianGrid strokeDasharray="3 3" />
          <mod.XAxis dataKey="time" />
          <mod.YAxis />
          <mod.Tooltip />
          <mod.Line type="monotone" dataKey="cpu" stroke="#8884d8" />
        </mod.LineChart>
      </mod.ResponsiveContainer>
    ),
  })),
  { ssr: false, loading: () => <div className="h-[300px] bg-gray-100 animate-pulse rounded" /> }
);

const MemoryMetricsChart = dynamic(
  () => import('recharts').then(mod => ({
    default: ({ data }: { data: Array<{ time: string; memory: number }> }) => (
      <mod.ResponsiveContainer width="100%" height={300}>
        <mod.LineChart data={data}>
          <mod.CartesianGrid strokeDasharray="3 3" />
          <mod.XAxis dataKey="time" />
          <mod.YAxis />
          <mod.Tooltip />
          <mod.Line type="monotone" dataKey="memory" stroke="#82ca9d" />
        </mod.LineChart>
      </mod.ResponsiveContainer>
    ),
  })),
  { ssr: false, loading: () => <div className="h-[300px] bg-gray-100 animate-pulse rounded" /> }
);

const NetworkMetricsChart = dynamic(
  () => import('recharts').then(mod => ({
    default: ({ data }: { data: Array<{ time: string; in: number; out: number }> }) => (
      <mod.ResponsiveContainer width="100%" height={300}>
        <mod.AreaChart data={data}>
          <mod.CartesianGrid strokeDasharray="3 3" />
          <mod.XAxis dataKey="time" />
          <mod.YAxis />
          <mod.Tooltip />
          <mod.Area type="monotone" dataKey="in" stackId="1" stroke="#8884d8" fill="#8884d8" />
          <mod.Area type="monotone" dataKey="out" stackId="1" stroke="#82ca9d" fill="#82ca9d" />
        </mod.AreaChart>
      </mod.ResponsiveContainer>
    ),
  })),
  { ssr: false, loading: () => <div className="h-[300px] bg-gray-100 animate-pulse rounded" /> }
);

interface VMMonitoringTabProps {
  cpuData: Array<{ time: string; cpu: number }>;
  memoryData: Array<{ time: string; memory: number }>;
  networkData: Array<{ time: string; in: number; out: number }>;
}

export function VMMonitoringTab({ cpuData, memoryData, networkData }: VMMonitoringTabProps) {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle>CPU Usage</CardTitle>
          <CardDescription>CPU utilization over time</CardDescription>
        </CardHeader>
        <CardContent>
          <CPUMetricsChart data={cpuData} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Memory Usage</CardTitle>
          <CardDescription>Memory utilization over time</CardDescription>
        </CardHeader>
        <CardContent>
          <MemoryMetricsChart data={memoryData} />
        </CardContent>
      </Card>

      <Card className="lg:col-span-2">
        <CardHeader>
          <CardTitle>Network Traffic</CardTitle>
          <CardDescription>Network inbound and outbound traffic</CardDescription>
        </CardHeader>
        <CardContent>
          <NetworkMetricsChart data={networkData} />
        </CardContent>
      </Card>
    </div>
  );
}


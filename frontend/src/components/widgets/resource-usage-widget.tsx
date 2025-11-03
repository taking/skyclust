'use client';

import * as React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Activity, Cpu, HardDrive, MemoryStick } from 'lucide-react';

interface ResourceUsageWidgetProps {
  data?: {
    cpu: { used: number; total: number; percentage: number };
    memory: { used: number; total: number; percentage: number };
    storage: { used: number; total: number; percentage: number };
  };
  isLoading?: boolean;
}

function ResourceUsageWidgetComponent({ data, isLoading }: ResourceUsageWidgetProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Activity className="mr-2 h-5 w-5" />
            Resource Usage
          </CardTitle>
          <CardDescription>Loading resource data...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="animate-pulse">
                <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                <div className="h-2 bg-gray-200 rounded"></div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  const mockData = data || {
    cpu: { used: 65, total: 100, percentage: 65 },
    memory: { used: 42, total: 100, percentage: 42 },
    storage: { used: 78, total: 100, percentage: 78 },
  };

  const getProgressColor = (percentage: number) => {
    if (percentage >= 90) return 'bg-red-500';
    if (percentage >= 70) return 'bg-yellow-500';
    return 'bg-green-500';
  };

  const getStatusText = (percentage: number) => {
    if (percentage >= 90) return 'Critical';
    if (percentage >= 70) return 'Warning';
    return 'Normal';
  };

  const getStatusColor = (percentage: number) => {
    if (percentage >= 90) return 'text-red-600';
    if (percentage >= 70) return 'text-yellow-600';
    return 'text-green-600';
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Activity className="mr-2 h-5 w-5" />
          Resource Usage
        </CardTitle>
        <CardDescription>Current system resource utilization</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* CPU Usage */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center">
                <Cpu className="mr-2 h-4 w-4 text-blue-500" />
                <span className="text-sm font-medium">CPU</span>
              </div>
              <div className="text-right">
                <span className="text-sm font-medium">{mockData.cpu.percentage}%</span>
                <span className={`ml-2 text-xs ${getStatusColor(mockData.cpu.percentage)}`}>
                  {getStatusText(mockData.cpu.percentage)}
                </span>
              </div>
            </div>
            <Progress 
              value={mockData.cpu.percentage} 
              className="h-2"
            />
            <div className="text-xs text-gray-500 mt-1">
              {mockData.cpu.used}% of {mockData.cpu.total}% available
            </div>
          </div>

          {/* Memory Usage */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center">
                <MemoryStick className="mr-2 h-4 w-4 text-green-500" />
                <span className="text-sm font-medium">Memory</span>
              </div>
              <div className="text-right">
                <span className="text-sm font-medium">{mockData.memory.percentage}%</span>
                <span className={`ml-2 text-xs ${getStatusColor(mockData.memory.percentage)}`}>
                  {getStatusText(mockData.memory.percentage)}
                </span>
              </div>
            </div>
            <Progress 
              value={mockData.memory.percentage} 
              className="h-2"
            />
            <div className="text-xs text-gray-500 mt-1">
              {mockData.memory.used}% of {mockData.memory.total}% available
            </div>
          </div>

          {/* Storage Usage */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center">
                <HardDrive className="mr-2 h-4 w-4 text-orange-500" />
                <span className="text-sm font-medium">Storage</span>
              </div>
              <div className="text-right">
                <span className="text-sm font-medium">{mockData.storage.percentage}%</span>
                <span className={`ml-2 text-xs ${getStatusColor(mockData.storage.percentage)}`}>
                  {getStatusText(mockData.storage.percentage)}
                </span>
              </div>
            </div>
            <Progress 
              value={mockData.storage.percentage} 
              className="h-2"
            />
            <div className="text-xs text-gray-500 mt-1">
              {mockData.storage.used}% of {mockData.storage.total}% available
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export const ResourceUsageWidget = React.memo(ResourceUsageWidgetComponent);

'use client';

import * as React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Server, Play, Pause, Square } from 'lucide-react';

interface VMStatusWidgetProps {
  data?: {
    total: number;
    running: number;
    stopped: number;
    starting: number;
    stopping: number;
  };
  isLoading?: boolean;
}

function VMStatusWidgetComponent({ data, isLoading }: VMStatusWidgetProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Server className="mr-2 h-5 w-5" />
            VM Status Overview
          </CardTitle>
          <CardDescription>Loading VM status...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
              <div className="h-4 bg-gray-200 rounded w-1/2"></div>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  const mockData = data || {
    total: 12,
    running: 8,
    stopped: 3,
    starting: 1,
    stopping: 0,
  };

  const runningPercentage = (mockData.running / mockData.total) * 100;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Server className="mr-2 h-5 w-5" />
          VM Status Overview
        </CardTitle>
        <CardDescription>
          {mockData.total} total virtual machines
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* Running VMs Progress */}
          <div>
            <div className="flex justify-between text-sm mb-2">
              <span className="flex items-center">
                <Play className="mr-1 h-3 w-3 text-green-500" />
                Running
              </span>
              <span className="font-medium">{mockData.running}</span>
            </div>
            <Progress value={runningPercentage} className="h-2" />
          </div>

          {/* Status Badges */}
          <div className="grid grid-cols-2 gap-2">
            <div className="flex items-center justify-between p-2 bg-green-50 rounded">
              <div className="flex items-center">
                <Play className="mr-1 h-3 w-3 text-green-600" />
                <span className="text-sm text-green-800">Running</span>
              </div>
              <Badge variant="secondary" className="bg-green-100 text-green-800">
                {mockData.running}
              </Badge>
            </div>

            <div className="flex items-center justify-between p-2 bg-red-50 rounded">
              <div className="flex items-center">
                <Square className="mr-1 h-3 w-3 text-red-600" />
                <span className="text-sm text-red-800">Stopped</span>
              </div>
              <Badge variant="secondary" className="bg-red-100 text-red-800">
                {mockData.stopped}
              </Badge>
            </div>

            <div className="flex items-center justify-between p-2 bg-yellow-50 rounded">
              <div className="flex items-center">
                <Play className="mr-1 h-3 w-3 text-yellow-600" />
                <span className="text-sm text-yellow-800">Starting</span>
              </div>
              <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
                {mockData.starting}
              </Badge>
            </div>

            <div className="flex items-center justify-between p-2 bg-orange-50 rounded">
              <div className="flex items-center">
                <Pause className="mr-1 h-3 w-3 text-orange-600" />
                <span className="text-sm text-orange-800">Stopping</span>
              </div>
              <Badge variant="secondary" className="bg-orange-100 text-orange-800">
                {mockData.stopping}
              </Badge>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export const VMStatusWidget = React.memo(VMStatusWidgetComponent);

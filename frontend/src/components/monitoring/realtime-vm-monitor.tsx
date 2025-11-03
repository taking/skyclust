'use client';

import { useEffect, useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Button } from '@/components/ui/button';
import { useRealtimeMonitoring } from '@/hooks/use-realtime-monitoring';
import { Play, Square, AlertTriangle, Activity } from 'lucide-react';
import { cn } from '@/lib/utils';

interface VMResourceData {
  vmId: string;
  cpu: number;
  memory: number;
  disk: number;
  timestamp: number;
}

interface VMStatusData {
  vmId: string;
  status: string;
  timestamp: number;
}

interface RealtimeVMMonitorProps {
  vmId: string;
  vmName: string;
  className?: string;
}

export function RealtimeVMMonitor({ vmId, vmName, className }: RealtimeVMMonitorProps) {
  const [resources, setResources] = useState<VMResourceData | null>(null);
  const [status, setStatus] = useState<string>('unknown');
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const {
    onVMResourceUpdate,
    onVMStatusUpdate,
    onVMError,
    subscribeToVM,
    unsubscribeFromVM,
    isConnected: checkConnection,
  } = useRealtimeMonitoring();

  // VM 구독
  useEffect(() => {
    subscribeToVM(vmId);
    setIsConnected(checkConnection());

    return () => {
      unsubscribeFromVM(vmId);
    };
  }, [vmId, subscribeToVM, unsubscribeFromVM, checkConnection]);

  // 리소스 업데이트 핸들러
  const handleResourceUpdate = useCallback((data: VMResourceData) => {
    if (data.vmId === vmId) {
      setResources(data);
      setLastUpdate(new Date(data.timestamp));
      setError(null);
    }
  }, [vmId]);

  // 상태 업데이트 핸들러
  const handleStatusUpdate = useCallback((data: VMStatusData) => {
    if (data.vmId === vmId) {
      setStatus(data.status);
      setLastUpdate(new Date(data.timestamp));
      setError(null);
    }
  }, [vmId]);

  // 에러 핸들러
  const handleError = useCallback((data: { vmId: string; error: string; timestamp: number }) => {
    if (data.vmId === vmId) {
      setError(data.error);
      setLastUpdate(new Date(data.timestamp));
    }
  }, [vmId]);

  // 이벤트 리스너 등록
  useEffect(() => {
    onVMResourceUpdate(handleResourceUpdate);
    onVMStatusUpdate(handleStatusUpdate);
    onVMError(handleError);
  }, [onVMResourceUpdate, onVMStatusUpdate, onVMError, handleResourceUpdate, handleStatusUpdate, handleError]);

  // 연결 상태 주기적 확인
  useEffect(() => {
    const interval = setInterval(() => {
      setIsConnected(checkConnection());
    }, 5000);

    return () => clearInterval(interval);
  }, [checkConnection]);

  const getStatusBadgeVariant = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
        return 'default';
      case 'stopped':
        return 'secondary';
      case 'pending':
        return 'outline';
      case 'error':
        return 'destructive';
      default:
        return 'outline';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
        return <Play className="h-4 w-4" />;
      case 'stopped':
        return <Square className="h-4 w-4" />;
      case 'error':
        return <AlertTriangle className="h-4 w-4" />;
      default:
        return <Activity className="h-4 w-4" />;
    }
  };

  const formatTimestamp = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString();
  };

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold">{vmName}</CardTitle>
          <div className="flex items-center space-x-2">
            <div className={cn(
              'h-2 w-2 rounded-full',
              isConnected ? 'bg-green-500' : 'bg-red-500'
            )} />
            <span className="text-sm text-muted-foreground">
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Badge variant={getStatusBadgeVariant(status)} className="flex items-center space-x-1">
            {getStatusIcon(status)}
            <span>{status}</span>
          </Badge>
          {lastUpdate && (
            <span className="text-xs text-muted-foreground">
              Last update: {lastUpdate.toLocaleTimeString()}
            </span>
          )}
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {error && (
          <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            <div className="flex items-center space-x-2">
              <AlertTriangle className="h-4 w-4" />
              <span>{error}</span>
            </div>
          </div>
        )}

        {resources && (
          <div className="space-y-3">
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-medium">CPU Usage</span>
                <span className="text-sm text-muted-foreground">{resources.cpu.toFixed(1)}%</span>
              </div>
              <Progress value={resources.cpu} className="h-2" />
            </div>

            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-medium">Memory Usage</span>
                <span className="text-sm text-muted-foreground">{resources.memory.toFixed(1)}%</span>
              </div>
              <Progress value={resources.memory} className="h-2" />
            </div>

            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-medium">Disk Usage</span>
                <span className="text-sm text-muted-foreground">{resources.disk.toFixed(1)}%</span>
              </div>
              <Progress value={resources.disk} className="h-2" />
            </div>
          </div>
        )}

        {!resources && !error && (
          <div className="text-center py-8 text-muted-foreground">
            <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">Waiting for real-time data...</p>
          </div>
        )}

        {!isConnected && (
          <div className="text-center py-4">
            <Button 
              variant="outline" 
              size="sm"
              onClick={() => {
                // 재연결 시도
                setIsConnected(checkConnection());
              }}
            >
              Reconnect
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

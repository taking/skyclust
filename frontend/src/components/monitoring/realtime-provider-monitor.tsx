'use client';

import { useEffect, useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useRealtimeMonitoring } from '@/hooks/use-realtime-monitoring';
import { Cloud, Activity, AlertTriangle, CheckCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ProviderStatusData {
  provider: string;
  status: string;
  timestamp: number;
}

interface ProviderInstanceData {
  provider: string;
  instances: unknown[];
  timestamp: number;
}

interface RealtimeProviderMonitorProps {
  provider: string;
  className?: string;
}

export function RealtimeProviderMonitor({ provider, className }: RealtimeProviderMonitorProps) {
  const [status, setStatus] = useState<string>('unknown');
  const [instances, setInstances] = useState<unknown[]>([]);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const {
    onProviderStatusUpdate,
    onProviderInstanceUpdate,
    subscribeToProvider,
    unsubscribeFromProvider,
    isConnected: checkConnection,
  } = useRealtimeMonitoring();

  // Provider 구독
  useEffect(() => {
    subscribeToProvider(provider);
    setIsConnected(checkConnection());

    return () => {
      unsubscribeFromProvider(provider);
    };
  }, [provider, subscribeToProvider, unsubscribeFromProvider, checkConnection]);

  // 상태 업데이트 핸들러
  const handleStatusUpdate = useCallback((data: ProviderStatusData) => {
    if (data.provider === provider) {
      setStatus(data.status);
      setLastUpdate(new Date(data.timestamp));
      setError(null);
    }
  }, [provider]);

  // 인스턴스 업데이트 핸들러
  const handleInstanceUpdate = useCallback((data: ProviderInstanceData) => {
    if (data.provider === provider) {
      setInstances(data.instances);
      setLastUpdate(new Date(data.timestamp));
      setError(null);
    }
  }, [provider]);

  // 이벤트 리스너 등록
  useEffect(() => {
    onProviderStatusUpdate(handleStatusUpdate);
    onProviderInstanceUpdate(handleInstanceUpdate);
  }, [onProviderStatusUpdate, onProviderInstanceUpdate, handleStatusUpdate, handleInstanceUpdate]);

  // 연결 상태 주기적 확인
  useEffect(() => {
    const interval = setInterval(() => {
      setIsConnected(checkConnection());
    }, 5000);

    return () => clearInterval(interval);
  }, [checkConnection]);

  const getStatusBadgeVariant = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
      case 'connected':
        return 'default';
      case 'inactive':
      case 'disconnected':
        return 'secondary';
      case 'error':
        return 'destructive';
      case 'maintenance':
        return 'outline';
      default:
        return 'outline';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
      case 'connected':
        return <CheckCircle className="h-4 w-4" />;
      case 'inactive':
      case 'disconnected':
        return <Cloud className="h-4 w-4" />;
      case 'error':
        return <AlertTriangle className="h-4 w-4" />;
      default:
        return <Activity className="h-4 w-4" />;
    }
  };

  const getInstanceStatusCount = () => {
    const statusCount = instances.reduce((acc: Record<string, number>, instance) => {
      const instanceObj = instance as { status?: string };
      const status = instanceObj.status || 'unknown';
      acc[status] = (acc[status] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);

    return statusCount;
  };

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold flex items-center space-x-2">
            <Cloud className="h-5 w-5" />
            <span>{provider.toUpperCase()}</span>
          </CardTitle>
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

        {instances.length > 0 && (
          <div className="space-y-3">
            <div>
              <h4 className="text-sm font-medium mb-2">Instance Status</h4>
              <div className="flex flex-wrap gap-2">
                {Object.entries(getInstanceStatusCount()).map(([status, count]) => (
                  <Badge key={status} variant="outline" className="text-xs">
                    {status}: {count}
                  </Badge>
                ))}
              </div>
            </div>

            <div>
              <h4 className="text-sm font-medium mb-2">Recent Instances</h4>
              <div className="space-y-2 max-h-32 overflow-y-auto">
                {instances.slice(0, 5).map((instance, index) => {
                  const instanceObj = instance as { name?: string; id?: string; status?: string };
                  return (
                    <div key={index} className="flex items-center justify-between text-xs">
                      <span className="truncate">{instanceObj.name || instanceObj.id}</span>
                      <Badge 
                        variant={getStatusBadgeVariant(instanceObj.status || 'unknown')} 
                        className="text-xs"
                      >
                        {instanceObj.status || 'unknown'}
                      </Badge>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        )}

        {instances.length === 0 && !error && (
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

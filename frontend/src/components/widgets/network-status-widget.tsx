'use client';

import * as React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Network, Shield, Layers, AlertCircle } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { networkService } from '@/services/network';
import { useWorkspaceStore } from '@/store/workspace';
import { useProviderStore } from '@/store/provider';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';
import { useCredentials } from '@/hooks/use-credentials';

interface NetworkStatusWidgetProps {
  credentialId?: string;
  region?: string;
  vpcId?: string;
  isLoading?: boolean;
}

function NetworkStatusWidgetComponent({ credentialId, region, vpcId, isLoading }: NetworkStatusWidgetProps) {
  const { currentWorkspace } = useWorkspaceStore();
  const { selectedProvider } = useProviderStore();

  // Fetch credentials using unified hook
  const { credentials } = useCredentials({
    workspaceId: currentWorkspace?.id,
  });

  const activeCredentialId = credentialId || credentials.find(c => c.provider === selectedProvider)?.id;
  const activeRegion = region || 'ap-northeast-2';

  const { data: vpcs = [], isLoading: isLoadingVPCs } = useQuery({
    queryKey: [...queryKeys.vpcs.all, 'widget', selectedProvider, activeCredentialId, activeRegion],
    queryFn: async () => {
      if (!selectedProvider || !activeCredentialId) return [];
      return networkService.listVPCs(selectedProvider, activeCredentialId, activeRegion);
    },
    enabled: !!selectedProvider && !!activeCredentialId && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 30000,
  });

  const activeVpcId = vpcId || vpcs[0]?.id;

  const { data: subnets = [], isLoading: isLoadingSubnets } = useQuery({
    queryKey: [...queryKeys.subnets.all, 'widget', selectedProvider, activeCredentialId, activeVpcId, activeRegion],
    queryFn: async () => {
      if (!selectedProvider || !activeCredentialId || !activeVpcId) return [];
      return networkService.listSubnets(selectedProvider, activeCredentialId, activeVpcId, activeRegion);
    },
    enabled: !!selectedProvider && !!activeCredentialId && !!activeVpcId && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 30000,
  });

  const { data: securityGroups = [], isLoading: isLoadingSecurityGroups } = useQuery({
    queryKey: [...queryKeys.securityGroups.all, 'widget', selectedProvider, activeCredentialId, activeVpcId, activeRegion],
    queryFn: async () => {
      if (!selectedProvider || !activeCredentialId || !activeVpcId) return [];
      return networkService.listSecurityGroups(selectedProvider, activeCredentialId, activeVpcId, activeRegion);
    },
    enabled: !!selectedProvider && !!activeCredentialId && !!activeVpcId && !!currentWorkspace,
    staleTime: CACHE_TIMES.REALTIME,
    gcTime: GC_TIMES.SHORT,
    refetchInterval: 30000,
  });

  const isLoadingData = isLoading || isLoadingVPCs || isLoadingSubnets || isLoadingSecurityGroups;

  if (isLoadingData) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Network className="mr-2 h-5 w-5" />
            Network Resources
          </CardTitle>
          <CardDescription>Loading network status...</CardDescription>
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

  if (!selectedProvider || !activeCredentialId) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Network className="mr-2 h-5 w-5" />
            Network Resources
          </CardTitle>
          <CardDescription>Select provider and credential</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-4">
            <AlertCircle className="h-8 w-8 text-gray-400" />
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Network className="mr-2 h-5 w-5" />
          Network Resources
        </CardTitle>
        <CardDescription>Network resource overview</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="grid grid-cols-3 gap-4">
            <div className="flex flex-col items-center text-center">
              <Network className="h-6 w-6 text-blue-600 mb-2" />
              <div className="text-2xl font-bold">{vpcs.length}</div>
              <div className="text-xs text-gray-500">VPCs</div>
            </div>
            <div className="flex flex-col items-center text-center">
              <Layers className="h-6 w-6 text-green-600 mb-2" />
              <div className="text-2xl font-bold">{subnets.length}</div>
              <div className="text-xs text-gray-500">Subnets</div>
            </div>
            <div className="flex flex-col items-center text-center">
              <Shield className="h-6 w-6 text-orange-600 mb-2" />
              <div className="text-2xl font-bold">{securityGroups.length}</div>
              <div className="text-xs text-gray-500">Security Groups</div>
            </div>
          </div>

          {vpcs.length > 0 && (
            <div className="space-y-2 pt-2 border-t">
              <div className="text-sm font-medium">VPCs</div>
              <div className="space-y-1">
                {vpcs.slice(0, 3).map((vpc) => (
                  <div key={vpc.id} className="flex items-center justify-between text-sm">
                    <span className="truncate">{vpc.name}</span>
                    <Badge variant={vpc.state === 'available' ? 'default' : 'secondary'} className="ml-2">
                      {vpc.state}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export const NetworkStatusWidget = React.memo(NetworkStatusWidgetComponent);


/**
 * VM Overview Tab Component
 * VM 상세 페이지의 Overview 탭 컴포넌트
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Server, Network, MapPin, Calendar, Monitor } from 'lucide-react';
import { toLocaleDateString, toLocaleTimeString } from '@/lib/utils/date-format';
import { useTranslation } from '@/hooks/use-translation';
import type { VM } from '@/lib/types';

interface VMOverviewTabProps {
  vm: VM;
}

export function VMOverviewTab({ vm }: VMOverviewTabProps) {
  const { locale } = useTranslation();
  
  return (
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
            {toLocaleDateString(vm.created_at, locale as 'ko' | 'en')}
          </div>
          <p className="text-xs text-muted-foreground">
            {toLocaleTimeString(vm.created_at, locale as 'ko' | 'en')}
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
  );
}


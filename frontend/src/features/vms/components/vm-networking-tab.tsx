/**
 * VM Networking Tab Component
 * VM 상세 페이지의 Networking 탭 컴포넌트
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import type { VM } from '@/lib/types';

interface VMNetworkingTabProps {
  vm: VM;
}

export function VMNetworkingTab({ vm }: VMNetworkingTabProps) {
  return (
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
  );
}


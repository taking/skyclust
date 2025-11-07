/**
 * VM Storage Tab Component
 * VM 상세 페이지의 Storage 탭 컴포넌트
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export function VMStorageTab() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Storage Configuration</CardTitle>
        <CardDescription>Attached storage and volumes</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="text-sm font-medium text-gray-700">Root Volume</label>
              <p className="text-lg font-mono">30 GB (gp3)</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-700">Additional Storage</label>
              <p className="text-lg font-mono">100 GB (gp3)</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-700">Total Storage</label>
              <p className="text-lg font-mono">130 GB</p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}


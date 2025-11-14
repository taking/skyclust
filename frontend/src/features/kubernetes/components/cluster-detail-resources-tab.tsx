/**
 * Cluster Detail Resources Tab Component
 * 클러스터 상세 리소스 탭 컴포넌트
 */

'use client';

import { Card, CardContent } from '@/components/ui/card';

export function ClusterDetailResourcesTab() {
  return (
    <Card>
      <CardContent className="py-12 text-center">
        <p className="text-muted-foreground">Not implemented</p>
        <p className="text-sm text-muted-foreground mt-2">
          Kubernetes 리소스 CRUD 기능은 추후 구현 예정입니다.
        </p>
      </CardContent>
    </Card>
  );
}


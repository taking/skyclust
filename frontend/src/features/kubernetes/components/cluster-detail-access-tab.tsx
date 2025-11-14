/**
 * Cluster Detail Access Tab Component
 * 클러스터 상세 액세스 탭 컴포넌트
 */

'use client';

import { Card, CardContent } from '@/components/ui/card';

export function ClusterDetailAccessTab() {
  return (
    <Card>
      <CardContent className="py-12 text-center">
        <p className="text-muted-foreground">Not implemented</p>
        <p className="text-sm text-muted-foreground mt-2">
          IAM 액세스 항목 기능은 추후 구현 예정입니다.
        </p>
      </CardContent>
    </Card>
  );
}


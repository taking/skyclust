/**
 * Cluster Detail Overview Tab Component
 * 클러스터 상세 개요 탭 컴포넌트
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { CopyableText } from '@/components/common/copyable-text';
import type { ProviderCluster, BaseCluster } from '@/lib/types';
import { isAWSCluster } from '@/lib/types';

interface ClusterDetailOverviewTabProps {
  cluster: ProviderCluster | BaseCluster;
}

export function ClusterDetailOverviewTab({ cluster }: ClusterDetailOverviewTabProps) {
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>개요</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* API 서버 엔드포인트 */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground">
              API 서버 엔드포인트
            </label>
            {cluster.endpoint ? (
              <CopyableText text={cluster.endpoint} maxLength={80} />
            ) : (
              <p className="text-sm text-muted-foreground">-</p>
            )}
          </div>

          {/* 클러스터 IAM 역할 ARN (AWS only) */}
          {isAWSCluster(cluster) && (
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">
                클러스터 IAM 역할 ARN
              </label>
              {cluster.role_arn ? (
                <CopyableText text={cluster.role_arn} maxLength={80} />
              ) : (
                <p className="text-sm text-muted-foreground">-</p>
              )}
            </div>
          )}

          {/* 생성일자 */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground">생성일자</label>
            <p className="text-sm">
              {cluster.created_at
                ? new Date(cluster.created_at).toLocaleString('ko-KR', {
                    year: 'numeric',
                    month: '2-digit',
                    day: '2-digit',
                    hour: '2-digit',
                    minute: '2-digit',
                  })
                : '-'}
            </p>
          </div>

          {/* 클러스터 ARN */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground">클러스터 ARN</label>
            {cluster.id ? (
              <CopyableText text={cluster.id} maxLength={80} />
            ) : (
              <p className="text-sm text-muted-foreground">-</p>
            )}
          </div>

          {/* 플랫폼 버전 (AWS only) */}
          {isAWSCluster(cluster) && (
            <div className="space-y-2">
              <label className="text-sm font-medium text-muted-foreground">플랫폼 버전</label>
              <p className="text-sm">{cluster.platform_version || '-'}</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}


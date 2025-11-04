/**
 * Cluster Metrics Tab Component
 * 클러스터 메트릭스 탭 컴포넌트
 */

'use client';

import { ClusterMetricsChart } from './cluster-metrics-chart';
import { NodeMetricsChart } from './node-metrics-chart';
import type { Node } from '@/lib/types';

interface ClusterMetricsTabProps {
  clusterName: string;
  nodes: Node[];
}

export function ClusterMetricsTab({ clusterName, nodes }: ClusterMetricsTabProps) {
  return (
    <div className="space-y-4">
      <ClusterMetricsChart clusterName={clusterName} />
      {nodes.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {nodes.slice(0, 4).map((node) => (
            <NodeMetricsChart key={node.id || node.name} node={node} />
          ))}
        </div>
      )}
    </div>
  );
}


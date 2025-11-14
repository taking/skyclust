/**
 * Cluster Detail Computing Tab Component
 * 클러스터 상세 컴퓨팅 탭 컴포넌트 (노드, 노드 그룹)
 */

'use client';

import { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ClusterNodePoolsTab } from './cluster-node-pools-tab';
import { ClusterNodeGroupsTab } from './cluster-node-groups-tab';
import { ClusterNodesTab } from './cluster-nodes-tab';
import type { NodePool, NodeGroup, Node } from '@/lib/types';

interface ClusterDetailComputingTabProps {
  nodePools: NodePool[];
  nodeGroups: NodeGroup[];
  nodes: Node[];
  isLoadingNodePools: boolean;
  isLoadingNodeGroups: boolean;
  isLoadingNodes: boolean;
  selectedProvider?: string;
  onCreateNodePoolClick?: () => void;
  onCreateNodeGroupClick?: () => void;
  onScaleNodePoolClick?: (nodePoolName: string, currentNodes: number) => void;
  onDeleteNodePoolClick?: (nodePoolName: string) => void;
  onDeleteNodeGroupClick?: (nodeGroupName: string) => void;
  isDeletingNodePool?: boolean;
  isDeletingNodeGroup?: boolean;
}

export function ClusterDetailComputingTab({
  nodePools,
  nodeGroups,
  nodes,
  isLoadingNodePools,
  isLoadingNodeGroups,
  isLoadingNodes,
  selectedProvider,
  onCreateNodePoolClick,
  onCreateNodeGroupClick,
  onScaleNodePoolClick,
  onDeleteNodePoolClick,
  onDeleteNodeGroupClick,
  isDeletingNodePool = false,
  isDeletingNodeGroup = false,
}: ClusterDetailComputingTabProps) {
  const [activeSubTab, setActiveSubTab] = useState<'nodes' | 'nodegroups' | 'nodepools'>('nodes');

  return (
    <Tabs value={activeSubTab} onValueChange={(v) => setActiveSubTab(v as typeof activeSubTab)}>
      <TabsList>
        <TabsTrigger value="nodes">노드</TabsTrigger>
        {selectedProvider === 'aws' && <TabsTrigger value="nodegroups">노드 그룹</TabsTrigger>}
        {selectedProvider !== 'aws' && <TabsTrigger value="nodepools">노드 풀</TabsTrigger>}
      </TabsList>

      <TabsContent value="nodes" className="mt-4">
        <ClusterNodesTab nodes={nodes} isLoading={isLoadingNodes} />
      </TabsContent>

      {selectedProvider === 'aws' && (
        <TabsContent value="nodegroups" className="mt-4">
          <ClusterNodeGroupsTab
            nodeGroups={nodeGroups}
            isLoading={isLoadingNodeGroups}
            onCreateClick={onCreateNodeGroupClick || (() => {})}
            onDeleteClick={onDeleteNodeGroupClick || (() => {})}
            isDeleting={isDeletingNodeGroup}
          />
        </TabsContent>
      )}

      {selectedProvider !== 'aws' && (
        <TabsContent value="nodepools" className="mt-4">
          <ClusterNodePoolsTab
            nodePools={nodePools}
            isLoading={isLoadingNodePools}
            onCreateClick={onCreateNodePoolClick || (() => {})}
            onScaleClick={onScaleNodePoolClick || (() => {})}
            onDeleteClick={onDeleteNodePoolClick || (() => {})}
            isDeleting={isDeletingNodePool}
          />
        </TabsContent>
      )}
    </Tabs>
  );
}


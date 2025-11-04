/**
 * Cluster Bulk Actions Hook
 * 클러스터 일괄 작업 로직 (삭제, 태그 추가)
 */

import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { kubernetesService } from '../services/kubernetes';
import type { KubernetesCluster, CloudProvider } from '@/lib/types';
import { queryKeys } from '@/lib/query-keys';

interface BulkOperationProgress {
  operation: 'delete' | 'tag';
  total: number;
  completed: number;
  failed: number;
  cancelled: number;
  isComplete: boolean;
  isCancelled: boolean;
}

export function useClusterBulkActions(
  selectedProvider?: CloudProvider,
  selectedCredentialId?: string
) {
  const queryClient = useQueryClient();
  const [bulkOperationProgress, setBulkOperationProgress] = useState<BulkOperationProgress | null>(null);
  const [isOperationCancelled, setIsOperationCancelled] = useState(false);

  const handleBulkDelete = async (
    clusterIds: string[],
    clusters: KubernetesCluster[],
    onSuccess?: (message: string) => void,
    onError?: (message: string) => void
  ) => {
    if (!selectedCredentialId || !selectedProvider) return;
    
    const clustersToDelete = clusters.filter(c => clusterIds.includes(c.id || c.name));
    const total = clustersToDelete.length;
    let completed = 0;
    let failed = 0;
    let cancelled = 0;

    setIsOperationCancelled(false);
    setBulkOperationProgress({
      operation: 'delete',
      total,
      completed: 0,
      failed: 0,
      cancelled: 0,
      isComplete: false,
      isCancelled: false,
    });

    const deletePromises = clustersToDelete.map(async (cluster) => {
      if (isOperationCancelled) {
        cancelled++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          cancelled,
          isComplete: completed + failed + cancelled === total,
        } : null);
        return;
      }
      
      try {
        await kubernetesService.deleteCluster(
          selectedProvider,
          cluster.name,
          selectedCredentialId,
          cluster.region
        );
        completed++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          completed,
          isComplete: completed + failed + cancelled === total,
        } : null);
      } catch (error) {
        failed++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          failed,
          isComplete: completed + failed + cancelled === total,
        } : null);
      }
    });

    try {
      await Promise.allSettled(deletePromises);
      
      setBulkOperationProgress(prev => prev ? {
        ...prev,
        isComplete: true,
        isCancelled: isOperationCancelled,
      } : null);
      
      if (isOperationCancelled) {
        onError?.(`Operation cancelled: ${completed} completed, ${failed} failed, ${cancelled} cancelled`);
      } else if (failed === 0) {
        onSuccess?.(`Successfully initiated deletion of ${completed} cluster(s)`);
      } else if (completed > 0) {
        onSuccess?.(`Initiated deletion of ${completed} cluster(s), ${failed} failed`);
      } else {
        onError?.(`Failed to delete all clusters`);
      }
      
      queryClient.invalidateQueries({ queryKey: queryKeys.kubernetesClusters.all });
      
      if (!isOperationCancelled) {
        setTimeout(() => {
          setBulkOperationProgress(null);
        }, 5000);
      }
    } catch (error) {
      onError?.(`Failed to delete some clusters: ${error instanceof Error ? error.message : 'Unknown error'}`);
      setBulkOperationProgress(null);
    }
  };

  const handleBulkTag = async (
    clusterIds: string[],
    clusters: KubernetesCluster[],
    tagKey: string,
    tagValue: string,
    onSuccess?: (message: string) => void,
    onError?: (message: string) => void
  ) => {
    if (!tagKey.trim() || !tagValue.trim() || !selectedProvider || !selectedCredentialId) return;
    
    setIsOperationCancelled(false);
    const clustersToTag = clusters.filter(c => clusterIds.includes(c.id || c.name));
    const total = clustersToTag.length;
    let completed = 0;
    let failed = 0;
    let cancelled = 0;

    setBulkOperationProgress({
      operation: 'tag',
      total,
      completed: 0,
      failed: 0,
      cancelled: 0,
      isComplete: false,
      isCancelled: false,
    });

    const tagPromises = clustersToTag.map(async (cluster) => {
      if (isOperationCancelled) {
        cancelled++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          cancelled,
          isComplete: completed + failed + cancelled === total,
        } : null);
        return;
      }
      
      try {
        const currentTags = cluster.tags || {};
        const updatedTags = {
          ...currentTags,
          [tagKey.trim()]: tagValue.trim(),
        };
        
        await kubernetesService.updateClusterTags(
          selectedProvider,
          cluster.name,
          selectedCredentialId,
          cluster.region,
          updatedTags
        );
        
        completed++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          completed,
          isComplete: completed + failed + (prev.cancelled || 0) === total,
        } : null);
      } catch (error) {
        failed++;
        setBulkOperationProgress(prev => prev ? {
          ...prev,
          failed,
          isComplete: completed + failed + (prev.cancelled || 0) === total,
        } : null);
      }
    });

    try {
      await Promise.allSettled(tagPromises);
      
      setBulkOperationProgress(prev => prev ? {
        ...prev,
        isComplete: true,
        isCancelled: isOperationCancelled,
      } : null);
      
      if (isOperationCancelled) {
        onError?.(`Operation cancelled: ${completed} completed, ${failed} failed, ${cancelled} cancelled`);
      } else if (failed === 0) {
        onSuccess?.(`Successfully added tag "${tagKey}: ${tagValue}" to ${completed} cluster(s)`);
      } else if (completed > 0) {
        onSuccess?.(`Added tag to ${completed} cluster(s), ${failed} failed`);
      } else {
        onError?.(`Failed to add tag to all clusters`);
      }
      
      queryClient.invalidateQueries({ queryKey: queryKeys.kubernetesClusters.all });
      
      if (!isOperationCancelled) {
        setTimeout(() => {
          setBulkOperationProgress(null);
        }, 5000);
      }
    } catch (error) {
      onError?.(`Failed to add tags: ${error instanceof Error ? error.message : 'Unknown error'}`);
      setBulkOperationProgress(null);
    }
  };

  const handleCancelOperation = () => {
    setIsOperationCancelled(true);
    setBulkOperationProgress(prev => prev ? {
      ...prev,
      isCancelled: true,
    } : null);
  };

  const clearProgress = () => {
    setBulkOperationProgress(null);
    setIsOperationCancelled(false);
  };

  return {
    bulkOperationProgress,
    handleBulkDelete,
    handleBulkTag,
    handleCancelOperation,
    clearProgress,
  };
}


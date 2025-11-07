/**
 * useResourceMutation Hook
 * 리소스 타입별 mutation을 자동으로 생성하는 훅
 * 
 * useStandardMutation을 래핑하여 리소스 타입에 따라
 * 자동으로 적절한 query key를 invalidate합니다.
 * 
 * @example
 * ```tsx
 * // VM 생성 mutation
 * const createVM = useResourceMutation({
 *   resourceType: 'vms',
 *   operation: 'create',
 *   mutationFn: (data) => vmService.createVM(data),
 *   successMessage: 'VM created successfully',
 *   onSuccess: () => setIsDialogOpen(false),
 * });
 * 
 * // VPC 삭제 mutation
 * const deleteVPC = useResourceMutation({
 *   resourceType: 'vpcs',
 *   operation: 'delete',
 *   mutationFn: (id) => networkService.deleteVPC(id),
 *   successMessage: 'VPC deleted successfully',
 * });
 * ```
 */

import { useStandardMutation, UseStandardMutationOptions } from './use-standard-mutation';
import { queryKeys } from '@/lib/query-keys';
import type { QueryKey } from '@tanstack/react-query';

/**
 * 리소스 타입
 */
export type ResourceType = 
  | 'vms'
  | 'vpcs'
  | 'subnets'
  | 'securityGroups'
  | 'clusters'
  | 'kubernetesClusters'
  | 'nodePools'
  | 'nodes'
  | 'credentials'
  | 'workspaces';

/**
 * Mutation 작업 타입
 */
export type MutationOperation = 'create' | 'update' | 'delete' | 'custom';

export interface UseResourceMutationOptions<TData, TVariables, TContext = unknown>
  extends Omit<UseStandardMutationOptions<TData, TVariables, TContext>, 'invalidateQueries'> {
  /**
   * 리소스 타입
   */
  resourceType: ResourceType;
  
  /**
   * Mutation 작업 타입
   * - create: 리소스 생성 (리스트 및 전체 무효화)
   * - update: 리소스 업데이트 (리스트, 전체, 상세 무효화)
   * - delete: 리소스 삭제 (리스트 및 전체 무효화)
   * - custom: 수동으로 invalidateQueries 지정 필요
   */
  operation?: MutationOperation;
  
  /**
   * 추가로 무효화할 query keys (operation이 'custom'일 때 필수)
   */
  additionalInvalidateQueries?: readonly QueryKey[];
  
  /**
   * 특정 리소스 ID (update/delete 시 상세 정보도 무효화)
   */
  resourceId?: string;
}

/**
 * 리소스 타입별 query keys 가져오기
 */
function getResourceQueryKeys(resourceType: ResourceType): typeof queryKeys[keyof typeof queryKeys] {
  switch (resourceType) {
    case 'vms':
      return queryKeys.vms;
    case 'vpcs':
      return queryKeys.vpcs;
    case 'subnets':
      return queryKeys.subnets;
    case 'securityGroups':
      return queryKeys.securityGroups;
    case 'clusters':
      return queryKeys.clusters;
    case 'kubernetesClusters':
      return queryKeys.kubernetesClusters;
    case 'nodePools':
      return queryKeys.nodePools;
    case 'nodes':
      return queryKeys.nodes;
    case 'credentials':
      return queryKeys.credentials;
    case 'workspaces':
      return queryKeys.workspaces;
    default:
      throw new Error(`Unknown resource type: ${resourceType}`);
  }
}

/**
 * 작업 타입에 따른 query keys 무효화 목록 생성
 */
function getInvalidateQueries(
  resourceType: ResourceType,
  operation: MutationOperation,
  resourceId?: string,
  additionalInvalidateQueries?: readonly QueryKey[]
): readonly QueryKey[] {
  const resourceKeys = getResourceQueryKeys(resourceType);
  const invalidateQueries: QueryKey[] = [];

  switch (operation) {
    case 'create':
      // 생성 시: 리스트 및 전체 무효화
      if ('all' in resourceKeys) {
        invalidateQueries.push(resourceKeys.all);
      }
      if ('lists' in resourceKeys && typeof resourceKeys.lists === 'function') {
        invalidateQueries.push(resourceKeys.lists());
      }
      break;

    case 'update':
      // 업데이트 시: 리스트, 전체, 상세 정보 무효화
      if ('all' in resourceKeys) {
        invalidateQueries.push(resourceKeys.all);
      }
      if ('lists' in resourceKeys && typeof resourceKeys.lists === 'function') {
        invalidateQueries.push(resourceKeys.lists());
      }
      if (resourceId && 'detail' in resourceKeys && typeof resourceKeys.detail === 'function') {
        invalidateQueries.push(resourceKeys.detail(resourceId));
      }
      break;

    case 'delete':
      // 삭제 시: 리스트 및 전체 무효화
      if ('all' in resourceKeys) {
        invalidateQueries.push(resourceKeys.all);
      }
      if ('lists' in resourceKeys && typeof resourceKeys.lists === 'function') {
        invalidateQueries.push(resourceKeys.lists());
      }
      // 삭제된 리소스의 상세 정보도 무효화
      if (resourceId && 'detail' in resourceKeys && typeof resourceKeys.detail === 'function') {
        invalidateQueries.push(resourceKeys.detail(resourceId));
      }
      break;

    case 'custom':
      // 커스텀: additionalInvalidateQueries 사용
      if (!additionalInvalidateQueries || additionalInvalidateQueries.length === 0) {
        // 기본적으로 all만 무효화
        if ('all' in resourceKeys) {
          invalidateQueries.push(resourceKeys.all);
        }
      }
      break;
  }

  // 추가 query keys 병합
  if (additionalInvalidateQueries && additionalInvalidateQueries.length > 0) {
    invalidateQueries.push(...additionalInvalidateQueries);
  }

  return invalidateQueries;
}

/**
 * useResourceMutation Hook
 * 
 * 리소스 타입별 mutation을 자동으로 생성합니다.
 * 적절한 query key를 자동으로 무효화합니다.
 */
export function useResourceMutation<TData, TVariables, TContext = unknown>({
  resourceType,
  operation = 'create',
  additionalInvalidateQueries,
  resourceId,
  mutationFn,
  successMessage,
  errorContext,
  onSuccess,
  onError,
  mutationOptions,
}: UseResourceMutationOptions<TData, TVariables, TContext>) {
  // 기본 에러 컨텍스트 생성
  const defaultErrorContext: Record<string, unknown> = {
    operation,
    resourceType,
    ...(resourceId && { resourceId }),
    ...errorContext,
  };

  // 작업 타입에 따른 query keys 무효화 목록 생성
  const invalidateQueries = getInvalidateQueries(
    resourceType,
    operation,
    resourceId,
    additionalInvalidateQueries
  );

  return useStandardMutation<TData, TVariables, TContext>({
    mutationFn,
    invalidateQueries,
    successMessage,
    errorContext: defaultErrorContext,
    onSuccess,
    onError,
    mutationOptions,
  });
}


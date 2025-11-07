/**
 * useStandardMutation Hook
 * 
 * 표준화된 mutation wrapper hook
 * 공통 패턴(쿼리 무효화, 성공/에러 처리, 토스트 메시지)을 통합
 * 
 * @example
 * ```tsx
 * const createMutation = useStandardMutation({
 *   mutationFn: (data) => service.create(data),
 *   invalidateQueries: [queryKeys.resources.all],
 *   successMessage: 'Resource created successfully',
 *   onSuccess: () => {
 *     setIsDialogOpen(false);
 *     form.reset();
 *   },
 *   errorContext: { operation: 'createResource', resource: 'Resource' },
 * });
 * ```
 */

import { useMutation, useQueryClient, UseMutationOptions } from '@tanstack/react-query';
import { useToast } from '@/hooks/use-toast';
import { useErrorHandler } from '@/hooks/use-error-handler';
import type { QueryKey } from '@tanstack/react-query';

export interface UseStandardMutationOptions<TData, TVariables, TContext = unknown> {
  /**
   * Mutation 함수
   */
  mutationFn: (variables: TVariables) => Promise<TData>;

  /**
   * 무효화할 쿼리 키 배열 (optional)
   */
  invalidateQueries?: readonly QueryKey[];

  /**
   * 성공 메시지
   */
  successMessage: string;

  /**
   * 에러 컨텍스트 정보 (operation, resource 등)
   */
  errorContext?: Record<string, unknown>;

  /**
   * 성공 시 추가 콜백
   */
  onSuccess?: (data: TData, variables: TVariables) => void;

  /**
   * 에러 시 추가 콜백
   */
  onError?: (error: unknown, variables: TVariables) => void;

  /**
   * React Query mutation 옵션 (추가 설정)
   */
  mutationOptions?: Omit<UseMutationOptions<TData, unknown, TVariables, TContext>, 'mutationFn' | 'onSuccess' | 'onError'> & {
    onSuccess?: (data: TData, variables: TVariables, context: TContext) => void;
    onError?: (error: unknown, variables: TVariables, context: TContext) => void;
  };
}

/**
 * useStandardMutation Hook
 * 
 * 표준화된 mutation을 생성합니다.
 * 자동으로 쿼리 무효화, 성공/에러 처리, 토스트 메시지를 처리합니다.
 */
export function useStandardMutation<TData, TVariables, TContext = unknown>({
  mutationFn,
  invalidateQueries,
  successMessage,
  errorContext,
  onSuccess,
  onError,
  mutationOptions,
}: UseStandardMutationOptions<TData, TVariables, TContext>) {
  const queryClient = useQueryClient();
  const { success } = useToast();
  const { handleError } = useErrorHandler();

  return useMutation<TData, unknown, TVariables, TContext>({
    mutationFn,
    ...mutationOptions,
    onSuccess: (data, variables, context) => {
      // Query invalidation
      if (invalidateQueries && invalidateQueries.length > 0) {
        invalidateQueries.forEach(queryKey => {
          queryClient.invalidateQueries({ queryKey });
        });
      }

      // Success message
      success(successMessage);

      // Custom success callback
      onSuccess?.(data, variables);

      // Call original onSuccess if provided in mutationOptions
      if (mutationOptions && 'onSuccess' in mutationOptions && mutationOptions.onSuccess) {
        mutationOptions.onSuccess(data, variables, context as TContext);
      }
    },
    onError: (error, variables, context) => {
      // Standard error handling
      handleError(error, errorContext);

      // Custom error callback
      onError?.(error, variables);

      // Call original onError if provided in mutationOptions
      if (mutationOptions && 'onError' in mutationOptions && mutationOptions.onError) {
        mutationOptions.onError(error, variables, context as TContext);
      }
    },
  });
}


/**
 * useToast 훅
 * 
 * react-hot-toast를 래핑한 토스트 알림 훅입니다.
 * 성공, 에러, 로딩, Promise 상태에 따른 토스트 메시지를 쉽게 표시할 수 있습니다.
 * 
 * @example
 * ```tsx
 * const { success, error, loading, promise } = useToast();
 * 
 * // 성공 메시지
 * success('작업이 완료되었습니다.');
 * 
 * // 에러 메시지
 * error('작업 중 오류가 발생했습니다.');
 * 
 * // 로딩 메시지
 * const toastId = loading('처리 중...');
 * 
 * // Promise 기반 토스트
 * promise(
 *   fetchData(),
 *   {
 *     loading: '데이터를 불러오는 중...',
 *     success: '데이터를 성공적으로 불러왔습니다.',
 *     error: '데이터를 불러오는데 실패했습니다.',
 *   }
 * );
 * ```
 */
import toast from 'react-hot-toast';

export const useToast = () => {
  /**
   * 성공 토스트 메시지 표시
   * @param message - 표시할 메시지
   */
  const success = (message: string) => {
    toast.success(message);
  };

  /**
   * 에러 토스트 메시지 표시
   * @param message - 표시할 메시지
   */
  const error = (message: string) => {
    toast.error(message);
  };

  /**
   * 로딩 토스트 메시지 표시
   * @param message - 표시할 메시지
   * @returns 토스트 ID (나중에 dismiss할 때 사용)
   */
  const loading = (message: string) => {
    return toast.loading(message);
  };

  /**
   * 토스트 메시지 닫기
   * @param toastId - 닫을 토스트 ID (없으면 모든 토스트 닫기)
   */
  const dismiss = (toastId?: string) => {
    if (toastId) {
      // 특정 토스트만 닫기
      toast.dismiss(toastId);
    } else {
      // 모든 토스트 닫기
      toast.dismiss();
    }
  };

  /**
   * Promise 기반 토스트 메시지
   * Promise의 상태에 따라 자동으로 로딩/성공/에러 메시지를 표시합니다.
   * 
   * @param promise - 실행할 Promise
   * @param messages - 각 상태별 메시지
   * @returns Promise (원본 Promise와 동일)
   * 
   * @example
   * ```tsx
   * promise(
   *   fetchData(),
   *   {
   *     loading: '데이터를 불러오는 중...',
   *     success: '데이터를 성공적으로 불러왔습니다.',
   *     error: '데이터를 불러오는데 실패했습니다.',
   *   }
   * );
   * ```
   */
  const promise = <T>(
    promise: Promise<T>,
    messages: {
      loading: string;
      success: string;
      error: string;
    }
  ) => {
    return toast.promise(promise, messages);
  };

  return {
    success,
    error,
    loading,
    dismiss,
    promise,
  };
};


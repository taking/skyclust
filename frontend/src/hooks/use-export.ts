/**
 * Export Hook
 * 데이터 내보내기 관련 React Query 훅
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { exportService } from '@/services/export';
import type { ExportRequest, ExportResult } from '@/lib/types/export';
import { toast } from 'react-hot-toast';
import { queryKeys } from '@/lib/query-keys';
import { CACHE_TIMES, GC_TIMES } from '@/lib/query-client';

export const useExportData = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: Omit<ExportRequest, 'user_id'>) => exportService.exportData(request),
    onSuccess: (data: ExportResult) => {
      toast.success(`내보내기가 시작되었습니다. (ID: ${data.id})`);
      queryClient.invalidateQueries({ queryKey: queryKeys.exports.history() });
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '내보내기 요청에 실패했습니다.';
      toast.error(message);
    },
  });
};

export const useExportStatus = (exportId: string, enabled: boolean = true) => {
  return useQuery({
    queryKey: queryKeys.exports.status(exportId),
    queryFn: () => exportService.getExportStatus(exportId),
    enabled: enabled && !!exportId,
    staleTime: 0, // 항상 최신 상태 확인 필요
    gcTime: GC_TIMES.SHORT, // 5 minutes - GC 시간
    refetchInterval: 2000, // 2초마다 refetch (진행 상황 추적)
    refetchIntervalInBackground: false, // 백그라운드 polling 비활성화
  });
};

export const useExportHistory = (limit: number = 20, offset: number = 0) => {
  return useQuery({
    queryKey: queryKeys.exports.history(),
    queryFn: () => exportService.getExportHistory(limit, offset),
    staleTime: CACHE_TIMES.MONITORING, // 1 minute - 히스토리는 비교적 안정적
    gcTime: GC_TIMES.MEDIUM, // 10 minutes - GC 시간
  });
};

export const useSupportedFormats = () => {
  return useQuery({
    queryKey: [...queryKeys.exports.all, 'formats'],
    queryFn: () => exportService.getSupportedFormats(),
    staleTime: CACHE_TIMES.STATIC, // 30 minutes - 지원 형식은 거의 변경되지 않음
    gcTime: GC_TIMES.LONG, // 30 minutes - GC 시간 (1시간 대신 30분으로 조정)
  });
};

export const useDownloadExport = () => {
  return useMutation({
    mutationFn: (exportId: string) => exportService.downloadExport(exportId),
    onSuccess: (blob: Blob, exportId: string) => {
      const filename = `export_${exportId}.${blob.type.split('/')[1]}`;
      exportService.downloadFile(blob, filename);
      toast.success('파일이 다운로드되었습니다.');
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '파일 다운로드에 실패했습니다.';
      toast.error(message);
    },
  });
};


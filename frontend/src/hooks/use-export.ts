/**
 * Export Hook
 * 데이터 내보내기 관련 React Query 훅
 */

import { useQuery } from '@tanstack/react-query';
import { exportService } from '@/services/export';
import type { ExportRequest, ExportResult } from '@/lib/types/export';
import { queryKeys, CACHE_TIMES, GC_TIMES } from '@/lib/query';
import { useStandardMutation } from './use-standard-mutation';

export const useExportData = () => {
  return useStandardMutation<ExportResult, Omit<ExportRequest, 'user_id'>>({
    mutationFn: (request: Omit<ExportRequest, 'user_id'>) => exportService.exportData(request),
    invalidateQueries: [
      queryKeys.exports.history(),
    ],
    successMessage: (data: ExportResult) => `내보내기가 시작되었습니다. (ID: ${data.id})`,
    errorContext: { operation: 'exportData', resource: 'Export' },
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
  return useStandardMutation<Blob, string>({
    mutationFn: (exportId: string) => exportService.downloadExport(exportId),
    invalidateQueries: [],
    successMessage: '파일이 다운로드되었습니다.',
    errorContext: { operation: 'downloadExport', resource: 'Export' },
    onSuccess: (blob: Blob, exportId: string) => {
      const filename = `export_${exportId}.${blob.type.split('/')[1]}`;
      exportService.downloadFile(blob, filename);
    },
  });
};


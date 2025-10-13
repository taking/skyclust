/**
 * Export Hook
 * 데이터 내보내기 관련 React Query 훅
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { exportService, ExportRequest, ExportResult } from '@/services/export';
import { toast } from 'react-hot-toast';

export const useExportData = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: Omit<ExportRequest, 'user_id'>) => exportService.exportData(request),
    onSuccess: (data: ExportResult) => {
      toast.success(`내보내기가 시작되었습니다. (ID: ${data.id})`);
      queryClient.invalidateQueries({ queryKey: ['exports', 'history'] });
    },
    onError: (error: unknown) => {
      const message = (error as { response?: { data?: { error?: string } } })?.response?.data?.error || '내보내기 요청에 실패했습니다.';
      toast.error(message);
    },
  });
};

export const useExportStatus = (exportId: string, enabled: boolean = true) => {
  return useQuery({
    queryKey: ['exports', 'status', exportId],
    queryFn: () => exportService.getExportStatus(exportId),
    enabled: enabled && !!exportId,
    refetchInterval: 2000,
  });
};

export const useExportHistory = (limit: number = 20, offset: number = 0) => {
  return useQuery({
    queryKey: ['exports', 'history', limit, offset],
    queryFn: () => exportService.getExportHistory(limit, offset),
  });
};

export const useSupportedFormats = () => {
  return useQuery({
    queryKey: ['exports', 'formats'],
    queryFn: () => exportService.getSupportedFormats(),
    staleTime: 5 * 60 * 1000, // 5분간 캐시
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

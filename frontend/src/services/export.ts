/**
 * Export Service
 * 데이터 내보내기 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import api from '@/lib/api';
import type {
  ExportRequest,
  ExportResult,
  SupportedFormats,
  ExportHistory,
} from '@/lib/types/export';

class ExportService extends BaseService {
  // 데이터 내보내기
  async exportData(request: Omit<ExportRequest, 'user_id'>): Promise<ExportResult> {
    return this.post<ExportResult>('/exports', request);
  }

  // 내보내기 상태 조회
  async getExportStatus(exportId: string): Promise<ExportResult> {
    return this.get<ExportResult>(`/exports/${exportId}/status`);
  }

  // 내보내기 파일 다운로드 (Blob 반환)
  async downloadExport(exportId: string): Promise<Blob> {
    const response = await api.get(`/exports/${exportId}/download`, {
      responseType: 'blob',
    });
    return response.data;
  }

  // 내보내기 이력 조회
  async getExportHistory(limit: number = 20, offset: number = 0): Promise<ExportHistory> {
    return this.get<ExportHistory>(`/exports/history?limit=${limit}&offset=${offset}`);
  }

  // 지원되는 형식 조회
  async getSupportedFormats(): Promise<SupportedFormats> {
    return this.get<SupportedFormats>('/exports/formats');
  }
}

// Utility functions (서비스와 별도로 export)
export const exportUtils = {
  // 파일 다운로드 헬퍼
  downloadFile: (blob: Blob, filename: string) => {
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  },

  // 파일 크기 포맷팅
  formatFileSize: (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  },

  // 상태별 색상 반환
  getStatusColor: (status: string): string => {
    switch (status) {
      case 'completed':
        return 'text-green-600';
      case 'processing':
        return 'text-blue-600';
      case 'pending':
        return 'text-yellow-600';
      case 'failed':
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  },

  // 상태별 배경 색상 반환
  getStatusBgColor: (status: string): string => {
    switch (status) {
      case 'completed':
        return 'bg-green-100';
      case 'processing':
        return 'bg-blue-100';
      case 'pending':
        return 'bg-yellow-100';
      case 'failed':
        return 'bg-red-100';
      default:
        return 'bg-gray-100';
    }
  },
};

export const exportService = Object.assign(
  new ExportService(),
  exportUtils
);

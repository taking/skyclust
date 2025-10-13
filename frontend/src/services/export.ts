/**
 * Export Service
 * 데이터 내보내기 관련 API 호출
 */

import { api } from '@/lib/api';

export interface ExportRequest {
  user_id: string;
  workspace_id?: string;
  type: 'vms' | 'workspaces' | 'credentials' | 'audit_logs' | 'costs';
  format: 'csv' | 'json' | 'xlsx' | 'pdf';
  filters?: Record<string, unknown>;
  date_from?: string;
  date_to?: string;
  include_deleted?: boolean;
}

export interface ExportResult {
  id: string;
  user_id: string;
  type: string;
  format: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  file_size?: number;
  download_url?: string;
  error?: string;
  created_at: string;
  completed_at?: string;
}

export interface ExportFormat {
  format: string;
  name: string;
  description: string;
  mime_type: string;
}

export interface ExportType {
  type: string;
  name: string;
  description: string;
}

export interface SupportedFormats {
  formats: ExportFormat[];
  types: ExportType[];
}

export interface ExportHistory {
  exports: ExportResult[];
  total: number;
  limit: number;
  offset: number;
}

export const exportService = {
  // 데이터 내보내기
  async exportData(request: Omit<ExportRequest, 'user_id'>): Promise<ExportResult> {
    const response = await api.post('/exports', request);
    return response.data.data;
  },

  // 내보내기 상태 조회
  async getExportStatus(exportId: string): Promise<ExportResult> {
    const response = await api.get(`/exports/${exportId}/status`);
    return response.data.data;
  },

  // 내보내기 파일 다운로드
  async downloadExport(exportId: string): Promise<Blob> {
    const response = await api.get(`/exports/${exportId}/download`, {
      responseType: 'blob',
    });
    return response.data;
  },

  // 내보내기 이력 조회
  async getExportHistory(limit: number = 20, offset: number = 0): Promise<ExportHistory> {
    const response = await api.get(`/exports/history?limit=${limit}&offset=${offset}`);
    return response.data.data;
  },

  // 지원되는 형식 조회
  async getSupportedFormats(): Promise<SupportedFormats> {
    const response = await api.get('/exports/formats');
    return response.data.data;
  },

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

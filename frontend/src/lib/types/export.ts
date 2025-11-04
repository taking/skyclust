/**
 * Export 관련 타입 정의
 */

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


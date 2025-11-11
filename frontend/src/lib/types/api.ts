/**
 * API 공통 타입 정의
 */

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
  code?: string;
  meta?: {
    page?: number;
    limit?: number;
    total?: number;
    total_pages?: number;
    has_next?: boolean;
    has_prev?: boolean;
  };
}


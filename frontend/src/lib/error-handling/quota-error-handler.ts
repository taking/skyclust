/**
 * Quota Error Handler
 * GPU quota 관련 에러 처리 유틸리티
 */

import { ServerError } from './types';
import { ServiceError } from '@/lib/types/service';

export interface GPUQuotaErrorDetails {
  instance_type: string;
  region: string;
  quota_code: string;
  current_quota: number;
  current_usage?: number;
  available_quota: number;
  required_count: number;
  quota_increase_url: string;
  available_regions?: Array<{
    region: string;
    available_quota: number;
    quota_value: number;
    current_usage?: number;
  }>;
}

/**
 * 에러가 GPU quota 에러인지 확인
 */
export function isGPUQuotaError(error: unknown): boolean {
  if (error instanceof ServiceError) {
    return error.code === 'PROVIDER_QUOTA_EXCEEDED';
  }
  if (error instanceof ServerError) {
    return error.code === 'PROVIDER_QUOTA_EXCEEDED';
  }
  if (error && typeof error === 'object' && 'code' in error) {
    return (error as { code?: string }).code === 'PROVIDER_QUOTA_EXCEEDED';
  }
  return false;
}

/**
 * GPU quota 에러의 details 추출
 */
export function extractGPUQuotaErrorDetails(error: unknown): GPUQuotaErrorDetails | null {
  if (!isGPUQuotaError(error)) {
    return null;
  }

  let details: Record<string, unknown> | undefined;

  if (error instanceof ServiceError) {
    details = error.details as Record<string, unknown> | undefined;
  } else if (error instanceof ServerError) {
    details = error.details;
  } else if (error && typeof error === 'object' && 'details' in error) {
    details = (error as { details?: Record<string, unknown> }).details;
  }

  if (!details) {
    return null;
  }

  // Details에서 GPU quota 정보 추출
  const quotaDetails: GPUQuotaErrorDetails = {
    instance_type: String(details.instance_type || ''),
    region: String(details.region || ''),
    quota_code: String(details.quota_code || ''),
    current_quota: Number(details.current_quota || 0),
    current_usage: details.current_usage !== undefined ? Number(details.current_usage) : undefined,
    available_quota: Number(details.available_quota || 0),
    required_count: Number(details.required_count || 0),
    quota_increase_url: String(details.quota_increase_url || ''),
  };

  // 사용 가능한 region 목록 추가
  if (details.available_regions && Array.isArray(details.available_regions)) {
    quotaDetails.available_regions = details.available_regions.map((region: unknown) => {
      if (region && typeof region === 'object') {
        const r = region as Record<string, unknown>;
        return {
          region: String(r.region || ''),
          available_quota: Number(r.available_quota || 0),
          quota_value: Number(r.quota_value || 0),
          current_usage: r.current_usage !== undefined ? Number(r.current_usage) : undefined,
        };
      }
      return null;
    }).filter((r): r is NonNullable<typeof r> => r !== null);
  }

  return quotaDetails;
}

/**
 * GPU quota 에러 메시지 생성
 */
export function getGPUQuotaErrorMessage(error: unknown): string {
  if (!isGPUQuotaError(error)) {
    return '';
  }

  // ServiceError나 ServerError에서 메시지 추출
  if (error instanceof Error) {
    return error.message;
  }

  if (error && typeof error === 'object' && 'message' in error) {
    return String((error as { message?: string }).message || '');
  }

  return 'GPU quota exceeded';
}


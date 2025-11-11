/**
 * API Configuration
 * API 버전 관리 및 설정 중앙화
 */

export const API_CONFIG = {
  BASE_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  API_PREFIX: '/api',
  VERSION: process.env.NEXT_PUBLIC_API_VERSION || 'v1',
  TIMEOUT: 30000, // 30 seconds
} as const;

/**
 * API URL 생성
 * 버전 관리 및 endpoint 정규화
 * 
 * @param endpoint - API endpoint (예: '/vms', 'vms', '/workspaces/123/members')
 * @param version - API 버전 (기본값: API_CONFIG.VERSION)
 * @returns 완전한 API URL
 * 
 * @example
 * getApiUrl('/vms') // '/api/v1/vms'
 * getApiUrl('vms') // '/api/v1/vms'
 * getApiUrl('/api/v1/vms') // '/api/v1/vms' (이미 완전한 URL인 경우)
 * getApiUrl('/vms', 'v2') // '/api/v2/vms'
 */
export function getApiUrl(endpoint: string, version?: string): string {
  const apiVersion = version || API_CONFIG.VERSION;
  
  // 이미 완전한 URL인 경우 (절대 경로로 시작하는 경우)
  if (endpoint.startsWith('/api/')) {
    return endpoint;
  }
  
  // endpoint 정규화 (앞뒤 슬래시 제거 후 정리)
  const normalizedEndpoint = endpoint
    .replace(/^\/+|\/+$/g, '') // 앞뒤 슬래시 제거
    .replace(/\/+/g, '/'); // 중복 슬래시 제거
  
  // endpoint가 비어있으면 버전만 반환
  if (!normalizedEndpoint) {
    return `${API_CONFIG.API_PREFIX}/${apiVersion}`;
  }
  
  return `${API_CONFIG.API_PREFIX}/${apiVersion}/${normalizedEndpoint}`;
}

/**
 * API 버전 확인
 */
export function getApiVersion(): string {
  return API_CONFIG.VERSION;
}

/**
 * API Base URL 확인
 */
export function getApiBaseUrl(): string {
  return API_CONFIG.BASE_URL;
}


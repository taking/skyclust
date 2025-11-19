/**
 * Routing Helpers
 * 개선된 RESTful URL 구조를 위한 헬퍼 함수들
 * 
 * 새로운 구조:
 * - Workspace Management: /w/{workspaceId}/{managementType}
 * - Workspace Resources: /w/{workspaceId}/workspaces/{id}/{subPath}
 * - Credential Resources: /w/{workspaceId}/c/{credentialId}/{resourceType}/*
 * - Filters: Query Parameter로 처리
 */

/**
 * Query String 생성 (빈 값 제외)
 */
function buildQueryString(filters?: Record<string, string | null | undefined>): string {
  if (!filters) return '';
  
  const params = new URLSearchParams();
  Object.entries(filters).forEach(([key, value]) => {
    if (value !== null && value !== undefined && value !== '') {
      params.set(key, value);
    }
  });
  
  const queryString = params.toString();
  return queryString ? `?${queryString}` : '';
}

/**
 * 리소스 타입 정의
 */
export type ResourceType = 'k8s' | 'networks' | 'compute' | 'azure';
export type ManagementType = 
  | 'dashboard' 
  | 'credentials' 
  | 'workspaces' 
  | 'profile' 
  | 'notifications' 
  | 'exports' 
  | 'cost-analysis';

/**
 * Workspace Management Path 생성
 * 
 * @example
 * buildWorkspaceManagementPath('ws-1', 'dashboard', { credentialId: 'cred-1', region: 'ap-northeast-3' })
 * // => '/w/ws-1/dashboard?credentialId=cred-1&region=ap-northeast-3'
 */
export function buildWorkspaceManagementPath(
  workspaceId: string,
  managementType: ManagementType,
  filters?: Record<string, string | null | undefined>
): string {
  const basePath = `/w/${workspaceId}/${managementType}`;
  const queryString = buildQueryString(filters);
  return `${basePath}${queryString}`;
}

/**
 * Workspace Resource Path 생성
 * 
 * @example
 * buildWorkspaceResourcePath('ws-1', 'ws-1', 'settings')
 * // => '/w/ws-1/workspaces/ws-1/settings'
 */
export function buildWorkspaceResourcePath(
  workspaceId: string,
  workspaceResourceId: string,
  subPath: 'overview' | 'settings' | 'members' | 'credentials'
): string {
  if (subPath === 'overview') {
    return `/w/${workspaceId}/workspaces/${workspaceResourceId}`;
  }
  return `/w/${workspaceId}/workspaces/${workspaceResourceId}/${subPath}`;
}

/**
 * Workspace Detail Path 생성
 * 
 * @example
 * buildWorkspaceDetailPath('ws-1', 'ws-1')
 * // => '/w/ws-1/workspaces/ws-1'
 */
export function buildWorkspaceDetailPath(
  workspaceId: string,
  workspaceResourceId: string
): string {
  return `/w/${workspaceId}/workspaces/${workspaceResourceId}`;
}

/**
 * Credential Resource Path 생성
 * 
 * @example
 * buildCredentialResourcePath('ws-1', 'cred-1', 'k8s', '/clusters', { region: 'ap-northeast-3' })
 * // => '/w/ws-1/c/cred-1/k8s/clusters?region=ap-northeast-3'
 */
export function buildCredentialResourcePath(
  workspaceId: string,
  credentialId: string,
  resourceType: ResourceType,
  resourcePath: string,
  filters?: Record<string, string | null | undefined>
): string {
  const normalizedPath = resourcePath.startsWith('/') ? resourcePath : `/${resourcePath}`;
  const basePath = `/w/${workspaceId}/c/${credentialId}/${resourceType}${normalizedPath}`;
  const queryString = buildQueryString(filters);
  return `${basePath}${queryString}`;
}

/**
 * Credential Resource Detail Path 생성
 * 
 * @example
 * buildCredentialResourceDetailPath('ws-1', 'cred-1', 'k8s', 'clusters', 'my-cluster', { region: 'ap-northeast-3' })
 * // => '/w/ws-1/c/cred-1/k8s/clusters/my-cluster?region=ap-northeast-3'
 */
export function buildCredentialResourceDetailPath(
  workspaceId: string,
  credentialId: string,
  resourceType: ResourceType,
  resourceCollection: string,
  resourceId: string,
  filters?: Record<string, string | null | undefined>
): string {
  return buildCredentialResourcePath(
    workspaceId,
    credentialId,
    resourceType,
    `/${resourceCollection}/${resourceId}`,
    filters
  );
}

/**
 * Credential Resource Create Path 생성
 * 
 * @example
 * buildCredentialResourceCreatePath('ws-1', 'cred-1', 'k8s', 'clusters', { region: 'ap-northeast-3' })
 * // => '/w/ws-1/c/cred-1/k8s/clusters/create?region=ap-northeast-3'
 */
export function buildCredentialResourceCreatePath(
  workspaceId: string,
  credentialId: string,
  resourceType: ResourceType,
  resourceCollection: string,
  filters?: Record<string, string | null | undefined>
): string {
  return buildCredentialResourcePath(
    workspaceId,
    credentialId,
    resourceType,
    `/${resourceCollection}/create`,
    filters
  );
}

/**
 * 현재 경로에서 필터만 업데이트한 새 경로 생성
 * 
 * @example
 * updatePathFilters('/w/ws-1/c/cred-1/k8s/clusters?region=ap-northeast-3', { region: 'ap-northeast-2' })
 * // => '/w/ws-1/c/cred-1/k8s/clusters?region=ap-northeast-2'
 */
export function updatePathFilters(
  currentPath: string,
  newFilters: Record<string, string | null | undefined>
): string {
  const [path, existingQuery] = currentPath.split('?');
  const existingParams = existingQuery ? new URLSearchParams(existingQuery) : new URLSearchParams();
  
  // 기존 필터 업데이트
  Object.entries(newFilters).forEach(([key, value]) => {
    if (value === null || value === undefined || value === '') {
      existingParams.delete(key);
    } else {
      existingParams.set(key, value);
    }
  });
  
  const queryString = existingParams.toString();
  return queryString ? `${path}?${queryString}` : path;
}

// ==========================================
// Legacy 함수들 (하위 호환성 유지 - 기존 코드 마이그레이션용)
// ==========================================

/**
 * @deprecated Use buildCredentialResourcePath instead
 */
export function buildResourcePath(
  workspaceId: string,
  credentialId: string,
  resourceType: 'kubernetes' | 'networks' | 'compute' | 'azure',
  resourcePath: string,
  filters?: Record<string, string | null | undefined>
): string {
  // kubernetes -> k8s 변환
  const newResourceType: ResourceType = resourceType === 'kubernetes' ? 'k8s' : resourceType as ResourceType;
  return buildCredentialResourcePath(workspaceId, credentialId, newResourceType, resourcePath, filters);
}

/**
 * @deprecated Use buildWorkspaceManagementPath instead
 */
export function buildManagementPath(
  workspaceId: string,
  managementType: string,
  filters?: Record<string, string | null | undefined>
): string {
  return buildWorkspaceManagementPath(workspaceId, managementType as ManagementType, filters);
}

/**
 * @deprecated Use buildWorkspaceResourcePath instead
 */
export function buildWorkspacePath(
  workspaceId: string,
  subPath: 'settings' | 'members'
): string {
  return buildWorkspaceResourcePath(workspaceId, workspaceId, subPath);
}

/**
 * @deprecated Use buildCredentialResourceDetailPath instead
 */
export function buildResourceDetailPath(
  workspaceId: string,
  credentialId: string,
  resourceType: 'kubernetes' | 'networks' | 'compute' | 'azure',
  resourceCollection: string,
  resourceId: string,
  filters?: Record<string, string | null | undefined>
): string {
  const newResourceType: ResourceType = resourceType === 'kubernetes' ? 'k8s' : resourceType as ResourceType;
  return buildCredentialResourceDetailPath(workspaceId, credentialId, newResourceType, resourceCollection, resourceId, filters);
}

/**
 * @deprecated Use buildCredentialResourceCreatePath instead
 */
export function buildResourceCreatePath(
  workspaceId: string,
  credentialId: string,
  resourceType: 'kubernetes' | 'networks' | 'compute' | 'azure',
  resourceCollection: string,
  filters?: Record<string, string | null | undefined>
): string {
  const newResourceType: ResourceType = resourceType === 'kubernetes' ? 'k8s' : resourceType as ResourceType;
  return buildCredentialResourceCreatePath(workspaceId, credentialId, newResourceType, resourceCollection, filters);
}

/**
 * 인증/설정 페이지 경로 생성
 * 
 * @example
 * buildAuthPath('login', { returnUrl: '/w/ws-1/dashboard' })
 * // => '/login?returnUrl=%2Fw%2Fws-1%2Fdashboard'
 */
export function buildAuthPath(
  authType: 'login' | 'register' | 'setup',
  filters?: Record<string, string | null | undefined>
): string {
  const basePath = `/${authType}`;
  const queryString = buildQueryString(filters);
  return `${basePath}${queryString}`;
}


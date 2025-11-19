/**
 * Workspace 전환 시 스마트 리다이렉트 로직
 * 
 * 페이지 타입에 따라 적절한 경로로 리다이렉트
 */

import { buildWorkspaceManagementPath, buildCredentialResourcePath, type ResourceType, type ManagementType } from './helpers';

/**
 * 페이지 타입 정의
 */
export type PageType = 
  | 'management'      // Dashboard, Workspaces, Credentials 등
  | 'resource-list'   // Clusters, VPCs, VMs 목록
  | 'resource-detail' // 특정 클러스터, VPC 상세
  | 'create'          // 리소스 생성 페이지
  | 'workspace-detail'; // Workspace 상세 페이지

/**
 * 경로 분석 결과
 */
export interface PathAnalysis {
  pageType: PageType;
  workspaceId?: string;
  credentialId?: string;
  resourceType?: ResourceType;
  resourceCollection?: string;
  resourceId?: string;
  managementType?: ManagementType;
  workspaceResourceId?: string;
  subPath?: string;
  queryParams?: string;
}

/**
 * 현재 경로를 분석하여 페이지 타입과 정보 추출
 * 
 * @param currentPath - 현재 경로 (예: '/w/ws-1/c/cred-1/k8s/clusters')
 * @returns 경로 분석 결과
 */
export function analyzePath(currentPath: string): PathAnalysis {
  // Query parameter 분리
  const [path, queryParams] = currentPath.split('?');
  
  // 경로 세그먼트 분리
  const segments = path.split('/').filter(Boolean);
  
  const analysis: PathAnalysis = {
    pageType: 'management',
    queryParams: queryParams || undefined,
  };

  // /w/{workspaceId} 패턴 확인
  if (segments[0] === 'w' && segments[1]) {
    analysis.workspaceId = segments[1];
    
    // Management 페이지: /w/{workspaceId}/{managementType}
    if (segments.length === 2 || (segments.length === 3 && !segments[2].match(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i))) {
      const managementType = segments[2] as ManagementType;
      if (['dashboard', 'workspaces', 'credentials', 'profile', 'notifications', 'exports', 'cost-analysis'].includes(managementType)) {
        analysis.pageType = 'management';
        analysis.managementType = managementType;
        return analysis;
      }
    }
    
    // Workspace Resource 페이지: /w/{workspaceId}/workspaces/{id}/{subPath}
    if (segments[2] === 'workspaces' && segments[3]) {
      analysis.pageType = 'workspace-detail';
      analysis.workspaceResourceId = segments[3];
      analysis.subPath = segments[4] as 'overview' | 'settings' | 'members' | 'credentials' | undefined;
      return analysis;
    }
    
    // Credential Resource 페이지: /w/{workspaceId}/c/{credentialId}/{resourceType}/{collection}...
    if (segments[2] === 'c' && segments[3] && segments[4]) {
      analysis.credentialId = segments[3];
      
      // Resource Type 확인
      const resourceType = segments[4] as ResourceType;
      if (['k8s', 'networks', 'compute', 'azure'].includes(resourceType)) {
        analysis.resourceType = resourceType;
        
        // Resource Collection 확인
        if (segments[5]) {
          analysis.resourceCollection = segments[5];
          
          // Create 페이지: .../{collection}/create
          if (segments[6] === 'create') {
            analysis.pageType = 'create';
            return analysis;
          }
          
          // Resource Detail 페이지: .../{collection}/{resourceId}
          if (segments[6] && segments[6] !== 'create') {
            analysis.pageType = 'resource-detail';
            analysis.resourceId = segments[6];
            return analysis;
          }
          
          // Resource List 페이지: .../{collection}
          analysis.pageType = 'resource-list';
          return analysis;
        }
      }
    }
  }
  
  // 기본값: Management (Dashboard)
  analysis.pageType = 'management';
  analysis.managementType = 'dashboard';
  return analysis;
}

/**
 * Workspace 전환 시 적절한 경로 생성 (동기 버전)
 * 
 * @param currentPath - 현재 경로
 * @param newWorkspaceId - 새로운 Workspace ID
 * @param firstCredential - 첫 번째 credential (선택적, 있으면 resource-list 페이지에서 사용)
 * @returns 새로운 경로
 */
export function getWorkspaceRedirectPath(
  currentPath: string,
  newWorkspaceId: string,
  firstCredential?: { id: string; provider: string } | null
): string {
  const analysis = analyzePath(currentPath);
  
  switch (analysis.pageType) {
    case 'management':
      // Management 페이지: 동일한 경로로 이동 (workspaceId만 교체)
      if (analysis.managementType) {
        const queryString = analysis.queryParams ? `?${analysis.queryParams}` : '';
        return buildWorkspaceManagementPath(newWorkspaceId, analysis.managementType) + queryString;
      }
      // 기본값: Dashboard
      return buildWorkspaceManagementPath(newWorkspaceId, 'dashboard');
    
    case 'workspace-detail':
      // Workspace 상세 페이지: 동일한 경로로 이동 (workspaceId만 교체)
      // workspaceResourceId는 보통 workspaceId와 동일하므로 newWorkspaceId 사용
      if (analysis.workspaceResourceId) {
        const basePath = `/w/${newWorkspaceId}/workspaces/${newWorkspaceId}`;
        const subPath = analysis.subPath && analysis.subPath !== 'overview' ? `/${analysis.subPath}` : '';
        const queryString = analysis.queryParams ? `?${analysis.queryParams}` : '';
        return `${basePath}${subPath}${queryString}`;
      }
      return buildWorkspaceManagementPath(newWorkspaceId, 'dashboard');
    
    case 'resource-list':
      // Resource 목록 페이지: 첫 번째 credential으로 동일한 리소스 타입 페이지로 이동
      if (analysis.resourceType && analysis.resourceCollection && firstCredential) {
        const queryString = analysis.queryParams ? `?${analysis.queryParams}` : '';
        return buildCredentialResourcePath(
          newWorkspaceId,
          firstCredential.id,
          analysis.resourceType,
          `/${analysis.resourceCollection}`,
          undefined
        ) + queryString;
      }
      // Credential이 없으면 Dashboard로 이동
      return buildWorkspaceManagementPath(newWorkspaceId, 'dashboard');
    
    case 'resource-detail':
    case 'create':
      // Resource 상세/생성 페이지: Dashboard로 이동 (안전)
      return buildWorkspaceManagementPath(newWorkspaceId, 'dashboard');
    
    default:
      // 기본값: Dashboard
      return buildWorkspaceManagementPath(newWorkspaceId, 'dashboard');
  }
}



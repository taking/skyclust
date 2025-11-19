/**
 * useResourceContext Hook
 * 개선된 URL 구조에서 workspaceId, credentialId를 추출하는 Hook
 * 
 * 새로운 구조:
 * - Workspace Management: /w/{workspaceId}/{managementType}
 * - Workspace Resources: /w/{workspaceId}/workspaces/{id}/{subPath}
 * - Credential Resources: /w/{workspaceId}/c/{credentialId}/{resourceType}/*
 */

import { useParams, useSearchParams, usePathname } from 'next/navigation';
import { useMemo } from 'react';

export interface ResourceContext {
  workspaceId: string | undefined;
  credentialId: string | undefined;
  workspaceResourceId: string | undefined; // /w/{workspaceId}/workspaces/{id}의 {id}
  region: string | undefined;
  availabilityZone: string | undefined;
  resourceGroup: string | undefined;
  isResourcePage: boolean;
  isManagementPage: boolean;
  isWorkspaceResourcePage: boolean; // /w/{workspaceId}/workspaces/{id}/* 페이지
  isAuthPage: boolean;
}

/**
 * Path Parameter와 Query Parameter에서 컨텍스트 정보 추출
 * 
 * @example
 * // 리소스 페이지: /w/{workspaceId}/c/{credentialId}/k8s/clusters?region=ap-northeast-3
 * const { workspaceId, credentialId, region, isResourcePage } = useResourceContext();
 * 
 * @example
 * // 관리 페이지: /w/{workspaceId}/dashboard?credentialId=xxx&region=yyy
 * const { workspaceId, credentialId, region, isManagementPage } = useResourceContext();
 * 
 * @example
 * // Workspace 리소스 페이지: /w/{workspaceId}/workspaces/{id}/settings
 * const { workspaceId, workspaceResourceId, isWorkspaceResourcePage } = useResourceContext();
 */
export function useResourceContext(): ResourceContext {
  const params = useParams();
  const searchParams = useSearchParams();
  const pathname = usePathname();
  
  // 새 구조: /w/{workspaceId}/... 또는 /w/{workspaceId}/c/{credentialId}/...
  const workspaceId = params.workspaceId as string | undefined;
  const credentialId = params.credentialId as string | undefined;
  const workspaceResourceId = params.id as string | undefined; // /workspaces/{id}의 id
  
  const region = searchParams.get('region') || undefined;
  const availabilityZone = searchParams.get('availability_zone') || undefined;
  const resourceGroup = searchParams.get('resourceGroup') || undefined;
  
  // 관리 페이지에서 credentialId가 query parameter로 올 수 있음
  const queryCredentialId = searchParams.get('credentialId') || undefined;
  
  // 최종 credentialId 결정: path parameter 우선, 없으면 query parameter
  const finalCredentialId = credentialId || queryCredentialId;
  
  return useMemo(() => {
    // Workspace 리소스 페이지: /w/{workspaceId}/workspaces/{id}/*
    const isWorkspaceResourcePage = !!workspaceId && !!workspaceResourceId && pathname.includes('/workspaces/');
    
    // Credential 리소스 페이지: /w/{workspaceId}/c/{credentialId}/*
    const isResourcePage = !!workspaceId && !!finalCredentialId && pathname.includes('/c/');
    
    // 관리 페이지: /w/{workspaceId}/{managementType}
    const isManagementPage = !!workspaceId && !isResourcePage && !isWorkspaceResourcePage;
    
    // 인증 페이지: /w/로 시작하지 않음
    const isAuthPage = !workspaceId;
    
    return {
      workspaceId,
      credentialId: finalCredentialId,
      workspaceResourceId,
      region,
      availabilityZone,
      resourceGroup,
      isResourcePage,
      isManagementPage,
      isWorkspaceResourcePage,
      isAuthPage,
    };
  }, [workspaceId, credentialId, finalCredentialId, workspaceResourceId, region, availabilityZone, resourceGroup, pathname]);
}

/**
 * 리소스 페이지에서 필수 컨텍스트 검증
 * workspaceId와 credentialId가 없으면 에러 발생
 */
export function useRequiredResourceContext(): Required<Pick<ResourceContext, 'workspaceId' | 'credentialId'>> & ResourceContext {
  const context = useResourceContext();
  
  if (!context.workspaceId || !context.credentialId) {
    throw new Error('useRequiredResourceContext must be used in a resource page with workspaceId and credentialId');
  }
  
  return {
    ...context,
    workspaceId: context.workspaceId,
    credentialId: context.credentialId,
  } as Required<Pick<ResourceContext, 'workspaceId' | 'credentialId'>> & ResourceContext;
}

/**
 * Workspace 리소스 페이지에서 필수 컨텍스트 검증
 */
export function useRequiredWorkspaceResourceContext(): Required<Pick<ResourceContext, 'workspaceId' | 'workspaceResourceId'>> & ResourceContext {
  const context = useResourceContext();
  
  if (!context.workspaceId || !context.workspaceResourceId) {
    throw new Error('useRequiredWorkspaceResourceContext must be used in a workspace resource page with workspaceId and workspaceResourceId');
  }
  
  return {
    ...context,
    workspaceId: context.workspaceId,
    workspaceResourceId: context.workspaceResourceId,
  } as Required<Pick<ResourceContext, 'workspaceId' | 'workspaceResourceId'>> & ResourceContext;
}


/**
 * Kubernetes Node Groups Page
 * Kubernetes Node Groups 관리 페이지
 * 
 * 새로운 라우팅 구조: /{workspaceId}/{credentialId}/kubernetes/node-groups
 * 
 * Note: node-pools와 node-groups는 동일한 통합 페이지를 사용합니다.
 * Provider에 따라 "Node Pools" 또는 "Node Groups"로 타이틀이 동적으로 변경됩니다.
 */

'use client';

import { NodePoolsGroupsPage } from '@/features/kubernetes';

export default NodePoolsGroupsPage;


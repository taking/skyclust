/**
 * Query Key Factory
 * 
 * React Query의 query key를 중앙에서 관리하고 타입 안전성을 보장합니다.
 * Query Key Factory 패턴을 사용하여 일관된 query key 구조를 유지합니다.
 * 
 * @example
 * ```tsx
 * // 사용 예시
 * const { data } = useQuery({
 *   queryKey: queryKeys.credentials.list(workspaceId),
 *   queryFn: () => credentialService.getCredentials(workspaceId),
 * });
 * 
 * // Invalidation
 * queryClient.invalidateQueries({ queryKey: queryKeys.credentials.all });
 * ```
 */

/**
 * Credentials Query Keys
 */
export const credentials = {
  all: ['credentials'] as const,
  lists: () => [...credentials.all, 'list'] as const,
  list: (workspaceId?: string) => [...credentials.lists(), workspaceId] as const,
  details: () => [...credentials.all, 'detail'] as const,
  detail: (id: string) => [...credentials.details(), id] as const,
} as const;

/**
 * Workspaces Query Keys
 */
export const workspaces = {
  all: ['workspaces'] as const,
  lists: () => [...workspaces.all, 'list'] as const,
  list: () => [...workspaces.lists()] as const,
  details: () => [...workspaces.all, 'detail'] as const,
  detail: (id: string) => [...workspaces.details(), id] as const,
  members: (workspaceId: string) => [...workspaces.detail(workspaceId), 'members'] as const,
} as const;

/**
 * VMs Query Keys
 */
export const vms = {
  all: ['vms'] as const,
  lists: () => [...vms.all, 'list'] as const,
  list: (workspaceId?: string) => [...vms.lists(), workspaceId] as const,
  details: () => [...vms.all, 'detail'] as const,
  detail: (id: string) => [...vms.details(), id] as const,
} as const;

/**
 * Kubernetes Clusters Query Keys
 */
export const kubernetesClusters = {
  all: ['kubernetes-clusters'] as const,
  lists: () => [...kubernetesClusters.all, 'list'] as const,
  list: (
    workspaceId?: string,
    provider?: string,
    credentialId?: string,
    region?: string
  ) => [
    ...kubernetesClusters.lists(),
    workspaceId,
    provider,
    credentialId,
    region,
  ] as const,
  details: () => [...kubernetesClusters.all, 'detail'] as const,
  detail: (name: string) => [...kubernetesClusters.details(), name] as const,
  nodePools: (clusterName: string, provider?: string, credentialId?: string, region?: string) =>
    [...kubernetesClusters.detail(clusterName), 'node-pools', provider, credentialId, region] as const,
  nodes: (clusterName: string, provider?: string, credentialId?: string, region?: string) =>
    [...kubernetesClusters.detail(clusterName), 'nodes', provider, credentialId, region] as const,
} as const;

/**
 * Kubernetes Metadata Query Keys
 */
export const kubernetesMetadata = {
  all: ['kubernetes-metadata'] as const,
  versions: (provider?: string, credentialId?: string, region?: string) =>
    [...kubernetesMetadata.all, 'versions', provider, credentialId, region] as const,
  regions: (provider?: string, credentialId?: string) =>
    [...kubernetesMetadata.all, 'regions', provider, credentialId] as const,
  availabilityZones: (provider?: string, credentialId?: string, region?: string) =>
    [...kubernetesMetadata.all, 'availability-zones', provider, credentialId, region] as const,
} as const;

/**
 * Clusters Query Keys (간단한 버전 - nodes 페이지에서 사용)
 */
export const clusters = {
  all: ['clusters'] as const,
  lists: () => [...clusters.all, 'list'] as const,
  list: (provider?: string, credentialId?: string, region?: string) =>
    [...clusters.lists(), provider, credentialId, region] as const,
} as const;

/**
 * Node Pools Query Keys
 */
export const nodePools = {
  all: ['node-pools'] as const,
  lists: () => [...nodePools.all, 'list'] as const,
  list: (provider?: string, credentialId?: string, region?: string, clusterName?: string) =>
    [...nodePools.lists(), provider, credentialId, region, clusterName] as const,
  details: () => [...nodePools.all, 'detail'] as const,
  detail: (id: string) => [...nodePools.details(), id] as const,
} as const;

/**
 * Nodes Query Keys
 */
export const nodes = {
  all: ['nodes'] as const,
  lists: () => [...nodes.all, 'list'] as const,
  list: (provider?: string, clusterName?: string, credentialId?: string, region?: string) =>
    [...nodes.lists(), provider, clusterName, credentialId, region] as const,
  details: () => [...nodes.all, 'detail'] as const,
  detail: (id: string) => [...nodes.details(), id] as const,
} as const;

/**
 * VPCs Query Keys
 */
export const vpcs = {
  all: ['vpcs'] as const,
  lists: () => [...vpcs.all, 'list'] as const,
  list: (provider?: string, credentialId?: string, region?: string) =>
    [...vpcs.lists(), provider, credentialId, region] as const,
  details: () => [...vpcs.all, 'detail'] as const,
  detail: (id: string) => [...vpcs.details(), id] as const,
} as const;

/**
 * Subnets Query Keys
 */
export const subnets = {
  all: ['subnets'] as const,
  lists: () => [...subnets.all, 'list'] as const,
  list: (provider?: string, credentialId?: string, vpcId?: string, region?: string) =>
    [...subnets.lists(), provider, credentialId, vpcId, region] as const,
  details: () => [...subnets.all, 'detail'] as const,
  detail: (id: string) => [...subnets.details(), id] as const,
} as const;

/**
 * Security Groups Query Keys
 */
export const securityGroups = {
  all: ['security-groups'] as const,
  lists: () => [...securityGroups.all, 'list'] as const,
  list: (provider?: string, credentialId?: string, vpcId?: string, region?: string) =>
    [...securityGroups.lists(), provider, credentialId, vpcId, region] as const,
  details: () => [...securityGroups.all, 'detail'] as const,
  detail: (id: string) => [...securityGroups.details(), id] as const,
} as const;

/**
 * Images Query Keys
 */
export const images = {
  all: ['images'] as const,
  lists: () => [...images.all, 'list'] as const,
  list: (provider?: string, credentialId?: string, region?: string) =>
    [...images.lists(), provider, credentialId, region] as const,
  details: () => [...images.all, 'detail'] as const,
  detail: (id: string) => [...images.details(), id] as const,
} as const;

/**
 * Snapshots Query Keys
 */
export const snapshots = {
  all: ['snapshots'] as const,
  lists: () => [...snapshots.all, 'list'] as const,
  list: (provider?: string, credentialId?: string, region?: string) =>
    [...snapshots.lists(), provider, credentialId, region] as const,
  details: () => [...snapshots.all, 'detail'] as const,
  detail: (id: string) => [...snapshots.details(), id] as const,
} as const;

/**
 * Notifications Query Keys
 */
export const notifications = {
  all: ['notifications'] as const,
  lists: () => [...notifications.all, 'list'] as const,
  list: (
    limit?: number,
    offset?: number,
    unreadOnly?: boolean,
    category?: string,
    priority?: string
  ) => [...notifications.lists(), limit, offset, unreadOnly, category, priority] as const,
  details: () => [...notifications.all, 'detail'] as const,
  detail: (id: string) => [...notifications.details(), id] as const,
  stats: () => [...notifications.all, 'stats'] as const,
  preferences: () => [...notifications.all, 'preferences'] as const,
} as const;

/**
 * Exports Query Keys
 */
export const exports = {
  all: ['exports'] as const,
  lists: () => [...exports.all, 'list'] as const,
  list: () => [...exports.lists()] as const,
  history: () => [...exports.all, 'history'] as const,
  details: () => [...exports.all, 'detail'] as const,
  detail: (id: string) => [...exports.details(), id] as const,
  status: (id: string) => [...exports.detail(id), 'status'] as const,
} as const;

/**
 * Cost Analysis Query Keys
 */
export const costAnalysis = {
  all: ['cost-analysis'] as const,
  summary: (workspaceId: string, period?: string) =>
    [...costAnalysis.all, 'summary', workspaceId, period] as const,
  predictions: (workspaceId: string, days?: number) =>
    [...costAnalysis.all, 'predictions', workspaceId, days] as const,
  trends: (workspaceId: string, period?: string) =>
    [...costAnalysis.all, 'trends', workspaceId, period] as const,
} as const;

/**
 * User Query Keys
 */
export const user = {
  all: ['user'] as const,
  me: () => [...user.all, 'me'] as const,
  details: () => [...user.all, 'detail'] as const,
  detail: (id: string) => [...user.details(), id] as const,
} as const;

/**
 * Dashboard Query Keys
 */
export const dashboard = {
  all: ['dashboard'] as const,
  summary: (workspaceId: string, credentialId?: string, region?: string) =>
    [...dashboard.all, 'summary', workspaceId, credentialId, region] as const,
} as const;

/**
 * Azure Resource Groups Query Keys
 */
export const azureResourceGroups = {
  all: ['azure-resource-groups'] as const,
  lists: () => [...azureResourceGroups.all, 'list'] as const,
  list: (credentialId?: string, limit?: number) =>
    [...azureResourceGroups.lists(), credentialId, limit] as const,
  details: () => [...azureResourceGroups.all, 'detail'] as const,
  detail: (name: string, credentialId?: string) =>
    [...azureResourceGroups.details(), name, credentialId] as const,
} as const;

/**
 * Query Keys 통합 객체
 */
export const queryKeys = {
  credentials,
  workspaces,
  vms,
  kubernetesClusters,
  kubernetesMetadata,
  clusters,
  nodePools,
  nodes,
  vpcs,
  subnets,
  securityGroups,
  images,
  snapshots,
  notifications,
  exports,
  costAnalysis,
  user,
  dashboard,
  azureResourceGroups,
} as const;

/**
 * Query Key 타입 추출 헬퍼
 */
export type QueryKey = readonly unknown[];


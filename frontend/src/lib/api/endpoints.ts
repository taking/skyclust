/**
 * API Endpoints
 * 모든 API 엔드포인트를 상수로 관리
 */

import { buildEndpointWithQuery } from '../query/builder';

type CloudProvider = 'aws' | 'gcp' | 'azure' | 'ncp';

/**
 * API 엔드포인트 헬퍼 함수
 */
export const API_ENDPOINTS = {
  // Auth endpoints
  auth: {
    login: () => 'auth/login',
    register: () => 'auth/register',
    logout: () => 'auth/logout',
    me: () => 'auth/me',
  },

  // User endpoints
  users: {
    detail: (id: string) => `users/${id}`,
    update: (id: string) => `users/${id}`,
  },

  // Credential endpoints
  credentials: {
    list: (workspaceId: string) => buildEndpointWithQuery('credentials', { workspace_id: workspaceId }),
    detail: (id: string) => `credentials/${id}`,
    create: () => 'credentials',
    update: (id: string) => `credentials/${id}`,
    delete: (id: string) => `credentials/${id}`,
    upload: () => 'credentials/upload',
  },

  // RBAC endpoints
  rbac: {
    // User role management
    assignRole: (userId: string) => `admin/rbac/users/${userId}/roles`,
    removeRole: (userId: string) => `admin/rbac/users/${userId}/roles`,
    getUserRoles: (userId: string) => `admin/rbac/users/${userId}/roles`,
    // Role permission management
    grantPermission: (role: string) => `admin/rbac/roles/${role}/permissions`,
    revokePermission: (role: string, permission: string) => `admin/rbac/roles/${role}/permissions/${permission}`,
    getRolePermissions: (role: string) => `admin/rbac/roles/${role}/permissions`,
    // User permission management
    checkUserPermission: (userId: string, permission: string) => buildEndpointWithQuery(`admin/rbac/users/${userId}/permissions/check`, { permission }),
    getUserEffectivePermissions: (userId: string) => `admin/rbac/users/${userId}/permissions/effective`,
  },

  // VM endpoints
  vms: {
    list: (workspaceId: string) => buildEndpointWithQuery('vms', { workspace_id: workspaceId }),
    detail: (id: string) => `vms/${id}`,
    create: () => 'vms',
    update: (id: string) => `vms/${id}`,
    delete: (id: string) => `vms/${id}`,
    start: (id: string) => `vms/${id}/start`,
    stop: (id: string) => `vms/${id}/stop`,
  },

  // Network endpoints
  network: {
    // VPC endpoints
    vpcs: {
      list: (provider: CloudProvider, credentialId: string, region?: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/vpcs`,
          { credential_id: credentialId, region }
        );
      },
      detail: (provider: CloudProvider, vpcId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/vpcs/${encodeURIComponent(vpcId)}`,
          { credential_id: credentialId, region }
        );
      },
      create: (provider: CloudProvider) => `${provider}/network/vpcs`,
      update: (provider: CloudProvider, vpcId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/vpcs/${encodeURIComponent(vpcId)}`,
          { credential_id: credentialId, region }
        );
      },
      delete: (provider: CloudProvider, vpcId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/vpcs/${encodeURIComponent(vpcId)}`,
          { credential_id: credentialId, region }
        );
      },
    },

    // Subnet endpoints
    subnets: {
      list: (provider: CloudProvider, credentialId: string, vpcId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/subnets`,
          { credential_id: credentialId, vpc_id: vpcId, region }
        );
      },
      detail: (provider: CloudProvider, subnetId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/subnets/${encodeURIComponent(subnetId)}`,
          { credential_id: credentialId, region }
        );
      },
      create: (provider: CloudProvider) => `${provider}/network/subnets`,
      update: (provider: CloudProvider, subnetId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/subnets/${encodeURIComponent(subnetId)}`,
          { credential_id: credentialId, region }
        );
      },
      delete: (provider: CloudProvider, subnetId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/subnets/${encodeURIComponent(subnetId)}`,
          { credential_id: credentialId, region }
        );
      },
    },

    // Security Group endpoints
    securityGroups: {
      list: (provider: CloudProvider, credentialId: string, vpcId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/security-groups`,
          { credential_id: credentialId, vpc_id: vpcId, region }
        );
      },
      detail: (provider: CloudProvider, securityGroupId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}`,
          { credential_id: credentialId, region }
        );
      },
      create: (provider: CloudProvider) => `${provider}/network/security-groups`,
      update: (provider: CloudProvider, securityGroupId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}`,
          { credential_id: credentialId, region }
        );
      },
      delete: (provider: CloudProvider, securityGroupId: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}`,
          { credential_id: credentialId, region }
        );
      },
      rules: {
        add: (provider: CloudProvider, securityGroupId: string, credentialId: string, region: string) => {
          return buildEndpointWithQuery(
            `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules`,
            { credential_id: credentialId, region }
          );
        },
        remove: (provider: CloudProvider, securityGroupId: string, credentialId: string, region: string, ruleId: string) => {
          return buildEndpointWithQuery(
            `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules`,
            { credential_id: credentialId, region, rule_id: ruleId }
          );
        },
        update: (provider: CloudProvider, securityGroupId: string, credentialId: string, region: string) => {
          return buildEndpointWithQuery(
            `${provider}/network/security-groups/${encodeURIComponent(securityGroupId)}/rules`,
            { credential_id: credentialId, region }
          );
        },
      },
    },
  },

  // Kubernetes endpoints
  kubernetes: {
    // Metadata endpoints (AWS only)
    metadata: {
      versions: (provider: CloudProvider, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/metadata/versions`,
          { credential_id: credentialId, region }
        );
      },
      regions: (provider: CloudProvider, credentialId: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/metadata/regions`,
          { credential_id: credentialId }
        );
      },
      availabilityZones: (provider: CloudProvider, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/metadata/availability-zones`,
          { credential_id: credentialId, region }
        );
      },
    },
    // Cluster endpoints
    clusters: {
      list: (provider: CloudProvider, credentialId: string, region?: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters`,
          { credential_id: credentialId, region }
        );
      },
      detail: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}`,
          { credential_id: credentialId, region }
        );
      },
      create: (provider: CloudProvider) => `${provider}/kubernetes/clusters`,
      delete: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}`,
          { credential_id: credentialId, region }
        );
      },
      kubeconfig: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/kubeconfig`,
          { credential_id: credentialId, region }
        );
      },
      upgrade: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/upgrade`,
          { credential_id: credentialId, region }
        );
      },
      upgradeStatus: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/upgrade/status`,
          { credential_id: credentialId, region }
        );
      },
      tags: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/tags`,
          { credential_id: credentialId, region }
        );
      },
      nodes: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/nodes`,
          { credential_id: credentialId, region }
        );
      },
    },

    // Node Pool endpoints (GKE, AKS, NKS)
    nodePools: {
      list: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/node-pools`,
          { credential_id: credentialId, region }
        );
      },
      detail: (provider: CloudProvider, clusterName: string, nodePoolName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/node-pools/${nodePoolName}`,
          { credential_id: credentialId, region }
        );
      },
      create: (provider: CloudProvider, clusterName: string) => `${provider}/kubernetes/clusters/${clusterName}/node-pools`,
      delete: (provider: CloudProvider, clusterName: string, nodePoolName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/node-pools/${nodePoolName}`,
          { credential_id: credentialId, region }
        );
      },
      scale: (provider: CloudProvider, clusterName: string, nodePoolName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/node-pools/${nodePoolName}/scale`,
          { credential_id: credentialId, region }
        );
      },
    },

    // Node Group endpoints (EKS specific)
    nodeGroups: {
      list: (provider: CloudProvider, clusterName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/node-groups`,
          { credential_id: credentialId, region }
        );
      },
      detail: (provider: CloudProvider, clusterName: string, nodeGroupName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/node-groups/${nodeGroupName}`,
          { credential_id: credentialId, region }
        );
      },
      create: (provider: CloudProvider, clusterName: string) => `${provider}/kubernetes/clusters/${clusterName}/node-groups`,
      delete: (provider: CloudProvider, clusterName: string, nodeGroupName: string, credentialId: string, region: string) => {
        return buildEndpointWithQuery(
          `${provider}/kubernetes/clusters/${clusterName}/node-groups/${nodeGroupName}`,
          { credential_id: credentialId, region }
        );
      },
    },
  },

  // Workspace endpoints
  workspaces: {
    list: () => 'workspaces',
    detail: (id: string) => `workspaces/${id}`,
    create: () => 'workspaces',
    update: (id: string) => `workspaces/${id}`,
    delete: (id: string) => `workspaces/${id}`,
    members: {
      list: (id: string) => `workspaces/${id}/members`,
      add: (id: string) => `workspaces/${id}/members`,
      remove: (id: string, memberId: string) => `workspaces/${id}/members/${memberId}`,
      update: (id: string, memberId: string) => `workspaces/${id}/members/${memberId}`,
    },
  },

  // Export endpoints
  exports: {
    create: () => '/exports',
    status: (id: string) => `/exports/${id}/status`,
    download: (id: string) => `/exports/${id}/download`,
    history: (limit: number = 20, offset: number = 0) => buildEndpointWithQuery('/exports/history', { limit, offset }),
    formats: () => '/exports/formats',
  },

  // Cost Analysis endpoints
  costAnalysis: {
    summary: (workspaceId: string, period: string = '30d') => buildEndpointWithQuery(`/cost-analysis/workspaces/${workspaceId}/summary`, { period }),
    predictions: (workspaceId: string, days: number = 30) => buildEndpointWithQuery(`/cost-analysis/workspaces/${workspaceId}/predictions`, { days }),
    budgetAlerts: (workspaceId: string, budgetLimit: number) => buildEndpointWithQuery(`/cost-analysis/workspaces/${workspaceId}/budget-alerts`, { budget_limit: budgetLimit }),
    trend: (workspaceId: string, period: string = '90d') => buildEndpointWithQuery(`/cost-analysis/workspaces/${workspaceId}/trend`, { period }),
    breakdown: (workspaceId: string, period: string = '30d', dimension: string = 'service') => buildEndpointWithQuery(`/cost-analysis/workspaces/${workspaceId}/breakdown`, { period, dimension }),
    comparison: (workspaceId: string, currentPeriod: string = '30d', comparePeriod: string = '30d') => buildEndpointWithQuery(`/cost-analysis/workspaces/${workspaceId}/comparison`, { current_period: currentPeriod, compare_period: comparePeriod }),
  },

  // Notification endpoints
  notifications: {
    list: (limit: number = 20, offset: number = 0, unreadOnly: boolean = false, category?: string, priority?: string) => {
      return buildEndpointWithQuery('/notifications', {
        limit,
        offset,
        unread_only: unreadOnly,
        category,
        priority,
      });
    },
    detail: (id: string) => `/notifications/${id}`,
    markAsRead: (id: string) => `/notifications/${id}/read`,
    markAllAsRead: () => '/notifications/read',
    delete: (id: string) => `/notifications/${id}`,
    deleteMultiple: () => '/notifications',
    preferences: () => '/notifications/preferences',
    updatePreferences: () => '/notifications/preferences',
    stats: () => '/notifications/stats',
    test: () => '/notifications/test',
  },

  // SSE endpoints
  sse: {
    connect: () => 'sse/events',
    subscribe: () => 'sse/subscribe',
    unsubscribe: () => 'sse/unsubscribe',
  },

  // Dashboard endpoints
  dashboard: {
    summary: (workspaceId: string, credentialId?: string, region?: string) => {
      return buildEndpointWithQuery(
        '/dashboard/summary',
        { workspace_id: workspaceId, credential_id: credentialId, region }
      );
    },
  },

  // Azure IAM endpoints
  azure: {
    iam: {
      resourceGroups: {
        list: (credentialId: string, limit?: number) => {
          return buildEndpointWithQuery(
            'azure/iam/resource-groups',
            { 
              credential_id: credentialId,
              // limit이 undefined이면 전달하지 않음 (클라이언트 사이드 페이징을 위해 모든 데이터 요청)
              ...(limit !== undefined && limit !== null && { limit })
            }
          );
        },
        detail: (name: string, credentialId: string) => {
          return buildEndpointWithQuery(
            `azure/iam/resource-groups/${encodeURIComponent(name)}`,
            { credential_id: credentialId }
          );
        },
        create: () => 'azure/iam/resource-groups',
        update: (name: string) => `azure/iam/resource-groups/${encodeURIComponent(name)}`,
        delete: (name: string, credentialId: string) => {
          return buildEndpointWithQuery(
            `azure/iam/resource-groups/${encodeURIComponent(name)}`,
            { credential_id: credentialId }
          );
        },
      },
    },
  },
} as const;


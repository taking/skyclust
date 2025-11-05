/**
 * SSE Event Data Types
 * 
 * Backend에서 전송하는 이벤트 데이터 구조와 일치하는 타입 정의
 * 필드명은 Backend와 일치 (camelCase: credentialId, clusterId, vpcId 등)
 */

/**
 * Base interface for resource events
 */
export interface ResourceEventBase {
  provider: string;
  credentialId: string; // Backend: credentialId (not credential_id)
  region: string;
}

/**
 * Kubernetes Cluster Event Data
 */
export interface KubernetesClusterEventData extends ResourceEventBase {
  clusterId?: string;      // Backend: clusterId (not cluster_id)
  cluster_name?: string;   // Backend에서 사용하는 필드명
  action?: 'created' | 'updated' | 'deleted' | 'list';
}

/**
 * Kubernetes Node Pool Event Data
 */
export interface KubernetesNodePoolEventData extends ResourceEventBase {
  clusterId?: string;
  nodePoolId?: string;
  action?: 'created' | 'updated' | 'deleted';
}

/**
 * Kubernetes Node Event Data
 */
export interface KubernetesNodeEventData extends ResourceEventBase {
  clusterId?: string;
  nodeId?: string;
  action?: 'created' | 'updated' | 'deleted';
}

/**
 * Network VPC Event Data
 */
export interface NetworkVPCEventData extends ResourceEventBase {
  vpcId?: string;         // Backend: vpcId (not vpc_id)
  action?: 'created' | 'updated' | 'deleted' | 'list';
}

/**
 * Network Subnet Event Data
 */
export interface NetworkSubnetEventData extends ResourceEventBase {
  vpcId?: string;
  subnetId?: string;      // Backend: subnetId (not subnet_id)
  action?: 'created' | 'updated' | 'deleted' | 'list';
}

/**
 * Network Security Group Event Data
 */
export interface NetworkSecurityGroupEventData extends ResourceEventBase {
  securityGroupId?: string;
  action?: 'created' | 'updated' | 'deleted' | 'list';
}

/**
 * VM Event Data
 */
export interface VMEventData {
  vmId?: string;          // Backend: vmId (not vm_id)
  workspaceId?: string;   // Backend: workspaceId (not workspace_id)
  provider?: string;
  region?: string;
  action?: 'created' | 'updated' | 'deleted' | 'list' | 'started' | 'stopped' | 'restarted';
}

/**
 * SSE Message Structure
 * Backend에서 전송하는 메시지 구조: { id, event, data: { data: <actualData> }, timestamp }
 */
export interface SSEMessage {
  id: string;
  event: string;
  data: {
    data: unknown; // 실제 이벤트 데이터가 여기에 있음
  };
  timestamp: string;
}

/**
 * Type Guards
 */
export function isResourceEventBase(data: unknown): data is ResourceEventBase {
  return (
    typeof data === 'object' &&
    data !== null &&
    'provider' in data &&
    'credentialId' in data &&
    'region' in data &&
    typeof (data as ResourceEventBase).provider === 'string' &&
    typeof (data as ResourceEventBase).credentialId === 'string' &&
    typeof (data as ResourceEventBase).region === 'string'
  );
}

export function isKubernetesClusterEventData(data: unknown): data is KubernetesClusterEventData {
  return isResourceEventBase(data);
}

export function isNetworkVPCEventData(data: unknown): data is NetworkVPCEventData {
  return isResourceEventBase(data);
}

export function isVMEventData(data: unknown): data is VMEventData {
  return (
    typeof data === 'object' &&
    data !== null &&
    ('vmId' in data || 'workspaceId' in data)
  );
}



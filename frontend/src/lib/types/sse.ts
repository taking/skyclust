/**
 * Server-Sent Events (SSE) 관련 타입 정의
 */

export interface SSEMessage {
  type: string;
  data: unknown;
  timestamp: number;
}

export interface SSEErrorInfo {
  type: 'SSE';
  readyState: number | undefined;
  url: string | undefined;
  timestamp: string;
  message?: string;
}

export interface SSECallbacks {
  onVMStatusUpdate?: (data: unknown) => void;
  onVMResourceUpdate?: (data: unknown) => void;
  onProviderStatusUpdate?: (data: unknown) => void;
  onProviderInstanceUpdate?: (data: unknown) => void;
  onSystemNotification?: (data: unknown) => void;
  onSystemAlert?: (data: unknown) => void;
  onConnected?: (data: unknown) => void;
  
  // Kubernetes 이벤트
  onKubernetesClusterCreated?: (data: unknown) => void;
  onKubernetesClusterUpdated?: (data: unknown) => void;
  onKubernetesClusterDeleted?: (data: unknown) => void;
  onKubernetesClusterList?: (data: unknown) => void;
  onKubernetesNodePoolCreated?: (data: unknown) => void;
  onKubernetesNodePoolUpdated?: (data: unknown) => void;
  onKubernetesNodePoolDeleted?: (data: unknown) => void;
  onKubernetesNodeCreated?: (data: unknown) => void;
  onKubernetesNodeUpdated?: (data: unknown) => void;
  onKubernetesNodeDeleted?: (data: unknown) => void;
  
  // Network 이벤트
  onNetworkVPCCreated?: (data: unknown) => void;
  onNetworkVPCUpdated?: (data: unknown) => void;
  onNetworkVPCDeleted?: (data: unknown) => void;
  onNetworkVPCList?: (data: unknown) => void;
  onNetworkSubnetCreated?: (data: unknown) => void;
  onNetworkSubnetUpdated?: (data: unknown) => void;
  onNetworkSubnetDeleted?: (data: unknown) => void;
  onNetworkSubnetList?: (data: unknown) => void;
  onNetworkSecurityGroupCreated?: (data: unknown) => void;
  onNetworkSecurityGroupUpdated?: (data: unknown) => void;
  onNetworkSecurityGroupDeleted?: (data: unknown) => void;
  onNetworkSecurityGroupList?: (data: unknown) => void;
  
  // VM 이벤트 (추가)
  onVMCreated?: (data: unknown) => void;
  onVMUpdated?: (data: unknown) => void;
  onVMDeleted?: (data: unknown) => void;
  onVMList?: (data: unknown) => void;
  
  // Azure Resource Group 이벤트
  onAzureResourceGroupCreated?: (data: unknown) => void;
  onAzureResourceGroupUpdated?: (data: unknown) => void;
  onAzureResourceGroupDeleted?: (data: unknown) => void;
  onAzureResourceGroupList?: (data: unknown) => void;
  
  // Dashboard Summary 이벤트
  onDashboardSummaryUpdated?: (data: unknown) => void;
  
  onError?: (error: Event | SSEErrorInfo) => void;
}


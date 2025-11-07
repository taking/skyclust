/**
 * Dashboard 타입 정의
 */

/**
 * 대시보드 요약 정보
 */
export interface DashboardSummary {
  workspace_id: string;
  vms: VMStats;
  clusters: ClusterStats;
  networks: NetworkStats;
  credentials: CredentialStats;
  members: MemberStats;
  last_updated: string;
}

/**
 * VM 통계 정보
 */
export interface VMStats {
  total: number;
  running: number;
  stopped: number;
  by_provider: Record<string, number>;
}

/**
 * Kubernetes 클러스터 통계 정보
 */
export interface ClusterStats {
  total: number;
  healthy: number;
  by_provider: Record<string, number>;
}

/**
 * 네트워크 리소스 통계 정보
 */
export interface NetworkStats {
  vpcs: number;
  subnets: number;
  security_groups: number;
}

/**
 * 자격 증명 통계 정보
 */
export interface CredentialStats {
  total: number;
  by_provider: Record<string, number>;
}

/**
 * 워크스페이스 멤버 통계 정보
 */
export interface MemberStats {
  total: number;
}


/**
 * Dashboard Service
 * 대시보드 관련 API 호출을 담당하는 서비스
 */

import { BaseService } from '@/lib/service-base';
import type { DashboardSummary } from '@/lib/types/dashboard';

/**
 * DashboardService 클래스
 * 대시보드 요약 정보를 조회하는 API를 제공합니다.
 */
class DashboardService extends BaseService {
  /**
   * 대시보드 요약 정보 조회
   * @param workspaceId - 워크스페이스 ID (필수)
   * @param credentialId - 자격 증명 ID (선택)
   * @param region - 리전 (선택)
   * @returns 대시보드 요약 정보
   */
  async getDashboardSummary(
    workspaceId: string,
    credentialId?: string,
    region?: string
  ): Promise<DashboardSummary> {
    const params = new URLSearchParams();
    params.set('workspace_id', workspaceId);
    
    if (credentialId) {
      params.set('credential_id', credentialId);
    }
    
    if (region) {
      params.set('region', region);
    }

    return this.get<DashboardSummary>(`/dashboard/summary?${params.toString()}`);
  }
}

export const dashboardService = new DashboardService();


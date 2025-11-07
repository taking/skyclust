/**
 * Cost Analysis Service
 * 비용 분석 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import { API_ENDPOINTS } from '@/lib/api-endpoints';
import type {
  CostData,
  CostSummary,
  CostPrediction,
  BudgetAlert,
  CostBreakdown,
  CostComparison,
} from '@/lib/types/cost-analysis';

class CostAnalysisService extends BaseService {
  // 비용 요약 조회
  async getCostSummary(workspaceId: string, period: string = '30d'): Promise<CostSummary> {
    return this.get<CostSummary>(API_ENDPOINTS.costAnalysis.summary(workspaceId, period));
  }

  // 비용 예측 조회
  async getCostPredictions(workspaceId: string, days: number = 30): Promise<CostPrediction[]> {
    return this.get<CostPrediction[]>(API_ENDPOINTS.costAnalysis.predictions(workspaceId, days));
  }

  // 예산 알림 조회
  async getBudgetAlerts(workspaceId: string, budgetLimit: number): Promise<BudgetAlert[]> {
    return this.get<BudgetAlert[]>(API_ENDPOINTS.costAnalysis.budgetAlerts(workspaceId, budgetLimit));
  }

  // 비용 트렌드 조회
  async getCostTrend(workspaceId: string, period: string = '90d'): Promise<{
    trend: string;
    growth_rate: number;
    total_cost: number;
    currency: string;
    period: string;
    start_date: string;
    end_date: string;
    daily_costs: CostData[];
  }> {
    return this.get(API_ENDPOINTS.costAnalysis.trend(workspaceId, period));
  }

  // 비용 분석 조회
  async getCostBreakdown(workspaceId: string, period: string = '30d', dimension: string = 'service'): Promise<CostBreakdown> {
    return this.get<CostBreakdown>(API_ENDPOINTS.costAnalysis.breakdown(workspaceId, period, dimension));
  }

  // 비용 비교 조회
  async getCostComparison(workspaceId: string, currentPeriod: string = '30d', comparePeriod: string = '30d'): Promise<CostComparison> {
    return this.get<CostComparison>(API_ENDPOINTS.costAnalysis.comparison(workspaceId, currentPeriod, comparePeriod));
  }
}

export const costAnalysisService = new CostAnalysisService();

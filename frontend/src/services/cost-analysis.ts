/**
 * Cost Analysis Service
 * 비용 분석 관련 API 호출
 */

import { api } from '@/lib/api';

export interface CostData {
  date: string;
  amount: number;
  currency: string;
  service: string;
  resource_id: string;
  resource_type: string;
  provider: string;
  region: string;
  workspace_id: string;
}

export interface CostSummary {
  total_cost: number;
  currency: string;
  period: string;
  start_date: string;
  end_date: string;
  by_service: Record<string, number>;
  by_provider: Record<string, number>;
  by_region: Record<string, number>;
  by_workspace: Record<string, number>;
  daily_costs: CostData[];
  trend: 'increasing' | 'decreasing' | 'stable';
  growth_rate: number;
}

export interface CostPrediction {
  date: string;
  predicted: number;
  confidence: number;
  lower_bound: number;
  upper_bound: number;
}

export interface BudgetAlert {
  id: string;
  workspace_id: string;
  budget_limit: number;
  current_cost: number;
  percentage: number;
  alert_level: 'warning' | 'critical';
  message: string;
  created_at: string;
}

export interface CostBreakdown {
  dimension: string;
  total_cost: number;
  currency: string;
  period: string;
  breakdown: Array<{
    name: string;
    value: number;
    percentage: number;
  }>;
}

export interface CostComparison {
  current_period: {
    period: string;
    total_cost: number;
    start_date: string;
    end_date: string;
  };
  compare_period: {
    period: string;
    total_cost: number;
    start_date: string;
    end_date: string;
  };
  comparison: {
    cost_change: number;
    percentage_change: number;
    trend: string;
    growth_rate: number;
  };
}

export const costAnalysisService = {
  // 비용 요약 조회
  async getCostSummary(workspaceId: string, period: string = '30d'): Promise<CostSummary> {
    const response = await api.get(`/cost-analysis/workspaces/${workspaceId}/summary?period=${period}`);
    return response.data.data;
  },

  // 비용 예측 조회
  async getCostPredictions(workspaceId: string, days: number = 30): Promise<CostPrediction[]> {
    const response = await api.get(`/cost-analysis/workspaces/${workspaceId}/predictions?days=${days}`);
    return response.data.data;
  },

  // 예산 알림 조회
  async getBudgetAlerts(workspaceId: string, budgetLimit: number): Promise<BudgetAlert[]> {
    const response = await api.get(`/cost-analysis/workspaces/${workspaceId}/budget-alerts?budget_limit=${budgetLimit}`);
    return response.data.data;
  },

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
    const response = await api.get(`/cost-analysis/workspaces/${workspaceId}/trend?period=${period}`);
    return response.data.data;
  },

  // 비용 분석 조회
  async getCostBreakdown(workspaceId: string, period: string = '30d', dimension: string = 'service'): Promise<CostBreakdown> {
    const response = await api.get(`/cost-analysis/workspaces/${workspaceId}/breakdown?period=${period}&dimension=${dimension}`);
    return response.data.data;
  },

  // 비용 비교 조회
  async getCostComparison(workspaceId: string, currentPeriod: string = '30d', comparePeriod: string = '30d'): Promise<CostComparison> {
    const response = await api.get(`/cost-analysis/workspaces/${workspaceId}/comparison?current_period=${currentPeriod}&compare_period=${comparePeriod}`);
    return response.data.data;
  },
};

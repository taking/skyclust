/**
 * Cost Analysis 관련 타입 정의
 */

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


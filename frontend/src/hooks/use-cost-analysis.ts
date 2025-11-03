import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { costAnalysisService } from '@/services/cost-analysis';
import type { CostSummary, CostPrediction, BudgetAlert, CostBreakdown, CostComparison } from '@/lib/types/cost-analysis';
import { useToast } from '@/hooks/use-toast';

export function useCostSummary(workspaceId: string, period: string = '30d') {
  return useQuery({
    queryKey: ['cost-summary', workspaceId, period],
    queryFn: () => costAnalysisService.getCostSummary(workspaceId, period),
    enabled: !!workspaceId,
    staleTime: 5 * 60 * 1000, // 5 minutes - 비용 요약 데이터
    gcTime: 10 * 60 * 1000, // 10 minutes - GC 시간
  });
}

export function useCostPredictions(workspaceId: string, days: number = 30) {
  return useQuery({
    queryKey: ['cost-predictions', workspaceId, days],
    queryFn: () => costAnalysisService.getCostPredictions(workspaceId, days),
    enabled: !!workspaceId,
    staleTime: 10 * 60 * 1000, // 10 minutes - 예측 데이터는 비교적 안정적
    gcTime: 30 * 60 * 1000, // 30 minutes - GC 시간
  });
}

export function useBudgetAlerts(workspaceId: string, budgetLimit: number) {
  return useQuery({
    queryKey: ['budget-alerts', workspaceId, budgetLimit],
    queryFn: () => costAnalysisService.getBudgetAlerts(workspaceId, budgetLimit),
    enabled: !!workspaceId && budgetLimit > 0,
    staleTime: 2 * 60 * 1000, // 2 minutes - 알림은 더 자주 업데이트 필요
    gcTime: 5 * 60 * 1000, // 5 minutes - GC 시간
    refetchInterval: 60000, // 1분마다 refetch (알림 중요성)
  });
}

export function useCostTrend(workspaceId: string, period: string = '90d') {
  return useQuery({
    queryKey: ['cost-trend', workspaceId, period],
    queryFn: () => costAnalysisService.getCostTrend(workspaceId, period),
    enabled: !!workspaceId,
    staleTime: 5 * 60 * 1000, // 5 minutes - 트렌드 데이터
    gcTime: 15 * 60 * 1000, // 15 minutes - GC 시간
  });
}

export function useCostBreakdown(workspaceId: string, period: string = '30d', dimension: string = 'service') {
  return useQuery({
    queryKey: ['cost-breakdown', workspaceId, period, dimension],
    queryFn: () => costAnalysisService.getCostBreakdown(workspaceId, period, dimension),
    enabled: !!workspaceId,
    staleTime: 5 * 60 * 1000, // 5 minutes - 비용 분석 데이터
    gcTime: 15 * 60 * 1000, // 15 minutes - GC 시간
  });
}

export function useCostComparison(workspaceId: string, currentPeriod: string = '30d', comparePeriod: string = '30d') {
  return useQuery({
    queryKey: ['cost-comparison', workspaceId, currentPeriod, comparePeriod],
    queryFn: () => costAnalysisService.getCostComparison(workspaceId, currentPeriod, comparePeriod),
    enabled: !!workspaceId,
    staleTime: 5 * 60 * 1000, // 5 minutes - 비교 데이터
    gcTime: 15 * 60 * 1000, // 15 minutes - GC 시간
  });
}

// 비용 분석 데이터 새로고침
export function useRefreshCostAnalysis() {
  const queryClient = useQueryClient();
  const { success } = useToast();

  return useMutation({
    mutationFn: async (workspaceId: string) => {
      // 모든 비용 분석 관련 쿼리 무효화
      await queryClient.invalidateQueries({
        queryKey: ['cost-summary', workspaceId],
      });
      await queryClient.invalidateQueries({
        queryKey: ['cost-predictions', workspaceId],
      });
      await queryClient.invalidateQueries({
        queryKey: ['budget-alerts', workspaceId],
      });
      await queryClient.invalidateQueries({
        queryKey: ['cost-trend', workspaceId],
      });
      await queryClient.invalidateQueries({
        queryKey: ['cost-breakdown', workspaceId],
      });
      await queryClient.invalidateQueries({
        queryKey: ['cost-comparison', workspaceId],
      });
    },
    onSuccess: () => {
      success('Cost analysis data refreshed successfully');
    },
  });
}

// 비용 분석 유틸리티 함수들
export const costAnalysisUtils = {
  // 통화 포맷팅
  formatCurrency: (amount: number, currency: string = 'USD'): string => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
    }).format(amount);
  },

  // 퍼센트 포맷팅
  formatPercentage: (value: number): string => {
    return `${value.toFixed(1)}%`;
  },

  // 트렌드 아이콘 반환
  getTrendIcon: (trend: string): string => {
    switch (trend) {
      case 'increasing':
        return '📈';
      case 'decreasing':
        return '📉';
      case 'stable':
        return '➡️';
      default:
        return '❓';
    }
  },

  // 트렌드 색상 반환
  getTrendColor: (trend: string): string => {
    switch (trend) {
      case 'increasing':
        return 'text-red-600';
      case 'decreasing':
        return 'text-green-600';
      case 'stable':
        return 'text-gray-600';
      default:
        return 'text-gray-400';
    }
  },

  // 예산 사용률 색상 반환
  getBudgetUsageColor: (percentage: number): string => {
    if (percentage >= 100) return 'text-red-600';
    if (percentage >= 80) return 'text-yellow-600';
    if (percentage >= 60) return 'text-blue-600';
    return 'text-green-600';
  },

  // 예산 사용률 배경 색상 반환
  getBudgetUsageBgColor: (percentage: number): string => {
    if (percentage >= 100) return 'bg-red-100';
    if (percentage >= 80) return 'bg-yellow-100';
    if (percentage >= 60) return 'bg-blue-100';
    return 'bg-green-100';
  },

  // 날짜 포맷팅
  formatDate: (dateString: string): string => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  },

  // 차트용 데이터 변환
  transformForChart: (data: Array<{ name: string; value: number; percentage?: number }>) => {
    return data.map(item => ({
      name: item.name,
      value: item.value,
      percentage: item.percentage || 0,
    }));
  },

  // 예측 신뢰도 색상 반환
  getConfidenceColor: (confidence: number): string => {
    if (confidence >= 0.8) return 'text-green-600';
    if (confidence >= 0.6) return 'text-yellow-600';
    return 'text-red-600';
  },
};


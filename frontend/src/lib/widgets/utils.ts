/**
 * Widget Utilities
 * 위젯 관련 유틸리티 함수
 */

import { WidgetType, WidgetConfig } from './types';
import { useTranslation } from '@/hooks/use-translation';

/**
 * 번역된 위젯 설정을 반환하는 함수
 */
export function useWidgetConfigs() {
  const { t } = useTranslation();

  const getWidgetConfig = (type: WidgetType): WidgetConfig => {
    const baseConfigs: Record<WidgetType, Omit<WidgetConfig, 'title' | 'description'>> = {
      'vm-status': {
        type: 'vm-status',
        defaultSize: 'medium',
        minSize: 'small',
        maxSize: 'large',
        category: 'overview',
        icon: 'Server',
      },
      'cost-chart': {
        type: 'cost-chart',
        defaultSize: 'large',
        minSize: 'medium',
        maxSize: 'xlarge',
        category: 'cost',
        icon: 'DollarSign',
      },
      'resource-usage': {
        type: 'resource-usage',
        defaultSize: 'medium',
        minSize: 'small',
        maxSize: 'large',
        category: 'monitoring',
        icon: 'Activity',
      },
      'recent-activity': {
        type: 'recent-activity',
        defaultSize: 'medium',
        minSize: 'small',
        maxSize: 'large',
        category: 'overview',
        icon: 'Clock',
      },
      'quick-actions': {
        type: 'quick-actions',
        defaultSize: 'small',
        minSize: 'small',
        maxSize: 'medium',
        category: 'management',
        icon: 'Zap',
      },
      'alerts': {
        type: 'alerts',
        defaultSize: 'medium',
        minSize: 'small',
        maxSize: 'large',
        category: 'monitoring',
        icon: 'AlertTriangle',
      },
      'performance-metrics': {
        type: 'performance-metrics',
        defaultSize: 'large',
        minSize: 'medium',
        maxSize: 'xlarge',
        category: 'monitoring',
        icon: 'TrendingUp',
      },
      'region-distribution': {
        type: 'region-distribution',
        defaultSize: 'medium',
        minSize: 'small',
        maxSize: 'large',
        category: 'overview',
        icon: 'MapPin',
      },
      'kubernetes-status': {
        type: 'kubernetes-status',
        defaultSize: 'medium',
        minSize: 'small',
        maxSize: 'large',
        category: 'overview',
        icon: 'Container',
      },
      'network-status': {
        type: 'network-status',
        defaultSize: 'medium',
        minSize: 'small',
        maxSize: 'large',
        category: 'overview',
        icon: 'Network',
      },
    };

    const translationKeys: Record<WidgetType, { title: string; description: string }> = {
      'vm-status': {
        title: t('widgets.vmStatus.title'),
        description: t('widgets.vmStatus.description'),
      },
      'cost-chart': {
        title: t('widgets.costChart.title'),
        description: t('widgets.costChart.description'),
      },
      'resource-usage': {
        title: t('widgets.resourceUsage.title'),
        description: t('widgets.resourceUsage.description'),
      },
      'recent-activity': {
        title: t('widgets.recentActivity.title'),
        description: t('widgets.recentActivity.description'),
      },
      'quick-actions': {
        title: t('widgets.quickActions.title'),
        description: t('widgets.quickActions.description'),
      },
      'alerts': {
        title: t('widgets.alerts.title'),
        description: t('widgets.alerts.description'),
      },
      'performance-metrics': {
        title: t('widgets.performanceMetrics.title'),
        description: t('widgets.performanceMetrics.description'),
      },
      'region-distribution': {
        title: t('widgets.regionDistribution.title'),
        description: t('widgets.regionDistribution.description'),
      },
      'kubernetes-status': {
        title: t('widgets.kubernetesStatus.title'),
        description: t('widgets.kubernetesStatus.description'),
      },
      'network-status': {
        title: t('widgets.networkStatus.title'),
        description: t('widgets.networkStatus.description'),
      },
    };

    return {
      ...baseConfigs[type],
      ...translationKeys[type],
    };
  };

  return { getWidgetConfig };
}

/**
 * 위젯 카테고리 번역 키 매핑
 */
export function getWidgetCategoryTranslationKey(category: string): string {
  const categoryMap: Record<string, string> = {
    all: 'widgets.addWidget.categories.all',
    overview: 'widgets.addWidget.categories.overview',
    monitoring: 'widgets.addWidget.categories.monitoring',
    cost: 'widgets.addWidget.categories.cost',
    management: 'widgets.addWidget.categories.management',
  };
  return categoryMap[category] || category;
}

/**
 * 위젯 크기 번역 키 매핑
 */
export function getWidgetSizeTranslationKey(size: string): string {
  const sizeMap: Record<string, string> = {
    small: 'widgets.addWidget.sizes.small',
    medium: 'widgets.addWidget.sizes.medium',
    large: 'widgets.addWidget.sizes.large',
    xlarge: 'widgets.addWidget.sizes.xlarge',
  };
  return sizeMap[size] || size;
}

/**
 * 위젯 아이콘 가져오기
 */
export function getWidgetIcon(iconName: string): string {
  return iconName;
}


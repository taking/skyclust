export interface WidgetData {
  id: string;
  type: WidgetType;
  title: string;
  description?: string;
  size: WidgetSize;
  position: { x: number; y: number };
  config: Record<string, unknown>;
  lastUpdated: string;
}

export type WidgetType = 
  | 'vm-status'
  | 'cost-chart'
  | 'resource-usage'
  | 'recent-activity'
  | 'quick-actions'
  | 'alerts'
  | 'performance-metrics'
  | 'region-distribution'
  | 'kubernetes-status'
  | 'network-status';

export type WidgetSize = 'small' | 'medium' | 'large' | 'xlarge';

export interface WidgetConfig {
  type: WidgetType;
  title: string;
  description: string;
  defaultSize: WidgetSize;
  minSize: WidgetSize;
  maxSize: WidgetSize;
  category: 'overview' | 'monitoring' | 'cost' | 'management';
  icon: string;
}

export const WIDGET_CONFIGS: Record<WidgetType, WidgetConfig> = {
  'vm-status': {
    type: 'vm-status',
    title: 'VM Status Overview',
    description: 'Overview of virtual machine statuses',
    defaultSize: 'medium',
    minSize: 'small',
    maxSize: 'large',
    category: 'overview',
    icon: 'Server',
  },
  'cost-chart': {
    type: 'cost-chart',
    title: 'Cost Analysis',
    description: 'Monthly cost trends and breakdown',
    defaultSize: 'large',
    minSize: 'medium',
    maxSize: 'xlarge',
    category: 'cost',
    icon: 'DollarSign',
  },
  'resource-usage': {
    type: 'resource-usage',
    title: 'Resource Usage',
    description: 'CPU, memory, and storage utilization',
    defaultSize: 'medium',
    minSize: 'small',
    maxSize: 'large',
    category: 'monitoring',
    icon: 'Activity',
  },
  'recent-activity': {
    type: 'recent-activity',
    title: 'Recent Activity',
    description: 'Latest system events and changes',
    defaultSize: 'medium',
    minSize: 'small',
    maxSize: 'large',
    category: 'overview',
    icon: 'Clock',
  },
  'quick-actions': {
    type: 'quick-actions',
    title: 'Quick Actions',
    description: 'Common management actions',
    defaultSize: 'small',
    minSize: 'small',
    maxSize: 'medium',
    category: 'management',
    icon: 'Zap',
  },
  'alerts': {
    type: 'alerts',
    title: 'Alerts & Notifications',
    description: 'System alerts and notifications',
    defaultSize: 'medium',
    minSize: 'small',
    maxSize: 'large',
    category: 'monitoring',
    icon: 'AlertTriangle',
  },
  'performance-metrics': {
    type: 'performance-metrics',
    title: 'Performance Metrics',
    description: 'Key performance indicators',
    defaultSize: 'large',
    minSize: 'medium',
    maxSize: 'xlarge',
    category: 'monitoring',
    icon: 'TrendingUp',
  },
  'region-distribution': {
    type: 'region-distribution',
    title: 'Region Distribution',
    description: 'Resource distribution across regions',
    defaultSize: 'medium',
    minSize: 'small',
    maxSize: 'large',
    category: 'overview',
    icon: 'MapPin',
  },
  'kubernetes-status': {
    type: 'kubernetes-status',
    title: 'Kubernetes Clusters',
    description: 'Kubernetes cluster status and overview',
    defaultSize: 'medium',
    minSize: 'small',
    maxSize: 'large',
    category: 'overview',
    icon: 'Container',
  },
  'network-status': {
    type: 'network-status',
    title: 'Network Resources',
    description: 'VPC, Subnet, and Security Group overview',
    defaultSize: 'medium',
    minSize: 'small',
    maxSize: 'large',
    category: 'overview',
    icon: 'Network',
  },
};

export const getWidgetSizeClasses = (size: WidgetSize): string => {
  const sizeMap = {
    small: 'col-span-1 row-span-1',
    medium: 'col-span-1 sm:col-span-2 row-span-1',
    large: 'col-span-1 sm:col-span-2 row-span-1 lg:row-span-2',
    xlarge: 'col-span-1 sm:col-span-2 lg:col-span-3 row-span-1 lg:row-span-2',
  };
  return sizeMap[size];
};

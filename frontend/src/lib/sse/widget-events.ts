/**
 * Widget SSE Events Mapping
 * 위젯별 필요한 SSE 이벤트 타입을 정의하고 관리합니다.
 */

import type { WidgetType } from '@/lib/widgets';

/**
 * 위젯 타입별로 필요한 SSE 이벤트 타입 목록
 */
export const WIDGET_EVENT_MAP: Record<WidgetType, string[]> = {
  'vm-status': [
    'vm-created',
    'vm-updated',
    'vm-deleted',
    'vm-list',
  ],
  'cost-chart': [
    'vm-created',
    'vm-updated',
    'vm-deleted',
    'kubernetes-cluster-created',
    'kubernetes-cluster-updated',
    'kubernetes-cluster-deleted',
  ],
  'resource-usage': [
    'vm-created',
    'vm-updated',
    'vm-deleted',
    'vm-list',
  ],
  'recent-activity': [
    'vm-created',
    'vm-updated',
    'vm-deleted',
    'kubernetes-cluster-created',
    'kubernetes-cluster-updated',
    'kubernetes-cluster-deleted',
    'network-vpc-created',
    'network-vpc-updated',
    'network-vpc-deleted',
    'network-subnet-created',
    'network-subnet-updated',
    'network-subnet-deleted',
    'network-security-group-created',
    'network-security-group-updated',
    'network-security-group-deleted',
  ],
  'quick-actions': [],
  'alerts': [
    'system-notification',
    'system-alert',
  ],
  'performance-metrics': [
    'vm-created',
    'vm-updated',
    'vm-deleted',
    'kubernetes-cluster-created',
    'kubernetes-cluster-updated',
    'kubernetes-cluster-deleted',
  ],
  'region-distribution': [
    'vm-created',
    'vm-updated',
    'vm-deleted',
    'kubernetes-cluster-created',
    'kubernetes-cluster-updated',
    'kubernetes-cluster-deleted',
    'network-vpc-created',
    'network-vpc-updated',
    'network-vpc-deleted',
  ],
  'kubernetes-status': [
    'kubernetes-cluster-created',
    'kubernetes-cluster-updated',
    'kubernetes-cluster-deleted',
    'kubernetes-cluster-list',
    'kubernetes-node-pool-created',
    'kubernetes-node-pool-updated',
    'kubernetes-node-pool-deleted',
    'kubernetes-node-created',
    'kubernetes-node-updated',
    'kubernetes-node-deleted',
  ],
  'network-status': [
    'network-vpc-created',
    'network-vpc-updated',
    'network-vpc-deleted',
    'network-vpc-list',
    'network-subnet-created',
    'network-subnet-updated',
    'network-subnet-deleted',
    'network-subnet-list',
    'network-security-group-created',
    'network-security-group-updated',
    'network-security-group-deleted',
    'network-security-group-list',
  ],
};

/**
 * 대시보드 요약 정보에 필요한 이벤트 타입 목록
 * 
 * 참고: Dashboard summary는 backend에서 자동으로 재계산되어
 * 'dashboard-summary-updated' 이벤트 하나만 구독하면 됩니다.
 * 개별 리소스 이벤트는 구독하지 않습니다.
 */
export const DASHBOARD_SUMMARY_EVENTS = [] as const;

/**
 * 위젯 목록에서 필요한 모든 이벤트 타입을 추출합니다.
 * @param widgets - 위젯 데이터 배열
 * @returns 필요한 이벤트 타입 Set
 */
export function getRequiredEventsForWidgets(widgets: Array<{ type: WidgetType }>): Set<string> {
  const events = new Set<string>();

  widgets.forEach((widget) => {
    const widgetEvents = WIDGET_EVENT_MAP[widget.type] || [];
    widgetEvents.forEach((eventType) => {
      events.add(eventType);
    });
  });

  return events;
}

/**
 * 대시보드 요약 정보에 필요한 이벤트 타입 Set을 반환합니다.
 * @returns 대시보드 요약 정보 이벤트 타입 Set
 */
export function getDashboardSummaryEvents(): Set<string> {
  return new Set(DASHBOARD_SUMMARY_EVENTS);
}

/**
 * 위젯 목록과 대시보드 요약 정보를 포함한 모든 필요한 이벤트 타입을 반환합니다.
 * @param widgets - 위젯 데이터 배열
 * @param includeSummary - 대시보드 요약 정보 이벤트 포함 여부
 * @returns 필요한 이벤트 타입 Set
 */
export function getAllRequiredEvents(
  widgets: Array<{ type: WidgetType }>,
  includeSummary: boolean = true
): Set<string> {
  const events = getRequiredEventsForWidgets(widgets);

  if (includeSummary) {
    const summaryEvents = getDashboardSummaryEvents();
    summaryEvents.forEach((eventType) => {
      events.add(eventType);
    });
  }

  return events;
}

/**
 * 이벤트 타입이 리소스 관련 이벤트인지 확인합니다.
 * @param eventType - 이벤트 타입
 * @returns 리소스 관련 이벤트 여부
 */
export function isResourceEvent(eventType: string): boolean {
  return (
    eventType.startsWith('vm-') ||
    eventType.startsWith('kubernetes-') ||
    eventType.startsWith('network-')
  );
}

/**
 * 이벤트 타입이 시스템 이벤트인지 확인합니다.
 * @param eventType - 이벤트 타입
 * @returns 시스템 이벤트 여부
 */
export function isSystemEvent(eventType: string): boolean {
  return eventType.startsWith('system-');
}


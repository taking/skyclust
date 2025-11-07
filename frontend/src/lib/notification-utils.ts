/**
 * Notification Utilities
 * 알림 관련 유틸리티 함수
 */

import * as React from 'react';
import { CheckCircle, AlertTriangle, AlertCircle, Info } from 'lucide-react';

/**
 * 알림 타입에 따른 아이콘 반환
 * 
 * @param type - 알림 타입 ('success', 'warning', 'error', 'info')
 * @returns 해당 타입의 아이콘 컴포넌트
 * 
 * @example
 * ```tsx
 * const icon = getNotificationTypeIcon('success');
 * ```
 */
export function getNotificationTypeIcon(type: string): React.ReactElement {
  switch (type) {
    case 'success':
      return React.createElement(CheckCircle, { className: 'h-4 w-4 text-green-600' });
    case 'warning':
      return React.createElement(AlertTriangle, { className: 'h-4 w-4 text-yellow-600' });
    case 'error':
      return React.createElement(AlertCircle, { className: 'h-4 w-4 text-red-600' });
    case 'info':
    default:
      return React.createElement(Info, { className: 'h-4 w-4 text-blue-600' });
  }
}

/**
 * 알림 타입에 따른 배지 색상 클래스 반환
 * 
 * @param type - 알림 타입 ('success', 'warning', 'error', 'info')
 * @returns Tailwind CSS 클래스 문자열
 * 
 * @example
 * ```tsx
 * const colorClass = getNotificationTypeColor('success');
 * // 'bg-green-100 text-green-800'
 * ```
 */
export function getNotificationTypeColor(type: string): string {
  switch (type) {
    case 'success':
      return 'bg-green-100 text-green-800';
    case 'warning':
      return 'bg-yellow-100 text-yellow-800';
    case 'error':
      return 'bg-red-100 text-red-800';
    case 'info':
    default:
      return 'bg-blue-100 text-blue-800';
  }
}

/**
 * 알림 우선순위에 따른 배지 색상 클래스 반환
 * 
 * @param priority - 알림 우선순위 ('urgent', 'high', 'medium', 'low')
 * @returns Tailwind CSS 클래스 문자열
 * 
 * @example
 * ```tsx
 * const colorClass = getNotificationPriorityColor('urgent');
 * // 'bg-red-100 text-red-800'
 * ```
 */
export function getNotificationPriorityColor(priority: string): string {
  switch (priority) {
    case 'urgent':
      return 'bg-red-100 text-red-800';
    case 'high':
      return 'bg-orange-100 text-orange-800';
    case 'medium':
      return 'bg-yellow-100 text-yellow-800';
    case 'low':
      return 'bg-green-100 text-green-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}


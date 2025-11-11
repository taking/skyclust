/**
 * Widgets Module
 * 모든 위젯 관련 기능을 중앙화하여 export
 */

export type { WidgetData, WidgetType, WidgetSize, WidgetConfig } from './types';
export { WIDGET_CONFIGS, getWidgetSizeClasses } from './types';
export { useWidgetConfigs, getWidgetCategoryTranslationKey, getWidgetSizeTranslationKey, getWidgetIcon } from './utils';


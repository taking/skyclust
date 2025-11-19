/**
 * VPCs Page Header Component
 * VPCs 페이지 헤더 컴포넌트
 */

'use client';

import { useWorkspaceStore } from '@/store/workspace';
import { useTranslation } from '@/hooks/use-translation';
import { Button } from '@/components/ui/button';
import { RefreshCw, Info, Plus } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip';

interface VPCsPageHeaderProps {
  onRefresh?: () => void;
  isRefreshing?: boolean;
  lastUpdated?: Date | null;
  onCreateClick?: () => void;
  disabled?: boolean;
}

export function VPCsPageHeader({ 
  onRefresh, 
  isRefreshing = false, 
  lastUpdated,
  onCreateClick,
  disabled = false,
}: VPCsPageHeaderProps) {
  const { currentWorkspace } = useWorkspaceStore();
  const { t } = useTranslation();

  // 마지막 업데이트 시간 포맷팅
  const formatLastUpdated = (date: Date | null | undefined): string => {
    if (!date) return '';
    
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    
    if (diffSec < 60) {
      return t('common.justNow') || '방금 전';
    } else if (diffMin < 60) {
      return t('common.minutesAgo', { minutes: diffMin }) || `${diffMin}분 전`;
    } else if (diffHour < 24) {
      return t('common.hoursAgo', { hours: diffHour }) || `${diffHour}시간 전`;
    } else {
      const diffDay = Math.floor(diffHour / 24);
      return t('common.daysAgo', { days: diffDay }) || `${diffDay}일 전`;
    }
  };

  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">{t('network.vpcs')}</h1>
        <div className="flex items-center gap-2">
          <p className="text-gray-600 mt-1">
            {currentWorkspace 
              ? t('network.manageVPCsWithWorkspace', { workspaceName: currentWorkspace.name }) 
              : t('network.manageVPCs')
            }
          </p>
          {lastUpdated && (
            <div className="flex items-center gap-1.5 mt-1">
              <span className="text-sm text-gray-500">
                ({t('common.lastUpdated')}: {formatLastUpdated(lastUpdated)})
              </span>
              <Tooltip delayDuration={200}>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    className="focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-1 rounded-sm transition-colors"
                    aria-label={t('sse.syncInfo') || 'Sync information'}
                  >
                    <Info className="h-3.5 w-3.5 text-gray-400 hover:text-gray-600 transition-colors" aria-hidden="true" />
                  </button>
                </TooltipTrigger>
                <TooltipContent 
                  side="right" 
                  className="max-w-xs bg-gray-900 text-white border-gray-700 z-50"
                >
                  <div className="space-y-2 text-xs">
                    <div className="font-semibold text-sm mb-1.5 pb-1.5 border-b border-gray-700">
                      {t('sse.dataSyncInfo') || '데이터 동기화 정보'}
                    </div>
                    
                    <div className="space-y-1.5">
                      <div>
                        <span className="font-medium text-gray-100">
                          {t('sse.syncInterval') || '동기화 주기'}:
                        </span>
                        <span className="ml-1.5 text-gray-300">
                          {t('sse.syncEvery5Minutes') || '5분마다'}
                        </span>
                      </div>
                      
                      <div>
                        <span className="font-medium text-gray-100">
                          {t('sse.lastSyncTime') || '마지막 동기화'}:
                        </span>
                        <span className="ml-1.5 text-gray-300">
                          {formatLastUpdated(lastUpdated)}
                        </span>
                      </div>
                    </div>
                    
                    <div className="pt-1.5 mt-1.5 border-t border-gray-700">
                      <p className="text-gray-300 leading-relaxed">
                        {t('sse.autoUpdateDescription') || '백엔드에서 5분마다 클라우드 서비스 제공자(CSP) API를 호출하여 최신 데이터를 조회합니다. 변경사항이 감지되면 자동으로 SSE 이벤트를 발행하여 실시간으로 업데이트됩니다.'}
                      </p>
                    </div>
                  </div>
                </TooltipContent>
              </Tooltip>
            </div>
          )}
        </div>
      </div>
      <div className="flex items-center space-x-2">
        {onRefresh && (
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw className={cn('mr-2 h-4 w-4', isRefreshing && 'animate-spin')} />
            {isRefreshing ? (t('common.refreshing') || '새로고침 중...') : (t('common.refresh') || '새로고침')}
          </Button>
        )}
        {onCreateClick && (
          <Button
            onClick={onCreateClick}
            disabled={disabled}
          >
            <Plus className="mr-2 h-4 w-4" />
            {t('network.createVPC') || 'Create VPC'}
          </Button>
        )}
      </div>
    </div>
  );
}

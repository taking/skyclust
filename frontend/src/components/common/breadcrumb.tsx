/**
 * Breadcrumb Component
 * 경로 표시를 위한 브레드크럼 컴포넌트
 */

'use client';

import * as React from 'react';
import Link from 'next/link';
import { usePathname, useSearchParams } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { ChevronRight, Home } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/hooks/use-translation';

export interface BreadcrumbItem {
  label: string;
  href: string;
  icon?: React.ComponentType<{ className?: string }>;
}

interface BreadcrumbProps {
  items?: BreadcrumbItem[];
  className?: string;
  /**
   * 동적 리소스 이름 (예: 클러스터 이름, VM 이름)
   */
  resourceName?: string;
  /**
   * 마지막 항목을 링크로 만들지 여부 (기본값: false)
   */
  linkLast?: boolean;
}

export function Breadcrumb({ items, className, resourceName, linkLast = false }: BreadcrumbProps) {
  const { t } = useTranslation();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { currentWorkspace } = useWorkspaceStore();

  // 1. 워크스페이스 ID 가져오기: URL 파라미터 또는 스토어에서
  const workspaceId = React.useMemo(() => {
    // URL에 workspaceId가 있으면 우선 사용, 없으면 스토어의 현재 워크스페이스 ID 사용
    const urlWorkspaceId = searchParams.get('workspaceId');
    return urlWorkspaceId || currentWorkspace?.id || '';
  }, [searchParams, currentWorkspace?.id]);

  // 2. Breadcrumb 항목 자동 생성: items가 제공되지 않으면 pathname에서 생성
  const breadcrumbItems = React.useMemo(() => {
    // items가 명시적으로 제공되면 그대로 사용
    if (items) return items;

    // 3. 1 depth 경로를 기본 하위 경로로 매핑 (예: /compute → /compute/vms)
    const defaultPaths: Record<string, string> = {
      '/compute': '/compute/vms',
      '/kubernetes': '/kubernetes/clusters',
      '/networks': '/networks/vpcs',
    };

    // 4. pathname을 세그먼트로 분리 (빈 문자열 제거)
    const paths = pathname.split('/').filter(Boolean);
    
    // 5. 홈 항목으로 시작하는 breadcrumb 배열 초기화
    const result: BreadcrumbItem[] = [
      { label: t('common.home'), href: '/dashboard' },
    ];

    // 6. 경로 세그먼트를 번역된 라벨로 매핑
    const pathLabels: Record<string, string> = {
      // 새로운 라우팅 구조 지원
      'w': '', // workspace 경로 구조의 일부이므로 라벨 없음
      'c': '', // credential 경로 구조의 일부이므로 라벨 없음
      'k8s': t('nav.kubernetes'),
      'kubernetes': t('nav.kubernetes'),
      'compute': t('nav.compute'),
      'networks': t('nav.networks'),
      'azure': t('nav.azure'),
      // 리소스 타입
      'vms': t('nav.vms'),
      'images': t('nav.images'),
      'snapshots': t('nav.snapshots'),
      'clusters': t('nav.clusters'),
      'node-pools': t('nav.nodePools'),
      'node-groups': t('nav.nodeGroups'),
      'nodes': t('nav.nodes'),
      'vpcs': t('nav.vpcs'),
      'subnets': t('nav.subnets'),
      'security-groups': t('nav.securityGroups'),
      // 관리 페이지
      'credentials': t('nav.credentials'),
      'workspaces': t('workspace.title'),
      'settings': t('common.settings'),
      'members': t('workspace.members'),
      'overview': t('workspace.overview') || 'Overview',
      'dashboard': t('nav.dashboard'),
      'profile': t('user.profile'),
      'cost-analysis': t('nav.costAnalysis'),
      'exports': t('nav.exports'),
      'notifications': t('nav.notifications'),
      'create': t('common.create'),
    };

    // 7. 현재 경로 추적 및 중복 href 방지를 위한 Set
    let currentPath = '';
    const seenHrefs = new Set<string>(); // 중복 방지를 위해 본 href 추적
    
    // 8. 각 경로 세그먼트를 순회하며 breadcrumb 항목 생성
    paths.forEach((segment) => {
      // 8-1. UUID 형식의 세그먼트는 breadcrumb에 추가하지 않음 (동적 라우트 파라미터)
      if (segment.match(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i)) {
        return;
      }
      
      // 8-2. 라벨이 빈 문자열인 세그먼트는 건너뛰기 (w, c 등 경로 구조의 일부)
      const label = pathLabels[segment];
      if (label === '') {
        return;
      }
      
      // 8-3. 현재 경로에 세그먼트 추가
      currentPath += `/${segment}`;
      
      // 8-4. 라벨이 없으면 첫 글자 대문자로 변환
      const displayLabel = label || segment.charAt(0).toUpperCase() + segment.slice(1);

      // 8-5. 1 depth 경로인지 확인하고 기본 하위 경로로 리다이렉트
      let href = currentPath;
      if (defaultPaths[currentPath]) {
        href = defaultPaths[currentPath];
      }

      // 8-6. 워크스페이스 ID가 있으면 쿼리 파라미터로 추가
      if (workspaceId && (defaultPaths[currentPath] || currentPath.startsWith('/compute') || currentPath.startsWith('/kubernetes') || currentPath.startsWith('/networks'))) {
        href = `${href}?workspaceId=${workspaceId}`;
      }

      // 8-7. 중복 href 방지: 이미 추가된 href면 스킵
      if (seenHrefs.has(href)) {
        return;
      }

      // 8-8. href를 Set에 추가하고 breadcrumb 항목에 추가
      seenHrefs.add(href);
      result.push({
        label: displayLabel,
        href,
      });
    });

    // 9. 동적 리소스 이름이 있으면 마지막 항목의 라벨을 업데이트 (예: 클러스터 이름)
    if (resourceName && result.length > 0) {
      result[result.length - 1].label = resourceName;
    }

    return result;
  }, [pathname, items, workspaceId, t, resourceName]);

  // 10. Breadcrumb 항목이 1개 이하면 렌더링하지 않음 (홈만 있는 경우)
  if (breadcrumbItems.length <= 1) {
    return null;
  }

  return (
    <nav
      className={cn('flex items-center space-x-1 text-sm text-muted-foreground overflow-hidden', className)}
      aria-label={t('common.breadcrumb')}
    >
      <ol className="flex items-center space-x-1 min-w-0 overflow-hidden">
        {breadcrumbItems.map((item, index) => {
          // 11. 현재 항목의 위치 확인
          const isLast = index === breadcrumbItems.length - 1;
          const isFirst = index === 0;
          
          return (
            <li 
              key={`${item.href}-${index}`} 
              className={cn(
                'flex items-center flex-shrink-0',
                !isFirst && !isLast && 'hidden sm:flex', // 중간 항목은 작은 화면에서 숨김 (반응형)
                isLast && 'min-w-0 overflow-hidden' // 마지막 항목만 truncate 가능
              )}
            >
              {isFirst ? (
                // 12. 첫 번째 항목: 홈 아이콘으로 표시
                <Link
                  href={item.href}
                  className="flex items-center hover:text-foreground transition-colors flex-shrink-0"
                  aria-label={t('common.home')}
                >
                  <Home className="h-4 w-4" />
                </Link>
              ) : (
                <>
                  {/* 13. 구분자: ChevronRight 아이콘 */}
                  <ChevronRight className="h-4 w-4 mx-1 text-muted-foreground flex-shrink-0" />
                  {isLast && !linkLast ? (
                    // 14. 마지막 항목: linkLast가 false면 span으로 표시 (현재 페이지)
                    <span 
                      className="font-medium text-foreground truncate" 
                      aria-current="page"
                      title={item.label}
                    >
                      {item.icon && <item.icon className="h-4 w-4 mr-1 inline flex-shrink-0" />}
                      <span className="truncate">{item.label}</span>
                    </span>
                  ) : (
                    // 15. 중간 항목 또는 linkLast가 true인 마지막 항목: Link로 표시
                    <Link
                      href={item.href}
                      className="hover:text-foreground transition-colors flex items-center truncate"
                      title={item.label}
                    >
                      {item.icon && <item.icon className="h-4 w-4 mr-1 flex-shrink-0" />}
                      <span className="truncate">{item.label}</span>
                    </Link>
                  )}
                </>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}




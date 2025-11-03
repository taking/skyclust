import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useWorkspaceStore } from '@/store/workspace';
import { workspaceService } from '@/services/workspace';
import { useQuery } from '@tanstack/react-query';

/**
 * URL 쿼리 파라미터에서 workspace ID를 읽어와서 store에 설정하는 hook
 * 새로고침 시에도 workspace 정보를 유지할 수 있도록 함
 */
export function useWorkspaceFromUrl() {
  const router = useRouter();
  const { currentWorkspace, setCurrentWorkspace, workspaces, setWorkspaces } = useWorkspaceStore();
  const [workspaceIdFromUrl, setWorkspaceIdFromUrl] = useState<string | null>(null);
  
  // URL에서 workspace ID 읽기 (클라이언트 사이드)
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search);
      const workspaceId = params.get('workspaceId');
      setWorkspaceIdFromUrl(workspaceId);
    }
  }, []); // 마운트 시 한 번만 실행

  // URL 변경 감지를 위한 이벤트 리스너
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleLocationChange = () => {
      const params = new URLSearchParams(window.location.search);
      const workspaceId = params.get('workspaceId');
      setWorkspaceIdFromUrl(workspaceId);
    };

    // popstate 이벤트 (뒤로/앞으로 가기)
    window.addEventListener('popstate', handleLocationChange);
    
    return () => {
      window.removeEventListener('popstate', handleLocationChange);
    };
  }, []);

  // Workspaces 목록 가져오기 (자주 변경되지 않으므로 긴 staleTime)
  const { data: fetchedWorkspaces = [], isLoading: isLoadingWorkspaces } = useQuery({
    queryKey: ['workspaces'],
    queryFn: workspaceService.getWorkspaces,
    staleTime: 10 * 60 * 1000, // 10분 - 워크스페이스 목록은 자주 변경되지 않음
    gcTime: 30 * 60 * 1000, // 30분 - GC 시간
    retry: 3,
    retryDelay: 1000,
  });

  // Workspaces 목록이 로드되면 store에 설정
  useEffect(() => {
    if (fetchedWorkspaces.length > 0 && workspaces.length === 0) {
      setWorkspaces(fetchedWorkspaces);
    }
  }, [fetchedWorkspaces, workspaces.length, setWorkspaces]);

  // URL에 workspace ID가 있고, 현재 선택된 workspace와 다르면 업데이트
  useEffect(() => {
    if (!fetchedWorkspaces.length) return;

    if (workspaceIdFromUrl) {
      const workspaceFromUrl = fetchedWorkspaces.find(w => w.id === workspaceIdFromUrl);
      
      if (workspaceFromUrl && currentWorkspace?.id !== workspaceIdFromUrl) {
        setCurrentWorkspace(workspaceFromUrl);
      } else if (!workspaceFromUrl) {
        // URL의 workspace ID가 유효하지 않으면 제거
        if (typeof window !== 'undefined') {
          const currentPath = window.location.pathname;
          const currentParams = new URLSearchParams(window.location.search);
          currentParams.delete('workspaceId');
          
          const newUrl = currentParams.toString() 
            ? `${currentPath}?${currentParams.toString()}`
            : currentPath;
          
          router.replace(newUrl, { scroll: false });
          setWorkspaceIdFromUrl(null);
        }
      }
    } else if (currentWorkspace) {
      // URL에 workspace ID가 없는데 store에 workspace가 있으면 URL 업데이트
      if (typeof window !== 'undefined') {
        const currentPath = window.location.pathname;
        const currentParams = new URLSearchParams(window.location.search);
        currentParams.set('workspaceId', currentWorkspace.id);
        
        router.replace(`${currentPath}?${currentParams.toString()}`, { scroll: false });
        setWorkspaceIdFromUrl(currentWorkspace.id);
      }
    } else {
      // URL에도 없고 store에도 없으면 첫 번째 workspace 선택 및 URL 업데이트
      const firstWorkspace = fetchedWorkspaces[0];
      if (firstWorkspace && typeof window !== 'undefined') {
        setCurrentWorkspace(firstWorkspace);
        const currentPath = window.location.pathname;
        const currentParams = new URLSearchParams(window.location.search);
        currentParams.set('workspaceId', firstWorkspace.id);
        
        router.replace(`${currentPath}?${currentParams.toString()}`, { scroll: false });
        setWorkspaceIdFromUrl(firstWorkspace.id);
      }
    }
  }, [workspaceIdFromUrl, currentWorkspace?.id, fetchedWorkspaces, setCurrentWorkspace, router]);

  /**
   * Workspace 변경 시 URL도 함께 업데이트
   */
  const changeWorkspace = (workspaceId: string) => {
    const workspace = fetchedWorkspaces.find(w => w.id === workspaceId);
    if (workspace && typeof window !== 'undefined') {
      setCurrentWorkspace(workspace);
      const currentPath = window.location.pathname;
      const currentParams = new URLSearchParams(window.location.search);
      currentParams.set('workspaceId', workspaceId);
      
      router.replace(`${currentPath}?${currentParams.toString()}`, { scroll: false });
      setWorkspaceIdFromUrl(workspaceId);
    }
  };

  return {
    currentWorkspace,
    workspaces: fetchedWorkspaces,
    isLoadingWorkspaces,
    changeWorkspace,
    workspaceIdFromUrl,
  };
}


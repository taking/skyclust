/**
 * Workspace Store
 * 워크스페이스 상태 관리 스토어
 * 
 * Zustand Devtools 지원 (개발 환경에서만 활성화)
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { StateCreator } from 'zustand';
import { Workspace } from '@/lib/types';

/**
 * 개발 환경 확인
 */
const isDevelopment = process.env.NODE_ENV === 'development';

interface WorkspaceState {
  currentWorkspace: Workspace | null;
  workspaces: Workspace[];
  setCurrentWorkspace: (workspace: Workspace | null) => void;
  setWorkspaces: (workspaces: Workspace[]) => void;
  addWorkspace: (workspace: Workspace) => void;
  updateWorkspace: (workspace: Workspace) => void;
  removeWorkspace: (workspaceId: string) => void;
}

const storeCreator: StateCreator<WorkspaceState> = (set) => ({
  // 초기 상태
  currentWorkspace: null,
  workspaces: [],
  
  /**
   * 현재 워크스페이스 설정
   * @param workspace - 선택할 워크스페이스 (null이면 선택 해제)
   */
  setCurrentWorkspace: (workspace) => set({ currentWorkspace: workspace }),
  
  /**
   * 워크스페이스 목록 설정
   * @param workspaces - 워크스페이스 배열
   */
  setWorkspaces: (workspaces) => set({ workspaces }),
  
  /**
   * 워크스페이스 추가
   * 기존 목록에 새로운 워크스페이스를 추가합니다.
   * @param workspace - 추가할 워크스페이스
   */
  addWorkspace: (workspace) =>
    set((state) => ({
      workspaces: [...state.workspaces, workspace],
    })),
  
  /**
   * 워크스페이스 업데이트
   * 목록과 현재 워크스페이스를 모두 업데이트합니다.
   * @param workspace - 업데이트할 워크스페이스
   */
  updateWorkspace: (workspace) =>
    set((state) => ({
      // 1. 워크스페이스 목록에서 해당 ID의 워크스페이스 업데이트
      workspaces: state.workspaces.map((w) =>
        w.id === workspace.id ? workspace : w
      ),
      // 2. 현재 워크스페이스가 업데이트 대상이면 함께 업데이트
      currentWorkspace:
        state.currentWorkspace?.id === workspace.id
          ? workspace
          : state.currentWorkspace,
    })),
  
  /**
   * 워크스페이스 제거
   * 목록에서 제거하고, 현재 워크스페이스가 제거 대상이면 다른 워크스페이스로 자동 전환합니다.
   * @param workspaceId - 제거할 워크스페이스 ID
   */
  removeWorkspace: (workspaceId) =>
    set((state) => {
      // 1. 워크스페이스 목록에서 제거
      const updatedWorkspaces = state.workspaces.filter((w) => w.id !== workspaceId);
      
      // 2. 현재 워크스페이스가 제거 대상인 경우
      if (state.currentWorkspace?.id === workspaceId) {
        // 2-1. 다른 워크스페이스가 있으면 첫 번째 워크스페이스로 전환
        if (updatedWorkspaces.length > 0) {
          return {
            workspaces: updatedWorkspaces,
            currentWorkspace: updatedWorkspaces[0],
          };
        }
        // 2-2. 워크스페이스가 없으면 선택 해제
        return {
          workspaces: updatedWorkspaces,
          currentWorkspace: null,
        };
      }
      
      // 3. 현재 워크스페이스가 제거 대상이 아니면 그대로 유지
      return {
        workspaces: updatedWorkspaces,
        currentWorkspace: state.currentWorkspace,
      };
    }),
});

// 개발 환경에서만 devtools 적용
export const useWorkspaceStore = create<WorkspaceState>()(
  isDevelopment 
    ? (devtools(storeCreator, { name: 'WorkspaceStore' }) as StateCreator<WorkspaceState>)
    : storeCreator
);

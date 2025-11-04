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
  currentWorkspace: null,
  workspaces: [],
  setCurrentWorkspace: (workspace) => set({ currentWorkspace: workspace }),
  setWorkspaces: (workspaces) => set({ workspaces }),
  addWorkspace: (workspace) =>
    set((state) => ({
      workspaces: [...state.workspaces, workspace],
    })),
  updateWorkspace: (workspace) =>
    set((state) => ({
      workspaces: state.workspaces.map((w) =>
        w.id === workspace.id ? workspace : w
      ),
      currentWorkspace:
        state.currentWorkspace?.id === workspace.id
          ? workspace
          : state.currentWorkspace,
    })),
  removeWorkspace: (workspaceId) =>
    set((state) => ({
      workspaces: state.workspaces.filter((w) => w.id !== workspaceId),
      currentWorkspace:
        state.currentWorkspace?.id === workspaceId
          ? null
          : state.currentWorkspace,
    })),
});

// 개발 환경에서만 devtools 적용
export const useWorkspaceStore = create<WorkspaceState>()(
  isDevelopment 
    ? (devtools(storeCreator, { name: 'WorkspaceStore' }) as StateCreator<WorkspaceState>)
    : storeCreator
);

/**
 * Workspace Store
 * 워크스페이스 상태 관리 스토어 - localStorage persist 추가
 * 
 * 개선 사항:
 * - localStorage에 마지막 선택된 workspace 저장
 * - 자동 선택 기능
 * - Workspace Switcher 지원
 */

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { devtools } from 'zustand/middleware';
import type { StateCreator } from 'zustand';
import { Workspace } from '@/lib/types';

const isDevelopment = process.env.NODE_ENV === 'development';

interface WorkspaceState {
  currentWorkspace: Workspace | null;
  workspaces: Workspace[];
  lastSelectedWorkspaceId: string | null;
  
  setCurrentWorkspace: (workspace: Workspace | null) => void;
  setWorkspaces: (workspaces: Workspace[]) => void;
  addWorkspace: (workspace: Workspace) => void;
  updateWorkspace: (workspace: Workspace) => void;
  removeWorkspace: (workspaceId: string) => void;
  
  autoSelectWorkspace: (workspaces: Workspace[]) => Workspace | null;
  getLastSelectedWorkspace: (workspaces: Workspace[]) => Workspace | null;
}

const storeCreator: StateCreator<WorkspaceState> = (set, get) => ({
  currentWorkspace: null,
  workspaces: [],
  lastSelectedWorkspaceId: null,
  
  setCurrentWorkspace: (workspace) => 
    set({ 
      currentWorkspace: workspace,
      lastSelectedWorkspaceId: workspace?.id || null,
    }),
  
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
    set((state) => {
      const updatedWorkspaces = state.workspaces.filter((w) => w.id !== workspaceId);
      
      if (state.currentWorkspace?.id === workspaceId) {
        const autoSelected = get().autoSelectWorkspace(updatedWorkspaces);
        return {
          workspaces: updatedWorkspaces,
          currentWorkspace: autoSelected,
          lastSelectedWorkspaceId: autoSelected?.id || null,
        };
      }
      
      return {
        workspaces: updatedWorkspaces,
        currentWorkspace: state.currentWorkspace,
      };
    }),
  
  autoSelectWorkspace: (workspaces) => {
    if (workspaces.length === 0) return null;
    
    const { lastSelectedWorkspaceId } = get();
    
    if (lastSelectedWorkspaceId) {
      const lastSelected = workspaces.find(w => w.id === lastSelectedWorkspaceId);
      if (lastSelected) {
        return lastSelected;
      }
    }
    
    return workspaces[0];
  },
  
  getLastSelectedWorkspace: (workspaces) => {
    const { lastSelectedWorkspaceId } = get();
    if (!lastSelectedWorkspaceId) return null;
    return workspaces.find(w => w.id === lastSelectedWorkspaceId) || null;
  },
});

const persistedStore = persist(
  storeCreator,
  {
    name: 'workspace-storage',
    storage: createJSONStorage(() => localStorage),
    partialize: (state) => ({
      lastSelectedWorkspaceId: state.lastSelectedWorkspaceId,
    }),
  }
);

export const useWorkspaceStore = create<WorkspaceState>()(
  isDevelopment 
    ? (devtools(persistedStore as StateCreator<WorkspaceState>, { name: 'WorkspaceStore' }) as StateCreator<WorkspaceState>)
    : persistedStore
);

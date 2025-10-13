import { create } from 'zustand';
import { Workspace } from '@/lib/types';

interface WorkspaceState {
  currentWorkspace: Workspace | null;
  workspaces: Workspace[];
  setCurrentWorkspace: (workspace: Workspace | null) => void;
  setWorkspaces: (workspaces: Workspace[]) => void;
  addWorkspace: (workspace: Workspace) => void;
  updateWorkspace: (workspace: Workspace) => void;
  removeWorkspace: (workspaceId: string) => void;
}

export const useWorkspaceStore = create<WorkspaceState>((set) => ({
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
}));

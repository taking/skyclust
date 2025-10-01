import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

// Types
export interface User {
  id: string
  email: string
  name: string
  createdAt: string
}

export interface Workspace {
  id: string
  name: string
  description?: string
  ownerId: string
  settings: Record<string, any>
  createdAt: string
  updatedAt: string
}

export interface VM {
  id: string
  name: string
  status: string
  type: string
  region: string
  provider: string
  publicIP?: string
  privateIP?: string
  createdAt: string
  tags: Record<string, string>
}

export interface Credential {
  id: string
  name: string
  provider: string
  workspaceId: string
  createdAt: string
  updatedAt: string
}

export interface Execution {
  id: string
  name: string
  status: 'pending' | 'running' | 'success' | 'failed'
  command: string
  output?: string
  error?: string
  startedAt: string
  completedAt?: string
}

// Auth Store
interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
}

interface AuthActions {
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  setUser: (user: User) => void
  setToken: (token: string) => void
  setError: (error: string | null) => void
  setLoading: (loading: boolean) => void
}

export const useAuthStore = create<AuthState & AuthActions>()(
  devtools(
    persist(
      immer((set, get) => ({
        // State
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,

        // Actions
        login: async (email: string, password: string) => {
          set((state) => {
            state.isLoading = true
            state.error = null
          })

          try {
            const response = await fetch('/api/v1/auth/login', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({ email, password }),
            })

            if (!response.ok) {
              const error = await response.json()
              throw new Error(error.error?.message || 'Login failed')
            }

            const data = await response.json()
            
            set((state) => {
              state.user = data.user
              state.token = data.token
              state.isAuthenticated = true
              state.isLoading = false
              state.error = null
            })
          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Login failed'
              state.isLoading = false
            })
            throw error
          }
        },

        logout: () => {
          set((state) => {
            state.user = null
            state.token = null
            state.isAuthenticated = false
            state.error = null
          })
        },

        setUser: (user: User) => {
          set((state) => {
            state.user = user
            state.isAuthenticated = true
          })
        },

        setToken: (token: string) => {
          set((state) => {
            state.token = token
          })
        },

        setError: (error: string | null) => {
          set((state) => {
            state.error = error
          })
        },

        setLoading: (loading: boolean) => {
          set((state) => {
            state.isLoading = loading
          })
        },
      })),
      {
        name: 'auth-storage',
        partialize: (state) => ({
          user: state.user,
          token: state.token,
          isAuthenticated: state.isAuthenticated,
        }),
      }
    ),
    { name: 'auth-store' }
  )
)

// Workspace Store
interface WorkspaceState {
  workspaces: Workspace[]
  currentWorkspace: Workspace | null
  isLoading: boolean
  error: string | null
}

interface WorkspaceActions {
  fetchWorkspaces: () => Promise<void>
  createWorkspace: (name: string, description?: string) => Promise<void>
  updateWorkspace: (id: string, updates: Partial<Workspace>) => Promise<void>
  deleteWorkspace: (id: string) => Promise<void>
  setCurrentWorkspace: (workspace: Workspace | null) => void
  setError: (error: string | null) => void
  setLoading: (loading: boolean) => void
}

export const useWorkspaceStore = create<WorkspaceState & WorkspaceActions>()(
  devtools(
    immer((set, get) => ({
      // State
      workspaces: [],
      currentWorkspace: null,
      isLoading: false,
      error: null,

      // Actions
      fetchWorkspaces: async () => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const { token } = useAuthStore.getState()
          const response = await fetch('/api/v1/workspaces', {
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          })

          if (!response.ok) {
            throw new Error('Failed to fetch workspaces')
          }

          const data = await response.json()
          
          set((state) => {
            state.workspaces = data.workspaces || []
            state.isLoading = false
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to fetch workspaces'
            state.isLoading = false
          })
        }
      },

      createWorkspace: async (name: string, description?: string) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const { token } = useAuthStore.getState()
          const response = await fetch('/api/v1/workspaces', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify({ name, description }),
          })

          if (!response.ok) {
            throw new Error('Failed to create workspace')
          }

          const workspace = await response.json()
          
          set((state) => {
            state.workspaces.push(workspace)
            state.isLoading = false
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to create workspace'
            state.isLoading = false
          })
        }
      },

      updateWorkspace: async (id: string, updates: Partial<Workspace>) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const { token } = useAuthStore.getState()
          const response = await fetch(`/api/v1/workspaces/${id}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify(updates),
          })

          if (!response.ok) {
            throw new Error('Failed to update workspace')
          }

          const updatedWorkspace = await response.json()
          
          set((state) => {
            const index = state.workspaces.findIndex(w => w.id === id)
            if (index !== -1) {
              state.workspaces[index] = updatedWorkspace
            }
            if (state.currentWorkspace?.id === id) {
              state.currentWorkspace = updatedWorkspace
            }
            state.isLoading = false
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to update workspace'
            state.isLoading = false
          })
        }
      },

      deleteWorkspace: async (id: string) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const { token } = useAuthStore.getState()
          const response = await fetch(`/api/v1/workspaces/${id}`, {
            method: 'DELETE',
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          })

          if (!response.ok) {
            throw new Error('Failed to delete workspace')
          }

          set((state) => {
            state.workspaces = state.workspaces.filter(w => w.id !== id)
            if (state.currentWorkspace?.id === id) {
              state.currentWorkspace = null
            }
            state.isLoading = false
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to delete workspace'
            state.isLoading = false
          })
        }
      },

      setCurrentWorkspace: (workspace: Workspace | null) => {
        set((state) => {
          state.currentWorkspace = workspace
        })
      },

      setError: (error: string | null) => {
        set((state) => {
          state.error = error
        })
      },

      setLoading: (loading: boolean) => {
        set((state) => {
          state.isLoading = loading
        })
      },
    })),
    { name: 'workspace-store' }
  )
)

// VM Store
interface VMState {
  vms: VM[]
  isLoading: boolean
  error: string | null
  filters: {
    provider?: string
    status?: string
    region?: string
  }
}

interface VMActions {
  fetchVMs: (workspaceId: string) => Promise<void>
  createVM: (workspaceId: string, vm: Partial<VM>) => Promise<void>
  deleteVM: (workspaceId: string, vmId: string) => Promise<void>
  startVM: (workspaceId: string, vmId: string) => Promise<void>
  stopVM: (workspaceId: string, vmId: string) => Promise<void>
  setFilters: (filters: Partial<VMState['filters']>) => void
  setError: (error: string | null) => void
  setLoading: (loading: boolean) => void
}

export const useVMStore = create<VMState & VMActions>()(
  devtools(
    immer((set, get) => ({
      // State
      vms: [],
      isLoading: false,
      error: null,
      filters: {},

      // Actions
      fetchVMs: async (workspaceId: string) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const { token } = useAuthStore.getState()
          const { filters } = get()
          
          const params = new URLSearchParams()
          if (filters.provider) params.append('provider', filters.provider)
          if (filters.status) params.append('status', filters.status)
          if (filters.region) params.append('region', filters.region)

          const response = await fetch(`/api/v1/workspaces/${workspaceId}/vms?${params}`, {
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          })

          if (!response.ok) {
            throw new Error('Failed to fetch VMs')
          }

          const data = await response.json()
          
          set((state) => {
            state.vms = data.vms || []
            state.isLoading = false
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to fetch VMs'
            state.isLoading = false
          })
        }
      },

      createVM: async (workspaceId: string, vm: Partial<VM>) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const { token } = useAuthStore.getState()
          const response = await fetch(`/api/v1/workspaces/${workspaceId}/vms`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify(vm),
          })

          if (!response.ok) {
            throw new Error('Failed to create VM')
          }

          const newVM = await response.json()
          
          set((state) => {
            state.vms.push(newVM.vm)
            state.isLoading = false
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to create VM'
            state.isLoading = false
          })
        }
      },

      deleteVM: async (workspaceId: string, vmId: string) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const { token } = useAuthStore.getState()
          const response = await fetch(`/api/v1/workspaces/${workspaceId}/vms/${vmId}`, {
            method: 'DELETE',
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          })

          if (!response.ok) {
            throw new Error('Failed to delete VM')
          }

          set((state) => {
            state.vms = state.vms.filter(vm => vm.id !== vmId)
            state.isLoading = false
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to delete VM'
            state.isLoading = false
          })
        }
      },

      startVM: async (workspaceId: string, vmId: string) => {
        try {
          const { token } = useAuthStore.getState()
          const response = await fetch(`/api/v1/workspaces/${workspaceId}/vms/${vmId}/start`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          })

          if (!response.ok) {
            throw new Error('Failed to start VM')
          }
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to start VM'
          })
        }
      },

      stopVM: async (workspaceId: string, vmId: string) => {
        try {
          const { token } = useAuthStore.getState()
          const response = await fetch(`/api/v1/workspaces/${workspaceId}/vms/${vmId}/stop`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          })

          if (!response.ok) {
            throw new Error('Failed to stop VM')
          }
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to stop VM'
          })
        }
      },

      setFilters: (filters: Partial<VMState['filters']>) => {
        set((state) => {
          state.filters = { ...state.filters, ...filters }
        })
      },

      setError: (error: string | null) => {
        set((state) => {
          state.error = error
        })
      },

      setLoading: (loading: boolean) => {
        set((state) => {
          state.isLoading = loading
        })
      },
    })),
    { name: 'vm-store' }
  )
)

// UI Store for global UI state
interface UIState {
  theme: 'light' | 'dark'
  sidebarCollapsed: boolean
  notifications: Notification[]
  modals: {
    createWorkspace: boolean
    createVM: boolean
    createCredential: boolean
  }
}

interface UIActions {
  setTheme: (theme: 'light' | 'dark') => void
  toggleSidebar: () => void
  addNotification: (notification: Omit<Notification, 'id'>) => void
  removeNotification: (id: string) => void
  openModal: (modal: keyof UIState['modals']) => void
  closeModal: (modal: keyof UIState['modals']) => void
  closeAllModals: () => void
}

interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message: string
  duration?: number
}

export const useUIStore = create<UIState & UIActions>()(
  devtools(
    persist(
      immer((set, get) => ({
        // State
        theme: 'light',
        sidebarCollapsed: false,
        notifications: [],
        modals: {
          createWorkspace: false,
          createVM: false,
          createCredential: false,
        },

        // Actions
        setTheme: (theme: 'light' | 'dark') => {
          set((state) => {
            state.theme = theme
          })
        },

        toggleSidebar: () => {
          set((state) => {
            state.sidebarCollapsed = !state.sidebarCollapsed
          })
        },

        addNotification: (notification: Omit<Notification, 'id'>) => {
          const id = Math.random().toString(36).substr(2, 9)
          set((state) => {
            state.notifications.push({ ...notification, id })
          })

          // Auto-remove notification after duration
          if (notification.duration !== 0) {
            setTimeout(() => {
              get().removeNotification(id)
            }, notification.duration || 5000)
          }
        },

        removeNotification: (id: string) => {
          set((state) => {
            state.notifications = state.notifications.filter(n => n.id !== id)
          })
        },

        openModal: (modal: keyof UIState['modals']) => {
          set((state) => {
            state.modals[modal] = true
          })
        },

        closeModal: (modal: keyof UIState['modals']) => {
          set((state) => {
            state.modals[modal] = false
          })
        },

        closeAllModals: () => {
          set((state) => {
            Object.keys(state.modals).forEach(key => {
              state.modals[key as keyof UIState['modals']] = false
            })
          })
        },
      })),
      {
        name: 'ui-storage',
        partialize: (state) => ({
          theme: state.theme,
          sidebarCollapsed: state.sidebarCollapsed,
        }),
      }
    ),
    { name: 'ui-store' }
  )
)

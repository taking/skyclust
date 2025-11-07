/**
 * Constants
 * 
 * 애플리케이션 전역 상수 정의
 * Event Names, UI Constants, Keyboard Shortcuts 등을 중앙에서 관리
 */

/**
 * Event Names
 * CustomEvent 이름 상수
 */
export const EVENTS = {
  // Create Dialog Events
  CREATE_DIALOG: {
    VM: 'open-create-vm-dialog',
    CLUSTER: 'open-create-cluster-dialog',
    VPC: 'open-create-vpc-dialog',
    SUBNET: 'open-create-subnet-dialog',
    SECURITY_GROUP: 'open-create-security-group-dialog',
    CREDENTIAL: 'open-create-credential-dialog',
  },
  // UI Events
  UI: {
    TOGGLE_SIDEBAR: 'toggle-sidebar',
    SHOW_KEYBOARD_SHORTCUTS: 'show-keyboard-shortcuts',
  },
} as const;

/**
 * UI Constants
 * 페이지네이션, 필터, 테이블 등 UI 관련 상수
 */
export const UI = {
  // Pagination
  PAGINATION: {
    DEFAULT_PAGE_SIZE: 20 as number,
    PAGE_SIZE_OPTIONS: [10, 20, 50, 100] as number[],
    DEFAULT_PAGE: 1 as number,
  },
  // Search
  SEARCH: {
    DEBOUNCE_MS: 300,
    MIN_QUERY_LENGTH: 1,
  },
  // Filter
  FILTER: {
    DEBOUNCE_MS: 300,
  },
  // Table
  TABLE: {
    DEFAULT_SORT_DIRECTION: 'asc' as const,
    MAX_ROWS_WITHOUT_PAGINATION: 50,
  },
} as const;

/**
 * Keyboard Shortcuts
 * 키보드 단축키 관련 상수
 */
export const KEYBOARD_SHORTCUTS = {
  // Navigation
  NAVIGATION: {
    DASHBOARD: 'h',
    COMPUTE: 'v',
    KUBERNETES: 'k',
    NETWORKS: 'n',
    CREDENTIALS: 'c',
  },
  // Actions
  ACTIONS: {
    CREATE_NEW: { key: 'N', shiftKey: true },
    MENU_LIST: { key: 'M', shiftKey: true },
    HELP: { key: '?', shiftKey: true },
  },
  // Common
  COMMON: {
    ESCAPE: 'Escape',
    DELETE: 'Delete',
    SAVE: { key: 's', ctrlKey: true },
    SEARCH: { key: 'k', ctrlKey: true },
  },
} as const;

/**
 * Storage Keys
 * localStorage/sessionStorage 키 상수
 */
export const STORAGE_KEYS = {
  ERROR_LOGS: 'skyclust-error-logs',
  OFFLINE_QUEUE: 'skyclust-offline-queue',
  AUTH_STORAGE: 'auth-storage',
  WORKSPACE_STORAGE: 'workspace-storage',
  CREDENTIAL_CONTEXT_STORAGE: 'credential-context-storage',
} as const;

/**
 * Time Constants
 * 시간 관련 상수 (밀리초 단위)
 */
export const TIME = {
  SECOND: 1000,
  MINUTE: 60 * 1000,
  HOUR: 60 * 60 * 1000,
  DAY: 24 * 60 * 60 * 1000,
  // Polling intervals
  POLLING: {
    REALTIME: 30 * 1000, // 30 seconds
    MONITORING: 60 * 1000, // 1 minute
    STANDARD: 5 * 60 * 1000, // 5 minutes
  },
  // Debounce
  DEBOUNCE: {
    SEARCH: 300,
    INPUT: 500,
    RESIZE: 250,
  },
  // Delays
  DELAY: {
    AUTH_HYDRATION: 300, // Auth hydration delay
    REGISTER_REDIRECT: 2000, // Register success redirect delay
    MIN_RECONNECT: 1000, // Minimum reconnection delay
    MAX_RECONNECT: 30000, // Maximum reconnection delay
  },
  // Retry
  RETRY: {
    DEFAULT_INTERVAL: 5000, // Default retry interval
    ERROR_HANDLER_DELAY: 1000, // Error handler retry delay
    DEFAULT_COUNT: 3, // Default retry count
  },
} as const;

/**
 * Connection Constants
 * 연결 관련 상수
 */
export const CONNECTION = {
  // SSE (Server-Sent Events)
  SSE: {
    MAX_RECONNECT_ATTEMPTS: 10,
    BASE_RECONNECT_DELAY: TIME.DELAY.MIN_RECONNECT,
    MAX_RECONNECT_DELAY: TIME.DELAY.MAX_RECONNECT,
  },
  // Offline Queue
  OFFLINE_QUEUE: {
    MAX_RETRIES: TIME.RETRY.DEFAULT_COUNT,
    RETRY_INTERVAL: TIME.RETRY.DEFAULT_INTERVAL,
    MAX_QUEUE_SIZE: 100,
  },
} as const;

/**
 * API Constants
 * API 관련 상수
 */
export const API = {
  // Request/Response
  REQUEST: {
    DEFAULT_TIMEOUT: 30000, // 30 seconds
    MAX_RETRIES: TIME.RETRY.DEFAULT_COUNT,
    RETRY_DELAY: TIME.RETRY.ERROR_HANDLER_DELAY,
  },
} as const;


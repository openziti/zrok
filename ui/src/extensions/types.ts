/**
 * zrok UI Extension System - Type Definitions
 *
 * This module defines the interfaces and types used by the extension system.
 * Extensions implement these interfaces to integrate with the zrok UI.
 */

import { ComponentType, ReactNode } from 'react';
import { Node, Edge } from '@xyflow/react';
import { User } from '../model/user';

/**
 * Main extension manifest interface.
 * Every extension must export a default object implementing this interface.
 */
export interface ExtensionManifest {
  /** Unique identifier for the extension (e.g., "acme-billing") */
  id: string;

  /** Human-readable display name */
  name: string;

  /** Semantic version string */
  version: string;

  /** Optional description */
  description?: string;

  /** Route extensions - add new pages to the UI */
  routes?: ExtensionRoute[];

  /** Navigation items - add buttons/links to the navbar */
  navItems?: ExtensionNavItem[];

  /** Panel extensions - extend or replace side panels */
  panelExtensions?: PanelExtension[];

  /** Custom graph node types */
  nodeTypes?: Record<string, ComponentType<any>>;

  /** Custom graph edge types */
  edgeTypes?: Record<string, ComponentType<any>>;

  /** Slot-based UI injections */
  slots?: Record<string, ComponentType<SlotProps>>;

  /** Initial state to add to the main store's extension namespace */
  initialState?: Record<string, unknown>;

  /**
   * Called when the extension is first loaded.
   * Use this for initialization, data fetching, subscriptions, etc.
   */
  onInit?: (context: ExtensionContext) => void | Promise<void>;

  /** Called when a user logs in */
  onUserLogin?: (user: User, context: ExtensionContext) => void;

  /** Called when a user logs out */
  onUserLogout?: (context: ExtensionContext) => void;
}

/**
 * Defines a route (page) added by an extension.
 */
export interface ExtensionRoute {
  /** URL path (e.g., "/billing", "/billing/invoices") */
  path: string;

  /** React component to render for this route */
  component: ComponentType<ExtensionRouteProps>;

  /** If true, only matches exact path (default: false) */
  exact?: boolean;

  /** If true, route requires authentication (default: true) */
  requiresAuth?: boolean;
}

/**
 * Props passed to extension route components.
 */
export interface ExtensionRouteProps {
  /** Current authenticated user (null if not authenticated) */
  user: User | null;

  /** Extension context for store access */
  context: ExtensionContext;

  /** Logout function */
  logout: () => void;
}

/**
 * Defines a navigation item added by an extension.
 */
export interface ExtensionNavItem {
  /** Unique identifier for this nav item */
  id: string;

  /** Display label */
  label: string;

  /** Icon component (optional) */
  icon?: ComponentType<{ fontSize?: 'small' | 'medium' | 'large' }>;

  /** Route path to navigate to (mutually exclusive with onClick) */
  path?: string;

  /** Custom click handler (mutually exclusive with path) */
  onClick?: () => void;

  /** Position in navbar: 'left' or 'right' (default: 'right') */
  position?: 'left' | 'right';

  /** Tooltip text */
  tooltip?: string;

  /** Sort order within position (lower = earlier, default: 0) */
  order?: number;

  /**
   * Visibility condition. Return false to hide the item.
   * Called with current user and extension state.
   */
  visible?: (user: User | null, extensionState: Record<string, unknown>) => boolean;
}

/**
 * Defines an extension to the side panel shown when selecting nodes.
 */
export interface PanelExtension {
  /**
   * Node types this panel applies to.
   * Use ["*"] for all node types, or specific types like ["account", "share"]
   */
  nodeTypes: string[];

  /**
   * How to position this extension relative to the base panel:
   * - 'before': Render above the base panel
   * - 'after': Render below the base panel
   * - 'tab': Add as a new tab (requires tabLabel)
   * - 'replace': Replace the entire base panel
   */
  position: 'before' | 'after' | 'tab' | 'replace';

  /** Component to render */
  component: ComponentType<PanelExtensionProps>;

  /** Tab label (required when position is 'tab') */
  tabLabel?: string;

  /** Tab icon (optional, used when position is 'tab') */
  tabIcon?: ComponentType;

  /** Sort order within position (lower = earlier, default: 0) */
  order?: number;
}

/**
 * Props passed to panel extension components.
 */
export interface PanelExtensionProps {
  /** The currently selected node */
  node: Node;

  /** Current authenticated user */
  user: User;

  /** Extension context */
  context: ExtensionContext;
}

/**
 * Props passed to slot components.
 */
export interface SlotProps {
  /** Current user (may be null for unauthenticated slots) */
  user?: User | null;

  /** Currently selected node (for node-related slots) */
  selectedNode?: Node | null;

  /** Extension context */
  context: ExtensionContext;

  /** Additional props passed to the slot */
  [key: string]: unknown;
}

/**
 * Context object provided to extensions for interacting with the zrok UI.
 */
export interface ExtensionContext {
  /** The extension's ID */
  extensionId: string;

  /**
   * Get the extension's state from the main store.
   * Returns undefined if no state has been set.
   */
  getState: <T = Record<string, unknown>>() => T | undefined;

  /**
   * Update the extension's state in the main store.
   * Performs a shallow merge with existing state.
   */
  setState: <T = Record<string, unknown>>(state: Partial<T>) => void;

  /**
   * Subscribe to changes in the extension's state.
   * Returns an unsubscribe function.
   */
  subscribe: <T = Record<string, unknown>>(
    selector: (state: T) => unknown,
    callback: (selectedValue: unknown, previousValue: unknown) => void
  ) => () => void;

  /**
   * Get the current authenticated user.
   */
  getUser: () => User | null;

  /**
   * Subscribe to user changes (login/logout).
   * Returns an unsubscribe function.
   */
  subscribeToUser: (callback: (user: User | null) => void) => () => void;

  /**
   * Get the currently selected node in the visualizer.
   */
  getSelectedNode: () => Node | null;

  /**
   * Subscribe to node selection changes.
   * Returns an unsubscribe function.
   */
  subscribeToSelectedNode: (callback: (node: Node | null) => void) => () => void;

  /**
   * Navigate to a route programmatically.
   */
  navigate: (path: string) => void;

  /**
   * Show a notification/toast message.
   */
  notify: (message: string, severity?: 'info' | 'success' | 'warning' | 'error') => void;
}

/**
 * Well-known slot names where extensions can inject UI.
 */
export const SLOTS = {
  // NavBar slots
  NAVBAR_LEFT: 'navbar-left',
  NAVBAR_RIGHT: 'navbar-right',
  NAVBAR_CENTER: 'navbar-center',

  // Account panel slots
  ACCOUNT_PANEL_TOP: 'account-panel-top',
  ACCOUNT_PANEL_BOTTOM: 'account-panel-bottom',
  ACCOUNT_PANEL_ACTIONS: 'account-panel-actions',

  // Environment panel slots
  ENVIRONMENT_PANEL_TOP: 'environment-panel-top',
  ENVIRONMENT_PANEL_BOTTOM: 'environment-panel-bottom',

  // Share panel slots
  SHARE_PANEL_TOP: 'share-panel-top',
  SHARE_PANEL_BOTTOM: 'share-panel-bottom',

  // Main console area slots
  CONSOLE_TOP: 'console-top',
  CONSOLE_BOTTOM: 'console-bottom',
  CONSOLE_SIDEBAR: 'console-sidebar',

  // Login page slots
  LOGIN_TOP: 'login-top',
  LOGIN_BOTTOM: 'login-bottom',
} as const;

export type SlotName = typeof SLOTS[keyof typeof SLOTS];

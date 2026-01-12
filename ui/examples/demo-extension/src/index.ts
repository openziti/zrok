/**
 * Demo Extension for zrok UI
 *
 * This extension demonstrates all the extension capabilities:
 * - Custom routes (pages)
 * - Navigation items
 * - Panel extensions (tabs)
 * - Slot injections
 * - State management
 * - Lifecycle hooks
 */

import { ExtensionManifest, SLOTS } from '../../../src/extensions';
import DemoPage from './DemoPage';
import DemoSettingsPage from './DemoSettingsPage';
import AccountBillingTab from './AccountBillingTab';
import DemoNavbarSlot from './DemoNavbarSlot';
import DemoIcon from './DemoIcon';

// Define the extension's state interface
export interface DemoExtensionState {
  counter: number;
  lastVisited: string | null;
  settings: {
    enableFeatureX: boolean;
    theme: 'light' | 'dark';
  };
}

// Initial state
const initialState: DemoExtensionState = {
  counter: 0,
  lastVisited: null,
  settings: {
    enableFeatureX: true,
    theme: 'light',
  },
};

const manifest: ExtensionManifest = {
  id: 'demo-extension',
  name: 'Demo Extension',
  version: '1.0.0',
  description: 'Demonstrates zrok UI extension capabilities',

  // Initial state for the extension
  initialState,

  // Add new routes (pages)
  routes: [
    {
      path: '/demo',
      component: DemoPage,
      requiresAuth: true,
    },
    {
      path: '/demo/settings',
      component: DemoSettingsPage,
      requiresAuth: true,
    },
  ],

  // Add navigation items to the navbar
  navItems: [
    {
      id: 'demo-nav',
      label: 'Demo',
      icon: DemoIcon,
      path: '/demo',
      position: 'right',
      tooltip: 'Demo Extension Page',
      order: 10,
    },
  ],

  // Extend existing panels with tabs
  panelExtensions: [
    {
      nodeTypes: ['account'],
      position: 'tab',
      tabLabel: 'Billing',
      component: AccountBillingTab,
      order: 1,
    },
  ],

  // Inject components into slots
  slots: {
    [SLOTS.NAVBAR_RIGHT]: DemoNavbarSlot,
  },

  // Lifecycle hooks
  onInit: async (context) => {
    console.log('[Demo Extension] Initializing...');

    // Example: Subscribe to user changes
    context.subscribeToUser((user) => {
      if (user) {
        console.log('[Demo Extension] User logged in:', user.email);
        context.setState({ lastVisited: new Date().toISOString() });
      }
    });

    // Example: Subscribe to node selection
    context.subscribeToSelectedNode((node) => {
      if (node) {
        console.log('[Demo Extension] Node selected:', node.type, node.id);
      }
    });

    console.log('[Demo Extension] Initialized successfully');
  },

  onUserLogin: (user, context) => {
    console.log('[Demo Extension] User login hook:', user.email);
    context.notify(`Welcome back!`, 'success');
  },

  onUserLogout: (context) => {
    console.log('[Demo Extension] User logout hook');
    // Reset extension state on logout
    context.setState({
      counter: 0,
      lastVisited: null,
    });
  },
};

export default manifest;

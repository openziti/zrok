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
 * - Script injection (build-time and runtime)
 */

import { ExtensionManifest, SLOTS, ScriptDefinition } from '../../../src/extensions';
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

/**
 * Build-time script injection examples
 *
 * These scripts are injected into index.html during the Vite build process.
 * Use headScripts for scripts that need to load early (analytics, polyfills).
 * Use bodyScripts for scripts that can load after page content.
 *
 * To enable build-time injection, uncomment the extensionScriptsPlugin
 * in vite.config.ts and pass this extension to it.
 */

// Scripts to inject in <head> - loaded early
const headScripts: ScriptDefinition[] = [
  // Example: Inline script for early initialization
  {
    id: 'demo-analytics-init',
    content: `
      // Demo Analytics - Initialize early
      window.demoAnalytics = window.demoAnalytics || [];
      window.demoAnalytics.push(['init', { extensionId: 'demo-extension' }]);
      console.log('[Demo Extension] Analytics initialized (build-time head script)');
    `,
  },
  // Example: External script with async loading
  // Uncomment to test external script loading:
  // {
  //   id: 'demo-external-sdk',
  //   src: 'https://example.com/sdk.js',
  //   async: true,
  // },
];

// Scripts to inject before </body> - loaded after page content
const bodyScripts: ScriptDefinition[] = [
  // Example: Inline script for deferred operations
  {
    id: 'demo-tracking-script',
    content: `
      // Demo tracking - runs after page loads
      document.addEventListener('DOMContentLoaded', function() {
        console.log('[Demo Extension] Page loaded, tracking initialized (build-time body script)');
        if (window.demoAnalytics) {
          window.demoAnalytics.push(['pageview', { page: window.location.pathname }]);
        }
      });
    `,
  },
];

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

  // Build-time script injection
  // These are injected into index.html when using extensionScriptsPlugin in vite.config.ts
  headScripts,
  bodyScripts,

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

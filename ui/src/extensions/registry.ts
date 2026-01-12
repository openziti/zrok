/**
 * zrok UI Extension Registry
 *
 * Central registry for managing extensions. Extensions register themselves
 * here during application startup, and the UI queries the registry to
 * discover routes, nav items, panel extensions, etc.
 */

import { ComponentType } from 'react';
import {
  ExtensionManifest,
  ExtensionRoute,
  ExtensionNavItem,
  PanelExtension,
  ExtensionContext,
  SlotProps,
} from './types';
import { createExtensionContext } from './context';

class ExtensionRegistry {
  private extensions: Map<string, ExtensionManifest> = new Map();
  private contexts: Map<string, ExtensionContext> = new Map();
  private initialized: Set<string> = new Set();

  /**
   * Register an extension with the registry.
   * Should be called during application startup before rendering.
   */
  register(manifest: ExtensionManifest): void {
    if (this.extensions.has(manifest.id)) {
      console.warn(
        `[Extensions] Extension "${manifest.id}" is already registered. ` +
        `The previous registration will be overwritten.`
      );
    }

    // Validate manifest
    this.validateManifest(manifest);

    this.extensions.set(manifest.id, manifest);
    console.log(`[Extensions] Registered extension: ${manifest.name} v${manifest.version}`);
  }

  /**
   * Unregister an extension.
   */
  unregister(extensionId: string): void {
    this.extensions.delete(extensionId);
    this.contexts.delete(extensionId);
    this.initialized.delete(extensionId);
  }

  /**
   * Initialize all registered extensions.
   * Called after the store is ready and user state is loaded.
   */
  async initializeAll(
    navigate: (path: string) => void,
    notify: (message: string, severity?: 'info' | 'success' | 'warning' | 'error') => void
  ): Promise<void> {
    for (const [id, manifest] of this.extensions) {
      if (this.initialized.has(id)) continue;

      try {
        const context = createExtensionContext(id, navigate, notify);
        this.contexts.set(id, context);

        if (manifest.onInit) {
          await manifest.onInit(context);
        }

        this.initialized.add(id);
        console.log(`[Extensions] Initialized: ${manifest.name}`);
      } catch (error) {
        console.error(`[Extensions] Failed to initialize ${manifest.name}:`, error);
      }
    }
  }

  /**
   * Get context for a specific extension.
   */
  getContext(extensionId: string): ExtensionContext | undefined {
    return this.contexts.get(extensionId);
  }

  /**
   * Get all registered extensions.
   */
  getAll(): ExtensionManifest[] {
    return Array.from(this.extensions.values());
  }

  /**
   * Get a specific extension by ID.
   */
  get(extensionId: string): ExtensionManifest | undefined {
    return this.extensions.get(extensionId);
  }

  /**
   * Get all routes from all extensions.
   */
  getRoutes(): Array<ExtensionRoute & { extensionId: string }> {
    const routes: Array<ExtensionRoute & { extensionId: string }> = [];

    for (const [id, manifest] of this.extensions) {
      if (manifest.routes) {
        for (const route of manifest.routes) {
          routes.push({ ...route, extensionId: id });
        }
      }
    }

    return routes;
  }

  /**
   * Get all nav items from all extensions, sorted by position and order.
   */
  getNavItems(position?: 'left' | 'right'): Array<ExtensionNavItem & { extensionId: string }> {
    const items: Array<ExtensionNavItem & { extensionId: string }> = [];

    for (const [id, manifest] of this.extensions) {
      if (manifest.navItems) {
        for (const item of manifest.navItems) {
          const itemPosition = item.position || 'right';
          if (!position || itemPosition === position) {
            items.push({ ...item, extensionId: id });
          }
        }
      }
    }

    // Sort by order (default 0), then by name for stability
    return items.sort((a, b) => {
      const orderA = a.order ?? 0;
      const orderB = b.order ?? 0;
      if (orderA !== orderB) return orderA - orderB;
      return a.label.localeCompare(b.label);
    });
  }

  /**
   * Get panel extensions for a specific node type.
   */
  getPanelExtensions(
    nodeType: string,
    position?: PanelExtension['position']
  ): Array<PanelExtension & { extensionId: string }> {
    const extensions: Array<PanelExtension & { extensionId: string }> = [];

    for (const [id, manifest] of this.extensions) {
      if (manifest.panelExtensions) {
        for (const ext of manifest.panelExtensions) {
          const matchesType = ext.nodeTypes.includes('*') || ext.nodeTypes.includes(nodeType);
          const matchesPosition = !position || ext.position === position;

          if (matchesType && matchesPosition) {
            extensions.push({ ...ext, extensionId: id });
          }
        }
      }
    }

    // Sort by order
    return extensions.sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
  }

  /**
   * Get all custom node types from all extensions.
   */
  getNodeTypes(): Record<string, ComponentType<any>> {
    const nodeTypes: Record<string, ComponentType<any>> = {};

    for (const manifest of this.extensions.values()) {
      if (manifest.nodeTypes) {
        Object.assign(nodeTypes, manifest.nodeTypes);
      }
    }

    return nodeTypes;
  }

  /**
   * Get all custom edge types from all extensions.
   */
  getEdgeTypes(): Record<string, ComponentType<any>> {
    const edgeTypes: Record<string, ComponentType<any>> = {};

    for (const manifest of this.extensions.values()) {
      if (manifest.edgeTypes) {
        Object.assign(edgeTypes, manifest.edgeTypes);
      }
    }

    return edgeTypes;
  }

  /**
   * Get components for a specific slot from all extensions.
   */
  getSlotComponents(slotName: string): Array<{
    component: ComponentType<SlotProps>;
    extensionId: string;
  }> {
    const components: Array<{
      component: ComponentType<SlotProps>;
      extensionId: string;
    }> = [];

    for (const [id, manifest] of this.extensions) {
      if (manifest.slots && manifest.slots[slotName]) {
        components.push({
          component: manifest.slots[slotName],
          extensionId: id,
        });
      }
    }

    return components;
  }

  /**
   * Get initial state for all extensions (used when setting up the store).
   */
  getInitialStates(): Record<string, Record<string, unknown>> {
    const states: Record<string, Record<string, unknown>> = {};

    for (const [id, manifest] of this.extensions) {
      if (manifest.initialState) {
        states[id] = manifest.initialState;
      }
    }

    return states;
  }

  /**
   * Notify all extensions of user login.
   */
  notifyUserLogin(user: any): void {
    for (const [id, manifest] of this.extensions) {
      if (manifest.onUserLogin) {
        const context = this.contexts.get(id);
        if (context) {
          try {
            manifest.onUserLogin(user, context);
          } catch (error) {
            console.error(`[Extensions] Error in ${manifest.name}.onUserLogin:`, error);
          }
        }
      }
    }
  }

  /**
   * Notify all extensions of user logout.
   */
  notifyUserLogout(): void {
    for (const [id, manifest] of this.extensions) {
      if (manifest.onUserLogout) {
        const context = this.contexts.get(id);
        if (context) {
          try {
            manifest.onUserLogout(context);
          } catch (error) {
            console.error(`[Extensions] Error in ${manifest.name}.onUserLogout:`, error);
          }
        }
      }
    }
  }

  /**
   * Validate an extension manifest.
   */
  private validateManifest(manifest: ExtensionManifest): void {
    if (!manifest.id || typeof manifest.id !== 'string') {
      throw new Error('Extension manifest must have a valid "id" string');
    }

    if (!manifest.name || typeof manifest.name !== 'string') {
      throw new Error(`Extension "${manifest.id}" must have a valid "name" string`);
    }

    if (!manifest.version || typeof manifest.version !== 'string') {
      throw new Error(`Extension "${manifest.id}" must have a valid "version" string`);
    }

    // Validate routes
    if (manifest.routes) {
      for (const route of manifest.routes) {
        if (!route.path || !route.path.startsWith('/')) {
          throw new Error(
            `Extension "${manifest.id}" has invalid route path: "${route.path}". ` +
            `Paths must start with "/".`
          );
        }
        if (!route.component) {
          throw new Error(
            `Extension "${manifest.id}" route "${route.path}" is missing a component.`
          );
        }
      }
    }

    // Validate panel extensions
    if (manifest.panelExtensions) {
      for (const ext of manifest.panelExtensions) {
        if (ext.position === 'tab' && !ext.tabLabel) {
          throw new Error(
            `Extension "${manifest.id}" has a tab panel extension without a tabLabel.`
          );
        }
      }
    }

    // Validate nav items
    if (manifest.navItems) {
      for (const item of manifest.navItems) {
        if (!item.id) {
          throw new Error(`Extension "${manifest.id}" has a nav item without an id.`);
        }
        if (!item.path && !item.onClick) {
          throw new Error(
            `Extension "${manifest.id}" nav item "${item.id}" must have either path or onClick.`
          );
        }
      }
    }
  }
}

// Singleton instance
export const extensionRegistry = new ExtensionRegistry();

// Export for use in extensions config
export default extensionRegistry;
